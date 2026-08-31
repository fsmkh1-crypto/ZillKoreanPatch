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

type koreanVisualToken struct {
	kind  string
	value string
}

type koreanVisualBreak struct {
	prefixEnd  int
	suffixStart int
}

// wrapKoreanVisualParagraphToLimit follows message.PreservesLayoutSemantics by
// construction. A generated boundary may either replace one complete semantic
// whitespace span or sit between two adjacent ordinary literal runes. Controls
// remain atomic and no new boundary is placed immediately beside <value:...>.
// Whitespace that is not selected as a wrap boundary is preserved byte-for-byte.
func (e *Engine) wrapKoreanVisualParagraphToLimit(text string, id int, mapping koreanslots.Mapping, limit int) (string, error) {
	if text == "" {
		return text, nil
	}
	tokens := koreanVisualTokens(text)
	if len(tokens) == 0 {
		return text, nil
	}

	lines := make([]string, 0, 4)
	current := make([]koreanVisualToken, 0, len(tokens))
	for _, token := range tokens {
		current = append(current, token)
		for {
			currentText := joinKoreanVisualTokens(current)
			width, err := e.measureKoreanRenderer(currentText, id, mapping)
			if err != nil {
				return "", err
			}
			if width <= limit {
				break
			}

			br, ok, err := e.latestFittingKoreanVisualBreak(current, id, mapping, limit)
			if err != nil {
				return "", err
			}
			if !ok {
				return "", fmt.Errorf("message %d dialogue token run exceeds %d units without a semantic-safe layout boundary: %q (%d units)", id, limit, currentText, width)
			}
			prefix := joinKoreanVisualTokens(current[:br.prefixEnd])
			if prefix == "" {
				return "", fmt.Errorf("message %d dialogue selected an empty visual prefix", id)
			}
			lines = append(lines, prefix)
			current = append([]koreanVisualToken(nil), current[br.suffixStart:]...)
		}
	}
	if len(current) != 0 {
		lines = append(lines, joinKoreanVisualTokens(current))
	}
	return strings.Join(lines, lineBreak), nil
}

func (e *Engine) latestFittingKoreanVisualBreak(tokens []koreanVisualToken, id int, mapping koreanslots.Mapping, limit int) (koreanVisualBreak, bool, error) {
	// Prefer the latest legal boundary, matching greedy upstream reflow intent.
	// Each candidate is measured because replacing a whitespace span removes its
	// rendered advance from the previous line.
	for split := len(tokens) - 1; split > 0; split-- {
		br, ok := koreanVisualBoundaryAt(tokens, split)
		if !ok || br.prefixEnd <= 0 || br.suffixStart >= len(tokens) {
			continue
		}
		prefix := joinKoreanVisualTokens(tokens[:br.prefixEnd])
		width, err := e.measureKoreanRenderer(prefix, id, mapping)
		if err != nil {
			return koreanVisualBreak{}, false, err
		}
		if width <= limit {
			return br, true, nil
		}
	}
	return koreanVisualBreak{}, false, nil
}

func koreanVisualBoundaryAt(tokens []koreanVisualToken, split int) (koreanVisualBreak, bool) {
	if split <= 0 || split >= len(tokens) {
		return koreanVisualBreak{}, false
	}

	// A complete whitespace span may be replaced by one generated boundary.
	if tokens[split-1].kind == "whitespace" || tokens[split].kind == "whitespace" {
		start := split
		for start > 0 && tokens[start-1].kind == "whitespace" {
			start--
		}
		end := split
		for end < len(tokens) && tokens[end].kind == "whitespace" {
			end++
		}
		if start == 0 || end == len(tokens) {
			return koreanVisualBreak{}, false
		}
		if isKoreanRuntimeValueToken(tokens[start-1]) || isKoreanRuntimeValueToken(tokens[end]) {
			return koreanVisualBreak{}, false
		}
		return koreanVisualBreak{prefixEnd: start, suffixStart: end}, true
	}

	// Zero-width insertion is legal only between adjacent ordinary literal
	// runes. Controls (including value substitutions) are deliberately excluded.
	if tokens[split-1].kind == "literal" && tokens[split].kind == "literal" {
		return koreanVisualBreak{prefixEnd: split, suffixStart: split}, true
	}
	return koreanVisualBreak{}, false
}

func koreanVisualTokens(text string) []koreanVisualToken {
	var tokens []koreanVisualToken
	cursor := 0
	appendPlain := func(plain string) {
		for _, r := range plain {
			kind := "literal"
			if unicode.IsSpace(r) {
				kind = "whitespace"
			}
			tokens = append(tokens, koreanVisualToken{kind: kind, value: string(r)})
		}
	}
	for _, loc := range controlTag.FindAllStringIndex(text, -1) {
		appendPlain(text[cursor:loc[0]])
		tokens = append(tokens, koreanVisualToken{kind: "control", value: text[loc[0]:loc[1]]})
		cursor = loc[1]
	}
	appendPlain(text[cursor:])
	return tokens
}

func joinKoreanVisualTokens(tokens []koreanVisualToken) string {
	var out strings.Builder
	for _, token := range tokens {
		out.WriteString(token.value)
	}
	return out.String()
}

func isKoreanRuntimeValueToken(token koreanVisualToken) bool {
	return token.kind == "control" && strings.HasPrefix(token.value, "<value:")
}
