// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"fmt"
	"os"
	"path/filepath"
)

// BuildKoreanAlphaISOOnly builds the mobile Korean beta ISO from the canonical
// reviewed Korean overlay. Retail message banks are authenticated and bound
// before either planning or compilation. Canonical rows that correspond to
// structural/no-text source records are projected out and retain their retail
// bytes; every materializable accepted Korean row keeps its reviewed Layout.
func BuildKoreanAlphaISOOnly(root, gameDir, isoPath, outputPath, version string, planBuilder KoreanAlphaPlanBuilder) (err error) {
	root, err = resolveExistingPath(root, "project root")
	if err != nil { return err }
	gameDir, err = resolveExistingPath(gameDir, "source PSP_GAME")
	if err != nil { return err }
	isoPath, err = resolveExistingPath(isoPath, "source ISO")
	if err != nil { return err }
	if planBuilder == nil { return fmt.Errorf("Korean beta build: nil plan builder") }

	outputPath, err = filepath.Abs(outputPath)
	if err != nil { return fmt.Errorf("resolve Korean beta output: %w", err) }
	if overlaps(isoPath, outputPath) || overlaps(gameDir, outputPath) {
		return fmt.Errorf("Korean beta output must not overlap retail inputs")
	}
	if err := validateReleaseDestination(outputPath, false); err != nil { return err }
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create Korean beta output parent: %w", err)
	}

	retailISO, isoManifest, err := openRetailISO(isoPath)
	if err != nil { return err }
	defer retailISO.Close()

	prepared, err := prepareKoreanMobilePayload(root, gameDir, version, "mobile", planBuilder)
	if err != nil { return err }
	defer prepared.close()

	bundle, err := os.MkdirTemp(filepath.Dir(outputPath), ".zill-korean-mobile.")
	if err != nil { return fmt.Errorf("create Korean beta staging directory: %w", err) }
	defer os.RemoveAll(bundle)
	staging := filepath.Join(bundle, "PSP_GAME")
	if err := copyTree(gameDir, staging); err != nil { return err }
	if err := os.WriteFile(filepath.Join(staging, "PARAM.SFO"), prepared.parameter, 0o644); err != nil { return err }
	for _, name := range []string{"BOOT.BIN", "EBOOT.BIN"} {
		if err := os.WriteFile(filepath.Join(staging, "SYSDIR", name), prepared.executable, 0o755); err != nil { return err }
	}
	for _, archive := range prepared.archives {
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
