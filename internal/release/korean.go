// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"fmt"
	"strings"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/message"
)

// compileKoreanBanks bridges the canonical sparse Korean corpus to the
// mapping-aware message compiler. It deliberately accepts layouts separately:
// canonical Korean text remains semantic translator-owned data while wrapping
// is machine-owned build output.
//
// Independent bank failures are accumulated before returning so one authenticated
// retail build exposes the full QA5 materialization/capacity failure set rather
// than stopping at the first section. No compiled output is returned on failure.
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
	projectionChecked := 0
	var failures []string
	for _, bank := range banks {
		if _, exists := seenSections[bank.Section]; exists {
			failures = append(failures, fmt.Sprintf("duplicate retail section %03d", bank.Section))
			continue
		}
		seenSections[bank.Section] = struct{}{}
		items := itemsBySection[bank.Section]
		if len(items) != len(bank.Records) {
			failures = append(failures, fmt.Sprintf("%s: source project has %d items for %d retail records", bank.Name, len(items), len(bank.Records)))
			continue
		}

		// Audit the Korean-specific semantic splitter against the untouched
		// upstream materializer on every translatable record in the authenticated
		// retail bank. This runs before Korean wording or glyph mapping enters the
		// comparison, so a fixed-line-break/fragment-topology divergence cannot be
		// misdiagnosed later as a translation or renderer problem.
		checked, err := message.VerifyKoreanProjectionCompatibility(bank)
		if err != nil {
			failures = append(failures, err.Error())
			continue
		}
		projectionChecked += checked

		data, err := message.CompileBankKorean(bank, items, replacementsBySection[bank.Section], mapping)
		if err != nil {
			failures = append(failures, err.Error())
			continue
		}
		// Never infer runtime safety from a previous successful playthrough. Every
		// generated bank must independently prove the widened table contract on
		// every build before it can enter the archive.
		if err := message.VerifyWideBank(bank.Name, data, len(bank.Records)); err != nil {
			failures = append(failures, err.Error())
			continue
		}
		if _, exists := compiled[bank.Name]; exists {
			failures = append(failures, fmt.Sprintf("duplicate retail bank name %s", bank.Name))
			continue
		}
		compiled[bank.Name] = data
	}
	for section, replacements := range replacementsBySection {
		if len(replacements) == 0 {
			continue
		}
		if _, ok := seenSections[section]; !ok {
			failures = append(failures, fmt.Sprintf("Korean section %03d has no retail bank", section))
		}
	}
	if len(failures) > 0 {
		return nil, fmt.Errorf("compile Korean banks failed:\n- %s", strings.Join(failures, "\n- "))
	}
	fmt.Printf("Korean projection compatibility audit: %d translatable retail record(s) byte-identical to upstream materialization.\n", projectionChecked)
	return compiled, nil
}
