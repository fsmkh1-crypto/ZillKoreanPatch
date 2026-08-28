// SPDX-License-Identifier: GPL-3.0-or-later

package zillfont

import (
	"testing"

	"github.com/HK47196/zill/internal/cp932"
)

func TestKoreanPlacementMatchesProvenMetricProfiles(t *testing.T) {
	cases := []struct {
		glyph Glyph
		x, y  int
	}{
		{Glyph{Index: 1, Key: cp932.GlyphKey(0xAC82), Width: 10, Height: 10, BearingX: 1, BearingY: -9, Advance: 12, Page: 1}, 0, 0},
		{Glyph{Index: 2, Key: cp932.GlyphKey(0xAD82), Width: 11, Height: 12, BearingX: 0, BearingY: -10, Advance: 12, Page: 1}, 1, 1},
	}
	for _, tc := range cases {
		x, y, err := KoreanPlacement(tc.glyph)
		if err != nil {
			t.Fatal(err)
		}
		if x != tc.x || y != tc.y {
			t.Fatalf("placement = (%d,%d), want (%d,%d)", x, y, tc.x, tc.y)
		}
	}
}

func TestKoreanPlacementRejectsBadAdvanceAndTooSmallCell(t *testing.T) {
	glyph := Glyph{Index: 1, Key: cp932.GlyphKey(0xAC82), Width: 10, Height: 10, BearingX: 1, BearingY: -9, Advance: 11, Page: 1}
	if _, _, err := KoreanPlacement(glyph); err == nil {
		t.Fatal("expected advance rejection")
	}
	glyph.Advance = 12
	glyph.Width = 9
	if _, _, err := KoreanPlacement(glyph); err == nil {
		t.Fatal("expected size rejection")
	}
}

func TestKoreanCompatibleKeysFiltersMetricsPrivateKeysAndSorts(t *testing.T) {
	paf := &PAF{Glyphs: []Glyph{
		{Index: 4, Key: cp932.GlyphKey(0xAD81), Width: 12, Height: 12, BearingX: 0, BearingY: -10, Advance: 12, Page: 1}, // CP932-shaped but not text
		{Index: 3, Key: cp932.GlyphKey(0xAD82), Width: 11, Height: 11, BearingX: 0, BearingY: -10, Advance: 12, Page: 1},
		{Index: 2, Key: cp932.GlyphKey(0xAC82), Width: 10, Height: 10, BearingX: 1, BearingY: -9, Advance: 12, Page: 1},
		{Index: 5, Key: cp932.GlyphKey(0xAE82), Width: 12, Height: 12, BearingX: 0, BearingY: -10, Advance: 10, Page: 1},
	}}
	got := paf.KoreanCompatibleKeys()
	if len(got) != 2 || got[0] != cp932.GlyphKey(0xAC82) || got[1] != cp932.GlyphKey(0xAD82) {
		t.Fatalf("keys = %#v", got)
	}
}

func TestPlaceKoreanRasterExpandsToRetailCell(t *testing.T) {
	glyph := Glyph{Index: 2, Key: cp932.GlyphKey(0xAD82), Width: 11, Height: 12, BearingX: 0, BearingY: -10, Advance: 12, Page: 1}
	source := Raster{Width: 10, Height: 10, Pixels: make([]uint8, 100)}
	source.Pixels[0] = 15
	got, err := PlaceKoreanRaster(glyph, source)
	if err != nil {
		t.Fatal(err)
	}
	if got.Width != 11 || got.Height != 12 || got.Pixels[1*11+1] != 15 {
		t.Fatalf("unexpected placed raster: %dx%d first=%d", got.Width, got.Height, got.Pixels[1*11+1])
	}
}
