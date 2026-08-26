// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/fixeddata"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/slotaudit"
)

const mobileAlphaPlaceholder = "[JP]"

// buildKoreanAlphaPlanMobile builds a development-only device-alpha plan.
// Planning happens before retail message banks are bound, so it must not depend
// on Record.Tokens. Accepted Korean rows contribute their runtime Korean/layout;
// untranslated visible Japanese rows contribute only a tiny ASCII marker. This
// mirrors the compiler-side placeholder project that is built after BindBanks.
func buildKoreanAlphaPlanMobile(root, gameDir string) (koreanslots.Plan, int, int, error) {
	font, err := loadRetailPAF(gameDir)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, err
	}
	source, sourceSummary, err := corpus.LoadProject(root)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, err
	}
	korean, koreanSummary, err := corpus.LoadKoreanProject(root, source)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, err
	}

	usedFixed, equipment, err := loadFixedRendererKeys(root)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, err
	}
	if _, err := loadAuthenticatedRetailEBOOT(gameDir); err != nil {
		return koreanslots.Plan{}, 0, 0, err
	}
	boot, err := loadAuthenticatedRetailBOOT(gameDir)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, err
	}
	bootScan, err := slotaudit.ScanCP932Literals(boot)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, fmt.Errorf("scan BOOT.BIN: %w", err)
	}
	bindata, err := loadRetailBindata(gameDir)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, err
	}
	if _, err := fixeddata.ApplyEquipment(bindata, equipment); err != nil {
		return koreanslots.Plan{}, 0, 0, fmt.Errorf("validate bindata.dat layout: %w", err)
	}
	bindataScan, err := slotaudit.ScanCP932Literals(bindata)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, fmt.Errorf("scan bindata.dat: %w", err)
	}

	// Placeholder alpha intentionally does not reserve heuristic BOOT/bindata
	// literal-scan keys. Those scans remain diagnostic only. Authenticated fixed
	// strings are still protected. This is a device-validation mode, not the
	// production-safe planner.
	reserved := make(map[cp932.GlyphKey]struct{})
	mergeRendererKeys(reserved, usedFixed)

	accepted := make(map[int]corpus.KoreanEntry, len(korean.Entries))
	for _, row := range korean.Entries {
		accepted[row.ID] = row
	}
	texts := make([]string, 0, len(source.Items))
	placeholderCount := 0
	for _, item := range source.Items {
		if row, ok := accepted[item.Record.ID]; ok {
			text := row.Korean
			if row.Layout != "" {
				text = row.Layout
			}
			texts = append(texts, text)
			continue
		}
		if item.Translation.Japanese == "" {
			texts = append(texts, "")
			continue
		}
		texts = append(texts, mobileAlphaPlaceholder)
		placeholderCount++
	}

	installed := font.DoubleByteKeys()
	stock := koreanslots.RequiredStockKeys(texts)
	custom := koreanslots.RequiredCustomRunes(texts)
	fmt.Printf("Korean alpha slot preflight: installed_double_byte=%d stock_required=%d fixed_reserved=%d boot_scan_keys=%d bindata_scan_keys=%d custom=%d placeholders=%d accepted_korean=%d total_records=%d\n",
		len(installed), len(stock), len(reserved), len(bootScan.Keys), len(bindataScan.Keys), len(custom), placeholderCount, koreanSummary.Records, sourceSummary.Records)
	plan, err := koreanslots.BuildPlan(texts, installed, rendererKeySetSlice(reserved))
	if err != nil {
		return koreanslots.Plan{}, 0, 0, fmt.Errorf("mobile alpha placeholder full-font plan allocation: %w", err)
	}
	fmt.Printf("Korean alpha slot allocation: candidates=%d custom=%d headroom=%d placeholders=%d\n",
		len(plan.Candidates), len(plan.CustomRunes), len(plan.Candidates)-len(plan.CustomRunes), placeholderCount)
	return plan, koreanSummary.Records, sourceSummary.Records, nil
}
