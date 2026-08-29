package main

import (
    "encoding/binary"
    "fmt"
    "os"
)

var gprNames = [32]string{
    "zero", "at", "v0", "v1", "a0", "a1", "a2", "a3",
    "t0", "t1", "t2", "t3", "t4", "t5", "t6", "t7",
    "s0", "s1", "s2", "s3", "s4", "s5", "s6", "s7",
    "t8", "t9", "k0", "k1", "gp", "sp", "fp", "ra",
}

type binaryImage struct {
    Data        []byte
    RuntimeBase uint32
    FileBias    uint32
    MaxInsns    int
}

func loadBinaryImage(path string, runtimeBase, fileBias uint32, maxInsns int) (*binaryImage, error) {
    b, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    return &binaryImage{Data: b, RuntimeBase: runtimeBase, FileBias: fileBias, MaxInsns: maxInsns}, nil
}

func (img *binaryImage) fileOffset(addr uint32) (uint32, bool) {
    if addr < img.RuntimeBase {
        return 0, false
    }
    off := addr - img.RuntimeBase + img.FileBias
    if uint64(off)+4 > uint64(len(img.Data)) {
        return 0, false
    }
    return off, true
}

func (img *binaryImage) decodeFunction(start uint32) (function, bool) {
    off, ok := img.fileOffset(start)
    if !ok {
        return function{}, false
    }
    out := function{Start: start}
    sawReturn := false
    for n := 0; n < img.MaxInsns; n++ {
        pos := uint64(off) + uint64(n*4)
        if pos+4 > uint64(len(img.Data)) {
            break
        }
        addr := start + uint32(n*4)
        word := binary.LittleEndian.Uint32(img.Data[pos : pos+4])
        i := decodeMIPS(addr, word)
        out.Insns = append(out.Insns, i)
        if sawReturn {
            break // return delay slot included
        }
        if i.Op == "jr" && i.Args == "ra" {
            sawReturn = true
        }
    }
    return out, len(out.Insns) > 0
}

func decodeMIPS(addr, w uint32) insn {
    op := w >> 26
    rs := (w >> 21) & 31
    rt := (w >> 16) & 31
    rd := (w >> 11) & 31
    sh := (w >> 6) & 31
    imm := uint16(w)
    simm := int16(imm)
    funct := w & 63
    mk := func(name, args string) insn {
        return insn{Addr: addr, Op: name, Args: args, Raw: fmt.Sprintf("0x%08X encoding=0x%08X %s %s", addr, w, name, args)}
    }
    r := func(x uint32) string { return gprNames[x] }
    mem := func() string { return fmt.Sprintf("%d(%s)", simm, r(rs)) }
    switch op {
    case 0:
        switch funct {
        case 0x08:
            return mk("jr", r(rs))
        case 0x09:
            return mk("jalr", fmt.Sprintf("%s,%s", r(rd), r(rs)))
        case 0x00:
            if w == 0 { return mk("nop", "") }
            return mk("sll", fmt.Sprintf("%s,%s,0x%X", r(rd), r(rt), sh))
        case 0x02:
            return mk("srl", fmt.Sprintf("%s,%s,0x%X", r(rd), r(rt), sh))
        case 0x03:
            return mk("sra", fmt.Sprintf("%s,%s,0x%X", r(rd), r(rt), sh))
        case 0x21:
            return mk("addu", fmt.Sprintf("%s,%s,%s", r(rd), r(rs), r(rt)))
        case 0x23:
            return mk("subu", fmt.Sprintf("%s,%s,%s", r(rd), r(rs), r(rt)))
        case 0x24:
            return mk("and", fmt.Sprintf("%s,%s,%s", r(rd), r(rs), r(rt)))
        case 0x25:
            if rt == 0 { return mk("move", fmt.Sprintf("%s,%s", r(rd), r(rs))) }
            if rs == 0 { return mk("move", fmt.Sprintf("%s,%s", r(rd), r(rt))) }
            return mk("or", fmt.Sprintf("%s,%s,%s", r(rd), r(rs), r(rt)))
        case 0x26:
            return mk("xor", fmt.Sprintf("%s,%s,%s", r(rd), r(rs), r(rt)))
        case 0x2A:
            return mk("slt", fmt.Sprintf("%s,%s,%s", r(rd), r(rs), r(rt)))
        case 0x2B:
            return mk("sltu", fmt.Sprintf("%s,%s,%s", r(rd), r(rs), r(rt)))
        }
    case 0x03:
        target := ((addr + 4) & 0xF0000000) | ((w & 0x03FFFFFF) << 2)
        return mk("jal", fmt.Sprintf("0x%08X", target))
    case 0x09:
        return mk("addiu", fmt.Sprintf("%s,%s,%d", r(rt), r(rs), simm))
    case 0x0A:
        return mk("slti", fmt.Sprintf("%s,%s,%d", r(rt), r(rs), simm))
    case 0x0B:
        return mk("sltiu", fmt.Sprintf("%s,%s,%d", r(rt), r(rs), simm))
    case 0x0C:
        return mk("andi", fmt.Sprintf("%s,%s,0x%X", r(rt), r(rs), imm))
    case 0x0D:
        if rs == 0 { return mk("li", fmt.Sprintf("%s,0x%X", r(rt), imm)) }
        return mk("ori", fmt.Sprintf("%s,%s,0x%X", r(rt), r(rs), imm))
    case 0x0E:
        return mk("xori", fmt.Sprintf("%s,%s,0x%X", r(rt), r(rs), imm))
    case 0x0F:
        return mk("lui", fmt.Sprintf("%s,0x%X", r(rt), imm))
    case 0x20:
        return mk("lb", fmt.Sprintf("%s,%s", r(rt), mem()))
    case 0x21:
        return mk("lh", fmt.Sprintf("%s,%s", r(rt), mem()))
    case 0x23:
        return mk("lw", fmt.Sprintf("%s,%s", r(rt), mem()))
    case 0x24:
        return mk("lbu", fmt.Sprintf("%s,%s", r(rt), mem()))
    case 0x25:
        return mk("lhu", fmt.Sprintf("%s,%s", r(rt), mem()))
    }
    return mk("op", fmt.Sprintf("0x%08X", w))
}
