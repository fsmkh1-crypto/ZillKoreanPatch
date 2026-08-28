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

// buildKoreanAlphaPlanMobile builds the production-safe mobile beta slot plan
// from the exact retail-bound source and runtime-materializable Korean project
// that the ISO compiler will use. Authenticated fixed renderer keys are reserved
// explicitly. BOOT/EBOOT/bindata are passed to koreanslots.BuildPlan for exact
// raw-byte exclusion; the broader CP932 literal scans remain diagnostics only so
// the same binary risk is not conservatively blocked twice.
func buildKoreanAlphaPlanMobile(root, gameDir string, source *corpus.Project, korean *corpus.KoreanProject) (koreanslots.Plan, int, int, error) {
	if source == nil || korean == nil {
		return koreanslots.Plan{}, 0, 0, fmt.Errorf("mobile beta planner: nil bound source or Korean project")
	}
	font, err := loadRetailPAF(gameDir)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, err
	}

	usedFixed, equipment, err := loadFixedRendererKeys(root)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, err
	}
	eboot, err := loadAuthenticatedRetailEBOOT(gameDir)
	if err != nil {
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

	// Only authenticated fixed-string ownership is a semantic reservation.
	// Binary occurrence safety is handled below by BuildPlan's exact-byte audit.
	reserved := make(map[cp932.GlyphKey]struct{})
	mergeRendererKeys(reserved, usedFixed)

	texts, err := korean.RuntimeTexts(source)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, err
	}
	installed := font.KoreanCompatibleKeys()
	stock := koreanslots.RequiredStockKeys(texts)
	custom := koreanslots.RequiredCustomRunes(texts)
	fmt.Printf("Korean beta slot preflight: installed_compatible=%d stock_required=%d fixed_reserved=%d boot_scan_keys=%d bindata_scan_keys=%d custom=%d materializable_korean=%d total_records=%d\n",
		len(installed), len(stock), len(usedFixed), len(bootScan.Keys), len(bindataScan.Keys), len(custom), len(korean.Entries), len(source.Items))
	plan, err := koreanslots.BuildPlan(texts, installed, rendererKeySetSlice(reserved), boot, eboot, bindata)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, fmt.Errorf("mobile beta full-font plan allocation: %w", err)
	}
	fmt.Printf("Korean beta slot allocation: candidates=%d custom=%d headroom=%d (CP932 literal scans diagnostic-only; BOOT/EBOOT/bindata exact-byte audited)\n",
		len(plan.Candidates), len(plan.CustomRunes), len(plan.Candidates)-len(plan.CustomRunes))
	return plan, len(korean.Entries), len(source.Items), nil
}
