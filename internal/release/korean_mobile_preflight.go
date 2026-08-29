// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"fmt"
	"path/filepath"

	"github.com/HK47196/zill/internal/corpus"
)

// PreflightKoreanAlphaISOOnly executes the authenticated, asset-backed mobile
// Korean build through every deterministic validation/generation step that can
// run before archive rebuilding and ISO authoring. It intentionally writes no
// patched PSP_GAME tree and no output ISO.
//
// The plan builder receives the same retail-bound source/Korean projects used by
// BuildKoreanAlphaISOOnly, so forensic results remain comparable with the real
// mobile build rather than being produced from a looser synthetic path.
func PreflightKoreanAlphaISOOnly(root, gameDir, isoPath, version string, planBuilder KoreanAlphaPlanBuilder) (err error) {
	root, err = resolveExistingPath(root, "project root")
	if err != nil { return err }
	gameDir, err = resolveExistingPath(gameDir, "source PSP_GAME")
	if err != nil { return err }
	isoPath, err = resolveExistingPath(isoPath, "source ISO")
	if err != nil { return err }
	if planBuilder == nil { return fmt.Errorf("Korean beta preflight: nil plan builder") }

	// Keep the source-ISO validation contract aligned with the real mobile build.
	retailISO, _, err := openRetailISO(isoPath)
	if err != nil { return err }
	if err := retailISO.Close(); err != nil {
		return fmt.Errorf("close retail ISO after preflight authentication: %w", err)
	}

	source, _, err := corpus.LoadProject(root)
	if err != nil { return err }
	canonicalKorean, _, err := corpus.LoadKoreanProject(root, source)
	if err != nil { return err }

	archives, err := openArchives(gameDir)
	if err != nil { return err }
	defer func() {
		for _, archive := range archives { _ = archive.pair.Close() }
	}()
	banks, owners, err := loadRetailBanks(archives)
	if err != nil { return err }
	if err := corpus.BindBanks(source, banks); err != nil { return err }

	korean, skippedStructural, err := BuildKoreanBetaProject(source, canonicalKorean)
	if err != nil { return err }
	fmt.Printf("Korean beta preflight projection: canonical=%d materializable=%d structural_retail=%d\n",
		len(canonicalKorean.Entries), len(korean.Entries), skippedStructural)

	plan, coverage, total, err := planBuilder(source, korean)
	if err != nil { return err }
	fmt.Printf("Korean preflight coverage: %d/%d records; custom glyphs: %d; reusable slots: %d\n",
		coverage, total, len(plan.CustomRunes), len(plan.Candidates))

	layouts := make(map[int]string)
	for _, row := range korean.Entries {
		if row.Layout != "" {
			layouts[row.ID] = row.Layout
		}
	}
	dynamicC5, err := validateKoreanRuntimeStorage(root, source, korean, layouts, plan.Mapping)
	if err != nil { return err }
	fmt.Printf("Korean C5 preflight storage check: no violation detected; %d dynamic-substitution record(s) remain runtime-QA risks.\n", len(dynamicC5))

	compiled, err := compileKoreanBanksWithPlan(source, korean, banks, plan, layouts)
	if err != nil { return err }
	if err := addBanks(owners, compiled); err != nil { return err }
	fmt.Printf("FORENSIC MOBILE_PREFLIGHT_BANKS compiled=%d authenticated_source_banks=%d\n", len(compiled), len(banks))

	fontReplacements, err := prepareKoreanMobileFontReplacements(root, archives, plan)
	if err != nil { return err }
	if len(fontReplacements) > 0 {
		var pa *archive
		for _, candidate := range archives {
			if candidate.name == "pa" { pa = candidate; break }
		}
		if pa == nil { return fmt.Errorf("Korean beta preflight: pa archive unavailable") }
		pa.replacements = append(pa.replacements, fontReplacements...)
	}
	fmt.Printf("FORENSIC MOBILE_PREFLIGHT_FONT replacements=%d\n", len(fontReplacements))

	executable, err := buildKoreanAlphaExecutable(root, gameDir, plan.Mapping)
	if err != nil { return err }
	parameter, err := buildKoreanAlphaSFO(root, gameDir, version)
	if err != nil { return err }
	fmt.Printf("FORENSIC MOBILE_PREFLIGHT_EXECUTABLE bytes=%d sfo_bytes=%d output_iso_written=false\n", len(executable), len(parameter))
	fmt.Printf("FORENSIC MOBILE_PREFLIGHT_OK game_dir=%q iso=%q output_iso_written=false\n", filepath.Clean(gameDir), filepath.Clean(isoPath))
	return nil
}
