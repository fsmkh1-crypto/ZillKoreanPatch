// SPDX-License-Identifier: GPL-3.0-or-later

// font-renderer-scan is a deliberately narrow forensic helper for the
// ULJM-05410 retail executable. It does not label any address as the font
// renderer. Instead it finds MIPS code windows that contain a 32-byte stride
// (sll ...,5), the record stride used by the authenticated 0x20-byte PAF glyph
// table, together with nearby halfword/word loads. The output is a candidate
// list for manual disassembly, not proof of a renderer lookup routine.
package main

import (
    "encoding/binary"
    "flag"
    "fmt"
    "os"
)

type instruction struct {
    offset int
    word   uint32
}

func opcode(w uint32) uint32 { return w >> 26 }
func funct(w uint32) uint32 { return w & 0x3f }
func rs(w uint32) uint32 { return (w >> 21) & 0x1f }
func rt(w uint32) uint32 { return (w >> 16) & 0x1f }
func rd(w uint32) uint32 { return (w >> 11) & 0x1f }
func shamt(w uint32) uint32 { return (w >> 6) & 0x1f }
func imm(w uint32) int16 { return int16(w & 0xffff) }

func isSLL5(w uint32) bool { return opcode(w) == 0 && funct(w) == 0 && shamt(w) == 5 && rd(w) != 0 }
func isLoad(w uint32) bool {
    switch opcode(w) {
    case 0x21, // lh
        0x23, // lw
        0x24, // lbu
        0x25: // lhu
        return true
    default:
        return false
    }
}

func opname(w uint32) string {
    switch {
    case isSLL5(w):
        return fmt.Sprintf("sll r%d,r%d,5", rd(w), rt(w))
    case opcode(w) == 0x21:
        return fmt.Sprintf("lh r%d,%d(r%d)", rt(w), imm(w), rs(w))
    case opcode(w) == 0x23:
        return fmt.Sprintf("lw r%d,%d(r%d)", rt(w), imm(w), rs(w))
    case opcode(w) == 0x24:
        return fmt.Sprintf("lbu r%d,%d(r%d)", rt(w), imm(w), rs(w))
    case opcode(w) == 0x25:
        return fmt.Sprintf("lhu r%d,%d(r%d)", rt(w), imm(w), rs(w))
    case opcode(w) == 0 && funct(w) == 0x21:
        return fmt.Sprintf("addu r%d,r%d,r%d", rd(w), rs(w), rt(w))
    case opcode(w) == 0 && funct(w) == 0x08:
        return fmt.Sprintf("jr r%d", rs(w))
    default:
        return fmt.Sprintf("word 0x%08x", w)
    }
}

func main() {
    radius := flag.Int("radius", 8, "instructions before/after a stride-32 instruction")
    flag.Parse()
    if flag.NArg() != 1 {
        fmt.Fprintln(os.Stderr, "usage: go run ./tools/forensics/font-renderer-scan.go [--radius N] RETAIL_EBOOT")
        os.Exit(2)
    }
    data, err := os.ReadFile(flag.Arg(0))
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    if len(data) < 4 || string(data[:4]) != "\x7fELF" {
        fmt.Fprintln(os.Stderr, "input is not an ELF executable")
        os.Exit(1)
    }

    candidates := 0
    for off := 0; off+4 <= len(data); off += 4 {
        w := binary.LittleEndian.Uint32(data[off : off+4])
        if !isSLL5(w) {
            continue
        }
        lo := off - *radius*4
        if lo < 0 { lo = 0 }
        hi := off + (*radius+1)*4
        if hi > len(data) { hi = len(data) }
        nearbyLoad := false
        for p := lo; p+4 <= hi; p += 4 {
            if isLoad(binary.LittleEndian.Uint32(data[p:p+4])) {
                nearbyLoad = true
                break
            }
        }
        if !nearbyLoad {
            continue
        }
        candidates++
        fmt.Printf("CANDIDATE file_offset=0x%X stride_insn=0x%08X\n", off, w)
        for p := lo; p+4 <= hi; p += 4 {
            q := binary.LittleEndian.Uint32(data[p:p+4])
            mark := " "
            if p == off { mark = ">" }
            fmt.Printf(" %s 0x%08X  0x%08X  %s\n", mark, p, q, opname(q))
        }
    }
    fmt.Printf("SUMMARY stride32_with_nearby_load=%d\n", candidates)
    fmt.Println("NOTE candidates are heuristic only; correlate with authenticated PAF parsing and call/data flow before assigning renderer semantics")
}
