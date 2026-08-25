// SPDX-License-Identifier: GPL-3.0-or-later

package koreanfont

import (
	"fmt"
	"sort"
	"strings"

	"github.com/HK47196/zill/internal/zillfont"
)

// Encode serializes source rasters into the deterministic catalog format used
// by release builds. Rendering remains an offline/preparation concern; this
// function makes generator output stable and reviewable in source control.
func Encode(sourceFont, renderRule string, rasters map[rune]zillfont.Raster) ([]byte, error) {
	if sourceFont == "" || renderRule == "" {
		return nil, fmt.Errorf("Korean raster catalog requires source font and render rule")
	}
	if len(rasters) == 0 {
		return nil, fmt.Errorf("Korean raster catalog has no glyphs")
	}
	runes := make([]rune, 0, len(rasters))
	for r, raster := range rasters {
		if raster.Width != zillfont.KoreanRasterWidth || raster.Height != zillfont.KoreanRasterHeight || len(raster.Pixels) != raster.Width*raster.Height {
			return nil, fmt.Errorf("Korean raster %U has invalid dimensions/pixel count", r)
		}
		for i, pixel := range raster.Pixels {
			if pixel > 0x0f {
				return nil, fmt.Errorf("Korean raster %U pixel %d exceeds 4bpp", r, i)
			}
		}
		runes = append(runes, r)
	}
	sort.Slice(runes, func(i, j int) bool { return runes[i] < runes[j] })
	quote := func(s string) string { return fmt.Sprintf("%q", s) }
	var out strings.Builder
	fmt.Fprintf(&out, "format = %s\nversion = %d\nwidth = %d\nheight = %d\nsource_font = %s\nrender_rule = %s\n\n[glyphs]\n", quote(catalogFormat), catalogVersion, zillfont.KoreanRasterWidth, zillfont.KoreanRasterHeight, quote(sourceFont), quote(renderRule))
	const hex = "0123456789abcdef"
	for _, r := range runes {
		raster := rasters[r]
		encoded := make([]byte, len(raster.Pixels))
		for i, pixel := range raster.Pixels { encoded[i] = hex[pixel] }
		fmt.Fprintf(&out, "%s = %s\n", quote(string(r)), quote(string(encoded)))
	}
	return []byte(out.String()), nil
}
