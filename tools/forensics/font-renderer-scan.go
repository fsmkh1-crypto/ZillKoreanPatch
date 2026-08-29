// SPDX-License-Identifier: GPL-3.0-or-later

// font-renderer-scan is a deliberately narrow forensic helper for the
// ULJM-05410 retail executable. It does not label any address as the font
// renderer. Instead it finds executable MIPS PT_LOAD windows where a 32-byte
// stride (sll ...,5) feeds an address-add and that derived record pointer is
// subsequently used by one or more loads from offsets inside a 0x20-byte
// record. The authenticated PAF table uses that stride. Candidate output is a
// static lead for manual disassembly, never runtime proof.
package main

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"sort"
)

type rendererInstruction struct {
	FileOffset uint64
	Vaddr      uint64
	Word       uint32
	Text       string
}

type rendererCandidate struct {
	FileOffset uint64
	Vaddr      uint64
	FieldLoads int
	Window     []rendererInstruction
}

func frOpcode(w uint32) uint32 { return w >> 26 }
func frFunct(w uint32) uint32  { return w & 0x3f }
func frRs(w uint32) uint32     { return (w >> 21) & 0x1f }
func frRt(w uint32) uint32     { return (w >> 16) & 0x1f }
func frRd(w uint32) uint32     { return (w >> 11) & 0x1f }
func frShamt(w uint32) uint32  { return (w >> 6) & 0x1f }
func frImm(w uint32) int16     { return int16(w) }

func frSLL5(w uint32) bool {
	return frOpcode(w) == 0 && frFunct(w) == 0 && frShamt(w) == 5 && frRd(w) != 0
}

func frAddu(w uint32) bool { return frOpcode(w) == 0 && frFunct(w) == 0x21 && frRd(w) != 0 }

func frLoad(w uint32) bool {
	switch frOpcode(w) {
	case 0x20, 0x21, 0x23, 0x24, 0x25:
		return true
	default:
		return false
	}
}

func frFieldLoadFrom(w uint32, base uint32) bool {
	return frLoad(w) && frRs(w) == base && frImm(w) >= 0 && frImm(w) < 0x20
}

func frText(w uint32) string {
	switch {
	case frSLL5(w):
		return fmt.Sprintf("sll r%d,r%d,5", frRd(w), frRt(w))
	case frAddu(w):
		return fmt.Sprintf("addu r%d,r%d,r%d", frRd(w), frRs(w), frRt(w))
	case frOpcode(w) == 0x20:
		return fmt.Sprintf("lb r%d,%d(r%d)", frRt(w), frImm(w), frRs(w))
	case frOpcode(w) == 0x21:
		return fmt.Sprintf("lh r%d,%d(r%d)", frRt(w), frImm(w), frRs(w))
	case frOpcode(w) == 0x23:
		return fmt.Sprintf("lw r%d,%d(r%d)", frRt(w), frImm(w), frRs(w))
	case frOpcode(w) == 0x24:
		return fmt.Sprintf("lbu r%d,%d(r%d)", frRt(w), frImm(w), frRs(w))
	case frOpcode(w) == 0x25:
		return fmt.Sprintf("lhu r%d,%d(r%d)", frRt(w), frImm(w), frRs(w))
	case frOpcode(w) == 0 && frFunct(w) == 0x08:
		return fmt.Sprintf("jr r%d", frRs(w))
	default:
		return fmt.Sprintf("word 0x%08x", w)
	}
}

func scanRendererCandidates(data []byte, radius int) ([]rendererCandidate, error) {
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

	var out []rendererCandidate
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
			if !frSLL5(stride) {
				continue
			}
			shiftReg := frRd(stride)
			hi := off + uint64(radius+1)*4
			if hi > end {
				hi = end
			}
			var recordBase uint32
			foundAdd := false
			for p := off + 4; p+4 <= hi; p += 4 {
				w := binary.LittleEndian.Uint32(data[p : p+4])
				if frAddu(w) && (frRs(w) == shiftReg || frRt(w) == shiftReg) {
					recordBase = frRd(w)
					foundAdd = true
					break
				}
			}
			if !foundAdd {
				continue
			}

			fieldLoads := 0
			for p := off + 4; p+4 <= hi; p += 4 {
				if frFieldLoadFrom(binary.LittleEndian.Uint32(data[p:p+4]), recordBase) {
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
			c := rendererCandidate{
				FileOffset: off,
				Vaddr:      prog.Vaddr + (off - start),
				FieldLoads: fieldLoads,
			}
			for p := (lo + 3) &^ 3; p+4 <= hi; p += 4 {
				w := binary.LittleEndian.Uint32(data[p : p+4])
				c.Window = append(c.Window, rendererInstruction{
					FileOffset: p,
					Vaddr:      prog.Vaddr + (p - start),
					Word:       w,
					Text:       frText(w),
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

func main() {
	radius := flag.Int("radius", 12, "instructions before/after a stride-32 instruction")
	limit := flag.Int("limit", 25, "maximum candidates to print (0 = all)")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: go run ./tools/forensics/font-renderer-scan.go [--radius N] [--limit N] RETAIL_EBOOT")
		os.Exit(2)
	}
	data, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	candidates, err := scanRendererCandidates(data, *radius)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	printCount := len(candidates)
	if *limit > 0 && printCount > *limit {
		printCount = *limit
	}
	for i := 0; i < printCount; i++ {
		c := candidates[i]
		fmt.Printf("CANDIDATE rank=%d file_offset=0x%X vaddr=0x%X linked_field_loads=%d heuristic_only=true\n", i+1, c.FileOffset, c.Vaddr, c.FieldLoads)
		for _, insn := range c.Window {
			mark := " "
			if insn.FileOffset == c.FileOffset {
				mark = ">"
			}
			fmt.Printf(" %s file_offset=0x%08X vaddr=0x%08X word=0x%08X text=%q\n", mark, insn.FileOffset, insn.Vaddr, insn.Word, insn.Text)
		}
	}
	fmt.Printf("SUMMARY executable_stride32_linked_record_candidates=%d printed=%d\n", len(candidates), printCount)
	fmt.Println("NOTE candidates are heuristic only; zero candidates do not disprove a renderer that computes record addresses differently")
}
