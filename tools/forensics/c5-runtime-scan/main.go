// SPDX-License-Identifier: GPL-3.0-or-later

// c5-runtime-scan prints heuristic executable candidates for the retained C5
// dialogue page/buffer contract. Candidate output is not semantic proof.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/HK47196/zill/internal/forensics/c5scan"
)

func main() {
	limit := flag.Int("limit", 25, "maximum candidates to print (0 = all)")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: go run ./tools/forensics/c5-runtime-scan [--limit N] RETAIL_EBOOT")
		os.Exit(2)
	}
	data, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	candidates, err := c5scan.Scan(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	count := len(candidates)
	printCount := count
	if *limit > 0 && printCount > *limit {
		printCount = *limit
	}
	for i := 0; i < printCount; i++ {
		candidate := candidates[i]
		fmt.Printf("CANDIDATE file_offset=0x%X score=%d reasons=%q\n", candidate.FileOffset, candidate.Score, strings.Join(candidate.Reasons, "; "))
		for _, insn := range candidate.Window {
			mark := " "
			if insn.FileOffset == candidate.FileOffset {
				mark = ">"
			}
			fmt.Printf(" %s 0x%08X  0x%08X  %s\n", mark, insn.FileOffset, insn.Word, insn.Text)
		}
	}
	fmt.Printf("SUMMARY heuristic_c5_candidates=%d printed=%d\n", count, printCount)
	fmt.Println("NOTE candidate patterns are heuristic only; verify message data/control flow and actual buffer ownership before assigning C5 semantics")
}
