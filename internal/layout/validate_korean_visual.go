// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/message"
	"github.com/HK47196/zill/internal/zillfont"
)

// DeriveKoreanEnglishVisualLayouts mirrors the upstream English character-profile
// reflow contract without rewriting canonical Korean. It measures the actual
// Korean renderer advances and fills each profile line up to profileAdvance,
// while retaining the English hard profileMaxLines limit.
func (e *Engine) DeriveKoreanEnglishVisualLayouts(source *corpus.Project, korean *corpus.KoreanProject, layouts map[int]string, mapping koreanslots.Mapping) (map[int]string, int, error) {
	if source == nil || korean == nil {
		return nil, 0, fmt.Errorf("Korean English visual derivation: nil project")
	}
	if len(mapping) == 0 && len(korean.Entries) != 0 {
		return nil, 0, fmt.Errorf("Korean English visual derivation: empty renderer mapping")
	}
	derived := make(map[int]string, len(layouts))
	for id, text := range layouts { derived[id] = text }
	count := 0
	for _, row := range korean.Entries {
		if !e.category(row.ID, "character-profile") || row.Layout != "" { continue }
		item, ok := source.Find(row.ID)
		if !ok { return nil, 0, fmt.Errorf("character-profile message %d lacks source", row.ID) }
		effective := effectiveKoreanText(row, derived)
		width, lines, err := e.koreanProfileMetrics(item.Record, effective, row.ID, mapping)
		if err != nil { return nil, 0, err }
		if width <= profileAdvance && lines <= profileMaxLines { continue }

		candidate, err := e.wrapKoreanProfileVisual(effective, row.ID, mapping)
		if err != nil { return nil, 0, err }
		if !message.PreservesLayoutSemantics(row.Korean, candidate) {
			return nil, 0, fmt.Errorf("message %d character-profile derived layout changes semantic/control text", row.ID)
		}
		postWidth, postLines, err := e.koreanProfileMetrics(item.Record, candidate, row.ID, mapping)
		if err != nil { return nil, 0, err }
		if postWidth > profileAdvance || postLines > profileMaxLines {
			return nil, 0, fmt.Errorf("message %d character-profile cannot be made English-contract safe by layout alone: width=%d/%d lines=%d/%d", row.ID, postWidth, profileAdvance, postLines, profileMaxLines)
		}
		derived[row.ID] = candidate
		count++
	}
	return derived, count, nil
}

// ValidateKoreanEnglishVisualContracts mirrors the release-blocking visual
// contracts in the upstream English Reflow path. Character profiles are hard
// failures upstream, so Korean enforces the same width and line-count limits
// using the renderer metrics that the Korean font build actually installs.
func (e *Engine) ValidateKoreanEnglishVisualContracts(source *corpus.Project, korean *corpus.KoreanProject, layouts map[int]string, mapping koreanslots.Mapping) error {
	if source == nil || korean == nil { return fmt.Errorf("Korean English visual-contract validation: nil project") }
	if len(mapping) == 0 && len(korean.Entries) != 0 { return fmt.Errorf("Korean English visual-contract validation: empty renderer mapping") }
	var failures []string
	for _, row := range korean.Entries {
		if !e.category(row.ID, "character-profile") { continue }
		item, ok := source.Find(row.ID)
		if !ok { failures = append(failures, fmt.Sprintf("character-profile message %d lacks source", row.ID)); continue }
		text := effectiveKoreanText(row, layouts)
		width, lines, err := e.koreanProfileMetrics(item.Record, text, row.ID, mapping)
		if err != nil { failures = append(failures, err.Error()); continue }
		if width > profileAdvance { failures = append(failures, fmt.Sprintf("message %d: profile line is %d units (maximum %d)", row.ID, width, profileAdvance)) }
		if lines > profileMaxLines { failures = append(failures, fmt.Sprintf("message %d: profile exceeds %d lines", row.ID, profileMaxLines)) }
	}
	if len(failures) != 0 {
		sort.Strings(failures)
		return fmt.Errorf("Korean upstream-English visual validation failed:\n- %s", strings.Join(failures, "\n- "))
	}
	return nil
}

func (e *Engine) koreanProfileMetrics(record corpus.Record, text string, id int, mapping koreanslots.Mapping) (int, int, error) {
	return e.koreanVisualMetrics(record, text, id, mapping, fmt.Sprintf("character-profile message %d projection", id))
}

// koreanVisualMetrics is the single Korean projection/fallback path for visual
// width and vertical metrics. Callers may retain context-specific error labels,
// but projection, renderer measurement and repository-only fallback semantics
// must not diverge between hard profile checks and warning-level audits.
func (e *Engine) koreanVisualMetrics(record corpus.Record, text string, id int, mapping koreanslots.Mapping, projectionError string) (int, int, error) {
	projection, err := message.Project(record)
	if err == nil {
		width, err := e.maxProjectedKoreanWidth(projection, text, id, mapping)
		if err != nil { return 0, 0, err }
		lines, err := maxProjectedKoreanFragmentLines(projection, text, id, mapping)
		if err != nil { return 0, 0, err }
		return width, lines, nil
	}
	if len(record.Raw) == 0 {
		return e.measureKoreanProfileText(text, id, mapping)
	}
	return 0, 0, fmt.Errorf("%s: %w", projectionError, err)
}

// measureKoreanRenderer follows the produced Korean PAF rather than nominal
// CP932 metrics of repurposed renderer keys. RendererRune is shared with the
// encoder so typography aliases have identical storage and visual widths.
func (e *Engine) measureKoreanRenderer(s string, id int, mapping koreanslots.Mapping) (int, error) {
	reserved := strings.Count(s, "<value:$28>") * e.playerNameAdvance
	for tag, advance := range e.postingAdvances[id] { reserved += strings.Count(s, tag) * advance }
	plain := visible(s)
	total := 0
	for i, raw := range plain {
		r := koreanslots.RendererRune(raw)
		if _, mapped := mapping[r]; mapped { total += zillfont.KoreanTargetAdvance; continue }
		key := uint16(r)
		if r > unicode.MaxASCII {
			encoded, err := cp932.Encode(string(r))
			if err != nil { return 0, fmt.Errorf("message %d character %q at %d (renderer %q): %w", id, raw, i, r, err) }
			if len(encoded) < 1 || len(encoded) > 2 { return 0, fmt.Errorf("message %d character %q has invalid CP932 width", id, raw) }
			key = uint16(encoded[0]); if len(encoded) == 2 { key |= uint16(encoded[1]) << 8 }
		}
		g, ok := e.glyphs[key]
		if !ok { return 0, fmt.Errorf("message %d character %q (renderer %q) has no installed-font glyph (%#04x)", id, raw, r, key) }
		total += g.Advance
	}
	return total + reserved, nil
}

func (e *Engine) measureKoreanProfileText(text string, id int, mapping koreanslots.Mapping) (int, int, error) {
	maximumWidth, maximumLines := 0, 0
	for _, page := range strings.Split(text, "<end>") {
		lines := strings.Split(page, lineBreak)
		if strings.TrimSpace(page) != "" { maximumLines = max(maximumLines, len(lines)) }
		for _, line := range lines {
			width, err := e.measureKoreanRenderer(line, id, mapping)
			if err != nil { return 0, 0, err }
			maximumWidth = max(maximumWidth, width)
		}
	}
	return maximumWidth, maximumLines, nil
}

func (e *Engine) maxProjectedKoreanWidth(p *message.Projection, text string, id int, mapping koreanslots.Mapping) (int, error) {
	parts, err := p.SplitSemanticKorean(text, mapping)
	if err != nil { return 0, fmt.Errorf("message %d Korean visual projection: %w", id, err) }
	maximum := 0
	for _, part := range parts {
		for _, page := range strings.Split(part, "<end>") {
			for _, line := range strings.Split(page, lineBreak) {
				width, err := e.measureKoreanRenderer(line, id, mapping)
				if err != nil { return 0, err }
				maximum = max(maximum, width)
			}
		}
	}
	return maximum, nil
}

func maxProjectedKoreanFragmentLines(p *message.Projection, text string, id int, mapping koreanslots.Mapping) (int, error) {
	parts, err := p.SplitSemanticKorean(text, mapping)
	if err != nil { return 0, fmt.Errorf("message %d Korean visual line projection: %w", id, err) }
	maximum := 0
	for _, part := range parts {
		for _, page := range strings.Split(part, "<end>") { maximum = max(maximum, strings.Count(page, lineBreak)+1) }
	}
	return maximum, nil
}

// wrapKoreanProfileVisual greedily fills each line using the actual Korean
// renderer advance. It prefers whitespace boundaries, matching the English
// profile reflow intent. Existing semantic <end>/<line-break> boundaries remain
// hard boundaries. A word containing <value:$28> is measured with the player-name
// reservation, so dynamic names cannot silently push a derived line over 300.
func (e *Engine) wrapKoreanProfileVisual(text string, id int, mapping koreanslots.Mapping) (string, error) {
	return wrapKoreanDelimitedParagraphs(text, func(paragraph string) (string, error) {
		return e.wrapKoreanProfileParagraph(paragraph, id, mapping)
	})
}

func (e *Engine) wrapKoreanProfileParagraph(text string, id int, mapping koreanslots.Mapping) (string, error) {
	if strings.TrimSpace(text) == "" { return text, nil }
	words := strings.Fields(text)
	if len(words) == 0 { return text, nil }
	lines := make([]string, 0, 8)
	current := ""
	flush := func() { if current != "" { lines = append(lines, current); current = "" } }

	for _, word := range words {
		candidate := word
		if current != "" { candidate = current + " " + word }
		width, err := e.measureKoreanRenderer(candidate, id, mapping)
		if err != nil { return "", err }
		if width <= profileAdvance { current = candidate; continue }
		if current != "" {
			// Never create a derived boundary immediately before a runtime value.
			// Move the preceding word with it when possible, giving the dispatcher
			// ordinary text on both sides of the line boundary.
			if strings.HasPrefix(word, "<value:") {
				if split := strings.LastIndex(current, " "); split >= 0 {
					prefix, tail := current[:split], current[split+1:]
					lines = append(lines, prefix)
					current = tail + " " + word
					width, err = e.measureKoreanRenderer(current, id, mapping)
					if err != nil { return "", err }
					if width <= profileAdvance { continue }
				}
			}
			flush()
		}
		wordWidth, err := e.measureKoreanRenderer(word, id, mapping)
		if err != nil { return "", err }
		if wordWidth > profileAdvance {
			return "", fmt.Errorf("message %d character-profile word exceeds %d units and cannot be whitespace-reflowed: %q (%d units)", id, profileAdvance, word, wordWidth)
		}
		current = word
	}
	flush()
	return strings.Join(lines, lineBreak), nil
}
