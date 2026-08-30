// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"crypto/sha256"
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

	// These are the same authenticated retail source fingerprints declared by
	// the upstream English static-font manifest. Korean font transforms must
	// start from exactly the same engine assets, not merely same-sized members.
	retailAtlasSourceSHA256 = "0d3d6d2648870e87a01636cdfc7cc7af8100ea40b71e5ed05f82ac197606584a"
	retailPAFSourceSHA256   = "95b48379092db4db72f890d5a221ba8c4094dd438cb4c4eba98eb5520c7b17aa"
)

func verifyKoreanFontRetailSources(atlas, paf []byte) error {
	if actual := fmt.Sprintf("%x", sha256.Sum256(atlas)); actual != retailAtlasSourceSHA256 {
		return fmt.Errorf("Korean font retail atlas fingerprint is %s, want upstream-English source %s", actual, retailAtlasSourceSHA256)
	}
	if actual := fmt.Sprintf("%x", sha256.Sum256(paf)); actual != retailPAFSourceSHA256 {
		return fmt.Errorf("Korean font retail PAF fingerprint is %s, want upstream-English source %s", actual, retailPAFSourceSHA256)
	}
	return nil
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
	var paArchive *archive
	for _, candidate := range archives {
		if candidate != nil && candidate.name == "pa" {
			if paArchive != nil {
				return paa.Replacement{}, false, fmt.Errorf("prepare Korean font: duplicate pa archive")
			}
			paArchive = candidate
		}
	}
	if paArchive == nil || paArchive.pair == nil {
		return paa.Replacement{}, false, fmt.Errorf("prepare Korean font: pa archive is unavailable")
	}
	members := paArchive.pair.Members()
	if retailPAFMemberIndex >= len(members) {
		return paa.Replacement{}, false, fmt.Errorf("prepare Korean font: retail pa archive has only %d members", len(members))
	}
	atlasMember := members[retailAtlasMemberIndex]
	pafMember := members[retailPAFMemberIndex]
	if atlasMember.Name != retailAtlasMemberName || atlasMember.Size != zillfont.RetailAtlasMemberSize {
		return paa.Replacement{}, false, fmt.Errorf("prepare Korean font: member %d is %q size %#x, want %q size %#x",
			retailAtlasMemberIndex, atlasMember.Name, atlasMember.Size, retailAtlasMemberName, zillfont.RetailAtlasMemberSize)
	}
	if pafMember.Name != retailPAFMemberName || pafMember.Size != zillfont.RetailPAFMemberSize {
		return paa.Replacement{}, false, fmt.Errorf("prepare Korean font: member %d is %q size %#x, want %q size %#x",
			retailPAFMemberIndex, pafMember.Name, pafMember.Size, retailPAFMemberName, zillfont.RetailPAFMemberSize)
	}

	atlas, err := paArchive.pair.Payload(retailAtlasMemberIndex)
	if err != nil {
		return paa.Replacement{}, false, err
	}
	jillbtn, err := paArchive.pair.Payload(retailPAFMemberIndex)
	if err != nil {
		return paa.Replacement{}, false, err
	}
	if err := verifyKoreanFontRetailSources(atlas, jillbtn); err != nil {
		return paa.Replacement{}, false, err
	}
	paf, err := zillfont.ParseAuthenticatedRetailPAF(jillbtn)
	if err != nil {
		return paa.Replacement{}, false, err
	}
	replacements, err := paf.ReplacementPlan(plan.Mapping)
	if err != nil {
		return paa.Replacement{}, false, err
	}

	catalogPath := filepath.Join(root, "release", "korean", "font", "glyphs.toml")
	catalogData, err := os.ReadFile(catalogPath)
	if err != nil {
		return paa.Replacement{}, false, fmt.Errorf("read Korean raster catalog: %w", err)
	}
	catalog, err := koreanfont.Parse(catalogData)
	if err != nil {
		return paa.Replacement{}, false, err
	}
	mappedRunes := make([]rune, 0, len(plan.Mapping))
	for r := range plan.Mapping {
		mappedRunes = append(mappedRunes, r)
	}
	if err := catalog.RequireRunes(mappedRunes); err != nil {
		return paa.Replacement{}, false, err
	}
	cellRasters, err := catalog.CellRasters(replacements)
	if err != nil {
		return paa.Replacement{}, false, err
	}
	patched, err := zillfont.PatchAuthenticatedRetailAtlas(atlas, jillbtn, plan.Mapping, cellRasters)
	if err != nil {
		return paa.Replacement{}, false, err
	}
	return paa.IndexReplacement(retailAtlasMemberIndex, patched), true, nil
}
