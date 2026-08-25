// SPDX-License-Identifier: GPL-3.0-or-later

// Package slotaudit finds renderer-key references outside parsed message banks.
// It intentionally prefers false positives over false negatives, but does not
// treat arbitrary binary byte-pair occurrence as text: compressed/image data
// would otherwise eliminate nearly every possible two-byte key by chance.
package slotaudit

import (
	"fmt"

	"github.com/HK47196/zill/internal/cp932"
)

// Literal is one NUL-terminated CP932-looking string recovered from binary data.
type Literal struct {
	Offset int
	Bytes  []byte
	Text   string
}

// Report summarizes recovered literals and the two-byte renderer keys they use.
type Report struct {
	Literals []Literal
	Keys     map[cp932.GlyphKey]struct{}
}

// ScanCP932Literals scans for the final printable CP932 run before each NUL.
// A run is accepted only if it has either at least two double-byte glyphs, or
// one double-byte glyph plus at least two printable ASCII bytes. This avoids
// treating isolated random Shift-JIS-looking pairs in executable/binary data as
// strong evidence while still recovering ordinary mixed Japanese UI strings.
func ScanCP932Literals(data []byte) (Report, error) {
	report := Report{Keys: make(map[cp932.GlyphKey]struct{})}
	start := 0
	doubleByte := 0
	ascii := 0

	reset := func(next int) {
		start = next
		doubleByte = 0
		ascii = 0
	}
	accept := func(end int) error {
		if end <= start || !(doubleByte >= 2 || doubleByte >= 1 && ascii >= 2) {
			return nil
		}
		candidate := append([]byte(nil), data[start:end]...)
		text, err := cp932.Decode(candidate)
		if err != nil {
			return nil // binary lookalike; not a trustworthy literal
		}
		for index := 0; index < len(candidate); index++ {
			b := candidate[index]
			if isLead(b) {
				if index+1 >= len(candidate) || !isTrail(candidate[index+1]) {
					return fmt.Errorf("internal scanner accepted incomplete CP932 pair at %#x", start+index)
				}
				key, err := cp932.GlyphKeyFromBytes(candidate[index : index+2])
				if err != nil {
					return fmt.Errorf("internal scanner key at %#x: %w", start+index, err)
				}
				report.Keys[key] = struct{}{}
				index++
			}
		}
		report.Literals = append(report.Literals, Literal{Offset: start, Bytes: candidate, Text: text})
		return nil
	}

	for index := 0; index < len(data); {
		b := data[index]
		switch {
		case b == 0:
			if err := accept(index); err != nil {
				return Report{}, err
			}
			index++
			reset(index)
		case isPrintableASCII(b):
			ascii++
			index++
		case b == '\t' || b == '\n' || b == '\r':
			index++
		case b >= 0xA1 && b <= 0xDF: // half-width kana; not a reusable two-byte key
			index++
		case isLead(b) && index+1 < len(data) && isTrail(data[index+1]):
			doubleByte++
			index += 2
		default:
			index++
			reset(index)
		}
	}
	return report, nil
}

func isPrintableASCII(b byte) bool { return b >= 0x20 && b <= 0x7E }
func isLead(b byte) bool           { return b >= 0x81 && b <= 0x9F || b >= 0xE0 && b <= 0xFC }
func isTrail(b byte) bool          { return b >= 0x40 && b <= 0x7E || b >= 0x80 && b <= 0xFC }
