// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/fixeddata"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/slotaudit"
)

// auditPR14HistoricalPolicies replays the planner policies used by the PR #14
// H0/B/A/Combined/Stable-minimal experiments against the CURRENT retail-bound
// runtime projection and one authenticated retail asset set. It is diagnostic
// only: resulting plans are never returned to the production builder.
//
// Historical EBOOT policy inputs are immutable fixtures. In particular H0 must
// not silently grow when the production Korean fixed-string overlay gains new
// reviewed U1/U2/U3 translations.
func auditPR14HistoricalPolicies(root, gameDir string, source *corpus.Project, korean *corpus.KoreanProject) error {
	if source == nil || korean == nil {
		return fmt.Errorf("PR14 policy audit: nil bound source or Korean project")
	}
	canonical, _, err := corpus.LoadKoreanProject(root, source)
	if err != nil {
		return fmt.Errorf("PR14 policy audit load canonical Korean: %w", err)
	}
	if len(canonical.Entries) < len(korean.Entries) {
		return fmt.Errorf("PR14 policy audit population inversion: canonical=%d materializable=%d", len(canonical.Entries), len(korean.Entries))
	}

	font, err := loadRetailPAF(gameDir)
	if err != nil { return err }
	usedFixed, equipment, err := loadFixedRendererKeys(root)
	if err != nil { return err }
	retailEBOOT, err := loadAuthenticatedRetailEBOOT(gameDir)
	if err != nil { return err }
	boot, err := loadAuthenticatedRetailBOOT(gameDir)
	if err != nil { return err }
	bootScan, err := slotaudit.ScanCP932Literals(boot)
	if err != nil { return fmt.Errorf("PR14 policy audit scan BOOT.BIN: %w", err) }
	bindata, err := loadRetailBindata(gameDir)
	if err != nil { return err }
	if _, err := fixeddata.ApplyEquipment(bindata, equipment); err != nil {
		return fmt.Errorf("PR14 policy audit validate bindata.dat layout: %w", err)
	}
	bindataScan, err := slotaudit.ScanCP932Literals(bindata)
	if err != nil { return fmt.Errorf("PR14 policy audit scan bindata.dat: %w", err) }

	reserved := make(map[cp932.GlyphKey]struct{})
	mergeRendererKeys(reserved, usedFixed)
	mergeRendererKeys(reserved, bootScan.Keys)
	mergeRendererKeys(reserved, bindataScan.Keys)
	reservedKeys := rendererKeySetSlice(reserved)
	installed := font.DoubleByteKeys()

	messageTexts, err := korean.RuntimeTexts(source)
	if err != nil { return err }

	// Historical H0 and expanded PR14 fixed-string inputs are frozen fixtures.
	h0Data, err := os.ReadFile(filepath.Join(root, "docs", "audit", "fixtures", "pr14-eboot-h0.toml"))
	if err != nil { return fmt.Errorf("PR14 policy audit read historical H0 EBOOT fixture: %w", err) }
	h0Fixed, err := fixeddata.ParseKoreanEBOOT(h0Data)
	if err != nil { return fmt.Errorf("PR14 policy audit parse historical H0 EBOOT fixture: %w", err) }
	fullData, err := os.ReadFile(filepath.Join(root, "docs", "audit", "fixtures", "pr14-eboot-full.toml"))
	if err != nil { return fmt.Errorf("PR14 policy audit read expanded EBOOT fixture: %w", err) }
	fullFixed, err := fixeddata.ParseKoreanEBOOT(fullData)
	if err != nil { return fmt.Errorf("PR14 policy audit parse expanded EBOOT fixture: %w", err) }
	currentData, err := os.ReadFile(filepath.Join(root, "release", "korean", "strings", "eboot.toml"))
	if err != nil { return fmt.Errorf("PR14 policy audit read current production EBOOT table: %w", err) }
	currentFixed, err := fixeddata.ParseKoreanEBOOT(currentData)
	if err != nil { return fmt.Errorf("PR14 policy audit parse current production EBOOT table: %w", err) }

	h0Texts := append([]string(nil), messageTexts...)
	h0Texts = append(h0Texts, fixeddata.KoreanEBOOTTexts(h0Fixed)...)
	fullTexts := append([]string(nil), messageTexts...)
	fullTexts = append(fullTexts, fixeddata.KoreanEBOOTTexts(fullFixed)...)

	h0Plan, err := koreanslots.BuildPlan(h0Texts, installed, reservedKeys)
	if err != nil { return fmt.Errorf("PR14 policy audit H0 plan: %w", err) }
	bPlan, err := koreanslots.BuildPlan(fullTexts, installed, reservedKeys)
	if err != nil { return fmt.Errorf("PR14 policy audit B plan: %w", err) }

	hanInstalled := make([]cp932.GlyphKey, 0, len(installed))
	for _, key := range installed {
		if isRoundTripHanKey(key) { hanInstalled = append(hanInstalled, key) }
	}
	aPlan, err := koreanslots.BuildPlan(h0Texts, hanInstalled, reservedKeys)
	if err != nil { return fmt.Errorf("PR14 policy audit A plan: %w", err) }
	combinedPlan, err := koreanslots.BuildPlan(fullTexts, hanInstalled, reservedKeys)
	if err != nil { return fmt.Errorf("PR14 policy audit Combined plan: %w", err) }
	stablePlan, stableRelocated, stableErr := replayStableMinimalPlan(h0Plan, fullTexts)
	if stableErr != nil {
		fmt.Printf("FORENSIC PR14_POLICY name=Stable-minimal replay_available=false reason=%q current_runtime_projection_replay=true\n", stableErr.Error())
	}

	blobs := []slotAuditBlob{
		{Name: "BOOT.BIN", Data: boot},
		{Name: "EBOOT.BIN", Data: retailEBOOT},
		{Name: "bindata.dat", Data: bindata},
	}
	fmt.Printf("FORENSIC PR14_POLICY_MATRIX current_runtime_projection_replay=true canonical_entries=%d materializable_entries=%d structural_retail=%d source_records=%d h0_eboot_fields=%d expanded_eboot_fields=%d current_production_eboot_fields=%d installed=%d han_installed=%d\n",
		len(canonical.Entries), len(korean.Entries), len(canonical.Entries)-len(korean.Entries), len(source.Items), len(h0Fixed), len(fullFixed), len(currentFixed), len(installed), len(hanInstalled))

	rows := []struct {
		name string
		plan koreanslots.Plan
		fixed fixeddata.KoreanEBOOTTranslations
	}{
		{name: "H0", plan: h0Plan, fixed: h0Fixed},
		{name: "B", plan: bPlan, fixed: fullFixed},
		{name: "A", plan: aPlan, fixed: h0Fixed},
		{name: "Combined", plan: combinedPlan, fixed: fullFixed},
	}
	if stableErr == nil {
		rows = append(rows, struct {
			name string
			plan koreanslots.Plan
			fixed fixeddata.KoreanEBOOTTranslations
		}{name: "Stable-minimal", plan: stablePlan, fixed: fullFixed})
	}

	for _, row := range rows {
		exact, auditErr := auditMobileExactByteReuse(row.plan, blobs...)
		if auditErr != nil { return fmt.Errorf("PR14 policy audit %s exact-byte audit: %w", row.name, auditErr) }
		ebootDigest, digestErr := koreanEBOOTEncodingDigest(row.fixed, row.plan.Mapping)
		if digestErr != nil { return fmt.Errorf("PR14 policy audit %s EBOOT digest: %w", row.name, digestErr) }
		fmt.Printf("FORENSIC PR14_POLICY name=%s mapping_sha256=%s changed_vs_h0=%d custom=%d candidates=%d mapped_exact_hit_records=%d eboot_encoding_sha256=%s current_runtime_projection_replay=true\n",
			row.name, mappingDigest(row.plan.Mapping), mappingDifferenceCount(h0Plan.Mapping, row.plan.Mapping),
			len(row.plan.CustomRunes), len(row.plan.Candidates), len(exact.MappedHits), ebootDigest)
	}

	fmt.Printf("FORENSIC PR14_RELATION h0_b_mapping_equal=%t a_combined_mapping_equal=%t stable_relocated=%d current_runtime_projection_replay=true\n",
		mappingDifferenceCount(h0Plan.Mapping, bPlan.Mapping) == 0,
		mappingDifferenceCount(aPlan.Mapping, combinedPlan.Mapping) == 0,
		stableRelocated)
	return nil
}

func replayStableMinimalPlan(base koreanslots.Plan, fullTexts []string) (koreanslots.Plan, int, error) {
	combinedCustom := koreanslots.RequiredCustomRunes(fullTexts)
	if !sameRuneSlice(base.CustomRunes, combinedCustom) {
		return koreanslots.Plan{}, 0, fmt.Errorf("expanded EBOOT changes custom rune set: base=%d combined=%d", len(base.CustomRunes), len(combinedCustom))
	}
	combinedStock := koreanslots.RequiredStockKeys(fullTexts)

	mapping := cloneKoreanMapping(base.Mapping)
	used := make(map[cp932.GlyphKey]rune, len(mapping))
	for r, key := range mapping { used[key] = r }
	for _, key := range combinedStock {
		if r, collision := used[key]; collision {
			return koreanslots.Plan{}, 0, fmt.Errorf("expanded EBOOT stock key 0x%04X collides with custom rune %U", uint16(key), r)
		}
	}

	spares := make([]cp932.GlyphKey, 0, len(base.Candidates)-len(mapping))
	for _, key := range base.Candidates {
		if _, inUse := used[key]; inUse { continue }
		if isRoundTripHanKey(key) { spares = append(spares, key) }
	}
	runes := append([]rune(nil), base.CustomRunes...)
	sort.Slice(runes, func(i, j int) bool { return runes[i] < runes[j] })
	relocated := 0
	for _, r := range runes {
		oldKey := mapping[r]
		if !isRendererPrivate87Key(oldKey) { continue }
		if relocated >= len(spares) {
			return koreanslots.Plan{}, relocated, fmt.Errorf("minimal87 needs more Han spare keys after %d relocations", relocated)
		}
		newKey := spares[relocated]
		delete(used, oldKey)
		mapping[r] = newKey
		used[newKey] = r
		relocated++
	}
	out := base
	out.CustomRunes = combinedCustom
	out.RequiredStock = combinedStock
	out.Mapping = mapping
	return out, relocated, nil
}

func mappingDifferenceCount(a, b koreanslots.Mapping) int {
	seen := make(map[rune]struct{}, len(a)+len(b))
	for r := range a { seen[r] = struct{}{} }
	for r := range b { seen[r] = struct{}{} }
	diff := 0
	for r := range seen {
		ak, aok := a[r]
		bk, bok := b[r]
		if aok != bok || ak != bk { diff++ }
	}
	return diff
}

func sameRuneSlice(a, b []rune) bool {
	if len(a) != len(b) { return false }
	for i := range a { if a[i] != b[i] { return false } }
	return true
}

func cloneKoreanMapping(in koreanslots.Mapping) koreanslots.Mapping {
	out := make(koreanslots.Mapping, len(in))
	for r, key := range in { out[r] = key }
	return out
}

func koreanEBOOTEncodingDigest(translations fixeddata.KoreanEBOOTTranslations, mapping koreanslots.Mapping) (string, error) {
	offsets := make([]uint64, 0, len(translations))
	for offset := range translations { offsets = append(offsets, offset) }
	sort.Slice(offsets, func(i, j int) bool { return offsets[i] < offsets[j] })
	h := sha256.New()
	var scratch [8]byte
	for _, offset := range offsets {
		binary.LittleEndian.PutUint64(scratch[:], offset)
		_, _ = h.Write(scratch[:])
		encoded, err := koreanslots.Encode(translations[offset].Replacement, mapping)
		if err != nil { return "", fmt.Errorf("field %#x: %w", offset, err) }
		binary.LittleEndian.PutUint64(scratch[:], uint64(len(encoded)))
		_, _ = h.Write(scratch[:])
		_, _ = h.Write(encoded)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
