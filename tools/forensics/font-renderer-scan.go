// SPDX-License-Identifier: GPL-3.0-or-later

// font-renderer-scan prints heuristic executable candidates for code that may
// index the authenticated 0x20-byte PAF glyph table. Candidate output is a
// static disassembly aid only, never runtime proof.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/HK47196/zill/internal/forensics/fontscan"
)

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
	candidates, err := fontscan.Scan(data, *radius)
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
		fmt.Printf("CANDIDATE rank=%d file_offset=0x%X vaddr=0x%X linked_field_loads=%d heuristic_only=true\n", i+1, c.FileOffset, c.VirtualAddress, c.FieldLoads)
		for _, insn := range c.Window {
			mark := " "
			if insn.FileOffset == c.FileOffset {
				mark = ">"
			}
			fmt.Printf(" %s file_offset=0x%08X vaddr=0x%08X word=0x%08X text=%q\n", mark, insn.FileOffset, insn.VirtualAddress, insn.Word, insn.Text)
		}
	}
	fmt.Printf("SUMMARY executable_stride32_linked_record_candidates=%d printed=%d\n", len(candidates), printCount)
	fmt.Println("NOTE candidates are heuristic only; zero candidates do not disprove a renderer that computes record addresses differently")
}
