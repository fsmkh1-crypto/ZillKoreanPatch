// SPDX-License-Identifier: GPL-3.0-or-later

package koreanfont

import (
	"strings"
	"testing"

	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/zillfont"
)

func testCatalogText(glyphs string) string {
	return `format = "zill-korean-raster-catalog"
version = 1
width = 10
height = 10
source_font = "test-font"
render_rule = "test-rule"

[glyphs]
` + glyphs
}

func TestParseCatalogAndPlaceCellRaster(t *testing.T) {
	pixels := strings.Repeat("0", 99) + "f"
	catalog, err := Parse([]byte(testCatalogText(`"가" = "` + pixels + `"`)))
	if err != nil {
		t.Fatal(err)
	}
	replacements := []zillfont.Replacement{{
		Rune: '가', Key: cp932.GlyphKey(0xAC82),
		Glyph: zillfont.Glyph{Index: 1, Key: cp932.GlyphKey(0xAC82), Width: 11, Height: 12, BearingX: 0, BearingY: -10, Advance: 12, Page: 1},
	}}
	got, err := catalog.CellRasters(replacements)
	if err != nil {
		t.Fatal(err)
	}
	raster := got['가']
	if raster.Width != 11 || raster.Height != 12 {
		t.Fatalf("placed raster = %dx%d", raster.Width, raster.Height)
	}
	// Source pixel (9,9) lands at (10,10) after the proven (1,1) placement.
	if raster.Pixels[10*11+10] != 15 {
		t.Fatalf("placed terminal pixel = %d", raster.Pixels[10*11+10])
	}
}

func TestCatalogRejectsUnknownFieldsAndBadPixels(t *testing.T) {
	_, err := Parse([]byte(testCatalogText(`"가" = "` + strings.Repeat("0", 99) + `z"`)))
	if err == nil || !strings.Contains(err.Error(), "not hexadecimal") {
		t.Fatalf("error = %v", err)
	}
	text := testCatalogText(`"가" = "` + strings.Repeat("0", 100) + `"`) + "extra = 1\n"
	if _, err := Parse([]byte(text)); err == nil {
		t.Fatal("expected unknown-field error")
	}
}

func TestCellRastersFailsClosedOnMissingGlyph(t *testing.T) {
	catalog, err := Parse([]byte(testCatalogText(`"가" = "` + strings.Repeat("0", 100) + `"`)))
	if err != nil {
		t.Fatal(err)
	}
	_, err = catalog.CellRasters([]zillfont.Replacement{{
		Rune: '나', Glyph: zillfont.Glyph{Width: 10, Height: 10, BearingX: 1, BearingY: -9, Advance: 12, Page: 1},
	}})
	if err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("error = %v", err)
	}
}
