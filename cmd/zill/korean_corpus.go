// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"io"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanslots"
)

func runKoreanCheck(root string, args []string, stdout, stderr io.Writer) int {
	if len(args) != 0 {
		fmt.Fprintln(stderr, "zill: usage: zill korean-check")
		return 2
	}
	source, sourceSummary, err := corpus.LoadProject(root)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-check: %v\n", err)
		return 1
	}
	korean, koreanSummary, err := corpus.LoadKoreanProject(root, source)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-check: %v\n", err)
		return 1
	}
	custom := koreanslots.RequiredCustomRunes(korean.Texts())
	fmt.Fprintf(stdout,
		"OK: %d accepted Korean records across %d section files; %d/%d source records covered; %d custom renderer glyphs currently required\n",
		koreanSummary.Records, koreanSummary.Sections, koreanSummary.Records, sourceSummary.Records, len(custom))
	return 0
}
