// SPDX-License-Identifier: GPL-3.0-or-later

package koreanslots

import (
	"testing"

	"github.com/HK47196/zill/internal/cp932"
)

func TestRequiredStockKeysUsesOnlyEncodableRuntimeRunes(t *testing.T) {
	got := RequiredStockKeys([]string{"日A한", "日B"})
	wantText := []string{"日", "A", "B"}
	want := make(map[cp932.GlyphKey]struct{})
	for _, text := range wantText {
		encoded, err := cp932.Encode(text)
		if err != nil {
			t.Fatal(err)
		}
		key, err := cp932.GlyphKeyFromBytes(encoded)
		if err != nil {
			t.Fatal(err)
		}
		want[key] = struct{}{}
	}
	if len(got) != len(want) {
		t.Fatalf("keys = %#v; want %d unique keys", got, len(want))
	}
	for _, key := range got {
		if _, ok := want[key]; !ok {
			t.Fatalf("unexpected key 0x%04X", uint16(key))
		}
	}
}

func TestRequiredStockKeysIsSortedAndUnique(t *testing.T) {
	got := RequiredStockKeys([]string{"日日AA"})
	if len(got) != 2 {
		t.Fatalf("keys = %#v", got)
	}
	if got[0] >= got[1] {
		t.Fatalf("keys not strictly sorted: %#v", got)
	}
}
