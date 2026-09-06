// SPDX-License-Identifier: GPL-3.0-or-later

package koreanfont

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/HK47196/zill/internal/zillfont"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// KoreanAlphaGamma is the Beta 1 B-profile alpha lift. Keep the 10x10 source
// raster and proven placement metrics unchanged; apply gamma only immediately
// before 4bpp quantization so geometry/reflow contracts do not move with the
// visual-weight adjustment.
const KoreanAlphaGamma = 0.60

// ProvenRenderRule names the repository-owned rasterization rule. The baseline
// is derived from the font ascent so its visual top-left matches the historical
// PoC's text origin (0,-2). Beta 1's B profile applies gamma 0.60 to normalized
// alpha immediately before rounding to 4bpp.
const ProvenRenderRule = "opentype-10px-72dpi-hinting-none-origin-0,-2-alpha-gamma-0.60-round-4bpp-v2"

// RenderRequired rasterizes every requested rune with a deterministic OpenType
// face configuration. Callers own fontData; no system font lookup is performed.
func RenderRequired(fontData []byte, runes []rune) (map[rune]zillfont.Raster, error) {
	parsed, err := opentype.Parse(fontData)
	if err != nil {
		return nil, fmt.Errorf("parse Korean source font: %w", err)
	}
	face, err := opentype.NewFace(parsed, &opentype.FaceOptions{Size: 10, DPI: 72, Hinting: font.HintingNone})
	if err != nil {
		return nil, fmt.Errorf("open Korean source font face: %w", err)
	}
	defer face.Close()

	unique := make(map[rune]struct{}, len(runes))
	out := make(map[rune]zillfont.Raster, len(runes))
	for _, r := range runes {
		if _, exists := unique[r]; exists {
			continue
		}
		unique[r] = struct{}{}
		raster, err := renderRune(face, r)
		if err != nil {
			return nil, err
		}
		out[r] = raster
	}
	return out, nil
}

func alphaTo4BPP(alpha uint8) uint8 {
	normalized := float64(alpha) / 255.0
	corrected := math.Pow(normalized, KoreanAlphaGamma)
	return uint8(math.Round(corrected * 15.0))
}

func renderRune(face font.Face, r rune) (zillfont.Raster, error) {
	if _, ok := face.GlyphAdvance(r); !ok {
		return zillfont.Raster{}, fmt.Errorf("Korean source font has no glyph for %U", r)
	}
	canvas := image.NewAlpha(image.Rect(0, 0, zillfont.KoreanRasterWidth, zillfont.KoreanRasterHeight))
	baseline := face.Metrics().Ascent + fixed.I(-2)
	drawer := font.Drawer{Dst: canvas, Src: image.NewUniform(color.Alpha{A: 0xff}), Face: face, Dot: fixed.Point26_6{X: 0, Y: baseline}}
	drawer.DrawString(string(r))

	pixels := make([]uint8, zillfont.KoreanRasterWidth*zillfont.KoreanRasterHeight)
	nonzero := false
	for y := 0; y < zillfont.KoreanRasterHeight; y++ {
		for x := 0; x < zillfont.KoreanRasterWidth; x++ {
			value := alphaTo4BPP(canvas.AlphaAt(x, y).A)
			pixels[y*zillfont.KoreanRasterWidth+x] = value
			if value != 0 {
				nonzero = true
			}
		}
	}
	if !nonzero {
		return zillfont.Raster{}, fmt.Errorf("Korean source font rendered empty glyph for %U", r)
	}
	return zillfont.Raster{Width: zillfont.KoreanRasterWidth, Height: zillfont.KoreanRasterHeight, Pixels: pixels}, nil
}
