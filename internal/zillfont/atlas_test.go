// SPDX-License-Identifier: GPL-3.0-or-later

package zillfont

import (
	"bytes"
	"testing"

	"github.com/HK47196/zill/internal/cp932"
)

func TestSwizzledByteOffset(t *testing.T) {
	cases := []struct {
		x, y int
		want int
	}{
		{0, 0, 0},
		{1, 0, 0},
		{2, 0, 1},
		{32, 0, 128},
		{0, 1, 16},
		{0, 8, 2048},
	}
	for _, tc := range cases {
		got, err := SwizzledByteOffset(tc.x, tc.y)
		if err != nil {
			t.Fatalf("(%d,%d): %v", tc.x, tc.y, err)
		}
		if got != tc.want {
			t.Fatalf("(%d,%d) = %#x, want %#x", tc.x, tc.y, got, tc.want)
		}
	}
	if _, err := SwizzledByteOffset(-1, 0); err == nil {
		t.Fatal("expected bounds error")
	}
}

func TestPatchAtlasCellWritesLowAndHighNibblesOnly(t *testing.T) {
	member := bytes.Repeat([]byte{0x55}, 0x80470)
	offset := gimStarts[0] + imageDataOffset
	member[offset] = 0xab
	before := append([]byte(nil), member...)
	glyph := Glyph{Index: 7, Key: cp932.GlyphKey(0xAC82), Width: 2, Height: 1, X: 0, Y: 0, Page: 0}
	raster := Raster{Width: 2, Height: 1, Pixels: []uint8{1, 2}}
	if err := PatchAtlasCell(member, glyph, raster); err != nil {
		t.Fatal(err)
	}
	if member[offset] != 0x21 {
		t.Fatalf("patched byte = %#02x, want 0x21", member[offset])
	}
	before[offset] = 0x21
	if !bytes.Equal(member, before) {
		t.Fatal("atlas patch changed bytes outside target nibble pair")
	}
}

func TestPatchAtlasCellRejectsMalformedRasterAndBounds(t *testing.T) {
	member := make([]byte, 0x80470)
	glyph := Glyph{Index: 1, Key: cp932.GlyphKey(0xAC82), Width: 1, Height: 1, X: 0, Y: 0, Page: 0}
	if err := PatchAtlasCell(member, glyph, Raster{Width: 1, Height: 1, Pixels: []uint8{16}}); err == nil {
		t.Fatal("expected 4bpp range error")
	}
	glyph.X = 512
	if err := PatchAtlasCell(member, glyph, Raster{Width: 1, Height: 1, Pixels: []uint8{1}}); err == nil {
		t.Fatal("expected atlas bounds error")
	}
	glyph.X = 0
	glyph.Page = 4
	if err := PatchAtlasCell(member, glyph, Raster{Width: 1, Height: 1, Pixels: []uint8{1}}); err == nil {
		t.Fatal("expected page error")
	}
}

func TestApplyRastersRejectsOverlapAndDoesNotMutateSource(t *testing.T) {
	member := make([]byte, 0x80470)
	replacements := []Replacement{
		{Rune: '가', Glyph: Glyph{Index: 1, Key: cp932.GlyphKey(0xAC82), Width: 1, Height: 1, X: 4, Y: 4, Page: 0}},
		{Rune: '나', Glyph: Glyph{Index: 2, Key: cp932.GlyphKey(0xAD82), Width: 1, Height: 1, X: 4, Y: 4, Page: 0}},
	}
	rasters := map[rune]Raster{
		'가': {Width: 1, Height: 1, Pixels: []uint8{1}},
		'나': {Width: 1, Height: 1, Pixels: []uint8{2}},
	}
	if _, err := ApplyRasters(member, replacements, rasters); err == nil {
		t.Fatal("expected overlap error")
	}
	if !bytes.Equal(member, make([]byte, len(member))) {
		t.Fatal("ApplyRasters mutated source member on failed plan")
	}
}
