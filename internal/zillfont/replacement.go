// SPDX-License-Identifier: GPL-3.0-or-later

package zillfont

import (
	"fmt"
	"sort"

	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
)

// Replacement identifies the existing retail atlas cell that must receive a
// custom Unicode glyph while keeping the renderer key and PAF tree unchanged.
type Replacement struct {
	Rune  rune
	Key   cp932.GlyphKey
	Glyph Glyph
}

// ReplacementPlan resolves every custom renderer mapping to its exact retail
// PAF glyph record. The result is sorted by Unicode rune so bitmap generation
// and archive patching are deterministic.
func (p *PAF) ReplacementPlan(mapping koreanslots.Mapping) ([]Replacement, error) {
	if p == nil {
		return nil, fmt.Errorf("font replacement plan: nil PAF")
	}
	byKey := make(map[cp932.GlyphKey]Glyph, len(p.Glyphs))
	for _, glyph := range p.Glyphs {
		if _, exists := byKey[glyph.Key]; exists {
			return nil, fmt.Errorf("font replacement plan: duplicate PAF key 0x%04X", uint16(glyph.Key))
		}
		byKey[glyph.Key] = glyph
	}

	runes := make([]rune, 0, len(mapping))
	for r := range mapping {
		runes = append(runes, r)
	}
	sort.Slice(runes, func(i, j int) bool { return runes[i] < runes[j] })

	out := make([]Replacement, 0, len(runes))
	seenKeys := make(map[cp932.GlyphKey]rune, len(runes))
	for _, r := range runes {
		key := mapping[r]
		if prior, exists := seenKeys[key]; exists {
			return nil, fmt.Errorf("font replacement plan: runes %U and %U share renderer key 0x%04X", prior, r, uint16(key))
		}
		seenKeys[key] = r
		glyph, ok := byKey[key]
		if !ok {
			return nil, fmt.Errorf("font replacement plan: renderer key 0x%04X for %U is not installed in PAF", uint16(key), r)
		}
		if !key.IsDoubleByte() {
			return nil, fmt.Errorf("font replacement plan: renderer key 0x%04X for %U is not a two-byte slot", uint16(key), r)
		}
		out = append(out, Replacement{Rune: r, Key: key, Glyph: glyph})
	}
	return out, nil
}
