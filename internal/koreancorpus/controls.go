// SPDX-License-Identifier: GPL-3.0-or-later

package koreancorpus

import (
	"fmt"
	"slices"
	"strings"
)

// validateControlContract catches translator-facing control corruption before
// retail banks are available. Line breaks are layout-authorable, but every
// other angle-bracket control (end markers, runtime values, etc.) must remain
// byte-for-byte identical and in the same order as the canonical Japanese
// display projection. The message compiler performs the authoritative retail
// projection check again at build time.
func validateControlContract(path string, id int, source, translated, field string) error {
	want, err := semanticControls(source)
	if err != nil {
		return fmt.Errorf("%s: ID %d: canonical Japanese controls are malformed: %w", path, id, err)
	}
	got, err := semanticControls(translated)
	if err != nil {
		return fmt.Errorf("%s: ID %d: %s controls are malformed: %w", path, id, field, err)
	}
	if !slices.Equal(got, want) {
		return fmt.Errorf("%s: ID %d: %s changes fixed control sequence: got %v, want %v", path, id, field, got, want)
	}
	return nil
}

func semanticControls(text string) ([]string, error) {
	var controls []string
	for cursor := 0; cursor < len(text); {
		open := strings.IndexByte(text[cursor:], '<')
		close := strings.IndexByte(text[cursor:], '>')
		if close >= 0 && (open < 0 || close < open) {
			return nil, fmt.Errorf("stray '>'")
		}
		if open < 0 {
			break
		}
		open += cursor
		end := strings.IndexByte(text[open+1:], '>')
		if end < 0 {
			return nil, fmt.Errorf("unclosed '<'")
		}
		end += open + 1
		tag := text[open : end+1]
		if strings.Contains(tag[1:len(tag)-1], "<") {
			return nil, fmt.Errorf("nested '<' in %q", tag)
		}
		if tag != "<line-break>" {
			controls = append(controls, tag)
		}
		cursor = end + 1
	}
	return controls, nil
}
