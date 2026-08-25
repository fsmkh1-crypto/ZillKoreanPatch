// SPDX-License-Identifier: GPL-3.0-or-later

package koreanfont

import (
	"bytes"
	"testing"

	"github.com/HK47196/zill/internal/zillfont"
)

func TestEncodeRoundTripDeterministic(t *testing.T) {
	pixelsA := make([]uint8, 100)
	pixelsB := make([]uint8, 100)
	for i := range pixelsA { pixelsA[i] = uint8(i % 16); pixelsB[i] = uint8((15-i) & 15) }
	rasters := map[rune]zillfont.Raster{
		'힣': {Width: 10, Height: 10, Pixels: pixelsB},
		'가': {Width: 10, Height: 10, Pixels: pixelsA},
	}
	first, err := Encode("fixture", "fixture-rule", rasters)
	if err != nil { t.Fatal(err) }
	second, err := Encode("fixture", "fixture-rule", rasters)
	if err != nil { t.Fatal(err) }
	if !bytes.Equal(first, second) { t.Fatal("catalog encoding is not deterministic") }
	catalog, err := Parse(first)
	if err != nil { t.Fatal(err) }
	for r, want := range rasters {
		got, ok := catalog.SourceRaster(r)
		if !ok { t.Fatalf("missing %U", r) }
		if !bytes.Equal(got.Pixels, want.Pixels) { t.Fatalf("pixels differ for %U", r) }
	}
}

func TestEncodeRejectsInvalidRaster(t *testing.T) {
	_, err := Encode("fixture", "rule", map[rune]zillfont.Raster{'가': {Width: 9, Height: 10, Pixels: make([]uint8, 90)}})
	if err == nil { t.Fatal("expected invalid dimensions to fail") }
	pixels := make([]uint8, 100); pixels[7] = 16
	_, err = Encode("fixture", "rule", map[rune]zillfont.Raster{'가': {Width: 10, Height: 10, Pixels: pixels}})
	if err == nil { t.Fatal("expected >4bpp pixel to fail") }
}
