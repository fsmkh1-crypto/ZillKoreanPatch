// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"io"
	"sort"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/fixeddata"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/slotaudit"
)

func rendererKeySetSlice(set map[cp932.GlyphKey]struct{}) []cp932.GlyphKey {
	out := make([]cp932.GlyphKey, 0, len(set))
	for key := range set {
		out = append(out, key)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// runKoreanSlotsV2 computes the actual slot allocation required by the current
// sparse Korean overlay. Accepted Korean rows replace Japanese source text;
// untranslated rows keep Japanese, so stock glyph reservations shrink as the
// translation progresses instead of preserving the whole Japanese corpus.
func runKoreanSlotsV2(root string, args []string, stdout, stderr io.Writer) int {
	gameDir, ok := parseRequiredGameDir("korean-slots", args, stderr)
	if !ok {
		return 2
	}
	font, err := loadRetailPAF(gameDir)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-slots: %v\n", err)
		return 1
	}
	source, sourceSummary, err := corpus.LoadProject(root)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-slots: %v\n", err)
		return 1
	}
	korean, koreanSummary, err := corpus.LoadKoreanProject(root, source)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-slots: %v\n", err)
		return 1
	}

	usedFixed, equipment, err := loadFixedRendererKeys(root)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-slots: %v\n", err)
		return 1
	}
	eboot, err := loadAuthenticatedRetailEBOOT(gameDir)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-slots: %v\n", err)
		return 1
	}
	boot, err := loadAuthenticatedRetailBOOT(gameDir)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-slots: %v\n", err)
		return 1
	}
	bootScan, err := slotaudit.ScanCP932Literals(boot)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-slots: scan BOOT.BIN: %v\n", err)
		return 1
	}
	bindata, err := loadRetailBindata(gameDir)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-slots: %v\n", err)
		return 1
	}
	if _, err := fixeddata.ApplyEquipment(bindata, equipment); err != nil {
		fmt.Fprintf(stderr, "zill: korean-slots: validate bindata.dat layout: %v\n", err)
		return 1
	}
	bindataScan, err := slotaudit.ScanCP932Literals(bindata)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-slots: scan bindata.dat: %v\n", err)
		return 1
	}

	reserved := make(map[cp932.GlyphKey]struct{})
	mergeRendererKeys(reserved, usedFixed)
	mergeRendererKeys(reserved, bootScan.Keys)
	mergeRendererKeys(reserved, bindataScan.Keys)

	texts, err := korean.RuntimeTexts(source)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-slots: %v\n", err)
		return 1
	}
	installed := font.DoubleByteKeys()
	plan, err := koreanslots.BuildPlan(texts, installed, rendererKeySetSlice(reserved), boot, eboot, bindata)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-slots: plan allocation: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Korean coverage: %d/%d records across %d section files\n", koreanSummary.Records, sourceSummary.Records, koreanSummary.Sections)
	fmt.Fprintf(stdout, "Installed two-byte slots: %d\n", len(installed))
	fmt.Fprintf(stdout, "Final runtime stock renderer keys required: %d\n", len(plan.RequiredStock))
	fmt.Fprintf(stdout, "Custom renderer glyphs required: %d\n", len(plan.CustomRunes))
	fmt.Fprintf(stdout, "Reusable candidates after fixed/structured/exact-byte audit: %d\n", len(plan.Candidates))
	fmt.Fprintf(stdout, "Allocated custom mappings: %d\n", len(plan.Mapping))
	fmt.Fprintln(stdout, "Safety status: DETERMINISTIC AUTHENTICATED PLAN; archive/UI/script resource classes not supplied to this audit remain outside the production-safe claim.")
	if len(plan.CustomRunes) > 0 {
		limit := min(16, len(plan.CustomRunes))
		fmt.Fprint(stdout, "First custom mappings:")
		for _, r := range plan.CustomRunes[:limit] {
			fmt.Fprintf(stdout, " %U=%04X", r, uint16(plan.Mapping[r]))
		}
		fmt.Fprintln(stdout)
	}
	return 0
}
