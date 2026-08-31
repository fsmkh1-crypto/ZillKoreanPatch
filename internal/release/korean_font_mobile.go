// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"crypto/sha256"
	"fmt"

	"github.com/HK47196/zill/internal/gamefmt/paa"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/zillfont"
)

// prepareKoreanMobileFontReplacements builds a complete Korean font result for
// device alpha testing. This intentionally follows the English patch model:
// atlas placement and PAF geometry are rebuilt as one coherent transform rather
// than trying to squeeze Korean into a small subset of retail-compatible cells.
func prepareKoreanMobileFontReplacements(root string, archives []*archive, plan koreanslots.Plan) ([]paa.Replacement, error) {
	if len(plan.Mapping) == 0 {
		return nil, nil
	}
	if len(plan.Mapping) != len(plan.CustomRunes) {
		return nil, fmt.Errorf("prepare mobile Korean font: mapping/custom rune mismatch: %d mappings for %d runes", len(plan.Mapping), len(plan.CustomRunes))
	}
	for _, r := range plan.CustomRunes {
		if _, ok := plan.Mapping[r]; !ok {
			return nil, fmt.Errorf("prepare mobile Korean font: required custom rune %U has no renderer-key mapping", r)
		}
	}
	if key, ok := plan.Mapping['명']; ok {
		fmt.Printf("Korean alpha glyph diagnostic: %U '명' -> renderer key 0x%04X\n", '명', uint16(key))
	} else {
		fmt.Printf("Korean alpha glyph diagnostic: %U '명' is not present in this mapping\n", '명')
	}

	inputs, err := loadKoreanFontInputs(root, archives, plan.Mapping, "prepare mobile Korean font")
	if err != nil {
		return nil, err
	}
	koreanRasters := make(map[rune]zillfont.Raster, len(plan.Mapping))
	for r := range plan.Mapping {
		raster, ok := inputs.catalog.SourceRaster(r)
		if !ok {
			return nil, fmt.Errorf("Korean raster catalog is missing %U", r)
		}
		koreanRasters[r] = raster
	}

	patchedAtlas, patchedPAF, err := zillfont.FullRepackAuthenticatedRetailFont(inputs.atlas, inputs.jillbtn, plan.Mapping, koreanRasters)
	if err != nil {
		return nil, err
	}
	if err := zillfont.VerifyFullRepackSemantics(inputs.atlas, inputs.jillbtn, patchedAtlas, patchedPAF, plan.Mapping, koreanRasters); err != nil {
		return nil, fmt.Errorf("prepare mobile Korean font postcondition audit: %w", err)
	}
	retailAtlasSHA := sha256.Sum256(inputs.atlas)
	retailPAFSHA := sha256.Sum256(inputs.jillbtn)
	patchedAtlasSHA := sha256.Sum256(patchedAtlas)
	patchedPAFSHA := sha256.Sum256(patchedPAF)
	fmt.Printf("FORENSIC MOBILE_FONT_FINGERPRINT retail_atlas_sha256=%x retail_paf_sha256=%x patched_atlas_sha256=%x patched_paf_sha256=%x mappings=%d\n",
		retailAtlasSHA, retailPAFSHA, patchedAtlasSHA, patchedPAFSHA, len(plan.Mapping))
	fmt.Printf("Korean mobile font semantic audit: %d PAF glyphs preserved/replaced under verified key/BST/metric/raster contracts.\n", zillfont.GlyphCount)
	return []paa.Replacement{
		paa.IndexReplacement(retailAtlasMemberIndex, patchedAtlas),
		paa.IndexReplacement(retailPAFMemberIndex, patchedPAF),
	}, nil
}
