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

// Candidate is one executable window with a control-prefix/copy-dispatch shape.
type Candidate struct {
	FileOffset uint64
	Score      int
	Reasons    []string
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
			c := scoreWindow(data, start, end, off)
			if c.Score >= 5 {
				out = append(out, c)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score { return out[i].Score > out[j].Score }
		return out[i].FileOffset < out[j].FileOffset
	})
	var dedup []Candidate
	for _, c := range out {
		near := false
		for _, p := range dedup {
			d := int64(c.FileOffset)-int64(p.FileOffset); if d < 0 { d = -d }
			if d <= 0x40 { near = true; break }
		}
		if !near { dedup = append(dedup, c) }
	}
	sort.Slice(dedup, func(i, j int) bool { return dedup[i].FileOffset < dedup[j].FileOffset })
	return dedup, nil
}

func scoreWindow(data []byte, segmentStart, segmentEnd, anchor uint64) Candidate {
	const radius = uint64(24)
	lo := anchor
	if lo >= segmentStart+radius*4 { lo -= radius*4 } else { lo = segmentStart }
	hi := anchor+(radius+1)*4; if hi > segmentEnd { hi = segmentEnd }
	c := Candidate{FileOffset: anchor}
	seenPrefix, seen15, seenByteLoad, seenBranch, seenInc, seenIndex := false, false, false, false, false, false
	for off := align4(lo); off+4 <= hi; off += 4 {
		w := binary.LittleEndian.Uint32(data[off:off+4])
		seenPrefix = seenPrefix || hasImmediate(w, 2)
		seen15 = seen15 || hasImmediate(w, 0x15)
		seenByteLoad = seenByteLoad || opcode(w) == 0x20 || opcode(w) == 0x24
		seenBranch = seenBranch || isBranch(w)
		seenInc = seenInc || opcode(w) == 0x09 && (int16(w) == 1 || int16(w) == 2)
		seenIndex = seenIndex || opcode(w) == 0 && funct(w) == 0 && (shamt(w) == 2 || shamt(w) == 3)
	}
	if seenPrefix { c.Score += 2; c.Reasons = append(c.Reasons, "immediate 0x02 control-prefix candidate") }
	if seen15 { c.Score += 3; c.Reasons = append(c.Reasons, "nearby immediate 0x15 focus opcode") }
	if seenByteLoad { c.Score += 2; c.Reasons = append(c.Reasons, "nearby byte load") }
	if seenBranch { c.Score++; c.Reasons = append(c.Reasons, "nearby conditional branch") }
	if seenInc { c.Score++; c.Reasons = append(c.Reasons, "nearby pointer increment") }
	if seenIndex { c.Score++; c.Reasons = append(c.Reasons, "nearby scaled-index shape") }
	return c
}

func align4(v uint64) uint64 { return (v+3)&^3 }
func opcode(w uint32) uint32 { return w>>26 }
func funct(w uint32) uint32 { return w&0x3f }
func shamt(w uint32) uint32 { return (w>>6)&0x1f }
func hasImmediate(w uint32, value uint16) bool {
	switch opcode(w) {
	case 0x08,0x09,0x0a,0x0b,0x0c,0x0d,0x0e:
		return uint16(w) == value
	default:
		return false
	}
}
func isBranch(w uint32) bool {
	switch opcode(w) { case 0x01,0x04,0x05,0x06,0x07: return true }; return false
}
