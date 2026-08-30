// SPDX-License-Identifier: GPL-3.0-or-later

// Package fontscan provides a deliberately heuristic MIPS scanner for locating
// executable code that may index the authenticated 0x20-byte PAF glyph table.
// A candidate is only a static disassembly lead, never runtime proof.
package fontscan

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"sort"
)

type Instruction struct {
	FileOffset     uint64
	VirtualAddress uint64
	Word           uint32
	Text           string
}

type Candidate struct {
	FileOffset     uint64
	VirtualAddress uint64
	FieldLoads     int
	Window         []Instruction
}

func opcode(w uint32) uint32 { return w >> 26 }
func funct(w uint32) uint32  { return w & 0x3f }
func rs(w uint32) uint32     { return (w >> 21) & 0x1f }
func rt(w uint32) uint32     { return (w >> 16) & 0x1f }
func rd(w uint32) uint32     { return (w >> 11) & 0x1f }
func shamt(w uint32) uint32  { return (w >> 6) & 0x1f }
func imm(w uint32) int16     { return int16(w) }

func isSLL5(w uint32) bool {
	return opcode(w) == 0 && funct(w) == 0 && shamt(w) == 5 && rd(w) != 0
}

func isAddu(w uint32) bool { return opcode(w) == 0 && funct(w) == 0x21 && rd(w) != 0 }

func isLoad(w uint32) bool {
	switch opcode(w) {
	case 0x20, 0x21, 0x23, 0x24, 0x25:
		return true
	default:
		return false
	}
}

func isFieldLoadFrom(w uint32, base uint32) bool {
	return isLoad(w) && rs(w) == base && imm(w) >= 0 && imm(w) < 0x20
}

func decode(w uint32) string {
	switch {
	case isSLL5(w):
		return fmt.Sprintf("sll r%d,r%d,5", rd(w), rt(w))
	case isAddu(w):
		return fmt.Sprintf("addu r%d,r%d,r%d", rd(w), rs(w), rt(w))
	case opcode(w) == 0x20:
		return fmt.Sprintf("lb r%d,%d(r%d)", rt(w), imm(w), rs(w))
	case opcode(w) == 0x21:
		return fmt.Sprintf("lh r%d,%d(r%d)", rt(w), imm(w), rs(w))
	case opcode(w) == 0x23:
		return fmt.Sprintf("lw r%d,%d(r%d)", rt(w), imm(w), rs(w))
	case opcode(w) == 0x24:
		return fmt.Sprintf("lbu r%d,%d(r%d)", rt(w), imm(w), rs(w))
	case opcode(w) == 0x25:
		return fmt.Sprintf("lhu r%d,%d(r%d)", rt(w), imm(w), rs(w))
	case opcode(w) == 0 && funct(w) == 0x08:
		return fmt.Sprintf("jr r%d", rs(w))
	default:
		return fmt.Sprintf("word 0x%08x", w)
	}
}

// Scan finds executable MIPS PT_LOAD windows where a 32-byte stride feeds an
// address add and the derived record pointer is used by loads from offsets
// inside a 0x20-byte record. Zero candidates do not disprove a renderer that
// computes PAF record addresses differently.
func Scan(data []byte, radius int) ([]Candidate, error) {
	f, err := elf.NewFile(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("open ELF: %w", err)
	}
	defer f.Close()
	if f.Class != elf.ELFCLASS32 || f.Data != elf.ELFDATA2LSB || f.Machine != elf.EM_MIPS {
		return nil, fmt.Errorf("expected 32-bit little-endian MIPS ELF")
	}
	if radius < 1 {
		radius = 1
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
		for off := (start + 3) &^ 3; off+4 <= end; off += 4 {
			stride := binary.LittleEndian.Uint32(data[off : off+4])
			if !isSLL5(stride) {
				continue
			}
			shiftReg := rd(stride)
			hi := off + uint64(radius+1)*4
			if hi > end {
				hi = end
			}
			var recordBase uint32
			foundAdd := false
			for p := off + 4; p+4 <= hi; p += 4 {
				w := binary.LittleEndian.Uint32(data[p : p+4])
				if isAddu(w) && (rs(w) == shiftReg || rt(w) == shiftReg) {
					recordBase = rd(w)
					foundAdd = true
					break
				}
			}
			if !foundAdd {
				continue
			}

			fieldLoads := 0
			for p := off + 4; p+4 <= hi; p += 4 {
				if isFieldLoadFrom(binary.LittleEndian.Uint32(data[p:p+4]), recordBase) {
					fieldLoads++
				}
			}
			if fieldLoads == 0 {
				continue
			}

			lo := off
			back := uint64(radius) * 4
			if lo >= start+back {
				lo -= back
			} else {
				lo = start
			}
			c := Candidate{FileOffset: off, VirtualAddress: prog.Vaddr + (off - start), FieldLoads: fieldLoads}
			for p := (lo + 3) &^ 3; p+4 <= hi; p += 4 {
				w := binary.LittleEndian.Uint32(data[p : p+4])
				c.Window = append(c.Window, Instruction{
					FileOffset: p,
					VirtualAddress: prog.Vaddr + (p - start),
					Word: w,
					Text: decode(w),
				})
			}
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].FieldLoads != out[j].FieldLoads {
			return out[i].FieldLoads > out[j].FieldLoads
		}
		return out[i].FileOffset < out[j].FileOffset
	})
	return out, nil
}
