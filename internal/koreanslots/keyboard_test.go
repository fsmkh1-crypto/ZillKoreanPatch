// SPDX-License-Identifier: GPL-3.0-or-later

package koreanslots

import (
	"testing"

	"github.com/HK47196/zill/internal/cp932"
)

func TestKeyboardInputReservedKeysCoverVisibleAlphanumericPage(t *testing.T) {
	reserved := KeyboardInputReservedKeys()
	set := make(map[cp932.GlyphKey]struct{}, len(reserved))
	for _, key := range reserved {
		if !key.IsDoubleByte() {
			t.Fatalf("keyboard reservation contains non-double-byte key %#04x", uint16(key))
		}
		set[key] = struct{}{}
	}
	for _, r := range []rune{'Ａ', 'Ｆ', 'Ｑ', 'Ｚ', 'ａ', 'ｚ', '０', '９', '？', '！', '．', '＝'} {
		encoded, err := cp932.Encode(string(r))
		if err != nil { t.Fatalf("encode %q: %v", r, err) }
		key, err := cp932.GlyphKeyFromBytes(encoded)
		if err != nil { t.Fatalf("key %q: %v", r, err) }
		if _, ok := set[key]; !ok { t.Errorf("keyboard reservation missing %q key=%#04x", r, uint16(key)) }
	}
}

func TestKeyboardReservationsCannotBeAllocatedToKoreanGlyph(t *testing.T) {
	keyboard := KeyboardInputReservedKeys()
	if len(keyboard) == 0 { t.Fatal("empty keyboard reservation") }
	installed := append([]cp932.GlyphKey(nil), keyboard...)
	installed = append(installed, cp932.GlyphKey(0x9f88))
	plan, err := BuildPlan([]string{"한"}, installed, keyboard)
	if err != nil { t.Fatalf("BuildPlan: %v", err) }
	for _, key := range keyboard {
		if got := plan.Mapping['한']; got == key { t.Fatalf("Hangul allocated to protected keyboard key %#04x", uint16(key)) }
	}
}
