// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"fmt"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/message"
)

// compileKoreanBanks bridges the canonical sparse Korean corpus to the
// mapping-aware message compiler. It deliberately accepts layouts separately:
// canonical Korean text remains semantic translator-owned data while wrapping
// is machine-owned build output.
func compileKoreanBanks(source *corpus.Project, korean *corpus.KoreanProject, banks []corpus.Bank,
	mapping koreanslots.Mapping, layouts map[int]string) (map[string][]byte, error) {
	if source == nil {
		return nil, fmt.Errorf("compile Korean banks: nil source project")
	}
	if korean == nil {
		return nil, fmt.Errorf("compile Korean banks: nil Korean project")
	}

	itemsBySection := make(map[int][]corpus.Item)
	for _, item := range source.Items {
		section := item.Record.ID / 10_000
		itemsBySection[section] = append(itemsBySection[section], item)
	}
	replacementsBySection := make(map[int]map[int]message.KoreanRecord)
	for _, row := range korean.Entries {
		section := row.ID / 10_000
		if replacementsBySection[section] == nil {
			replacementsBySection[section] = make(map[int]message.KoreanRecord)
		}
		replacement := message.KoreanRecord{Text: row.Korean}
		if layout, ok := layouts[row.ID]; ok {
			replacement.Layout = layout
		}
		replacementsBySection[section][row.ID] = replacement
	}

	compiled := make(map[string][]byte, len(banks))
	seenSections := make(map[int]struct{}, len(banks))
	for _, bank := range banks {
		if _, exists := seenSections[bank.Section]; exists {
			return nil, fmt.Errorf("compile Korean banks: duplicate retail section %03d", bank.Section)
		}
		seenSections[bank.Section] = struct{}{}
		items := itemsBySection[bank.Section]
		if len(items) != len(bank.Records) {
			return nil, fmt.Errorf("%s: source project has %d items for %d retail records", bank.Name, len(items), len(bank.Records))
		}
		data, err := message.CompileBankKorean(bank, items, replacementsBySection[bank.Section], mapping)
		if err != nil {
			return nil, err
		}
		if _, exists := compiled[bank.Name]; exists {
			return nil, fmt.Errorf("compile Korean banks: duplicate retail bank name %s", bank.Name)
		}
		compiled[bank.Name] = data
	}
	for section, replacements := range replacementsBySection {
		if len(replacements) == 0 {
			continue
		}
		if _, ok := seenSections[section]; !ok {
			return nil, fmt.Errorf("compile Korean banks: Korean section %03d has no retail bank", section)
		}
	}
	return compiled, nil
}
