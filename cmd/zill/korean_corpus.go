// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"io"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreancorpus"
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

	// The release/font path still consumes corpus.KoreanProject. Until those
	// callers are consolidated onto koreancorpus, make the CI-facing checker
	// pass every non-empty accepted overlay through the stricter loader as well.
	// Both loaders must agree on semantic Korean and machine-owned layout.
	if koreanSummary.Records != 0 {
		strict, err := koreancorpus.Load(root, source)
		if err != nil {
			fmt.Fprintf(stderr, "zill: korean-check: strict corpus validation: %v\n", err)
			return 1
		}
		if len(strict.Records) != len(korean.Entries) {
			fmt.Fprintf(stderr, "zill: korean-check: loader disagreement: production loader has %d records, strict loader has %d\n", len(korean.Entries), len(strict.Records))
			return 1
		}
		for index, entry := range korean.Entries {
			record := strict.Records[index]
			if record.ID != entry.ID || record.Japanese != entry.Japanese || record.Text != entry.Korean || record.Layout != entry.Layout {
				fmt.Fprintf(stderr, "zill: korean-check: loader disagreement at index %d: production ID %d, strict ID %d\n", index, entry.ID, record.ID)
				return 1
			}
		}
	}

	custom := koreanslots.RequiredCustomRunes(korean.Texts())
	fmt.Fprintf(stdout,
		"OK: %d accepted Korean records across %d section files; %d/%d source records covered; %d custom renderer glyphs currently required\n",
		koreanSummary.Records, koreanSummary.Sections, koreanSummary.Records, sourceSummary.Records, len(custom))
	return 0
}
