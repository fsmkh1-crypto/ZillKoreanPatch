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

// ValidateKoreanEnglishVisualContracts mirrors the release-blocking visual
// contracts in the upstream English Reflow path. English treats most line-width
// ceilings as authoring warnings, not engine failures; those remain QA signals
// and are deliberately not promoted here. Character profiles are different:
// English fails the build when a projected line exceeds profileAdvance or a
// fragment exceeds profileMaxLines, so Korean must enforce those same hard
// conditions using the renderer metrics that the Korean font build will
// actually install.
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
		projection, err := message.Project(item.Record)
		if err != nil {
			failures = append(failures, fmt.Sprintf("character-profile message %d projection: %v", row.ID, err))
			continue
		}
		text := effectiveKoreanText(row, layouts)
		width, err := e.maxProjectedKoreanWidth(projection, text, row.ID, mapping)
		if err != nil {
			failures = append(failures, err.Error())
			continue
		}
		if width > profileAdvance {
			failures = append(failures, fmt.Sprintf("message %d: profile line is %d units (maximum %d)", row.ID, width, profileAdvance))
		}
		lines, err := maxProjectedKoreanFragmentLines(projection, text, row.ID, mapping)
		if err != nil {
			failures = append(failures, err.Error())
			continue
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
