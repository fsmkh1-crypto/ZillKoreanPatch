// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/fixeddata"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/release"
	"github.com/HK47196/zill/internal/slotaudit"
)

// buildKoreanAlphaPlanMobile builds a development-only device-alpha plan. It
// uses the same accepted Korean overlay as production but replaces untranslated
// Japanese rows in memory with a tiny ASCII marker, freeing the Japanese CP932
// repertoire solely for renderer validation. Production corpus files are never
// modified by this path.
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
	alphaProject, placeholderCount, err := release.BuildKoreanAlphaPlaceholderProject(source, korean)
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

	reserved := make(map[cp932.GlyphKey]struct{})
	mergeRendererKeys(reserved, usedFixed)
	mergeRendererKeys(reserved, bootScan.Keys)
	mergeRendererKeys(reserved, bindataScan.Keys)

	texts, err := alphaProject.RuntimeTexts(source)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, err
	}
	installed := font.DoubleByteKeys()
	stock := koreanslots.RequiredStockKeys(texts)
	plan, err := koreanslots.BuildPlan(texts, installed, rendererKeySetSlice(reserved))
	if err != nil {
		return koreanslots.Plan{}, 0, 0, fmt.Errorf("mobile alpha placeholder full-font plan allocation: %w", err)
	}
	fmt.Printf("Korean alpha slot stats: installed_double_byte=%d stock_required=%d reserved=%d candidates=%d custom=%d placeholders=%d accepted_korean=%d total_records=%d\n",
		len(installed), len(stock), len(reserved), len(plan.Candidates), len(plan.CustomRunes), placeholderCount, koreanSummary.Records, sourceSummary.Records)
	return plan, koreanSummary.Records, sourceSummary.Records, nil
}
