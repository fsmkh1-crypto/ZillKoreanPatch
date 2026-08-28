// SPDX-License-Identifier: GPL-3.0-or-later

package zillfont

import (
	"fmt"
	"sort"

	"github.com/HK47196/zill/internal/cp932"
)

const (
	KoreanRasterWidth   = 10
	KoreanRasterHeight  = 10
	KoreanTargetBearingX = 1
	KoreanTargetBearingY = -9
	KoreanTargetAdvance  = 12
)

// KoreanPlacement returns where the proven 10x10 Korean raster must be pasted
// inside an existing retail glyph cell so its effective bearing is (1,-9) with
// advance 12. The PAF record itself remains unchanged.
func KoreanPlacement(glyph Glyph) (pasteX, pasteY int, err error) {
	if glyph.Advance != KoreanTargetAdvance {
		return 0, 0, fmt.Errorf("glyph %d key 0x%04X advance %d, want %d", glyph.Index, uint16(glyph.Key), glyph.Advance, KoreanTargetAdvance)
	}
	pasteX = KoreanTargetBearingX - int(glyph.BearingX)
	pasteY = KoreanTargetBearingY - int(glyph.BearingY)
	if pasteX < 0 || pasteY < 0 || pasteX+KoreanRasterWidth > int(glyph.Width) || pasteY+KoreanRasterHeight > int(glyph.Height) {
		return 0, 0, fmt.Errorf("glyph %d key 0x%04X cell %dx%d bearing=(%d,%d) cannot host %dx%d Korean raster at effective bearing (%d,%d)",
			glyph.Index, uint16(glyph.Key), glyph.Width, glyph.Height, glyph.BearingX, glyph.BearingY,
			KoreanRasterWidth, KoreanRasterHeight, KoreanTargetBearingX, KoreanTargetBearingY)
	}
	if glyph.Page >= uint32(len(gimStarts)) {
		return 0, 0, fmt.Errorf("glyph %d key 0x%04X references unsupported page %d", glyph.Index, uint16(glyph.Key), glyph.Page)
	}
	return pasteX, pasteY, nil
}

// KoreanCompatibleKeys returns installed round-trip CP932 text renderer keys
// whose existing PAF metrics can host the proven Korean raster without changing
// PAF metadata. Renderer-private/UI keys are excluded before metric filtering.
// The result is sorted and therefore safe to feed directly to deterministic
// slot allocation.
func (p *PAF) KoreanCompatibleKeys() []cp932.GlyphKey {
	if p == nil {
		return nil
	}
	keys := make([]cp932.GlyphKey, 0, len(p.Glyphs))
	for _, glyph := range p.Glyphs {
		if !glyph.Key.IsRoundTripText() {
			continue
		}
		if _, _, err := KoreanPlacement(glyph); err != nil {
			continue
		}
		keys = append(keys, glyph.Key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

// PlaceKoreanRaster converts one proven 10x10 raster into the exact dimensions
// of its selected retail cell using the placement derived from PAF metrics.
func PlaceKoreanRaster(glyph Glyph, source Raster) (Raster, error) {
	if source.Width != KoreanRasterWidth || source.Height != KoreanRasterHeight {
		return Raster{}, fmt.Errorf("Korean source raster is %dx%d, want %dx%d", source.Width, source.Height, KoreanRasterWidth, KoreanRasterHeight)
	}
	pasteX, pasteY, err := KoreanPlacement(glyph)
	if err != nil {
		return Raster{}, err
	}
	return PlaceRaster(int(glyph.Width), int(glyph.Height), pasteX, pasteY, source)
}
