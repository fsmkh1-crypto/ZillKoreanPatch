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

// prepareKoreanMobileFontReplacements is the device-alpha font path. It may
// rewrite bearing/advance metadata for selected retail cells, so both the atlas
// member and jillbtn.par must be rebuilt together.
func prepareKoreanMobileFontReplacements(root string, archives []*archive, plan koreanslots.Plan) ([]paa.Replacement, error) {
	if len(plan.Mapping) == 0 {
		return nil, nil
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
	paf, err := zillfont.ParseAuthenticatedRetailPAF(jillbtn)
	if err != nil {
		return nil, err
	}
	replacements, err := paf.MetricRewriteReplacementPlan(plan.Mapping)
	if err != nil {
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
	cellRasters, err := catalog.CellRasters(replacements)
	if err != nil {
		return nil, err
	}
	patchedAtlas, patchedPAF, err := zillfont.PatchAuthenticatedRetailFontWithMetricRewrite(atlas, jillbtn, plan.Mapping, cellRasters)
	if err != nil {
		return nil, err
	}
	return []paa.Replacement{
		paa.IndexReplacement(retailAtlasMemberIndex, patchedAtlas),
		paa.IndexReplacement(retailPAFMemberIndex, patchedPAF),
	}, nil
}
