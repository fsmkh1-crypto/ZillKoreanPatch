// SPDX-License-Identifier: GPL-3.0-or-later

package cp932

import "testing"

func TestGlyphKeyLittleEndian(t *testing.T) {
	key, err := GlyphKeyFromBytes([]byte{0x82, 0xAC})
	if err != nil {
		t.Fatal(err)
	}
	if key != GlyphKey(0xAC82) {
		t.Fatalf("got 0x%04X, want 0xAC82", uint16(key))
	}
	got, err := key.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != 0x82 || got[1] != 0xAC {
		t.Fatalf("round trip = % X, want 82 AC", got)
	}
}

func TestGlyphKeyRejectsInvalidTrail(t *testing.T) {
	if _, err := GlyphKeyFromBytes([]byte{0x82, 0x7F}); err == nil {
		t.Fatal("expected invalid Shift-JIS trail byte to fail")
	}
}

func TestGlyphKeyRoundTripTextRejectsPrivateShapedKey(t *testing.T) {
	if !GlyphKey(0xAC82).IsRoundTripText() { // bytes 82 AC: ordinary Shift-JIS text
		t.Fatal("expected 82 AC to be accepted as round-trip text")
	}
	if GlyphKey(0xAD81).IsRoundTripText() { // bytes 81 AD: valid byte shape, undefined as text
		t.Fatal("expected renderer-private/undefined 81 AD key to be rejected")
	}
	if GlyphKey(0x0041).IsRoundTripText() {
		t.Fatal("single-byte key must not be accepted as a reusable two-byte text slot")
	}
}
