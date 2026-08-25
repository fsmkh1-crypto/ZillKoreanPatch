// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"fmt"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanslots"
)

// compileKoreanBanksWithPlan is the release boundary between authenticated slot
// planning and message-bank materialization. Keeping the full plan here makes it
// harder for callers to accidentally compile Korean text with an unrelated or
// ad-hoc mapping.
func compileKoreanBanksWithPlan(source *corpus.Project, korean *corpus.KoreanProject, banks []corpus.Bank,
	plan koreanslots.Plan, layouts map[int]string) (map[string][]byte, error) {
	texts, err := korean.RuntimeTexts(source)
	if err != nil {
		return nil, err
	}
	required := koreanslots.RequiredCustomRunes(texts)
	if len(required) != len(plan.CustomRunes) {
		return nil, fmt.Errorf("compile Korean banks: slot plan custom rune count %d does not match runtime requirement %d", len(plan.CustomRunes), len(required))
	}
	for index, r := range required {
		if plan.CustomRunes[index] != r {
			return nil, fmt.Errorf("compile Korean banks: slot plan custom runes do not match current runtime text at index %d", index)
		}
		if _, ok := plan.Mapping[r]; !ok {
			return nil, fmt.Errorf("compile Korean banks: slot plan has no mapping for required rune %U", r)
		}
	}
	return compileKoreanBanks(source, korean, banks, plan.Mapping, layouts)
}
