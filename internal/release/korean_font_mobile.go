// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/HK47196/zill/internal/gamefmt/paa"
	"github.com/HK47196/zill/internal/koreanfont"
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

	var paArchive *archive
	for _, candidate := range archives {
		if candidate != nil && candidate.name == "pa" {
			if paArchive != nil {
				return nil, fmt.Errorf("prepare mobile Korean font: duplicate pa archive")
			}
			paArchive = candidate
		}
	}
	if paArchive == nil || paArchive.pair == nil {
		return nil, fmt.Errorf("prepare mobile Korean font: pa archive is unavailable")
	}
	members := paArchive.pair.Members()
	if retailPAFMemberIndex >= len(members) {
		return nil, fmt.Errorf("prepare mobile Korean font: retail pa archive has only %d members", len(members))
	}
	atlasMember := members[retailAtlasMemberIndex]
	pafMember := members[retailPAFMemberIndex]
	if atlasMember.Name != retailAtlasMemberName || atlasMember.Size != zillfont.RetailAtlasMemberSize {
		return nil, fmt.Errorf("prepare mobile Korean font: unexpected atlas member %d %q size %#x", retailAtlasMemberIndex, atlasMember.Name, atlasMember.Size)
	}
	if pafMember.Name != retailPAFMemberName || pafMember.Size != zillfont.RetailPAFMemberSize {
		return nil, fmt.Errorf("prepare mobile Korean font: unexpected PAF member %d %q size %#x", retailPAFMemberIndex, pafMember.Name, pafMember.Size)
	}

	atlas, err := paArchive.pair.Payload(retailAtlasMemberIndex)
	if err != nil {
		return nil, err
	}
	jillbtn, err := paArchive.pair.Payload(retailPAFMemberIndex)
	if err != nil {
		return nil, err
	}
	if _, err := zillfont.ParseAuthenticatedRetailPAF(jillbtn); err != nil {
		return nil, err
	}

	catalogData, err := os.ReadFile(filepath.Join(root, "release", "korean", "font", "glyphs.toml"))
	if err != nil {
		return nil, fmt.Errorf("read Korean raster catalog: %w", err)
	}
	catalog, err := koreanfont.Parse(catalogData)
	if err != nil {
		return nil, err
	}
	mappedRunes := make([]rune, 0, len(plan.Mapping))
	for r := range plan.Mapping {
		mappedRunes = append(mappedRunes, r)
	}
	if err := catalog.RequireRunes(mappedRunes); err != nil {
		return nil, err
	}
	koreanRasters := make(map[rune]zillfont.Raster, len(plan.Mapping))
	for r := range plan.Mapping {
		raster, ok := catalog.SourceRaster(r)
		if !ok {
			return nil, fmt.Errorf("Korean raster catalog is missing %U", r)
		}
		koreanRasters[r] = raster
	}

	patchedAtlas, patchedPAF, err := zillfont.FullRepackAuthenticatedRetailFont(atlas, jillbtn, plan.Mapping, koreanRasters)
	if err != nil {
		return nil, err
	}
	return []paa.Replacement{
		paa.IndexReplacement(retailAtlasMemberIndex, patchedAtlas),
		paa.IndexReplacement(retailPAFMemberIndex, patchedPAF),
	}, nil
}
