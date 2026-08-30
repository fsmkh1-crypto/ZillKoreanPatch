// SPDX-License-Identifier: GPL-3.0-or-later

// Package valuescan provides a deliberately heuristic MIPS scanner for locating
// executable regions that may decode message substitution controls of the form
// 0x02 <opcode>. A candidate is only a disassembly target, never runtime proof.
package valuescan

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"sort"
)

// Instruction is one decoded word retained around a candidate so authenticated
// retail runs leave enough evidence to inspect the candidate without guessing a
// file-offset-to-runtime-address mapping afterward.
type Instruction struct {
	FileOffset     uint64
	VirtualAddress uint64
	Word           uint32
	Text           string
}

// Candidate is one executable window with a control-prefix/copy-dispatch shape.
type Candidate struct {
	FileOffset     uint64
	VirtualAddress uint64
	Score          int
	Reasons        []string
	Window         []Instruction
	Focus15Linked  bool
}

// Scan searches executable PT_LOAD segments for register-linked evidence of a
// 0x02 control-prefix comparison and, because this scanner is specifically the
// $15 forensic gate, a register-linked 0x15 comparison in the same retained
// window. Supporting evidence can rank a qualifying candidate but can never
// substitute for the focus-opcode linkage itself.
func Scan(data []byte) ([]Candidate, error) {
	f, err := elf.NewFile(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("open ELF: %w", err)
	}
	defer f.Close()
	if f.Class != elf.ELFCLASS32 || f.Data != elf.ELFDATA2LSB || f.Machine != elf.EM_MIPS {
		return nil, fmt.Errorf("expected 32-bit little-endian MIPS ELF")
	}
	var out []Candidate
	for _, prog := range f.Progs {
		if prog.Type != elf.PT_LOAD || prog.Flags&elf.PF_X == 0 || prog.Filesz < 4 {
			continue
		}
		start, end := prog.Off, prog.Off+prog.Filesz
		if end > uint64(len(data)) {
			return nil, fmt.Errorf("executable segment [%#x,%#x) exceeds file size %#x", start, end, len(data))
		}
		for off := align4(start); off+4 <= end; off += 4 {
			word := binary.LittleEndian.Uint32(data[off : off+4])
			if !loadsLiteral(word, 2) {
				continue
			}
			c := scoreWindow(data, start, end, prog.Vaddr, off)
			if c.Focus15Linked && c.Score >= 6 {
				out = append(out, c)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].FileOffset < out[j].FileOffset
	})
	var dedup []Candidate
	for _, c := range out {
		near := false
		for _, p := range dedup {
			d := int64(c.FileOffset) - int64(p.FileOffset)
			if d < 0 {
				d = -d
			}
			if d <= 0x40 {
				near = true
				break
			}
		}
		if !near {
			dedup = append(dedup, c)
		}
	}
	sort.Slice(dedup, func(i, j int) bool { return dedup[i].FileOffset < dedup[j].FileOffset })
	return dedup, nil
}

func scoreWindow(data []byte, segmentStart, segmentEnd, segmentVaddr, anchor uint64) Candidate {
	const radius = uint64(24)
	lo := anchor
	if lo >= segmentStart+radius*4 {
		lo -= radius * 4
	} else {
		lo = segmentStart
	}
	hi := anchor + (radius+1)*4
	if hi > segmentEnd {
		hi = segmentEnd
	}
	c := Candidate{
		FileOffset:     anchor,
		VirtualAddress: segmentVaddr + (anchor - segmentStart),
	}

	literal2Regs := make(map[uint32]struct{})
	literal15Regs := make(map[uint32]struct{})
	byteLoadRegs := make(map[uint32]uint32) // destination -> base register
	prefixLoadRegs := make(map[uint32]struct{})
	focusLoadRegs := make(map[uint32]struct{})
	linkedPrefix := false
	linked15 := false
	linkedIncrement := false
	linkedIndex := false

	words := make([]uint32, 0, (hi-lo)/4)
	for off := align4(lo); off+4 <= hi; off += 4 {
		w := binary.LittleEndian.Uint32(data[off : off+4])
		words = append(words, w)
		vaddr := segmentVaddr + (off - segmentStart)
		c.Window = append(c.Window, Instruction{
			FileOffset:     off,
			VirtualAddress: vaddr,
			Word:           w,
			Text:           decodeAt(w, vaddr),
		})
		if loadsLiteral(w, 2) {
			literal2Regs[rt(w)] = struct{}{}
		}
		if loadsLiteral(w, 0x15) {
			literal15Regs[rt(w)] = struct{}{}
		}
		if isByteLoad(w) {
			byteLoadRegs[rt(w)] = rs(w)
		}
	}

	// Require actual branch register linkage instead of treating unrelated
	// immediates and byte loads in the same routine as dispatcher evidence.
	for _, w := range words {
		if !isEqualityBranch(w) {
			continue
		}
		a, b := rs(w), rt(w)
		if registerPair(a, b, byteLoadRegs, literal2Regs) {
			linkedPrefix = true
			if _, ok := byteLoadRegs[a]; ok {
				prefixLoadRegs[a] = struct{}{}
			}
			if _, ok := byteLoadRegs[b]; ok {
				prefixLoadRegs[b] = struct{}{}
			}
		}
		if registerPair(a, b, byteLoadRegs, literal15Regs) {
			linked15 = true
			if _, ok := byteLoadRegs[a]; ok {
				focusLoadRegs[a] = struct{}{}
			}
			if _, ok := byteLoadRegs[b]; ok {
				focusLoadRegs[b] = struct{}{}
			}
		}
	}

	if linkedPrefix {
		c.Score += 5
		c.Reasons = append(c.Reasons, "byte-loaded register compared with literal 0x02 register")
	}
	if linked15 {
		c.Score += 4
		c.Reasons = append(c.Reasons, "byte-loaded register compared with literal 0x15 register")
		c.Focus15Linked = true
	}

	// Supporting evidence is useful only when tied to the same loaded stream or
	// opcode register. It can rank an already focus-qualified candidate, but it
	// cannot independently make a $15 forensic candidate reportable.
	for _, w := range words {
		if opcode(w) == 0x09 && (simm(w) == 1 || simm(w) == 2) && rs(w) == rt(w) {
			for loadReg := range prefixLoadRegs {
				if byteLoadRegs[loadReg] == rt(w) {
					linkedIncrement = true
				}
			}
			for loadReg := range focusLoadRegs {
				if byteLoadRegs[loadReg] == rt(w) {
					linkedIncrement = true
				}
			}
		}
		if opcode(w) == 0 && funct(w) == 0 && (shamt(w) == 2 || shamt(w) == 3) {
			if _, ok := focusLoadRegs[rt(w)]; ok {
				linkedIndex = true
			}
		}
	}
	if linkedIncrement {
		c.Score++
		c.Reasons = append(c.Reasons, "linked source pointer increment")
	}
	if linkedIndex {
		c.Score += 2
		c.Reasons = append(c.Reasons, "0x15-loaded opcode register used as scaled index")
	}
	return c
}

func registerPair(a, b uint32, byteLoads map[uint32]uint32, literalRegs map[uint32]struct{}) bool {
	_, aLoad := byteLoads[a]
	_, bLoad := byteLoads[b]
	_, aLiteral := literalRegs[a]
	_, bLiteral := literalRegs[b]
	return (aLoad && bLiteral) || (bLoad && aLiteral)
}

func align4(v uint64) uint64 { return (v + 3) &^ 3 }
func opcode(w uint32) uint32 { return w >> 26 }
func rs(w uint32) uint32     { return (w >> 21) & 0x1f }
func rt(w uint32) uint32     { return (w >> 16) & 0x1f }
func rd(w uint32) uint32     { return (w >> 11) & 0x1f }
func shamt(w uint32) uint32  { return (w >> 6) & 0x1f }
func funct(w uint32) uint32  { return w & 0x3f }
func simm(w uint32) int16    { return int16(w) }
func uimm(w uint32) uint16   { return uint16(w) }
func jtarget(w uint32, pc uint64) uint64 {
	return ((pc + 4) & 0xf0000000) | uint64((w&0x03ffffff)<<2)
}
func btarget(w uint32, pc uint64) uint64 {
	return uint64(int64(pc+4) + int64(simm(w))*4)
}

func loadsLiteral(w uint32, value uint16) bool {
	// The zero-register source makes these actual literal constructions rather
	// than arbitrary arithmetic that happens to contain the same immediate.
	if rs(w) != 0 {
		return false
	}
	switch opcode(w) {
	case 0x08, 0x09, 0x0d: // addi/addiu/ori rt,r0,imm
		return uimm(w) == value
	default:
		return false
	}
}

func isByteLoad(w uint32) bool { return opcode(w) == 0x20 || opcode(w) == 0x24 }
func isEqualityBranch(w uint32) bool { return opcode(w) == 0x04 || opcode(w) == 0x05 }

func decodeAt(w uint32, pc uint64) string {
	switch opcode(w) {
	case 0x02:
		return fmt.Sprintf("j 0x%x", jtarget(w, pc))
	case 0x03:
		return fmt.Sprintf("jal 0x%x", jtarget(w, pc))
	case 0x04:
		return fmt.Sprintf("beq r%d,r%d,0x%x", rs(w), rt(w), btarget(w, pc))
	case 0x05:
		return fmt.Sprintf("bne r%d,r%d,0x%x", rs(w), rt(w), btarget(w, pc))
	case 0x06:
		return fmt.Sprintf("blez r%d,0x%x", rs(w), btarget(w, pc))
	case 0x07:
		return fmt.Sprintf("bgtz r%d,0x%x", rs(w), btarget(w, pc))
	case 0x08:
		return fmt.Sprintf("addi r%d,r%d,%d", rt(w), rs(w), simm(w))
	case 0x09:
		return fmt.Sprintf("addiu r%d,r%d,%d", rt(w), rs(w), simm(w))
	case 0x0a:
		return fmt.Sprintf("slti r%d,r%d,%d", rt(w), rs(w), simm(w))
	case 0x0b:
		return fmt.Sprintf("sltiu r%d,r%d,%d", rt(w), rs(w), simm(w))
	case 0x0c:
		return fmt.Sprintf("andi r%d,r%d,0x%x", rt(w), rs(w), uimm(w))
	case 0x0d:
		return fmt.Sprintf("ori r%d,r%d,0x%x", rt(w), rs(w), uimm(w))
	case 0x0e:
		return fmt.Sprintf("xori r%d,r%d,0x%x", rt(w), rs(w), uimm(w))
	case 0x0f:
		return fmt.Sprintf("lui r%d,0x%x", rt(w), uimm(w))
	case 0x20:
		return fmt.Sprintf("lb r%d,%d(r%d)", rt(w), simm(w), rs(w))
	case 0x21:
		return fmt.Sprintf("lh r%d,%d(r%d)", rt(w), simm(w), rs(w))
	case 0x23:
		return fmt.Sprintf("lw r%d,%d(r%d)", rt(w), simm(w), rs(w))
	case 0x24:
		return fmt.Sprintf("lbu r%d,%d(r%d)", rt(w), simm(w), rs(w))
	case 0x25:
		return fmt.Sprintf("lhu r%d,%d(r%d)", rt(w), simm(w), rs(w))
	case 0x28:
		return fmt.Sprintf("sb r%d,%d(r%d)", rt(w), simm(w), rs(w))
	case 0x29:
		return fmt.Sprintf("sh r%d,%d(r%d)", rt(w), simm(w), rs(w))
	case 0x2b:
		return fmt.Sprintf("sw r%d,%d(r%d)", rt(w), simm(w), rs(w))
	case 0:
		switch funct(w) {
		case 0:
			return fmt.Sprintf("sll r%d,r%d,%d", rd(w), rt(w), shamt(w))
		case 0x08:
			return fmt.Sprintf("jr r%d", rs(w))
		case 0x09:
			return fmt.Sprintf("jalr r%d,r%d", rd(w), rs(w))
		}
	}
	return fmt.Sprintf("word 0x%08x", w)
}

func decode(w uint32) string { return decodeAt(w, 0) }
