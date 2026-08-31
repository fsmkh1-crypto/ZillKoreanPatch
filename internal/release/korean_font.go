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

const (
	retailAtlasMemberIndex = 13611
	retailPAFMemberIndex   = 13612
	retailAtlasMemberName  = "font/zillfont.par"
	retailPAFMemberName    = "2d/font/jillbtn.par"
)

type koreanFontInputs struct {
	atlas   []byte
	jillbtn []byte
	paf     *zillfont.PAF
	catalog *koreanfont.Catalog
}

// loadKoreanFontInputs is the single authenticated source path for every Korean
// font build mode. Desktop and mobile may choose different transforms, but they
// must not independently rediscover archives, authenticate retail members, or
// interpret the raster catalog.
func loadKoreanFontInputs(root string, archives []*archive, mapping koreanslots.Mapping, label string) (koreanFontInputs, error) {
	var out koreanFontInputs
	var paArchive *archive
	for _, candidate := range archives {
		if candidate != nil && candidate.name == "pa" {
			if paArchive != nil {
				return out, fmt.Errorf("%s: duplicate pa archive", label)
			}
			paArchive = candidate
		}
	}
	if paArchive == nil || paArchive.pair == nil {
		return out, fmt.Errorf("%s: pa archive is unavailable", label)
	}
	members := paArchive.pair.Members()
	if retailPAFMemberIndex >= len(members) {
		return out, fmt.Errorf("%s: retail pa archive has only %d members", label, len(members))
	}
	atlasMember := members[retailAtlasMemberIndex]
	pafMember := members[retailPAFMemberIndex]
	if atlasMember.Name != retailAtlasMemberName || atlasMember.Size != zillfont.RetailAtlasMemberSize {
		return out, fmt.Errorf("%s: member %d is %q size %#x, want %q size %#x",
			label, retailAtlasMemberIndex, atlasMember.Name, atlasMember.Size, retailAtlasMemberName, zillfont.RetailAtlasMemberSize)
	}
	if pafMember.Name != retailPAFMemberName || pafMember.Size != zillfont.RetailPAFMemberSize {
		return out, fmt.Errorf("%s: member %d is %q size %#x, want %q size %#x",
			label, retailPAFMemberIndex, pafMember.Name, pafMember.Size, retailPAFMemberName, zillfont.RetailPAFMemberSize)
	}

	atlas, err := paArchive.pair.Payload(retailAtlasMemberIndex)
	if err != nil {
		return out, err
	}
	jillbtn, err := paArchive.pair.Payload(retailPAFMemberIndex)
	if err != nil {
		return out, err
	}
	paf, err := zillfont.AuthenticateRetailFont(atlas, jillbtn)
	if err != nil {
		return out, err
	}

	catalogPath := filepath.Join(root, "release", "korean", "font", "glyphs.toml")
	catalogData, err := os.ReadFile(catalogPath)
	if err != nil {
		return out, fmt.Errorf("read Korean raster catalog: %w", err)
	}
	catalog, err := koreanfont.Parse(catalogData)
	if err != nil {
		return out, err
	}
	mappedRunes := make([]rune, 0, len(mapping))
	for r := range mapping {
		mappedRunes = append(mappedRunes, r)
	}
	if err := catalog.RequireRunes(mappedRunes); err != nil {
		return out, err
	}
	out.atlas = atlas
	out.jillbtn = jillbtn
	out.paf = paf
	out.catalog = catalog
	return out, nil
}

// prepareKoreanFontReplacement builds the single PAA replacement needed for
// custom Korean glyphs. It deliberately starts from authenticated retail font
// members rather than the existing English static font transform; composing
// those two transforms must be an explicit build-mode decision, never an
// accidental duplicate replacement of the same archive member.
func prepareKoreanFontReplacement(root string, archives []*archive, plan koreanslots.Plan) (paa.Replacement, bool, error) {
	if len(plan.Mapping) == 0 {
		return paa.Replacement{}, false, nil
	}
	inputs, err := loadKoreanFontInputs(root, archives, plan.Mapping, "prepare Korean font")
	if err != nil {
		return paa.Replacement{}, false, err
	}
	replacements, err := inputs.paf.ReplacementPlan(plan.Mapping)
	if err != nil {
		return paa.Replacement{}, false, err
	}
	cellRasters, err := inputs.catalog.CellRasters(replacements)
	if err != nil {
		return paa.Replacement{}, false, err
	}
	patched, err := zillfont.PatchAuthenticatedRetailAtlas(inputs.atlas, inputs.jillbtn, plan.Mapping, cellRasters)
	if err != nil {
		return paa.Replacement{}, false, err
	}
	return paa.IndexReplacement(retailAtlasMemberIndex, patched), true, nil
}
