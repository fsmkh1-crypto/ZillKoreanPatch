// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/message"
)

// DeriveKoreanEnglishDialogueLayouts mirrors the upstream English Reflow path
// for verified narrow dialogue/in-world-guidance records. Korean previously
// mirrored only the warning emitted after English reflow, which allowed an
// unbroken Korean line to reach the device even though English would have
// inserted layout breaks first. Canonical Korean is never rewritten: derived
// breaks live only in the build-local layout map.
func (e *Engine) DeriveKoreanEnglishDialogueLayouts(source *corpus.Project, korean *corpus.KoreanProject, layouts map[int]string, mapping koreanslots.Mapping) (map[int]string, int, error) {
	if source == nil || korean == nil {
		return nil, 0, fmt.Errorf("Korean English dialogue derivation: nil project")
	}
	if len(mapping) == 0 && len(korean.Entries) != 0 {
		return nil, 0, fmt.Errorf("Korean English dialogue derivation: empty renderer mapping")
	}
	derived := make(map[int]string, len(layouts))
	for id, text := range layouts {
		derived[id] = text
	}
	count := 0
	for _, row := range korean.Entries {
		// narrowText is inherited from upstream English category data and is
		// deliberately restricted to verified dialogue/in-world-guidance ranges.
		if !e.narrowText(row.ID) || row.Layout != "" {
			continue
		}
		item, ok := source.Find(row.ID)
		if !ok {
			return nil, 0, fmt.Errorf("dialogue message %d lacks source", row.ID)
		}
		effective := effectiveKoreanText(row, derived)
		width, _, err := e.koreanWarningMetrics(item.Record, effective, row.ID, mapping)
		if err != nil {
			return nil, 0, err
		}
		limit := e.advanceLimit(row.ID)
		if width <= limit {
			continue
		}
		candidate, err := e.wrapKoreanVisualToLimit(effective, row.ID, mapping, limit)
		if err != nil {
			return nil, 0, err
		}
		if !message.PreservesLayoutSemantics(row.Korean, candidate) {
			return nil, 0, fmt.Errorf("message %d dialogue derived layout changes semantic/control text", row.ID)
		}
		postWidth, _, err := e.koreanWarningMetrics(item.Record, candidate, row.ID, mapping)
		if err != nil {
			return nil, 0, err
		}
		if postWidth > limit {
			return nil, 0, fmt.Errorf("message %d dialogue cannot be made English-reflow safe by layout alone: width=%d/%d", row.ID, postWidth, limit)
		}
		derived[row.ID] = candidate
		count++
	}
	return derived, count, nil
}

// AuditKoreanEnglishDialogueResiduals hard-checks the exact record population
// owned by DeriveKoreanEnglishDialogueLayouts after derivation. It deliberately
// mirrors the derivation eligibility predicate instead of widening U5's scope to
// unrelated authoring warnings whose actual consumer contract is not proven.
func (e *Engine) AuditKoreanEnglishDialogueResiduals(source *corpus.Project, korean *corpus.KoreanProject, layouts map[int]string, mapping koreanslots.Mapping) (checked int, overflowIDs []int, err error) {
	if source == nil || korean == nil {
		return 0, nil, fmt.Errorf("Korean English dialogue residual audit: nil project")
	}
	if len(mapping) == 0 && len(korean.Entries) != 0 {
		return 0, nil, fmt.Errorf("Korean English dialogue residual audit: empty renderer mapping")
	}
	for _, row := range korean.Entries {
		if !e.narrowText(row.ID) || row.Layout != "" {
			continue
		}
		item, ok := source.Find(row.ID)
		if !ok {
			return checked, overflowIDs, fmt.Errorf("dialogue message %d lacks source", row.ID)
		}
		checked++
		effective := effectiveKoreanText(row, layouts)
		width, _, metricErr := e.koreanWarningMetrics(item.Record, effective, row.ID, mapping)
		if metricErr != nil {
			return checked, overflowIDs, metricErr
		}
		if width > e.advanceLimit(row.ID) {
			overflowIDs = append(overflowIDs, row.ID)
		}
	}
	return checked, overflowIDs, nil
}

func (e *Engine) wrapKoreanVisualToLimit(text string, id int, mapping koreanslots.Mapping, limit int) (string, error) {
	pages := strings.Split(text, "<end>")
	for pi, page := range pages {
		paragraphs := strings.Split(page, lineBreak)
		for i, paragraph := range paragraphs {
			wrapped, err := e.wrapKoreanVisualParagraphToLimit(paragraph, id, mapping, limit)
			if err != nil {
				return "", err
			}
			paragraphs[i] = wrapped
		}
		pages[pi] = strings.Join(paragraphs, lineBreak)
	}
	return strings.Join(pages, "<end>"), nil
}

type koreanVisualRun struct {
	text       string
	whitespace bool
}

func splitKoreanVisualRuns(text string) []koreanVisualRun {
	runes := []rune(text)
	if len(runes) == 0 {
		return nil
	}
	start := 0
	space := unicode.IsSpace(runes[0])
	runs := make([]koreanVisualRun, 0, 8)
	for i := 1; i < len(runes); i++ {
		nextSpace := unicode.IsSpace(runes[i])
		if nextSpace == space {
			continue
		}
		runs = append(runs, koreanVisualRun{text: string(runes[start:i]), whitespace: space})
		start = i
		space = nextSpace
	}
	return append(runs, koreanVisualRun{text: string(runes[start:]), whitespace: space})
}

// lastBreakableWhitespaceRun returns the final complete whitespace span that
// separates two non-whitespace runs. Replacing that whole span with
// <line-break> is accepted by PreservesLayoutSemantics; normalizing or partly
// consuming the span is not.
func lastBreakableWhitespaceRun(text string) (prefix, tail string, ok bool) {
	runes := []rune(text)
	lastStart, lastEnd := -1, -1
	for i := 0; i < len(runes); {
		if !unicode.IsSpace(runes[i]) {
			i++
			continue
		}
		start := i
		for i < len(runes) && unicode.IsSpace(runes[i]) {
			i++
		}
		if start > 0 && i < len(runes) {
			lastStart, lastEnd = start, i
		}
	}
	if lastStart < 0 {
		return "", "", false
	}
	return string(runes[:lastStart]), string(runes[lastEnd:]), true
}

func (e *Engine) wrapKoreanVisualParagraphToLimit(text string, id int, mapping koreanslots.Mapping, limit int) (string, error) {
	if strings.TrimSpace(text) == "" {
		return text, nil
	}
	runs := splitKoreanVisualRuns(text)
	lines := make([]string, 0, 4)
	current := ""
	pendingWhitespace := ""
	hasWord := false

	for _, run := range runs {
		if run.whitespace {
			pendingWhitespace += run.text
			continue
		}

		word := run.text
		candidate := current + pendingWhitespace + word
		width, err := e.measureKoreanRenderer(candidate, id, mapping)
		if err != nil {
			return "", err
		}
		if width <= limit {
			current = candidate
			pendingWhitespace = ""
			hasWord = true
			continue
		}

		if !hasWord {
			return "", fmt.Errorf("message %d dialogue word exceeds %d units and cannot be whitespace-reflowed: %q (%d units)", id, limit, word, width)
		}

		// Keep the forensic rule used by the established profile wrapper: never
		// create a fresh line boundary immediately before a runtime substitution.
		// Instead move the previous natural-text run together with the value token
		// by replacing the preceding complete whitespace span.
		if strings.HasPrefix(word, "<value:") {
			prefix, tail, ok := lastBreakableWhitespaceRun(current)
			if !ok {
				return "", fmt.Errorf("message %d dialogue cannot wrap safely before runtime substitution %q", id, word)
			}
			moved := tail + pendingWhitespace + word
			movedWidth, err := e.measureKoreanRenderer(moved, id, mapping)
			if err != nil {
				return "", err
			}
			if movedWidth > limit {
				return "", fmt.Errorf("message %d dialogue runtime-substitution group exceeds %d units: %q (%d units)", id, limit, moved, movedWidth)
			}
			lines = append(lines, prefix)
			current = moved
			pendingWhitespace = ""
			hasWord = true
			continue
		}

		wordWidth, err := e.measureKoreanRenderer(word, id, mapping)
		if err != nil {
			return "", err
		}
		if wordWidth > limit {
			return "", fmt.Errorf("message %d dialogue word exceeds %d units and cannot be whitespace-reflowed: %q (%d units)", id, limit, word, wordWidth)
		}

		// The entire pending whitespace span is replaced by a layout break. All
		// whitespace that is not selected as a break remains byte-for-byte intact.
		lines = append(lines, current)
		current = word
		pendingWhitespace = ""
		hasWord = true
	}

	current += pendingWhitespace
	if current != "" {
		lines = append(lines, current)
	}
	return strings.Join(lines, lineBreak), nil
}
