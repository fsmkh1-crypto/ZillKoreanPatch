// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"fmt"
	"strings"

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

func (e *Engine) wrapKoreanVisualParagraphToLimit(text string, id int, mapping koreanslots.Mapping, limit int) (string, error) {
	if strings.TrimSpace(text) == "" {
		return text, nil
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return text, nil
	}
	lines := make([]string, 0, 4)
	current := ""
	flush := func() {
		if current != "" {
			lines = append(lines, current)
			current = ""
		}
	}
	for _, word := range words {
		candidate := word
		if current != "" {
			candidate = current + " " + word
		}
		width, err := e.measureKoreanRenderer(candidate, id, mapping)
		if err != nil {
			return "", err
		}
		if width <= limit {
			current = candidate
			continue
		}
		if current != "" {
			// Preserve the forensic rule already used by profile wrapping: do not
			// create a fresh boundary immediately before a runtime substitution.
			if strings.HasPrefix(word, "<value:") {
				if split := strings.LastIndex(current, " "); split >= 0 {
					prefix, tail := current[:split], current[split+1:]
					lines = append(lines, prefix)
					current = tail + " " + word
					w, err := e.measureKoreanRenderer(current, id, mapping)
					if err != nil {
						return "", err
					}
					if w <= limit {
						continue
					}
				}
			}
			flush()
		}
		wordWidth, err := e.measureKoreanRenderer(word, id, mapping)
		if err != nil {
			return "", err
		}
		if wordWidth > limit {
			return "", fmt.Errorf("message %d dialogue word exceeds %d units and cannot be whitespace-reflowed: %q (%d units)", id, limit, word, wordWidth)
		}
		current = word
	}
	flush()
	return strings.Join(lines, lineBreak), nil
}
