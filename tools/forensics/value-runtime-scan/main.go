// SPDX-License-Identifier: GPL-3.0-or-later

// value-runtime-scan prints heuristic executable candidates for message
// substitution controls of the form 0x02 <opcode>. Candidate output is a static
// disassembly aid only and must not be treated as runtime proof.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/HK47196/zill/internal/forensics/valuescan"
)

func main() {
	limit := flag.Int("limit", 10, "maximum candidates to print (0 = all)")
	window := flag.Int64("window", 0x20, "maximum byte distance from candidate anchor to print (negative = full retained window)")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: go run ./tools/forensics/value-runtime-scan [--limit N] [--window BYTES] RETAIL_EBOOT")
		os.Exit(2)
	}

	data, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	candidates, err := valuescan.Scan(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	printCount := len(candidates)
	if *limit > 0 && printCount > *limit {
		printCount = *limit
	}
	for i := 0; i < printCount; i++ {
		candidate := candidates[i]
		fmt.Printf("CANDIDATE rank=%d file_offset=0x%X vaddr=0x%X score=%d reasons=%q heuristic_only=true\n",
			i+1, candidate.FileOffset, candidate.VirtualAddress, candidate.Score, strings.Join(candidate.Reasons, "; "))
		for _, insn := range candidate.Window {
			delta := int64(insn.FileOffset) - int64(candidate.FileOffset)
			if *window >= 0 && (delta < -*window || delta > *window) {
				continue
			}
			mark := " "
			if insn.FileOffset == candidate.FileOffset {
				mark = ">"
			}
			fmt.Printf(" %s delta=%+d file_offset=0x%08X vaddr=0x%08X word=0x%08X kind=%s text=%q\n",
				mark, delta, insn.FileOffset, insn.VirtualAddress, insn.Word, kind(insn.Text), insn.Text)
		}
	}
	fmt.Printf("SUMMARY focus_opcode=$15 heuristic_value_candidates=%d printed=%d\n", len(candidates), printCount)
	fmt.Println("NOTE candidate patterns are heuristic only; zero candidates do not disprove a table-driven/shared substitution dispatcher")
}

func kind(text string) string {
	switch {
	case strings.HasPrefix(text, "jal "), strings.HasPrefix(text, "jalr "):
		return "call"
	case strings.HasPrefix(text, "j "), strings.HasPrefix(text, "jr "):
		return "jump"
	case strings.HasPrefix(text, "beq "), strings.HasPrefix(text, "bne "), strings.HasPrefix(text, "blez "), strings.HasPrefix(text, "bgtz "):
		return "branch"
	case strings.HasPrefix(text, "lb "), strings.HasPrefix(text, "lbu "), strings.HasPrefix(text, "lh "), strings.HasPrefix(text, "lhu "), strings.HasPrefix(text, "lw "):
		return "load"
	case strings.HasPrefix(text, "sb "), strings.HasPrefix(text, "sh "), strings.HasPrefix(text, "sw "):
		return "store"
	case strings.HasPrefix(text, "lui "), strings.HasPrefix(text, "ori "), strings.HasPrefix(text, "addiu "):
		return "address-or-immediate"
	default:
		return "other"
	}
}
