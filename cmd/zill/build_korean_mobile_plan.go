// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bytes"
	"crypto/sha256"
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
// This experiment intentionally starts from the exact H0 plan, including H0's
// two pre-existing Korean EBOOT strings, and changes only mappings that use the
// CP932 0x87 special-character lead byte. No new title/character-creation EBOOT
// translation bundle is present on this branch.
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
	stock := koreanslots.RequiredStockKeys(texts)
	custom := koreanslots.RequiredCustomRunes(texts)
	fmt.Printf("Korean beta slot preflight: installed_double_byte=%d stock_required=%d fixed_reserved=%d boot_scan_keys=%d bindata_scan_keys=%d total_reserved=%d custom=%d materializable_korean=%d fixed_korean=%d total_records=%d\n",
		len(installed), len(stock), len(usedFixed), len(bootScan.Keys), len(bindataScan.Keys), len(reserved), len(custom), len(korean.Entries), len(fixedKorean), len(source.Items))

	// This is byte-for-byte the H0 planning input. Only after the plan exists do
	// we relocate renderer-private 0x87 mappings, so every unaffected rune keeps
	// its H0 renderer key.
	plan, err := koreanslots.BuildPlan(texts, installed, rendererKeySetSlice(reserved))
	if err != nil {
		return koreanslots.Plan{}, 0, 0, fmt.Errorf("mobile beta H0 slot allocation: %w", err)
	}
	blobs := []slotAuditBlob{
		{Name: "BOOT.BIN", Data: boot},
		{Name: "EBOOT.BIN", Data: eboot},
		{Name: "bindata.dat", Data: bindata},
	}
	audit, err := auditMobileExactByteReuse(plan, blobs...)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, err
	}
	fmt.Printf("FORENSIC MOBILE_EXACT_BYTE_AUDIT phase=h0 candidates=%d exact_hit_candidates=%d mapped_hit_records=%d\n",
		len(plan.Candidates), audit.CandidateHits, len(audit.MappedHits))
	for _, hit := range audit.MappedHits {
		encoded, _ := hit.Key.Bytes()
		fmt.Printf("FORENSIC MOBILE_EXACT_BYTE_MAPPING phase=h0 rune=%q unicode=%U key=%02X %02X blob=%s offsets=%v\n",
			string(hit.Rune), hit.Rune, encoded[0], encoded[1], hit.Blob, hit.Offsets)
	}
	h0Digest := mappingDigest(plan.Mapping)

	mapping := make(koreanslots.Mapping, len(plan.Mapping))
	used := make(map[cp932.GlyphKey]rune, len(plan.Mapping))
	for r, key := range plan.Mapping {
		mapping[r] = key
		used[key] = r
	}

	spares := make([]cp932.GlyphKey, 0, len(plan.Candidates)-len(mapping))
	for _, key := range plan.Candidates {
		if _, inUse := used[key]; inUse {
			continue
		}
		if isRoundTripHanKey(key) {
			spares = append(spares, key)
		}
	}

	runes := append([]rune(nil), plan.CustomRunes...)
	sort.Slice(runes, func(i, j int) bool { return runes[i] < runes[j] })
	spareIndex := 0
	relocated := 0
	for _, r := range runes {
		oldKey := mapping[r]
		if !isRendererPrivate87Key(oldKey) {
			continue
		}
		if spareIndex >= len(spares) {
			return koreanslots.Plan{}, 0, 0, fmt.Errorf("mobile beta minimal87: need more safe spare keys after %d relocations", relocated)
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
		fmt.Printf("FORENSIC MINIMAL87_RELOCATE rune=%q unicode=%U old=%02X %02X nominal_old=%q new=%02X %02X nominal_new=%q\n",
			string(r), r, oldBytes[0], oldBytes[1], oldNominal, newBytes[0], newBytes[1], newNominal)
	}
	plan.Mapping = mapping

	// The H0 audit above is useful for historical comparison, but the ISO is
	// built with the relocated mapping. Re-run exact-byte ownership against the
	// final plan so a newly selected Minimal87 spare cannot escape the retail
	// BOOT/EBOOT/BINDATA collision report merely because it was absent from H0.
	finalAudit, err := auditMobileExactByteReuse(plan, blobs...)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, err
	}
	fmt.Printf("FORENSIC MOBILE_EXACT_BYTE_AUDIT phase=final candidates=%d exact_hit_candidates=%d mapped_hit_records=%d\n",
		len(plan.Candidates), finalAudit.CandidateHits, len(finalAudit.MappedHits))
	for _, hit := range finalAudit.MappedHits {
		encoded, _ := hit.Key.Bytes()
		fmt.Printf("FORENSIC MOBILE_EXACT_BYTE_MAPPING phase=final rune=%q unicode=%U key=%02X %02X blob=%s offsets=%v\n",
			string(hit.Rune), hit.Rune, encoded[0], encoded[1], hit.Blob, hit.Offsets)
	}

	finalDigest := mappingDigest(plan.Mapping)
	fmt.Printf("FORENSIC MAPPING_FINGERPRINT h0_sha256=%s final_sha256=%s changed=%d fixed_korean=%d new_eboot_bundle=false\n",
		h0Digest, finalDigest, relocated, len(fixedKorean))
	fmt.Printf("Korean beta minimal87 slot allocation: candidates=%d custom=%d headroom=%d relocated_private87=%d\n",
		len(plan.Candidates), len(plan.CustomRunes), len(plan.Candidates)-len(plan.CustomRunes), relocated)
	return plan, len(korean.Entries), len(source.Items), nil
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

func mappingDigest(mapping koreanslots.Mapping) string {
	runes := make([]rune, 0, len(mapping))
	for r := range mapping {
		runes = append(runes, r)
	}
	sort.Slice(runes, func(i, j int) bool { return runes[i] < runes[j] })
	h := sha256.New()
	for _, r := range runes {
		fmt.Fprintf(h, "%08X=%04X;", uint32(r), uint16(mapping[r]))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
