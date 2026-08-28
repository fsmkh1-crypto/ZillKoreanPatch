// SPDX-License-Identifier: GPL-3.0-or-later

package cp932

import (
	"bytes"
	"fmt"
)

// GlyphKey is the renderer lookup key stored in zillfont.paf.
//
// It is NOT a Unicode code point. For a two-byte Shift-JIS sequence 82 AC,
// the PAF key is the little-endian integer 0xAC82.
type GlyphKey uint16

// GlyphKeyFromBytes converts exactly one CP932 character encoding (one or two
// bytes) to the renderer's little-endian lookup key.
func GlyphKeyFromBytes(encoded []byte) (GlyphKey, error) {
	switch len(encoded) {
	case 1:
		if !validSingle(encoded[0]) {
			return 0, fmt.Errorf("invalid CP932 single byte 0x%02X", encoded[0])
		}
		return GlyphKey(encoded[0]), nil
	case 2:
		if !validLead(encoded[0]) || !validTrail(encoded[1]) {
			return 0, fmt.Errorf("invalid CP932 double byte %02X %02X", encoded[0], encoded[1])
		}
		return GlyphKey(uint16(encoded[0]) | uint16(encoded[1])<<8), nil
	default:
		return 0, fmt.Errorf("glyph key requires one CP932 character, got %d bytes", len(encoded))
	}
}

// Bytes returns the byte sequence that the game renderer uses to look up k.
func (k GlyphKey) Bytes() ([]byte, error) {
	lo, hi := byte(k), byte(uint16(k)>>8)
	if hi == 0 {
		if !validSingle(lo) {
			return nil, fmt.Errorf("invalid one-byte glyph key 0x%04X", uint16(k))
		}
		return []byte{lo}, nil
	}
	if !validLead(lo) || !validTrail(hi) {
		return nil, fmt.Errorf("invalid two-byte glyph key 0x%04X", uint16(k))
	}
	return []byte{lo, hi}, nil
}

// IsDoubleByte reports whether k has the byte shape of a two-byte CP932 key.
// This is intentionally only a structural check: the retail PAF also contains
// renderer-private/UI keys that use syntactically valid lead/trail bytes.
func (k GlyphKey) IsDoubleByte() bool {
	lo, hi := byte(k), byte(uint16(k)>>8)
	return hi != 0 && validLead(lo) && validTrail(hi)
}

// IsRoundTripText reports whether k is an actual two-byte text character rather
// than merely a renderer key with CP932-shaped bytes. A production Korean text
// slot must decode to exactly one Unicode rune and encode back to the identical
// byte pair. This excludes PAF-private/UI keys (for example button/icon entries)
// from being repurposed as Hangul message bytes.
func (k GlyphKey) IsRoundTripText() bool {
	if !k.IsDoubleByte() {
		return false
	}
	encoded, err := k.Bytes()
	if err != nil {
		return false
	}
	decoded, err := Decode(encoded)
	if err != nil || len([]rune(decoded)) != 1 {
		return false
	}
	roundTrip, err := Encode(decoded)
	return err == nil && bytes.Equal(roundTrip, encoded)
}

func validSingle(b byte) bool {
	return b <= 0x7F || b >= 0xA1 && b <= 0xDF
}

func validLead(b byte) bool {
	return b >= 0x81 && b <= 0x9F || b >= 0xE0 && b <= 0xFC
}

func validTrail(b byte) bool {
	return b >= 0x40 && b <= 0x7E || b >= 0x80 && b <= 0xFC
}
