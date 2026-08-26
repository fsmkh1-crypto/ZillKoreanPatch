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

// buildKoreanAlphaPlanMobile keeps authenticated structured reservations but,
// like the English patch's complete font transform, allocates custom runes from
// the full installed two-byte PAF repertoire rather than requiring a Korean rune
// to fit the retail cell geometry that happened to belong to that key. The full
// Korean font repacker rewrites atlas placement and PAF geometry coherently.
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

	reserved := make(map[cp932.GlyphKey]struct{})
	mergeRendererKeys(reserved, usedFixed)
	mergeRendererKeys(reserved, bootScan.Keys)
	mergeRendererKeys(reserved, bindataScan.Keys)

	texts, err := korean.RuntimeTexts(source)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, err
	}
	plan, err := koreanslots.BuildPlan(
		texts,
		font.DoubleByteKeys(),
		rendererKeySetSlice(reserved),
	)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, fmt.Errorf("mobile alpha full-font plan allocation: %w", err)
	}
	return plan, koreanSummary.Records, sourceSummary.Records, nil
}
