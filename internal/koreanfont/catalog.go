// SPDX-License-Identifier: GPL-3.0-or-later

// Package koreanfont loads deterministic pre-rendered Korean glyph rasters used
// by the production patch builder. Rendering is intentionally separated from
// release compilation so builds never depend on whichever system font happens
// to be installed on the maintainer machine.
package koreanfont

import (
	"bytes"
	"fmt"
	"unicode/utf8"

	"github.com/HK47196/zill/internal/zillfont"
	"github.com/pelletier/go-toml/v2"
)

const (
	catalogFormat  = "zill-korean-raster-catalog"
	catalogVersion = 1
)

type rawCatalog struct {
	Format      string            `toml:"format"`
	Version     int               `toml:"version"`
	Width       int               `toml:"width"`
	Height      int               `toml:"height"`
	SourceFont  string            `toml:"source_font"`
	RenderRule  string            `toml:"render_rule"`
	Glyphs      map[string]string `toml:"glyphs"`
}

// Catalog is an authenticated-by-source-control set of unpacked 4bpp source
// rasters. Each raster is the proven 10x10 Korean visual box; placement into a
// selected retail PAF cell is derived separately from that cell's metrics.
type Catalog struct {
	SourceFont string
	RenderRule string
	rasters    map[rune]zillfont.Raster
}

// Parse loads a strict TOML raster catalog. Pixels are encoded as one lowercase
// or uppercase hexadecimal nibble per pixel in row-major order.
func Parse(data []byte) (*Catalog, error) {
	var raw rawCatalog
	decoder := toml.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("Korean raster catalog: %w", err)
	}
	if raw.Format != catalogFormat || raw.Version != catalogVersion {
		return nil, fmt.Errorf("unsupported Korean raster catalog format/version")
	}
	if raw.Width != zillfont.KoreanRasterWidth || raw.Height != zillfont.KoreanRasterHeight {
		return nil, fmt.Errorf("Korean raster catalog dimensions %dx%d, want %dx%d", raw.Width, raw.Height, zillfont.KoreanRasterWidth, zillfont.KoreanRasterHeight)
	}
	if raw.SourceFont == "" || raw.RenderRule == "" {
		return nil, fmt.Errorf("Korean raster catalog requires source_font and render_rule")
	}
	if len(raw.Glyphs) == 0 {
		return nil, fmt.Errorf("Korean raster catalog has no glyphs")
	}

	catalog := &Catalog{SourceFont: raw.SourceFont, RenderRule: raw.RenderRule, rasters: make(map[rune]zillfont.Raster, len(raw.Glyphs))}
	wantPixels := raw.Width * raw.Height
	for key, encoded := range raw.Glyphs {
		if utf8.RuneCountInString(key) != 1 {
			return nil, fmt.Errorf("Korean raster catalog glyph key %q is not exactly one Unicode rune", key)
		}
		r, _ := utf8.DecodeRuneInString(key)
		if _, exists := catalog.rasters[r]; exists {
			return nil, fmt.Errorf("Korean raster catalog duplicates %U", r)
		}
		if len(encoded) != wantPixels {
			return nil, fmt.Errorf("Korean raster catalog %U has %d nibbles, want %d", r, len(encoded), wantPixels)
		}
		pixels := make([]uint8, wantPixels)
		for index := 0; index < wantPixels; index++ {
			value, ok := hexNibble(encoded[index])
			if !ok {
				return nil, fmt.Errorf("Korean raster catalog %U pixel %d is not hexadecimal", r, index)
			}
			pixels[index] = value
		}
		catalog.rasters[r] = zillfont.Raster{Width: raw.Width, Height: raw.Height, Pixels: pixels}
	}
	return catalog, nil
}

func hexNibble(value byte) (uint8, bool) {
	switch {
	case value >= '0' && value <= '9':
		return value - '0', true
	case value >= 'a' && value <= 'f':
		return value - 'a' + 10, true
	case value >= 'A' && value <= 'F':
		return value - 'A' + 10, true
	default:
		return 0, false
	}
}

// SourceRaster returns an independent copy of one canonical 10x10 raster.
func (catalog *Catalog) SourceRaster(r rune) (zillfont.Raster, bool) {
	if catalog == nil {
		return zillfont.Raster{}, false
	}
	raster, ok := catalog.rasters[r]
	if !ok {
		return zillfont.Raster{}, false
	}
	raster.Pixels = append([]uint8(nil), raster.Pixels...)
	return raster, true
}

// CellRasters selects every rune in a replacement plan and places its canonical
// source raster into the exact retail cell dimensions implied by PAF metrics.
func (catalog *Catalog) CellRasters(replacements []zillfont.Replacement) (map[rune]zillfont.Raster, error) {
	if catalog == nil {
		return nil, fmt.Errorf("Korean raster catalog is nil")
	}
	out := make(map[rune]zillfont.Raster, len(replacements))
	for _, replacement := range replacements {
		source, ok := catalog.SourceRaster(replacement.Rune)
		if !ok {
			return nil, fmt.Errorf("Korean raster catalog is missing %U", replacement.Rune)
		}
		placed, err := zillfont.PlaceKoreanRaster(replacement.Glyph, source)
		if err != nil {
			return nil, fmt.Errorf("place Korean raster %U: %w", replacement.Rune, err)
		}
		out[replacement.Rune] = placed
	}
	return out, nil
}
