// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"

	"github.com/HK47196/zill/internal/corpus"
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
//
// The production planner conservatively reserves CP932-looking literals found
// in BOOT.BIN and bindata.dat. That is appropriate for release safety, but it
// prevents a renderer smoke test because those broad binary reservations consume
// most installed two-byte keys even after message text is replaced by [JP]. The
// placeholder alpha therefore authenticates and scans those binaries for
// diagnostics, but reserves only explicitly modelled fixed renderer strings.
// Some non-message Japanese UI may render incorrectly in this throw-away alpha;
// that is acceptable because the purpose is to validate Korean font/message/ISO
// integration, not to ship a mixed-language build.
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

	texts, err := alphaProject.RuntimeTexts(source)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, err
	}
	installed := font.DoubleByteKeys()
	stock := koreanslots.RequiredStockKeys(texts)
	custom := koreanslots.RequiredCustomRunes(texts)

	// Print the preflight before allocation so a capacity failure still reports
	// which reservation class is responsible.
	fmt.Printf("Korean alpha slot preflight: installed_double_byte=%d stock_required=%d fixed_reserved=%d boot_scan_keys=%d bindata_scan_keys=%d custom=%d placeholders=%d accepted_korean=%d total_records=%d\n",
		len(installed), len(stock), installedReferenceCount(installed, usedFixed), installedReferenceCount(installed, bootScan.Keys), installedReferenceCount(installed, bindataScan.Keys), len(custom), placeholderCount, koreanSummary.Records, sourceSummary.Records)

	plan, err := koreanslots.BuildPlan(texts, installed, rendererKeySetSlice(usedFixed))
	if err != nil {
		return koreanslots.Plan{}, 0, 0, fmt.Errorf("mobile alpha placeholder full-font plan allocation: %w", err)
	}
	fmt.Printf("Korean alpha slot result: candidates=%d allocated=%d headroom=%d\n",
		len(plan.Candidates), len(plan.Mapping), len(plan.Candidates)-len(plan.Mapping))
	return plan, koreanSummary.Records, sourceSummary.Records, nil
}
