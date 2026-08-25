// SPDX-License-Identifier: GPL-3.0-or-later

package zillfont

import "fmt"

const (
	AtlasSize       = 512
	atlasRowBytes   = AtlasSize / 2 // 4bpp
	imageDataOffset = 0x80
)

var gimStarts = [...]int{0xC0, 0x201B0, 0x402A0, 0x60390}

// Raster is one unpacked 4bpp glyph bitmap. Pixels are row-major alpha/index
// values in the inclusive range 0..15. The low-level atlas writer deliberately
// requires the raster to match the target PAF cell exactly; placement, padding,
// baseline, and font-rendering policy live above this package boundary.
type Raster struct {
	Width  int
	Height int
	Pixels []uint8
}

func (r Raster) validate() error {
	if r.Width <= 0 || r.Height <= 0 {
		return fmt.Errorf("invalid raster dimensions %dx%d", r.Width, r.Height)
	}
	if len(r.Pixels) != r.Width*r.Height {
		return fmt.Errorf("raster has %d pixels, want %d for %dx%d", len(r.Pixels), r.Width*r.Height, r.Width, r.Height)
	}
	for index, pixel := range r.Pixels {
		if pixel > 0x0f {
			return fmt.Errorf("raster pixel %d has 4bpp value %#x", index, pixel)
		}
	}
	return nil
}

// SwizzledByteOffset converts one 512-wide 4bpp pixel coordinate into a byte
// offset relative to the GIM image-data payload. Two horizontal pixels share a
// byte; even X occupies the low nibble and odd X the high nibble.
func SwizzledByteOffset(pixelX, pixelY int) (int, error) {
	if pixelX < 0 || pixelX >= AtlasSize || pixelY < 0 || pixelY >= AtlasSize {
		return 0, fmt.Errorf("atlas coordinate (%d,%d) is outside %dx%d", pixelX, pixelY, AtlasSize, AtlasSize)
	}
	byteX := pixelX >> 1
	blocksPerRow := atlasRowBytes / 16
	return (pixelY/8)*blocksPerRow*128 + (byteX/16)*128 + (pixelY&7)*16 + (byteX & 15), nil
}

// PatchAtlasCell replaces exactly one existing PAF glyph cell in an authenticated
// font/zillfont.par member. It changes atlas texels only; PAF keys, metrics, BST
// links, and record count remain untouched.
func PatchAtlasCell(member []byte, glyph Glyph, raster Raster) error {
	if err := raster.validate(); err != nil {
		return err
	}
	if glyph.Page >= uint32(len(gimStarts)) {
		return fmt.Errorf("glyph %d key 0x%04X references unsupported GIM page %d", glyph.Index, uint16(glyph.Key), glyph.Page)
	}
	if int(glyph.Width) != raster.Width || int(glyph.Height) != raster.Height {
		return fmt.Errorf("glyph %d key 0x%04X cell is %dx%d but raster is %dx%d", glyph.Index, uint16(glyph.Key), glyph.Width, glyph.Height, raster.Width, raster.Height)
	}
	if int(glyph.X)+raster.Width > AtlasSize || int(glyph.Y)+raster.Height > AtlasSize {
		return fmt.Errorf("glyph %d key 0x%04X cell (%d,%d %dx%d) exceeds atlas bounds", glyph.Index, uint16(glyph.Key), glyph.X, glyph.Y, raster.Width, raster.Height)
	}

	for row := 0; row < raster.Height; row++ {
		for column := 0; column < raster.Width; column++ {
			pixelX := int(glyph.X) + column
			pixelY := int(glyph.Y) + row
			swizzled, err := SwizzledByteOffset(pixelX, pixelY)
			if err != nil {
				return err
			}
			offset := gimStarts[glyph.Page] + imageDataOffset + swizzled
			if offset < 0 || offset >= len(member) {
				return fmt.Errorf("glyph %d key 0x%04X atlas byte offset %#x exceeds member size %#x", glyph.Index, uint16(glyph.Key), offset, len(member))
			}
			pixel := raster.Pixels[row*raster.Width+column]
			if pixelX&1 == 0 {
				member[offset] = (member[offset] & 0xf0) | pixel
			} else {
				member[offset] = (member[offset] & 0x0f) | (pixel << 4)
			}
		}
	}
	return nil
}

// ApplyRasters applies a complete replacement plan to a copy of the retail font
// member. Missing or extra rasters fail closed, as do overlapping target pixels.
func ApplyRasters(member []byte, replacements []Replacement, rasters map[rune]Raster) ([]byte, error) {
	if len(rasters) != len(replacements) {
		return nil, fmt.Errorf("font raster set has %d entries for %d replacements", len(rasters), len(replacements))
	}
	out := append([]byte(nil), member...)
	seenRunes := make(map[rune]struct{}, len(replacements))
	occupied := make(map[[3]int]rune)
	for _, replacement := range replacements {
		if _, duplicate := seenRunes[replacement.Rune]; duplicate {
			return nil, fmt.Errorf("duplicate font replacement for %U", replacement.Rune)
		}
		seenRunes[replacement.Rune] = struct{}{}
		raster, ok := rasters[replacement.Rune]
		if !ok {
			return nil, fmt.Errorf("missing raster for %U", replacement.Rune)
		}
		for row := 0; row < raster.Height; row++ {
			for column := 0; column < raster.Width; column++ {
				coordinate := [3]int{int(replacement.Glyph.Page), int(replacement.Glyph.X) + column, int(replacement.Glyph.Y) + row}
				if prior, exists := occupied[coordinate]; exists {
					return nil, fmt.Errorf("font cells for %U and %U overlap at page %d (%d,%d)", prior, replacement.Rune, coordinate[0], coordinate[1], coordinate[2])
				}
				occupied[coordinate] = replacement.Rune
			}
		}
		if err := PatchAtlasCell(out, replacement.Glyph, raster); err != nil {
			return nil, fmt.Errorf("patch %U: %w", replacement.Rune, err)
		}
	}
	for r := range rasters {
		if _, ok := seenRunes[r]; !ok {
			return nil, fmt.Errorf("raster supplied for unmapped rune %U", r)
		}
	}
	return out, nil
}
