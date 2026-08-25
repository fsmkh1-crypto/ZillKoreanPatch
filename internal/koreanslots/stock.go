// SPDX-License-Identifier: GPL-3.0-or-later

package koreanslots

import (
	"sort"

	"github.com/HK47196/zill/internal/cp932"
)

// RequiredStockKeys returns the sorted unique stock renderer keys still needed
// by a set of final runtime texts. Runes that are not CP932-encodable are
// intentionally ignored here because they are handled by custom Korean slot
// allocation instead.
//
// This helper is useful for partial Korean rollouts: callers can pass Korean
// text for replaced records and Japanese source text for records not yet
// translated, so Japanese glyphs cease being reserved as soon as their final
// runtime text no longer needs them.
func RequiredStockKeys(texts []string) []cp932.GlyphKey {
	set := make(map[cp932.GlyphKey]struct{})
	for _, text := range texts {
		for _, r := range text {
			encoded, err := cp932.Encode(string(r))
			if err != nil {
				continue
			}
			key, err := cp932.GlyphKeyFromBytes(encoded)
			if err != nil {
				continue
			}
			set[key] = struct{}{}
		}
	}
	out := make([]cp932.GlyphKey, 0, len(set))
	for key := range set {
		out = append(out, key)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
