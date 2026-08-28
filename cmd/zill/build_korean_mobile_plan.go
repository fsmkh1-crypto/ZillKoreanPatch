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

// buildKoreanAlphaPlanMobile builds the mobile beta slot plan from the exact
// retail-bound source and runtime-materializable Korean project used by the ISO
// compiler. The mobile font path performs a full authenticated atlas+PAF repack,
// so it does not require the original retail cell geometry used by the older
// atlas-only desktop path. It still restricts message-byte allocation to PAF
// keys that round-trip as real CP932 text; renderer-private/UI keys can share
// Shift-JIS-shaped byte ranges and must never be emitted as Hangul text.
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

	texts, err := korean.RuntimeTexts(source)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, err
	}
	fixedKorean, err := loadKoreanFixedEBOOT(root)
	if err != nil {
		return koreanslots.Plan{}, 0, 0, err
	}
	texts = append(texts, fixeddata.KoreanEBOOTTexts(fixedKorean)...)
	installedAll := font.DoubleByteKeys()
	installedText := font.TextDoubleByteKeys()
	stock := koreanslots.RequiredStockKeys(texts)
	custom := koreanslots.RequiredCustomRunes(texts)
	fmt.Printf("Korean beta slot preflight: installed_double_byte=%d roundtrip_text_double_byte=%d excluded_private_or_undefined=%d stock_required=%d fixed_reserved=%d boot_scan_keys=%d bindata_scan_keys=%d total_reserved=%d custom=%d materializable_korean=%d fixed_korean=%d total_records=%d\n",
		len(installedAll), len(installedText), len(installedAll)-len(installedText), len(stock), len(usedFixed), len(bootScan.Keys), len(bindataScan.Keys), len(reserved), len(custom), len(korean.Entries), len(fixedKorean), len(source.Items))
	plan, err := koreanslots.BuildPlan(texts, installedText, rendererKeySetSlice(reserved))
	if err != nil {
		return koreanslots.Plan{}, 0, 0, fmt.Errorf("mobile beta text-only slot allocation: %w", err)
	}
	fmt.Printf("Korean beta slot allocation: candidates=%d custom=%d headroom=%d (full PAF repack; round-trip text keys only; fixed + BOOT/bindata CP932 literal ownership reserved)\n",
		len(plan.Candidates), len(plan.CustomRunes), len(plan.Candidates)-len(plan.CustomRunes))

	glyphIndex := make(map[cp932.GlyphKey]int, len(font.Glyphs))
	for _, glyph := range font.Glyphs {
		glyphIndex[glyph.Key] = glyph.Index
	}
	for _, id := range []int{10007, 10010} {
		row, ok := korean.Find(id)
		if !ok {
			fmt.Printf("FORENSIC ERROR id=%d reason=missing_korean_row\n", id)
			continue
		}
		field, text := "korean", row.Korean
		if row.Layout != "" {
			field, text = "layout", row.Layout
		}
		fmt.Printf("FORENSIC RECORD_TEXT id=%d field=%s text=%q\n", id, field, text)
		seen := make(map[rune]struct{})
		for _, r := range text {
			key, mapped := plan.Mapping[r]
			if !mapped {
				continue
			}
			if _, duplicate := seen[r]; duplicate {
				continue
			}
			seen[r] = struct{}{}
			encoded, keyErr := key.Bytes()
			if keyErr != nil {
				fmt.Printf("FORENSIC MAP id=%d rune=%q unicode=U+%04X key=%04X error=%q\n", id, r, r, uint16(key), keyErr)
				continue
			}
			nominal, decodeErr := cp932.Decode(encoded)
			if decodeErr != nil {
				nominal = "<decode-error>"
			}
			fmt.Printf("FORENSIC MAP id=%d rune=%q unicode=U+%04X key=%04X bytes=% X paf_index=%d nominal_cp932=%q roundtrip=%t\n",
				id, r, r, uint16(key), encoded, glyphIndex[key], nominal, key.IsRoundTripText())
		}
	}
	return plan, len(korean.Entries), len(source.Items), nil
}
