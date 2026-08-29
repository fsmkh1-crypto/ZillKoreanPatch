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
}

// Scan searches executable PT_LOAD segments for windows containing an immediate
// 0x02 plus nearby byte reads and control-flow. Immediate 0x15 is scored strongly
// when present, but is not required because a real implementation may use a jump
// table or indexed dispatch with no literal 0x15 in the handler.
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
			if !hasImmediate(word, 2) {
				continue
			}
			c := scoreWindow(data, start, end, prog.Vaddr, off)
			if c.Score >= 5 {
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
	seenPrefix, seen15, seenByteLoad, seenBranch, seenInc, seenIndex := false, false, false, false, false, false
	for off := align4(lo); off+4 <= hi; off += 4 {
		w := binary.LittleEndian.Uint32(data[off : off+4])
		c.Window = append(c.Window, Instruction{
			FileOffset:     off,
			VirtualAddress: segmentVaddr + (off - segmentStart),
			Word:           w,
			Text:           decode(w),
		})
		seenPrefix = seenPrefix || hasImmediate(w, 2)
		seen15 = seen15 || hasImmediate(w, 0x15)
		seenByteLoad = seenByteLoad || opcode(w) == 0x20 || opcode(w) == 0x24
		seenBranch = seenBranch || isBranch(w)
		seenInc = seenInc || opcode(w) == 0x09 && (int16(w) == 1 || int16(w) == 2)
		seenIndex = seenIndex || opcode(w) == 0 && funct(w) == 0 && (shamt(w) == 2 || shamt(w) == 3)
	}
	if seenPrefix {
		c.Score += 2
		c.Reasons = append(c.Reasons, "immediate 0x02 control-prefix candidate")
	}
	if seen15 {
		c.Score += 3
		c.Reasons = append(c.Reasons, "nearby immediate 0x15 focus opcode")
	}
	if seenByteLoad {
		c.Score += 2
		c.Reasons = append(c.Reasons, "nearby byte load")
	}
	if seenBranch {
		c.Score++
		c.Reasons = append(c.Reasons, "nearby conditional branch")
	}
	if seenInc {
		c.Score++
		c.Reasons = append(c.Reasons, "nearby pointer increment")
	}
	if seenIndex {
		c.Score++
		c.Reasons = append(c.Reasons, "nearby scaled-index shape")
	}
	return c
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

func hasImmediate(w uint32, value uint16) bool {
	switch opcode(w) {
	case 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e:
		return uint16(w) == value
	default:
		return false
	}
}
func isBranch(w uint32) bool {
	switch opcode(w) {
	case 0x01, 0x04, 0x05, 0x06, 0x07:
		return true
	}
	return false
}

func decode(w uint32) string {
	switch opcode(w) {
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
	case 0x20:
		return fmt.Sprintf("lb r%d,%d(r%d)", rt(w), simm(w), rs(w))
	case 0x24:
		return fmt.Sprintf("lbu r%d,%d(r%d)", rt(w), simm(w), rs(w))
	case 0x28:
		return fmt.Sprintf("sb r%d,%d(r%d)", rt(w), simm(w), rs(w))
	case 0x04:
		return fmt.Sprintf("beq r%d,r%d,%d", rs(w), rt(w), simm(w))
	case 0x05:
		return fmt.Sprintf("bne r%d,r%d,%d", rs(w), rt(w), simm(w))
	case 0x06:
		return fmt.Sprintf("blez r%d,%d", rs(w), simm(w))
	case 0x07:
		return fmt.Sprintf("bgtz r%d,%d", rs(w), simm(w))
	case 0:
		if funct(w) == 0 {
			return fmt.Sprintf("sll r%d,r%d,%d", rd(w), rt(w), shamt(w))
		}
	}
	return fmt.Sprintf("word 0x%08x", w)
}
