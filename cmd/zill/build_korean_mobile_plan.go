// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bytes"
	"fmt"
	"sort"
	"unicode"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/fixeddata"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/slotaudit"
)

// buildKoreanAlphaPlanMobile builds the mobile beta slot plan from the exact
// retail-bound source and runtime-materializable Korean project used by the ISO
// compiler. The mobile font path performs a full authenticated atlas+PAF repack,
// so every installed double-byte renderer key is eligible; it is not constrained
// by the retail cell geometry required by the older atlas-only desktop path.
//
// Authenticated fixed-string ownership and CP932 literals recovered from retail
// BOOT.BIN/bindata.dat are reserved before custom Korean glyph allocation. The
// scanner intentionally accepts only plausible NUL-terminated CP932 strings;
// arbitrary raw byte-pair occurrence across binary blobs is not treated as slot
// ownership because that would eliminate nearly every candidate by chance.
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

	// Preserve the exact H0 message-corpus allocation first. Adding EBOOT text to
	// BuildPlan before allocation can reshuffle every rune->key assignment, because
	// Allocate pairs sorted runes with sorted candidates by index. Runtime A/B
	// showed that a global remap changes automatic wrapping and, combined with the
	// EBOOT overlay, can reproduce the opening freeze. Keep the proven H0 mapping
	// stable and only relocate renderer-private keys below.
	messageTexts, err := korean.RuntimeTexts(source)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, err
	}
	installed := font.DoubleByteKeys()
	basePlan, err := koreanslots.BuildPlan(messageTexts, installed, rendererKeySetSlice(reserved))
	if err != nil {
		return koreanslots.Plan{}, 0, 0, fmt.Errorf("mobile beta H0 slot allocation: %w", err)
	}

	fixedKorean, err := loadKoreanFixedEBOOT(root)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, err
	}
	combinedTexts := append([]string(nil), messageTexts...)
	combinedTexts = append(combinedTexts, fixeddata.KoreanEBOOTTexts(fixedKorean)...)
	combinedCustom := koreanslots.RequiredCustomRunes(combinedTexts)
	if !sameRunes(basePlan.CustomRunes, combinedCustom) {
		return koreanslots.Plan{}, 0, 0, fmt.Errorf("mobile beta stable mapping: EBOOT overlay changes custom rune set (%d message-only, %d combined); explicit extension required", len(basePlan.CustomRunes), len(combinedCustom))
	}
	combinedStock := koreanslots.RequiredStockKeys(combinedTexts)

	mapping := make(koreanslots.Mapping, len(basePlan.Mapping))
	used := make(map[cp932.GlyphKey]rune, len(basePlan.Mapping))
	for r, key := range basePlan.Mapping {
		mapping[r] = key
		used[key] = r
	}
	// A newly required stock key may not be repurposed by the preserved H0 map.
	for _, key := range combinedStock {
		if r, collision := used[key]; collision {
			return koreanslots.Plan{}, 0, 0, fmt.Errorf("mobile beta stable mapping: EBOOT stock key 0x%04X is owned by custom rune %U", uint16(key), r)
		}
	}

	// Renderer evidence: H0 mapped '게' to SJIS 87 45 (nominal ⑥) and '깃' to
	// 87 4D (nominal ⑭); the game intercepted those keys as UI icons even though
	// the PAF rasters were correct. Preserve every other H0 assignment and move
	// only mappings in the CP932 0x87 special-character row to unused, reversible
	// CJK Han slots. This is intentionally much narrower than the rejected global
	// Han-only remap.
	spares := make([]cp932.GlyphKey, 0, len(basePlan.Candidates)-len(mapping))
	for _, key := range basePlan.Candidates {
		if _, inUse := used[key]; inUse {
			continue
		}
		if isRoundTripHanKey(key) {
			spares = append(spares, key)
		}
	}
	spareIndex := 0
	runes := append([]rune(nil), basePlan.CustomRunes...)
	sort.Slice(runes, func(i, j int) bool { return runes[i] < runes[j] })
	relocated := 0
	for _, r := range runes {
		oldKey := mapping[r]
		if !isRendererPrivate87Key(oldKey) {
			continue
		}
		if spareIndex >= len(spares) {
			return koreanslots.Plan{}, 0, 0, fmt.Errorf("mobile beta stable mapping: need more safe spare keys after %d relocations", relocated)
		}
		newKey := spares[spareIndex]
		spareIndex++
		delete(used, oldKey)
		mapping[r] = newKey
		used[newKey] = r
		relocated++
		oldBytes, _ := oldKey.Bytes()
		newBytes, _ := newKey.Bytes()
		oldNominal, _ := cp932.Decode(oldBytes)
		newNominal, _ := cp932.Decode(newBytes)
		fmt.Printf("FORENSIC STABLE_RELOCATE rune=%q unicode=%U old=%02X %02X nominal_old=%q new=%02X %02X nominal_new=%q\n",
			string(r), r, oldBytes[0], oldBytes[1], oldNominal, newBytes[0], newBytes[1], newNominal)
	}

	plan := basePlan
	plan.CustomRunes = combinedCustom
	plan.RequiredStock = combinedStock
	plan.Mapping = mapping
	fmt.Printf("Korean beta stable slot plan: h0_candidates=%d custom=%d headroom=%d relocated_private87=%d fixed_korean=%d materializable_korean=%d total_records=%d\n",
		len(plan.Candidates), len(plan.CustomRunes), len(plan.Candidates)-len(plan.CustomRunes), relocated, len(fixedKorean), len(korean.Entries), len(source.Items))
	return plan, len(korean.Entries), len(source.Items), nil
}

func sameRunes(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func isRendererPrivate87Key(key cp932.GlyphKey) bool {
	encoded, err := key.Bytes()
	return err == nil && len(encoded) == 2 && encoded[0] == 0x87
}

func isRoundTripHanKey(key cp932.GlyphKey) bool {
	encoded, err := key.Bytes()
	if err != nil || len(encoded) != 2 {
		return false
	}
	decoded, err := cp932.Decode(encoded)
	if err != nil {
		return false
	}
	runes := []rune(decoded)
	if len(runes) != 1 || !unicode.Is(unicode.Han, runes[0]) {
		return false
	}
	roundTrip, err := cp932.Encode(decoded)
	return err == nil && bytes.Equal(roundTrip, encoded)
}
