// SPDX-License-Identifier: GPL-3.0-or-later

package zillfont

import (
	"encoding/binary"
	"fmt"
	"sort"

	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
)

// KoreanMetricRewriteKeys returns installed two-byte renderer keys whose atlas
// cells are physically large enough for the proven 10x10 Korean raster. Unlike
// KoreanCompatibleKeys, these slots may have their bearing/advance metadata
// rewritten for Korean use while keeping key, cell size, atlas position, page,
// BST ordering, and record count unchanged.
func (p *PAF) KoreanMetricRewriteKeys() []cp932.GlyphKey {
	if p == nil {
		return nil
	}
	keys := make([]cp932.GlyphKey, 0, len(p.Glyphs))
	for _, glyph := range p.Glyphs {
		if !glyph.Key.IsDoubleByte() || glyph.Page >= uint32(len(gimStarts)) {
			continue
		}
		if int(glyph.Width) < KoreanRasterWidth || int(glyph.Height) < KoreanRasterHeight {
			continue
		}
		if int(glyph.X)+int(glyph.Width) > AtlasSize || int(glyph.Y)+int(glyph.Height) > AtlasSize {
			continue
		}
		keys = append(keys, glyph.Key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

// MetricRewriteReplacementPlan resolves a mapping to retail cells and presents
// those cells with the Korean target metrics. The physical atlas cell geometry
// remains unchanged; only bearing/advance are altered in the returned Glyphs.
func (p *PAF) MetricRewriteReplacementPlan(mapping koreanslots.Mapping) ([]Replacement, error) {
	replacements, err := p.ReplacementPlan(mapping)
	if err != nil {
		return nil, err
	}
	for i := range replacements {
		glyph := &replacements[i].Glyph
		if glyph.Page >= uint32(len(gimStarts)) {
			return nil, fmt.Errorf("metric rewrite: glyph %d key 0x%04X uses unsupported page %d", glyph.Index, uint16(glyph.Key), glyph.Page)
		}
		if int(glyph.Width) < KoreanRasterWidth || int(glyph.Height) < KoreanRasterHeight {
			return nil, fmt.Errorf("metric rewrite: glyph %d key 0x%04X cell %dx%d is smaller than %dx%d Korean raster", glyph.Index, uint16(glyph.Key), glyph.Width, glyph.Height, KoreanRasterWidth, KoreanRasterHeight)
		}
		glyph.BearingX = KoreanTargetBearingX
		glyph.BearingY = KoreanTargetBearingY
		glyph.Advance = KoreanTargetAdvance
	}
	return replacements, nil
}

// PatchAuthenticatedRetailFontWithMetricRewrite patches both the retail atlas
// and jillbtn PAF metadata for the mapped Korean slots. Only bearing X/Y and
// advance are rewritten in PAF records; renderer keys, atlas coordinates,
// dimensions, pages, tree links, and record count remain byte-for-byte retail.
func PatchAuthenticatedRetailFontWithMetricRewrite(zillfontMember, jillbtnMember []byte, mapping koreanslots.Mapping, rasters map[rune]Raster) ([]byte, []byte, error) {
	if err := verifyRetailMember("font/zillfont.par", zillfontMember, RetailAtlasMemberSize, retailAtlasSHA256); err != nil {
		return nil, nil, err
	}
	if err := verifyRetailMember("2d/font/jillbtn.par", jillbtnMember, RetailPAFMemberSize, retailPAFSHA256); err != nil {
		return nil, nil, err
	}
	paf, err := ParseAuthenticatedRetailPAF(jillbtnMember)
	if err != nil {
		return nil, nil, err
	}
	replacements, err := paf.MetricRewriteReplacementPlan(mapping)
	if err != nil {
		return nil, nil, err
	}
	patchedAtlas, err := ApplyRasters(zillfontMember, replacements, rasters)
	if err != nil {
		return nil, nil, err
	}
	patchedPAF := append([]byte(nil), jillbtnMember...)
	for _, replacement := range replacements {
		offset := RetailPAFOffset + RecordOffset + replacement.Glyph.Index*RecordStride
		if offset < RetailPAFOffset || offset+RecordStride > len(patchedPAF) {
			return nil, nil, fmt.Errorf("metric rewrite: glyph %d PAF record is out of range", replacement.Glyph.Index)
		}
		binary.LittleEndian.PutUint16(patchedPAF[offset+8:offset+10], uint16(int16(KoreanTargetBearingX)))
		binary.LittleEndian.PutUint16(patchedPAF[offset+10:offset+12], uint16(int16(KoreanTargetBearingY)))
		binary.LittleEndian.PutUint32(patchedPAF[offset+12:offset+16], KoreanTargetAdvance)
	}
	return patchedAtlas, patchedPAF, nil
}
