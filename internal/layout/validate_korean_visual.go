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
// reflow contract without rewriting canonical Korean. Character-profile body
// rows are fixed to at most profileAdvance units and profileMaxLines lines. The
// Korean renderer installs custom glyphs at advance 12, so a rune-count guess is
// not authoritative: every candidate is remeasured with the actual renderer
// mapping before it is accepted.
//
// Authored Korean Layout is authoritative and is never replaced. For semantic-
// only rows, try the widest conservative hard-wrap first and tighten only when
// required by a dynamic reservation such as <value:$28>. If no semantic-
// preserving layout satisfies both English hard limits, fail closed rather than
// shortening translation text automatically.
func (e *Engine) DeriveKoreanEnglishVisualLayouts(source *corpus.Project, korean *corpus.KoreanProject, layouts map[int]string, mapping koreanslots.Mapping) (map[int]string, int, error) {
	if source == nil || korean == nil {
		return nil, 0, fmt.Errorf("Korean English visual derivation: nil project")
	}
	if len(mapping) == 0 && len(korean.Entries) != 0 {
		return nil, 0, fmt.Errorf("Korean English visual derivation: empty renderer mapping")
	}
	derived := make(map[int]string, len(layouts))
	for id, text := range layouts {
		derived[id] = text
	}
	count := 0
	for _, row := range korean.Entries {
		if !e.category(row.ID, "character-profile") || row.Layout != "" {
			continue
		}
		item, ok := source.Find(row.ID)
		if !ok {
			return nil, 0, fmt.Errorf("character-profile message %d lacks source", row.ID)
		}
		effective := effectiveKoreanText(row, derived)
		width, lines, err := e.koreanProfileMetrics(item.Record, effective, row.ID, mapping)
		if err != nil {
			return nil, 0, err
		}
		if width <= profileAdvance && lines <= profileMaxLines {
			continue
		}

		accepted := ""
		lastWidth, lastLines := width, lines
		for hard := 24; hard >= 12; hard-- {
			candidate := wrapKoreanProfileVisual(effective, hard)
			if !message.PreservesLayoutSemantics(row.Korean, candidate) {
				continue
			}
			candidateWidth, candidateLines, err := e.koreanProfileMetrics(item.Record, candidate, row.ID, mapping)
			if err != nil {
				return nil, 0, err
			}
			lastWidth, lastLines = candidateWidth, candidateLines
			if candidateWidth <= profileAdvance && candidateLines <= profileMaxLines {
				accepted = candidate
				break
			}
		}
		if accepted == "" {
			return nil, 0, fmt.Errorf("message %d character-profile cannot be made English-contract safe by layout alone: width=%d/%d lines=%d/%d", row.ID, lastWidth, profileAdvance, lastLines, profileMaxLines)
		}
		derived[row.ID] = accepted
		count++
	}
	return derived, count, nil
}

// ValidateKoreanEnglishVisualContracts mirrors the release-blocking visual
// contracts in the upstream English Reflow path. English treats most line-width
// ceilings as authoring warnings, not engine failures; those remain QA signals
// and are deliberately not promoted here. Character profiles are different:
// English fails the build when a projected line exceeds profileAdvance or a
// fragment exceeds profileMaxLines, so Korean must enforce those same hard
// conditions using the renderer metrics that the Korean font build will
// actually install.
//
// With authenticated retail bytes, this uses the exact source projection. In
// repository-only CI Record.Raw is intentionally absent, so source tokens cannot
// be reconstructed; in that case we conservatively validate the effective
// annotated Korean text directly. The asset-bound build remains the exact gate.
func (e *Engine) ValidateKoreanEnglishVisualContracts(source *corpus.Project, korean *corpus.KoreanProject, layouts map[int]string, mapping koreanslots.Mapping) error {
	if source == nil || korean == nil {
		return fmt.Errorf("Korean English visual-contract validation: nil project")
	}
	if len(mapping) == 0 && len(korean.Entries) != 0 {
		return fmt.Errorf("Korean English visual-contract validation: empty renderer mapping")
	}
	var failures []string
	for _, row := range korean.Entries {
		if !e.category(row.ID, "character-profile") {
			continue
		}
		item, ok := source.Find(row.ID)
		if !ok {
			failures = append(failures, fmt.Sprintf("character-profile message %d lacks source", row.ID))
			continue
		}
		text := effectiveKoreanText(row, layouts)
		width, lines, err := e.koreanProfileMetrics(item.Record, text, row.ID, mapping)
		if err != nil {
			failures = append(failures, err.Error())
			continue
		}
		if width > profileAdvance {
			failures = append(failures, fmt.Sprintf("message %d: profile line is %d units (maximum %d)", row.ID, width, profileAdvance))
		}
		if lines > profileMaxLines {
			failures = append(failures, fmt.Sprintf("message %d: profile exceeds %d lines", row.ID, profileMaxLines))
		}
	}
	if len(failures) != 0 {
		sort.Strings(failures)
		return fmt.Errorf("Korean upstream-English visual validation failed:\n- %s", strings.Join(failures, "\n- "))
	}
	return nil
}

func (e *Engine) koreanProfileMetrics(record corpus.Record, text string, id int, mapping koreanslots.Mapping) (int, int, error) {
	projection, err := message.Project(record)
	if err == nil {
		width, err := e.maxProjectedKoreanWidth(projection, text, id, mapping)
		if err != nil {
			return 0, 0, err
		}
		lines, err := maxProjectedKoreanFragmentLines(projection, text, id, mapping)
		if err != nil {
			return 0, 0, err
		}
		return width, lines, nil
	}
	if len(record.Raw) == 0 {
		// Asset-free source projects deliberately have no token stream. Treat
		// the whole annotated profile as one semantic fragment; this cannot hide
		// a width/line overflow. Authenticated mobile/desktop builds still use the
		// exact source projection above.
		return e.measureKoreanProfileText(text, id, mapping)
	}
	return 0, 0, fmt.Errorf("character-profile message %d projection: %w", id, err)
}

// measureKoreanRenderer mirrors Engine.measure but follows the produced Korean
// PAF rather than the nominal CP932 metric of a repurposed renderer key. A
// mapped custom rune is rewritten by the authenticated Korean font transform to
// KoreanTargetAdvance even when the retail/English metric attached to that key
// was narrower punctuation. Unmapped stock text keeps the English metric table.
func (e *Engine) measureKoreanRenderer(s string, id int, mapping koreanslots.Mapping) (int, error) {
	reserved := strings.Count(s, "<value:$28>") * e.playerNameAdvance
	for tag, advance := range e.postingAdvances[id] {
		reserved += strings.Count(s, tag) * advance
	}
	plain := visible(s)
	total := 0
	for i, r := range plain {
		if _, mapped := mapping[r]; mapped {
			total += zillfont.KoreanTargetAdvance
			continue
		}
		key := uint16(r)
		if r > unicode.MaxASCII {
			encoded, err := cp932.Encode(string(r))
			if err != nil {
				return 0, fmt.Errorf("message %d character %q at %d: %w", id, r, i, err)
			}
			if len(encoded) < 1 || len(encoded) > 2 {
				return 0, fmt.Errorf("message %d character %q has invalid CP932 width", id, r)
			}
			key = uint16(encoded[0])
			if len(encoded) == 2 {
				key |= uint16(encoded[1]) << 8
			}
		}
		g, ok := e.glyphs[key]
		if !ok {
			return 0, fmt.Errorf("message %d character %q has no installed-font glyph (%#04x)", id, r, key)
		}
		total += g.Advance
	}
	return total + reserved, nil
}

func (e *Engine) measureKoreanProfileText(text string, id int, mapping koreanslots.Mapping) (int, int, error) {
	maximumWidth := 0
	maximumLines := 0
	for _, page := range strings.Split(text, "<end>") {
		lines := strings.Split(page, lineBreak)
		if strings.TrimSpace(page) != "" {
			maximumLines = max(maximumLines, len(lines))
		}
		for _, line := range lines {
			width, err := e.measureKoreanRenderer(line, id, mapping)
			if err != nil {
				return 0, 0, err
			}
			maximumWidth = max(maximumWidth, width)
		}
	}
	return maximumWidth, maximumLines, nil
}

func (e *Engine) maxProjectedKoreanWidth(p *message.Projection, text string, id int, mapping koreanslots.Mapping) (int, error) {
	parts, err := p.SplitSemanticKorean(text, mapping)
	if err != nil {
		return 0, fmt.Errorf("message %d Korean visual projection: %w", id, err)
	}
	maximum := 0
	for _, part := range parts {
		for _, page := range strings.Split(part, "<end>") {
			for _, line := range strings.Split(page, lineBreak) {
				width, err := e.measureKoreanRenderer(line, id, mapping)
				if err != nil {
					return 0, err
				}
				maximum = max(maximum, width)
			}
		}
	}
	return maximum, nil
}

// maxProjectedKoreanFragmentLines is the Korean counterpart of maxFragmentLines.
// It must traverse the same Korean renderer-aware semantic projection used by
// materialization. Falling back to the stock CP932 split either rejects valid
// Hangul or, in the legacy helper, silently turns a projection error into zero
// lines and can hide a real profile overflow.
func maxProjectedKoreanFragmentLines(p *message.Projection, text string, id int, mapping koreanslots.Mapping) (int, error) {
	parts, err := p.SplitSemanticKorean(text, mapping)
	if err != nil {
		return 0, fmt.Errorf("message %d Korean visual line projection: %w", id, err)
	}
	maximum := 0
	for _, part := range parts {
		for _, page := range strings.Split(part, "<end>") {
			maximum = max(maximum, strings.Count(page, lineBreak)+1)
		}
	}
	return maximum, nil
}

// wrapKoreanProfileVisual is deliberately separate from C5/C22 wrapping. The
// profile box has a 300-unit visual limit and an 8-line hard limit, so its reflow
// should fill substantially more of each line than the storage-oriented wrappers.
// Runtime value adjacency is protected exactly as in the storage wrapper.
func wrapKoreanProfileVisual(text string, hard int) string {
	text = stripDerivedValueAdjacencyBreaks(text)
	soft := hard - 4
	if soft < 1 {
		soft = 1
	}
	var out strings.Builder
	out.Grow(len(text) + len(text)/8)
	lineRunes := 0
	cursor := 0
	protectNextPlainRune := true
	for _, loc := range controlTag.FindAllStringIndex(text, -1) {
		appendKoreanProfilePlain(&out, text[cursor:loc[0]], &lineRunes, protectNextPlainRune, soft, hard)
		protectNextPlainRune = false
		tag := text[loc[0]:loc[1]]
		out.WriteString(tag)
		if tag == lineBreak {
			lineRunes = 0
			protectNextPlainRune = true
		} else {
			protectNextPlainRune = strings.HasPrefix(tag, "<value:")
		}
		cursor = loc[1]
	}
	if cursor < len(text) {
		appendKoreanProfilePlain(&out, text[cursor:], &lineRunes, protectNextPlainRune, soft, hard)
	}
	return out.String()
}

func appendKoreanProfilePlain(out *strings.Builder, text string, lineRunes *int, protectLeading bool, soft, hard int) {
	firstEmitted := true
	for _, r := range text {
		space := r == ' ' || r == '\t' || r == '\r' || r == '\n'
		if space {
			if *lineRunes == 0 {
				if protectLeading && firstEmitted {
					out.WriteRune(r)
					*lineRunes++
					firstEmitted = false
				}
				continue
			}
			if *lineRunes >= soft && !(protectLeading && firstEmitted) {
				out.WriteString(lineBreak)
				*lineRunes = 0
				continue
			}
			out.WriteRune(' ')
			*lineRunes++
			firstEmitted = false
			continue
		}
		if *lineRunes >= hard && !(protectLeading && firstEmitted) {
			out.WriteString(lineBreak)
			*lineRunes = 0
		}
		out.WriteRune(r)
		*lineRunes++
		firstEmitted = false
	}
}
