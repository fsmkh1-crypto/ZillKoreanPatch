// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bytes"
	"fmt"
	"unicode"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/fixeddata"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/slotaudit"
)

// buildKoreanAlphaPlanMobile builds the mobile beta slot plan from the exact
// retail-bound source and runtime-materializable Korean project used by the ISO
// compiler. The mobile font path performs a full authenticated atlas+PAF repack.
//
// This experiment deliberately restricts custom Korean allocation to installed
// renderer keys whose CP932 bytes decode to exactly one CJK Han rune and encode
// back to the identical byte pair. It keeps the historical H1 message/layout
// state unchanged so device testing isolates renderer-slot selection alone.
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
	fixedKorean, err := loadKoreanFixedEBOOT(root)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, err
	}
	texts = append(texts, fixeddata.KoreanEBOOTTexts(fixedKorean)...)

	installed := font.DoubleByteKeys()
	hanInstalled := make([]cp932.GlyphKey, 0, len(installed))
	for _, key := range installed {
		encoded, err := key.Bytes()
		if err != nil {
			continue
		}
		decoded, err := cp932.Decode(encoded)
		if err != nil {
			continue
		}
		runes := []rune(decoded)
		if len(runes) != 1 || !unicode.Is(unicode.Han, runes[0]) {
			continue
		}
		roundTrip, err := cp932.Encode(decoded)
		if err != nil || !bytes.Equal(roundTrip, encoded) {
			continue
		}
		hanInstalled = append(hanInstalled, key)
	}

	stock := koreanslots.RequiredStockKeys(texts)
	custom := koreanslots.RequiredCustomRunes(texts)
	fmt.Printf("Korean beta slot preflight: installed_double_byte=%d han_roundtrip_installed=%d stock_required=%d fixed_reserved=%d boot_scan_keys=%d bindata_scan_keys=%d total_reserved=%d custom=%d materializable_korean=%d fixed_korean=%d total_records=%d\n",
		len(installed), len(hanInstalled), len(stock), len(usedFixed), len(bootScan.Keys), len(bindataScan.Keys), len(reserved), len(custom), len(korean.Entries), len(fixedKorean), len(source.Items))

	plan, err := koreanslots.BuildPlan(texts, hanInstalled, rendererKeySetSlice(reserved))
	if err != nil {
		return koreanslots.Plan{}, 0, 0, fmt.Errorf("mobile beta Han-only slot allocation: %w", err)
	}
	fmt.Printf("FORENSIC SAFE_SLOT_PLAN policy=single_han_roundtrip candidates=%d custom=%d headroom=%d enough=%t\n",
		len(plan.Candidates), len(plan.CustomRunes), len(plan.Candidates)-len(plan.CustomRunes), len(plan.Candidates) >= len(plan.CustomRunes))

	for _, r := range []rune{'게', '깃'} {
		key, ok := plan.Mapping[r]
		if !ok {
			fmt.Printf("FORENSIC SAFE_SLOT rune=%q unicode=U+%04X mapped=false\n", r, r)
			continue
		}
		encoded, keyErr := key.Bytes()
		nominal := ""
		if keyErr == nil {
			nominal, _ = cp932.Decode(encoded)
		}
		fmt.Printf("FORENSIC SAFE_SLOT rune=%q unicode=U+%04X key=%04X bytes=% X nominal_cp932=%q\n", r, r, uint16(key), encoded, nominal)
	}

	return plan, len(korean.Entries), len(source.Items), nil
}
