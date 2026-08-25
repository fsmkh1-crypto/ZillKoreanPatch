// SPDX-License-Identifier: GPL-3.0-or-later

package koreanslots

import (
	"sort"

	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/slotaudit"
)

// Plan describes a deterministic custom-glyph allocation for final runtime text.
// Reserved contains renderer keys that must not be repurposed for reasons outside
// the runtime message corpus (fixed strings, structured binary scans, and other
// caller-authenticated resource audits).
type Plan struct {
	CustomRunes   []rune
	RequiredStock []cp932.GlyphKey
	Candidates    []cp932.GlyphKey
	Mapping       Mapping
}

// BuildPlan derives the production slot allocation from final runtime text.
// It preserves CP932 glyphs still needed by that text, applies caller-supplied
// reservations, then conservatively removes any key whose exact two-byte
// sequence occurs in an authenticated binary blob before allocating custom
// glyphs. All sets are normalized so equivalent inputs produce the same plan.
func BuildPlan(texts []string, installed, reserved []cp932.GlyphKey, authenticatedBlobs ...[]byte) (Plan, error) {
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

	var err error
	candidates, err = slotaudit.ExcludeExactByteReferences(candidates, authenticatedBlobs...)
	if err != nil {
		return Plan{}, err
	}
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
