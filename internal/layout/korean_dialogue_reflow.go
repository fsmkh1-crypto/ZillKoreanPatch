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

// wrapKoreanVisualParagraphToLimit is insertion-only: it never collapses,
// normalizes, trims or replaces canonical text. The first U5 implementation
// used strings.Fields and therefore changed messages containing repeated or
// edge whitespace (caught by the full-corpus census at message 70018). Here
// natural runes and complete control tags are immutable tokens; wrapping only
// inserts <line-break> between safe token boundaries.
func (e *Engine) wrapKoreanVisualParagraphToLimit(text string, id int, mapping koreanslots.Mapping, limit int) (string, error) {
	if text == "" {
		return text, nil
	}
	tokens := koreanVisualTokens(text)
	if len(tokens) == 0 {
		return text, nil
	}

	lines := make([]string, 0, 4)
	current := make([]string, 0, len(tokens))
	for _, token := range tokens {
		current = append(current, token)
		for {
			width, err := e.measureKoreanRenderer(strings.Join(current, ""), id, mapping)
			if err != nil {
				return "", err
			}
			if width <= limit {
				break
			}

			// The line fit before the newest token in the ordinary case. If that
			// boundary touches a runtime value token, walk left until the entire
			// value adjacency group moves together to the next line. This preserves
			// the forensic rule used elsewhere in the Korean pipeline: do not invent
			// a fresh boundary immediately before or after <value:...>.
			split := latestSafeKoreanVisualBoundary(current)
			if split <= 0 || split >= len(current) {
				return "", fmt.Errorf("message %d dialogue token run exceeds %d units without a safe layout boundary: %q (%d units)", id, limit, strings.Join(current, ""), width)
			}
			prefix := strings.Join(current[:split], "")
			prefixWidth, err := e.measureKoreanRenderer(prefix, id, mapping)
			if err != nil {
				return "", err
			}
			if prefixWidth > limit {
				// A safe split can be earlier than the first fitting split when value
				// adjacency is involved. Search for a fitting safe prefix explicitly.
				found := false
				for candidate := split - 1; candidate > 0; candidate-- {
					if !safeKoreanVisualBoundary(current, candidate) {
						continue
					}
					candidateText := strings.Join(current[:candidate], "")
					candidateWidth, err := e.measureKoreanRenderer(candidateText, id, mapping)
					if err != nil {
						return "", err
					}
					if candidateWidth <= limit {
						split, prefix, found = candidate, candidateText, true
						break
					}
				}
				if !found {
					return "", fmt.Errorf("message %d dialogue cannot find a renderer-safe semantic boundary within %d units", id, limit)
				}
			}
			lines = append(lines, prefix)
			current = append([]string(nil), current[split:]...)
		}
	}
	if len(current) != 0 {
		lines = append(lines, strings.Join(current, ""))
	}
	return strings.Join(lines, lineBreak), nil
}

func koreanVisualTokens(text string) []string {
	var tokens []string
	cursor := 0
	for _, loc := range controlTag.FindAllStringIndex(text, -1) {
		for _, r := range text[cursor:loc[0]] {
			tokens = append(tokens, string(r))
		}
		tokens = append(tokens, text[loc[0]:loc[1]])
		cursor = loc[1]
	}
	for _, r := range text[cursor:] {
		tokens = append(tokens, string(r))
	}
	return tokens
}

func latestSafeKoreanVisualBoundary(tokens []string) int {
	for split := len(tokens) - 1; split > 0; split-- {
		if safeKoreanVisualBoundary(tokens, split) {
			return split
		}
	}
	return -1
}

func safeKoreanVisualBoundary(tokens []string, split int) bool {
	if split <= 0 || split >= len(tokens) {
		return false
	}
	return !isKoreanRuntimeValueToken(tokens[split-1]) && !isKoreanRuntimeValueToken(tokens[split])
}

func isKoreanRuntimeValueToken(token string) bool {
	return strings.HasPrefix(token, "<value:")
}
