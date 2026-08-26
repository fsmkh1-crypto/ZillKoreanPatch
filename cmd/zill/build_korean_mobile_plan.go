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

// buildKoreanAlphaPlanMobile builds a development-only device-alpha plan from
// the exact retail-bound source project that the ISO compiler will use. The
// caller must BindBanks before invoking this function; keeping planning and
// compilation on one bound source prevents token/state drift between the two.
func buildKoreanAlphaPlanMobile(root, gameDir string, source *corpus.Project, korean *corpus.KoreanProject) (koreanslots.Plan, int, int, error) {
	if source == nil || korean == nil {
		return koreanslots.Plan{}, 0, 0, fmt.Errorf("mobile alpha planner: nil bound source or Korean project")
	}
	font, err := loadRetailPAF(gameDir)
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

	// Placeholder alpha intentionally does not reserve heuristic BOOT/bindata
	// literal-scan keys. Those scans remain diagnostic only. Authenticated fixed
	// strings are still protected. This is a device-validation mode, not the
	// production-safe planner.
	reserved := make(map[cp932.GlyphKey]struct{})
	mergeRendererKeys(reserved, usedFixed)

	texts, err := alphaProject.RuntimeTexts(source)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, err
	}
	installed := font.DoubleByteKeys()
	stock := koreanslots.RequiredStockKeys(texts)
	custom := koreanslots.RequiredCustomRunes(texts)
	fmt.Printf("Korean alpha slot preflight: installed_double_byte=%d stock_required=%d fixed_reserved=%d boot_scan_keys=%d bindata_scan_keys=%d custom=%d placeholders=%d accepted_korean=%d total_records=%d\n",
		len(installed), len(stock), len(reserved), len(bootScan.Keys), len(bindataScan.Keys), len(custom), placeholderCount, len(korean.Entries), len(source.Items))
	plan, err := koreanslots.BuildPlan(texts, installed, rendererKeySetSlice(reserved))
	if err != nil {
		return koreanslots.Plan{}, 0, 0, fmt.Errorf("mobile alpha placeholder full-font plan allocation: %w", err)
	}
	fmt.Printf("Korean alpha slot allocation: candidates=%d custom=%d headroom=%d placeholders=%d\n",
		len(plan.Candidates), len(plan.CustomRunes), len(plan.Candidates)-len(plan.CustomRunes), placeholderCount)
	return plan, len(korean.Entries), len(source.Items), nil
}
