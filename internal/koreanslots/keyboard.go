// SPDX-License-Identifier: GPL-3.0-or-later

package koreanslots

import "github.com/HK47196/zill/internal/cp932"

// KeyboardInputReservedKeys returns the stock renderer keys owned by the retail
// name/input keyboard's fullwidth ASCII page. Device evidence shows that these
// cells are live input choices: when a custom Korean glyph occupies one of these
// slots, the displayed Hangul is also inserted into the name buffer. Therefore
// these are structured renderer/input ownership, not arbitrary whole-blob byte
// aliases, and must never be repurposed by Korean custom-glyph allocation.
//
// The Japanese input UI exposes the ordinary fullwidth ASCII repertoire for its
// alphanumeric/symbol page. Reserve U+FF01..U+FF5E as a complete page-level
// contract rather than chasing only the currently visible collisions.
func KeyboardInputReservedKeys() []cp932.GlyphKey {
	runes := make([]rune, 0, 0x5e)
	for r := rune(0xff01); r <= 0xff5e; r++ {
		runes = append(runes, r)
	}
	return RequiredStockKeys([]string{string(runes)})
}
