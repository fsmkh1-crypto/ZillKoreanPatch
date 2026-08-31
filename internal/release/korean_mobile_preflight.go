// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"fmt"
	"path/filepath"
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

	retailISO, _, err := openRetailISO(isoPath)
	if err != nil { return err }
	if err := retailISO.Close(); err != nil {
		return fmt.Errorf("close retail ISO after preflight authentication: %w", err)
	}

	prepared, err := prepareKoreanMobilePayload(root, gameDir, version, "preflight", planBuilder)
	if err != nil { return err }
	defer prepared.close()

	fmt.Printf("FORENSIC MOBILE_PREFLIGHT_OK game_dir=%q iso=%q output_iso_written=false\n", filepath.Clean(gameDir), filepath.Clean(isoPath))
	return nil
}
