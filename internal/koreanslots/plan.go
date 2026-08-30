// SPDX-License-Identifier: GPL-3.0-or-later

package koreanslots

import (
	"sort"

	"github.com/HK47196/zill/internal/cp932"
)

// Plan describes a deterministic custom-glyph allocation for final runtime text.
// Reserved contains renderer keys that must not be repurposed for reasons outside
// the runtime message corpus. Callers may reserve only structured evidence that
// the engine actually consumes as rendered text (for example fixed strings or
// authenticated CP932 literal scans), never arbitrary whole-blob byte aliases.
type Plan struct {
	CustomRunes   []rune
	RequiredStock []cp932.GlyphKey
	Candidates    []cp932.GlyphKey
	Mapping       Mapping
}

// BuildPlan derives the production slot allocation from final runtime text.
// It preserves CP932 glyphs still needed by that text and applies caller-supplied
// structured renderer reservations before allocating custom glyphs. Whole-blob
// exact-byte exclusion is deliberately not part of this API: arbitrary machine
// code/data byte pairs are not renderer ownership evidence.
func BuildPlan(texts []string, installed, reserved []cp932.GlyphKey) (Plan, error) {
	custom := RequiredCustomRunes(texts)
	stock := RequiredStockKeys(texts)

	blocked := make(map[cp932.GlyphKey]struct{}, len(stock)+len(reserved))
	for _, key := range stock {
		blocked[key] = struct{}{}
	}
	for _, key := range reserved {
		blocked[key] = struct{}{}
	}

	installedSet := make(map[cp932.GlyphKey]struct{}, len(installed))
	for _, key := range installed {
		// Allocate performs the authoritative renderer-key validation below. Keep
		// invalid keys here so malformed planner input cannot be silently ignored.
		installedSet[key] = struct{}{}
	}
	candidates := make([]cp932.GlyphKey, 0, len(installedSet))
	for key := range installedSet {
		if _, keep := blocked[key]; !keep {
			candidates = append(candidates, key)
		}
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i] < candidates[j] })

	mapping, err := Allocate(custom, candidates)
	if err != nil {
		return Plan{}, err
	}
	return Plan{
		CustomRunes:   custom,
		RequiredStock: stock,
		Candidates:    candidates,
		Mapping:       mapping,
	}, nil
}
