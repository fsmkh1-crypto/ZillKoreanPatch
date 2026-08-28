// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/gamefmt/elfpatch"
	"github.com/HK47196/zill/internal/gamefmt/sfo"
	"github.com/HK47196/zill/internal/koreanslots"
)

const (
	koreanAlphaISOName   = "zill-korean-beta.iso"
	koreanAlphaPatchName = "zill-korean-beta.xdelta"
)

// BuildKoreanAlpha builds the current Korean beta integration release. The
// legacy function name is kept for compatibility with the mobile/debug builder.
// Accepted Korean records are compiled with the authenticated slot plan while
// source records outside the canonical Korean overlay retain their retail
// Japanese bytes. The English static font/string transforms are deliberately
// excluded: Korean uses its own renderer-slot/font path instead.
func BuildKoreanAlpha(root, gameDir, isoPath, version string, plan koreanslots.Plan) (result Result, err error) {
	root, err = resolveExistingPath(root, "project root")
	if err != nil { return result, err }
	gameDir, err = resolveExistingPath(gameDir, "source PSP_GAME")
	if err != nil { return result, err }
	isoPath, err = resolveExistingPath(isoPath, "source ISO")
	if err != nil { return result, err }

	buildDirectory := filepath.Join(root, "build")
	gameDestination := filepath.Join(buildDirectory, "PSP_GAME-korean-beta")
	isoDestination := filepath.Join(buildDirectory, koreanAlphaISOName)
	patchDestination := filepath.Join(buildDirectory, koreanAlphaPatchName)
	if overlaps(gameDir, gameDestination) { return result, fmt.Errorf("source and output PSP_GAME trees must not overlap") }
	for _, output := range []string{gameDestination, isoDestination, patchDestination} {
		if overlaps(isoPath, output) { return result, fmt.Errorf("retail ISO and output %s must not overlap", output) }
	}
	if err := validateReleaseDestination(gameDestination, true); err != nil { return result, err }
	for _, output := range []string{isoDestination, patchDestination} {
		if err := validateReleaseDestination(output, false); err != nil { return result, err }
	}

	xdelta, err := findXdelta()
	if err != nil { return result, err }
	retailISO, isoManifest, err := openRetailISO(isoPath)
	if err != nil { return result, err }
	defer retailISO.Close()

	source, _, err := corpus.LoadProject(root)
	if err != nil { return result, err }
	korean, _, err := corpus.LoadKoreanProject(root, source)
	if err != nil { return result, err }

	archives, err := openArchives(gameDir)
	if err != nil { return result, err }
	defer func() { for _, archive := range archives { _ = archive.pair.Close() } }()
	banks, owners, err := loadRetailBanks(archives)
	if err != nil { return result, err }
	if err := corpus.BindBanks(source, banks); err != nil { return result, err }

	layouts := make(map[int]string)
	for _, row := range korean.Entries {
		if row.Layout != "" { layouts[row.ID] = row.Layout }
	}
	compiled, err := compileKoreanBanksWithPlan(source, korean, banks, plan, layouts)
	if err != nil { return result, err }
	if err := addBanks(owners, compiled); err != nil { return result, err }
	fontReplacement, ok, err := prepareKoreanFontReplacement(root, archives, plan)
	if err != nil { return result, err }
	if ok {
		var pa *archive
		for _, candidate := range archives { if candidate.name == "pa" { pa = candidate; break } }
		if pa == nil { return result, fmt.Errorf("Korean beta build: pa archive unavailable") }
		pa.replacements = append(pa.replacements, fontReplacement)
	}

	executable, err := buildKoreanAlphaExecutable(root, gameDir)
	if err != nil { return result, err }
	parameter, err := buildKoreanAlphaSFO(root, gameDir, version)
	if err != nil { return result, err }

	if err := os.MkdirAll(buildDirectory, 0o755); err != nil { return result, fmt.Errorf("create build directory: %w", err) }
	bundle, err := os.MkdirTemp(buildDirectory, ".korean-beta.stage.")
	if err != nil { return result, fmt.Errorf("create Korean beta staging directory: %w", err) }
	defer os.RemoveAll(bundle)
	staging := filepath.Join(bundle, "PSP_GAME")
	if err := copyTree(gameDir, staging); err != nil { return result, err }
	if err := os.WriteFile(filepath.Join(staging, "PARAM.SFO"), parameter, 0o644); err != nil { return result, err }
	for _, name := range []string{"BOOT.BIN", "EBOOT.BIN"} {
		if err := os.WriteFile(filepath.Join(staging, "SYSDIR", name), executable, 0o755); err != nil { return result, err }
	}
	for _, archive := range archives {
		usrdir := filepath.Join(staging, "USRDIR")
		if err := archive.pair.Rebuild(filepath.Join(usrdir, archive.name+".bin"), filepath.Join(usrdir, archive.name+".arc"), archive.replacements...); err != nil {
			return result, fmt.Errorf("rebuild %s archive: %w", archive.name, err)
		}
	}
	stagedISO := filepath.Join(bundle, koreanAlphaISOName)
	if err := authorTranslatedISO(stagedISO, retailISO, isoManifest, staging); err != nil { return result, err }
	if err := retailISO.Close(); err != nil { return result, fmt.Errorf("close retail ISO before xdelta encoding: %w", err) }
	stagedPatch := filepath.Join(bundle, koreanAlphaPatchName)
	if err := createAndVerifyPatch(xdelta, isoPath, stagedISO, stagedPatch); err != nil { return result, err }
	cleanup, err := publishAll([]publishItem{{staging: staging, destination: gameDestination}, {staging: stagedISO, destination: isoDestination}, {staging: stagedPatch, destination: patchDestination}})
	if err != nil { return result, fmt.Errorf("publish Korean beta release: %w", err) }
	result.GameDirectory, result.ISO, result.Patch = gameDestination, isoDestination, patchDestination
	for _, warning := range cleanup { result.Warnings = append(result.Warnings, warning.Error()) }
	return result, nil
}

func buildKoreanAlphaExecutable(root, gameDir string) ([]byte, error) {
	eboot, err := os.ReadFile(filepath.Join(gameDir, "SYSDIR", "EBOOT.BIN"))
	if err != nil { return nil, fmt.Errorf("read retail EBOOT.BIN: %w", err) }
	if actual := fmt.Sprintf("%x", sha256.Sum256(eboot)); actual != sourceEBOOTSHA256 { return nil, fmt.Errorf("retail EBOOT.BIN does not match supported ULJM05410 1.03 fingerprint") }
	source, err := os.ReadFile(filepath.Join(gameDir, "SYSDIR", "BOOT.BIN"))
	if err != nil { return nil, fmt.Errorf("read retail BOOT.BIN: %w", err) }
	manifestData, err := read(root, "patches", "executable", "manifest.toml")
	if err != nil { return nil, err }
	manifest, err := elfpatch.ParseManifest(manifestData)
	if err != nil { return nil, err }
	return elfpatch.Apply(source, manifest)
}

func buildKoreanAlphaSFO(root, gameDir, version string) ([]byte, error) {
	source, err := os.ReadFile(filepath.Join(gameDir, "PARAM.SFO"))
	if err != nil { return nil, fmt.Errorf("read retail PARAM.SFO: %w", err) }
	manifestData, err := read(root, "patches", "system", "param-sfo.toml")
	if err != nil { return nil, err }
	manifest, err := sfo.ParseManifest(manifestData)
	if err != nil { return nil, err }
	return sfo.Apply(source, manifest, fmt.Sprintf("Zill O'll Infinite Plus [Korean Beta %s]", version))
}
