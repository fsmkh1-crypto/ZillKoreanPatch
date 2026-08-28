// SPDX-License-Identifier: GPL-3.0-or-later

// Package zillfont parses the retail Zill O'll Infinite Plus font metadata
// needed by the Korean slot-reuse PoC. It intentionally models the confirmed
// ULJM-05410 v1.03 PAF shape rather than pretending the format is generic.
package zillfont

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/HK47196/zill/internal/cp932"
)

const (
	PAFSize      = 0x149d0
	RecordOffset = 0x30
	RecordStride = 0x20
	GlyphCount   = 2637
	BSTRoot      = 1318
	ExpectedVer  = 0x000c0201
)

// Glyph is the part of one PAF record needed for deterministic slot reuse.
type Glyph struct {
	Index    int
	Key      cp932.GlyphKey
	Width    uint8
	Height   uint8
	X        uint16
	Y        uint16
	BearingX int16
	BearingY int16
	Advance  uint32
	Left     int32
	Right    int32
	Page     uint32
}

// PAF is the validated retail font metadata view.
type PAF struct {
	Version uint32
	Glyphs  []Glyph
}

// ParsePAF parses the exact retail PAF shape confirmed from the authenticated
// ULJM-05410 v1.03 jillbtn.par member.
func ParsePAF(data []byte) (*PAF, error) {
	if len(data) != PAFSize {
		return nil, fmt.Errorf("PAF size %#x, want %#x", len(data), PAFSize)
	}
	if !bytes.Equal(data[:4], []byte{'p', 'a', 'f', 0}) {
		return nil, fmt.Errorf("invalid PAF magic % X", data[:4])
	}
	version := binary.LittleEndian.Uint32(data[4:8])
	if version != ExpectedVer {
		return nil, fmt.Errorf("PAF version %#08x, want %#08x", version, ExpectedVer)
	}
	count := binary.LittleEndian.Uint32(data[8:12])
	if count != GlyphCount {
		return nil, fmt.Errorf("PAF glyph count %d, want %d", count, GlyphCount)
	}
	root := binary.LittleEndian.Uint32(data[12:16])
	if root != BSTRoot {
		return nil, fmt.Errorf("PAF BST root %d, want %d", root, BSTRoot)
	}

	glyphs := make([]Glyph, 0, GlyphCount)
	var previous cp932.GlyphKey
	for index := 0; index < GlyphCount; index++ {
		offset := RecordOffset + index*RecordStride
		tailEnd := offset + RecordStride
		if tailEnd > len(data) {
			return nil, fmt.Errorf("glyph %d extends past PAF", index)
		}
		key := cp932.GlyphKey(binary.LittleEndian.Uint16(data[offset : offset+2]))
		if index > 0 && key <= previous {
			return nil, fmt.Errorf("glyph keys are not strictly ascending at %d: %#04x <= %#04x", index, uint16(key), uint16(previous))
		}
		previous = key
		if binary.LittleEndian.Uint32(data[offset+0x1c:tailEnd]) != 0 {
			return nil, fmt.Errorf("glyph %d reserved tail is nonzero", index)
		}
		glyphs = append(glyphs, Glyph{
			Index:    index,
			Key:      key,
			Width:    data[offset+2],
			Height:   data[offset+3],
			X:        binary.LittleEndian.Uint16(data[offset+4 : offset+6]),
			Y:        binary.LittleEndian.Uint16(data[offset+6 : offset+8]),
			BearingX: int16(binary.LittleEndian.Uint16(data[offset+8 : offset+10])),
			BearingY: int16(binary.LittleEndian.Uint16(data[offset+10 : offset+12])),
			Advance:  binary.LittleEndian.Uint32(data[offset+12 : offset+16]),
			Left:     int32(binary.LittleEndian.Uint32(data[offset+0x10 : offset+0x14])),
			Right:    int32(binary.LittleEndian.Uint32(data[offset+0x14 : offset+0x18])),
			Page:     binary.LittleEndian.Uint32(data[offset+0x18 : offset+0x1c]),
		})
	}
	if RecordOffset+GlyphCount*RecordStride != len(data) {
		return nil, fmt.Errorf("PAF record table does not end at EOF")
	}
	return &PAF{Version: version, Glyphs: glyphs}, nil
}

// DoubleByteKeys returns every installed renderer key whose bytes have the
// structural shape of a two-byte CP932 sequence. This intentionally includes
// renderer-private/UI keys and therefore must not by itself define safe Korean
// message slots.
func (p *PAF) DoubleByteKeys() []cp932.GlyphKey {
	out := make([]cp932.GlyphKey, 0, len(p.Glyphs))
	for _, glyph := range p.Glyphs {
		if glyph.Key.IsDoubleByte() {
			out = append(out, glyph.Key)
		}
	}
	return out
}

// TextDoubleByteKeys returns installed two-byte keys that are real CP932 text:
// the byte pair decodes to exactly one rune and re-encodes byte-identically.
// Retail PAF-private entries used for button/icon rendering are excluded even
// when their raw bytes happen to look like a legal Shift-JIS lead/trail pair.
func (p *PAF) TextDoubleByteKeys() []cp932.GlyphKey {
	out := make([]cp932.GlyphKey, 0, len(p.Glyphs))
	for _, glyph := range p.Glyphs {
		if glyph.Key.IsRoundTripText() {
			out = append(out, glyph.Key)
		}
	}
	return out
}

// PageCounts reports all PAF records by GIM page.
func (p *PAF) PageCounts() map[uint32]int {
	out := make(map[uint32]int)
	for _, glyph := range p.Glyphs {
		out[glyph.Page]++
	}
	return out
}
