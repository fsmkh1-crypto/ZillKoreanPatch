// SPDX-License-Identifier: GPL-3.0-or-later

// Package c5scan provides a deliberately heuristic MIPS scanner for locating
// retail executable regions that may implement the retained C5 dialogue page
// contract. It does not assign C5 semantics by itself.
package c5scan

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"sort"
)

// Candidate is one executable instruction window whose local immediate/byte-op
// pattern resembles the retained C5 contract closely enough to inspect further.
type Candidate struct {
	FileOffset uint64
	Score      int
	Reasons    []string
	Window     []Instruction
}

// Instruction is a compact decoded MIPS instruction used only for forensic
// reporting. Unknown instructions retain their raw word.
type Instruction struct {
	FileOffset uint64
	Word       uint32
	Text       string
}

// Scan searches executable PT_LOAD bytes only. Candidates are heuristic and
// must be verified by data/control flow before being treated as runtime proof.
func Scan(data []byte) ([]Candidate, error) {
	f, err := elf.NewFile(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("open ELF: %w", err)
	}
	defer f.Close()
	if f.Class != elf.ELFCLASS32 || f.Data != elf.ELFDATA2LSB {
		return nil, fmt.Errorf("expected 32-bit little-endian ELF, got class=%v data=%v", f.Class, f.Data)
	}
	if f.Machine != elf.EM_MIPS {
		return nil, fmt.Errorf("expected MIPS ELF, got machine=%v", f.Machine)
	}

	var out []Candidate
	for _, prog := range f.Progs {
		if prog.Type != elf.PT_LOAD || prog.Flags&elf.PF_X == 0 || prog.Filesz < 4 {
			continue
		}
		start := prog.Off
		end := prog.Off + prog.Filesz
		if end > uint64(len(data)) {
			return nil, fmt.Errorf("executable segment [%#x,%#x) exceeds file size %#x", start, end, len(data))
		}
		for off := align4(start); off+4 <= end; off += 4 {
			word := binary.LittleEndian.Uint32(data[off : off+4])
			if !hasCapacityImmediate(word) {
				continue
			}
			candidate := scoreWindow(data, start, end, off)
			if candidate.Score >= 5 {
				out = append(out, candidate)
			}
		}
	}

	// Multiple capacity-immediate instructions in one routine can generate
	// overlapping windows. Keep the strongest representative within 0x40 bytes.
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].FileOffset < out[j].FileOffset
	})
	var dedup []Candidate
	for _, candidate := range out {
		near := false
		for _, prior := range dedup {
			delta := int64(candidate.FileOffset) - int64(prior.FileOffset)
			if delta < 0 {
				delta = -delta
			}
			if delta <= 0x40 {
				near = true
				break
			}
		}
		if !near {
			dedup = append(dedup, candidate)
		}
	}
	sort.Slice(dedup, func(i, j int) bool { return dedup[i].FileOffset < dedup[j].FileOffset })
	return dedup, nil
}

func scoreWindow(data []byte, segmentStart, segmentEnd, anchor uint64) Candidate {
	const radius = uint64(24) // instructions
	lo := anchor
	if lo >= radius*4+segmentStart {
		lo -= radius * 4
	} else {
		lo = segmentStart
	}
	hi := anchor + (radius+1)*4
	if hi > segmentEnd {
		hi = segmentEnd
	}
	lo = align4(lo)

	candidate := Candidate{FileOffset: anchor}
	seenCapacity := false
	seenLineCount := false
	seenPageCount := false
	seenByteOp := false
	seenBranch := false
	seenCopyShape := false

	for off := lo; off+4 <= hi; off += 4 {
		word := binary.LittleEndian.Uint32(data[off : off+4])
		candidate.Window = append(candidate.Window, Instruction{FileOffset: off, Word: word, Text: decode(word)})
		if hasCapacityImmediate(word) {
			seenCapacity = true
		}
		if hasSmallImmediate(word, 3) {
			seenLineCount = true
		}
		if hasSmallImmediate(word, 9) {
			seenPageCount = true
		}
		if isByteLoadStore(word) {
			seenByteOp = true
		}
		if isConditionalBranch(word) {
			seenBranch = true
		}
		if isPointerIncrement(word) || isByteStore(word) {
			seenCopyShape = seenCopyShape || isByteStore(word)
		}
	}

	if seenCapacity {
		candidate.Score += 2
		candidate.Reasons = append(candidate.Reasons, "capacity immediate 0x100/0xff (or signed -0x100)")
	}
	if seenLineCount {
		candidate.Score += 2
		candidate.Reasons = append(candidate.Reasons, "nearby immediate 3")
	}
	if seenPageCount {
		candidate.Score++
		candidate.Reasons = append(candidate.Reasons, "nearby immediate 9")
	}
	if seenByteOp {
		candidate.Score += 2
		candidate.Reasons = append(candidate.Reasons, "nearby byte load/store")
	}
	if seenBranch {
		candidate.Score++
		candidate.Reasons = append(candidate.Reasons, "nearby conditional branch")
	}
	if seenCopyShape {
		candidate.Score++
		candidate.Reasons = append(candidate.Reasons, "nearby byte-store/copy-loop shape")
	}
	return candidate
}

func align4(v uint64) uint64 { return (v + 3) &^ 3 }
func opcode(w uint32) uint32 { return w >> 26 }
func rs(w uint32) uint32     { return (w >> 21) & 0x1f }
func rt(w uint32) uint32     { return (w >> 16) & 0x1f }
func rd(w uint32) uint32     { return (w >> 11) & 0x1f }
func shamt(w uint32) uint32  { return (w >> 6) & 0x1f }
func funct(w uint32) uint32  { return w & 0x3f }
func uimm(w uint32) uint16   { return uint16(w) }
func simm(w uint32) int16    { return int16(w) }

func hasCapacityImmediate(w uint32) bool {
	switch opcode(w) {
	case 0x08, 0x09, 0x0a, 0x0b: // addi/addiu/slti/sltiu, signed immediate
		return simm(w) == 0x100 || simm(w) == 0xff || simm(w) == -0x100
	case 0x0c, 0x0d, 0x0e: // andi/ori/xori, unsigned immediate
		return uimm(w) == 0x100 || uimm(w) == 0xff
	default:
		return false
	}
}

func hasSmallImmediate(w uint32, value uint16) bool {
	switch opcode(w) {
	case 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e:
		return uimm(w) == value
	default:
		return false
	}
}

func isByteLoadStore(w uint32) bool {
	switch opcode(w) {
	case 0x20, 0x24, 0x28: // lb/lbu/sb
		return true
	default:
		return false
	}
}

func isByteStore(w uint32) bool { return opcode(w) == 0x28 }

func isConditionalBranch(w uint32) bool {
	switch opcode(w) {
	case 0x01, 0x04, 0x05, 0x06, 0x07:
		return true
	default:
		return false
	}
}

func isPointerIncrement(w uint32) bool {
	return opcode(w) == 0x09 && (simm(w) == 1 || simm(w) == -1)
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
		switch funct(w) {
		case 0x00:
			return fmt.Sprintf("sll r%d,r%d,%d", rd(w), rt(w), shamt(w))
		case 0x08:
			return fmt.Sprintf("jr r%d", rs(w))
		case 0x21:
			return fmt.Sprintf("addu r%d,r%d,r%d", rd(w), rs(w), rt(w))
		}
	}
	return fmt.Sprintf("word 0x%08x", w)
}
