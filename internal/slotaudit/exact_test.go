// SPDX-License-Identifier: GPL-3.0-or-later

package slotaudit

import (
	"testing"

	"github.com/HK47196/zill/internal/cp932"
)

func TestExcludeExactByteReferencesRejectsIsolatedPair(t *testing.T) {
	keys := []cp932.GlyphKey{0xA1E1, 0xA1E9, 0xB8E2}
	blob := []byte{0x00, 0xE9, 0xA1, 0x00}
	got, err := ExcludeExactByteReferences(keys, blob)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != 0xA1E1 || got[1] != 0xB8E2 {
		t.Fatalf("filtered keys = %#v", got)
	}
}

func TestExcludeExactByteReferencesChecksEveryBlob(t *testing.T) {
	keys := []cp932.GlyphKey{0xA1E1, 0xA1E9}
	first := []byte{0x11, 0x22}
	second := []byte{0xE1, 0xA1}
	got, err := ExcludeExactByteReferences(keys, first, second)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != 0xA1E9 {
		t.Fatalf("filtered keys = %#v", got)
	}
}

func TestExcludeExactByteReferencesRejectsInvalidKey(t *testing.T) {
	_, err := ExcludeExactByteReferences([]cp932.GlyphKey{0x0101}, []byte("anything"))
	if err == nil {
		t.Fatal("expected invalid key error")
	}
}
