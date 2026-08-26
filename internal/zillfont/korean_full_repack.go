// SPDX-License-Identifier: GPL-3.0-or-later

package zillfont

import (
	"encoding/binary"
	"fmt"
	"sort"

	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
)

const repackUnit = 16

// ExtractAtlasCell copies one glyph bitmap out of an authenticated retail atlas.
// It is the inverse of PatchAtlasCell and lets the Korean full-font transform
// preserve every stock glyph while moving it to a new packed position.
func ExtractAtlasCell(member []byte, glyph Glyph) (Raster, error) {
	if glyph.Width == 0 || glyph.Height == 0 {
		return Raster{}, nil
	}
	if glyph.Page >= uint32(len(gimStarts)) {
		return Raster{}, fmt.Errorf("glyph %d key 0x%04X references unsupported GIM page %d", glyph.Index, uint16(glyph.Key), glyph.Page)
	}
	if int(glyph.X)+int(glyph.Width) > AtlasSize || int(glyph.Y)+int(glyph.Height) > AtlasSize {
		return Raster{}, fmt.Errorf("glyph %d key 0x%04X cell exceeds atlas bounds", glyph.Index, uint16(glyph.Key))
	}
	pixels := make([]uint8, int(glyph.Width)*int(glyph.Height))
	for row := 0; row < int(glyph.Height); row++ {
		for column := 0; column < int(glyph.Width); column++ {
			pixelX := int(glyph.X) + column
			pixelY := int(glyph.Y) + row
			swizzled, err := SwizzledByteOffset(pixelX, pixelY)
			if err != nil {
				return Raster{}, err
			}
			offset := gimStarts[glyph.Page] + imageDataOffset + swizzled
			if offset < 0 || offset >= len(member) {
				return Raster{}, fmt.Errorf("glyph %d key 0x%04X atlas byte offset %#x exceeds member size %#x", glyph.Index, uint16(glyph.Key), offset, len(member))
			}
			value := member[offset]
			if pixelX&1 == 0 {
				pixels[row*int(glyph.Width)+column] = value & 0x0f
			} else {
				pixels[row*int(glyph.Width)+column] = value >> 4
			}
		}
	}
	return Raster{Width: int(glyph.Width), Height: int(glyph.Height), Pixels: pixels}, nil
}

type fullRepackEntry struct {
	glyph  Glyph
	raster Raster
	bw     int
	bh     int
}

// FullRepackAuthenticatedRetailFont follows the same architectural idea as the
// English patch's frozen complete font transform: rebuild the atlas and all PAF
// geometry as one coherent result instead of requiring Korean to fit pre-existing
// retail cells. Keys, BST links and record count stay unchanged. Stock glyph
// bitmaps/metrics are preserved; only mapped custom slots receive Korean rasters
// and Korean metrics.
func FullRepackAuthenticatedRetailFont(atlasMember, pafMember []byte, mapping koreanslots.Mapping, koreanRasters map[rune]Raster) ([]byte, []byte, error) {
	if err := verifyRetailMember("font/zillfont.par", atlasMember, RetailAtlasMemberSize, retailAtlasSHA256); err != nil {
		return nil, nil, err
	}
	if err := verifyRetailMember("2d/font/jillbtn.par", pafMember, RetailPAFMemberSize, retailPAFSHA256); err != nil {
		return nil, nil, err
	}
	paf, err := ParseAuthenticatedRetailPAF(pafMember)
	if err != nil {
		return nil, nil, err
	}
	if len(koreanRasters) != len(mapping) {
		return nil, nil, fmt.Errorf("full Korean repack has %d rasters for %d mappings", len(koreanRasters), len(mapping))
	}

	byKey := make(map[cp932.GlyphKey]rune, len(mapping))
	for r, key := range mapping {
		if prior, exists := byKey[key]; exists {
			return nil, nil, fmt.Errorf("full Korean repack maps %U and %U to key 0x%04X", prior, r, uint16(key))
		}
		byKey[key] = r
	}

	entries := make([]fullRepackEntry, 0, len(paf.Glyphs))
	for _, original := range paf.Glyphs {
		glyph := original
		var raster Raster
		if r, custom := byKey[glyph.Key]; custom {
			var ok bool
			raster, ok = koreanRasters[r]
			if !ok {
				return nil, nil, fmt.Errorf("full Korean repack is missing raster for %U", r)
			}
			if err := raster.validate(); err != nil {
				return nil, nil, fmt.Errorf("full Korean repack raster %U: %w", r, err)
			}
			if raster.Width > 0xff || raster.Height > 0xff {
				return nil, nil, fmt.Errorf("full Korean repack raster %U is too large: %dx%d", r, raster.Width, raster.Height)
			}
			glyph.Width = uint8(raster.Width)
			glyph.Height = uint8(raster.Height)
			glyph.BearingX = KoreanTargetBearingX
			glyph.BearingY = KoreanTargetBearingY
			glyph.Advance = KoreanTargetAdvance
		} else {
			raster, err = ExtractAtlasCell(atlasMember, original)
			if err != nil {
				return nil, nil, err
			}
		}
		bw, bh := 0, 0
		if glyph.Width > 0 && glyph.Height > 0 {
			bw = (int(glyph.Width) + repackUnit - 1) / repackUnit
			bh = (int(glyph.Height) + repackUnit - 1) / repackUnit
		}
		entries = append(entries, fullRepackEntry{glyph: glyph, raster: raster, bw: bw, bh: bh})
	}

	// Pack larger block rectangles first. A 16-pixel allocation grid keeps every
	// destination aligned to the PSP atlas swizzle block while still allowing the
	// PAF cell itself to keep its exact glyph dimensions.
	order := make([]int, 0, len(entries))
	for i := range entries {
		if entries[i].bw > 0 && entries[i].bh > 0 {
			order = append(order, i)
		}
	}
	sort.Slice(order, func(i, j int) bool {
		a, b := entries[order[i]], entries[order[j]]
		if a.bw*a.bh != b.bw*b.bh {
			return a.bw*a.bh > b.bw*b.bh
		}
		if a.bh != b.bh {
			return a.bh > b.bh
		}
		if a.bw != b.bw {
			return a.bw > b.bw
		}
		return a.glyph.Index < b.glyph.Index
	})

	const grid = AtlasSize / repackUnit
	var used [4][grid][grid]bool
	blocksUsed := 0
	for _, entryIndex := range order {
		entry := &entries[entryIndex]
		if entry.bw > grid || entry.bh > grid {
			return nil, nil, fmt.Errorf("full Korean repack glyph %d key 0x%04X requires %dx%d atlas blocks", entry.glyph.Index, uint16(entry.glyph.Key), entry.bw, entry.bh)
		}
		placed := false
		for page := 0; page < len(gimStarts) && !placed; page++ {
			for y := 0; y+entry.bh <= grid && !placed; y++ {
				for x := 0; x+entry.bw <= grid; x++ {
					fits := true
					for yy := y; yy < y+entry.bh && fits; yy++ {
						for xx := x; xx < x+entry.bw; xx++ {
							if used[page][yy][xx] {
								fits = false
								break
							}
						}
					}
					if !fits {
						continue
					}
					for yy := y; yy < y+entry.bh; yy++ {
						for xx := x; xx < x+entry.bw; xx++ {
							used[page][yy][xx] = true
						}
					}
					entry.glyph.Page = uint32(page)
					entry.glyph.X = uint16(x * repackUnit)
					entry.glyph.Y = uint16(y * repackUnit)
					blocksUsed += entry.bw * entry.bh
					placed = true
					break
				}
			}
		}
		if !placed {
			return nil, nil, fmt.Errorf("full Korean repack exhausted four atlas pages after %d/%d 16px blocks", blocksUsed, len(gimStarts)*grid*grid)
		}
	}

	patchedAtlas := append([]byte(nil), atlasMember...)
	pageBytes := AtlasSize * AtlasSize / 2
	for page := range gimStarts {
		start := gimStarts[page] + imageDataOffset
		end := start + pageBytes
		if start < 0 || end > len(patchedAtlas) {
			return nil, nil, fmt.Errorf("full Korean repack page %d image payload is outside atlas member", page)
		}
		clear(patchedAtlas[start:end])
	}
	for _, entry := range entries {
		if entry.glyph.Width == 0 || entry.glyph.Height == 0 {
			continue
		}
		if err := PatchAtlasCell(patchedAtlas, entry.glyph, entry.raster); err != nil {
			return nil, nil, fmt.Errorf("full Korean repack write glyph %d: %w", entry.glyph.Index, err)
		}
	}

	patchedPAF := append([]byte(nil), pafMember...)
	for _, entry := range entries {
		offset := RetailPAFOffset + RecordOffset + entry.glyph.Index*RecordStride
		if offset < RetailPAFOffset || offset+RecordStride > len(patchedPAF) {
			return nil, nil, fmt.Errorf("full Korean repack glyph %d PAF record is out of range", entry.glyph.Index)
		}
		patchedPAF[offset+2] = entry.glyph.Width
		patchedPAF[offset+3] = entry.glyph.Height
		binary.LittleEndian.PutUint16(patchedPAF[offset+4:offset+6], entry.glyph.X)
		binary.LittleEndian.PutUint16(patchedPAF[offset+6:offset+8], entry.glyph.Y)
		bx, by := entry.glyph.BearingX, entry.glyph.BearingY
		binary.LittleEndian.PutUint16(patchedPAF[offset+8:offset+10], uint16(bx))
		binary.LittleEndian.PutUint16(patchedPAF[offset+10:offset+12], uint16(by))
		binary.LittleEndian.PutUint32(patchedPAF[offset+12:offset+16], entry.glyph.Advance)
		binary.LittleEndian.PutUint32(patchedPAF[offset+0x18:offset+0x1c], entry.glyph.Page)
	}
	if _, err := ParsePAF(patchedPAF[RetailPAFOffset:]); err != nil {
		return nil, nil, fmt.Errorf("validate full Korean repack PAF: %w", err)
	}
	return patchedAtlas, patchedPAF, nil
}
