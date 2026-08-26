// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanslots"
)

// KoreanAlphaPlanBuilder receives the exact retail-bound source and Korean
// project that the ISO compiler will use. The callback must not mutate either
// project. Keeping plan generation inside this bound-source lifetime prevents
// planner/compiler drift.
type KoreanAlphaPlanBuilder func(source *corpus.Project, korean *corpus.KoreanProject) (koreanslots.Plan, int, int, error)

// BuildKoreanAlphaISOOnly builds only the development Korean ISO. Retail message
// banks are authenticated and bound before either planning or placeholder
// projection. The same bound source instance is then used by the planner and
// compiler, matching the validated ordering of the normal release builder.
func BuildKoreanAlphaISOOnly(root, gameDir, isoPath, outputPath, version string, planBuilder KoreanAlphaPlanBuilder) (err error) {
	root, err = resolveExistingPath(root, "project root")
	if err != nil { return err }
	gameDir, err = resolveExistingPath(gameDir, "source PSP_GAME")
	if err != nil { return err }
	isoPath, err = resolveExistingPath(isoPath, "source ISO")
	if err != nil { return err }
	if planBuilder == nil { return fmt.Errorf("Korean alpha build: nil plan builder") }

	outputPath, err = filepath.Abs(outputPath)
	if err != nil { return fmt.Errorf("resolve Korean alpha output: %w", err) }
	if overlaps(isoPath, outputPath) || overlaps(gameDir, outputPath) {
		return fmt.Errorf("Korean alpha output must not overlap retail inputs")
	}
	if err := validateReleaseDestination(outputPath, false); err != nil { return err }
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create Korean alpha output parent: %w", err)
	}

	retailISO, isoManifest, err := openRetailISO(isoPath)
	if err != nil { return err }
	defer retailISO.Close()

	source, _, err := corpus.LoadProject(root)
	if err != nil { return err }
	korean, _, err := corpus.LoadKoreanProject(root, source)
	if err != nil { return err }

	archives, err := openArchives(gameDir)
	if err != nil { return err }
	defer func() {
		for _, archive := range archives { _ = archive.pair.Close() }
	}()
	banks, owners, err := loadRetailBanks(archives)
	if err != nil { return err }
	if err := corpus.BindBanks(source, banks); err != nil { return err }

	plan, coverage, total, err := planBuilder(source, korean)
	if err != nil { return err }
	fmt.Printf("Korean coverage: %d/%d records; custom glyphs: %d; reusable slots: %d\n", coverage, total, len(plan.CustomRunes), len(plan.Candidates))

	// Placeholder projection uses the same authenticated retail token streams
	// already consumed by the planner above.
	alphaProject, placeholderCount, err := BuildKoreanAlphaPlaceholderProject(source, korean)
	if err != nil { return err }
	fmt.Printf("Korean alpha compiler placeholders: %d\n", placeholderCount)

	layouts := make(map[int]string)
	for _, row := range korean.Entries {
		if row.Layout != "" { layouts[row.ID] = row.Layout }
	}
	compiled, err := compileKoreanBanksWithPlan(source, alphaProject, banks, plan, layouts)
	if err != nil { return err }
	if err := addBanks(owners, compiled); err != nil { return err }

	fontReplacements, err := prepareKoreanMobileFontReplacements(root, archives, plan)
	if err != nil { return err }
	if len(fontReplacements) > 0 {
		var pa *archive
		for _, candidate := range archives {
			if candidate.name == "pa" { pa = candidate; break }
		}
		if pa == nil { return fmt.Errorf("Korean alpha build: pa archive unavailable") }
		pa.replacements = append(pa.replacements, fontReplacements...)
	}

	executable, err := buildKoreanAlphaExecutable(root, gameDir)
	if err != nil { return err }
	parameter, err := buildKoreanAlphaSFO(root, gameDir, version)
	if err != nil { return err }

	bundle, err := os.MkdirTemp(filepath.Dir(outputPath), ".zill-korean-mobile.")
	if err != nil { return fmt.Errorf("create Korean alpha staging directory: %w", err) }
	defer os.RemoveAll(bundle)
	staging := filepath.Join(bundle, "PSP_GAME")
	if err := copyTree(gameDir, staging); err != nil { return err }
	if err := os.WriteFile(filepath.Join(staging, "PARAM.SFO"), parameter, 0o644); err != nil { return err }
	for _, name := range []string{"BOOT.BIN", "EBOOT.BIN"} {
		if err := os.WriteFile(filepath.Join(staging, "SYSDIR", name), executable, 0o755); err != nil { return err }
	}
	for _, archive := range archives {
		usrdir := filepath.Join(staging, "USRDIR")
		if err := archive.pair.Rebuild(
			filepath.Join(usrdir, archive.name+".bin"),
			filepath.Join(usrdir, archive.name+".arc"),
			archive.replacements...,
		); err != nil {
			return fmt.Errorf("rebuild %s archive: %w", archive.name, err)
		}
	}
	if err := authorTranslatedISO(outputPath, retailISO, isoManifest, staging); err != nil { return err }
	return nil
}
