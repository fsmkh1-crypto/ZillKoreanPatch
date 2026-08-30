// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"crypto/sha256"
	"fmt"

	"github.com/HK47196/zill/internal/fixeddata"
	"github.com/HK47196/zill/internal/gamefmt/elfpatch"
	"github.com/HK47196/zill/internal/koreanslots"
)

func applyKoreanFixedEBOOT(root string, source []byte, mapping koreanslots.Mapping) ([]byte, error) {
	data, err := read(root, "release", "korean", "strings", "eboot.toml")
	if err != nil {
		return nil, fmt.Errorf("read Korean EBOOT fixed strings: %w", err)
	}
	translations, err := fixeddata.ParseKoreanEBOOT(data)
	if err != nil {
		return nil, err
	}
	result, err := fixeddata.ApplyKoreanEBOOT(source, translations, mapping)
	if err != nil {
		return nil, err
	}

	// The sparse Korean overlay runs after the authenticated runtime manifest.
	// Re-verify every guarded runtime-patch span here so a later localization
	// write can never silently clobber message-arena/wide-offset instructions.
	manifestData, err := read(root, "patches", "executable", "manifest.toml")
	if err != nil {
		return nil, fmt.Errorf("read executable manifest for post-overlay verification: %w", err)
	}
	manifest, err := elfpatch.ParseManifest(manifestData)
	if err != nil {
		return nil, err
	}
	if err := elfpatch.VerifyApplied(result, manifest); err != nil {
		return nil, fmt.Errorf("verify executable runtime patches after Korean overlay: %w", err)
	}
	sourceSHA := sha256.Sum256(source)
	resultSHA := sha256.Sum256(result)
	fmt.Printf("FORENSIC KOREAN_EBOOT_FINGERPRINT source_sha256=%x patched_sha256=%x fields=%d mappings=%d\n",
		sourceSHA, resultSHA, len(translations), len(mapping))
	return result, nil
}
