// SPDX-License-Identifier: GPL-3.0-or-later

// Package zillfont parses the retail Zill O'll Infinite Plus font metadata
// needed by the Korean slot-reuse PoC. It intentionally models the known
// ULJM-05410 v1.03 PAF shape rather than pretending the format is generic.
package zillfont

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/HK47196/zill/internal/cp932"
)

const (
	PAFSize       = 0x149c0
	RecordOffset  = 0x30
	RecordStride  = 0x20
	GlyphCount    = 2637
	BSTRoot       = 1318
	LastCoreSize  = 0x10
	ExpectedVer   = 0x000c0201
)

// Glyph is the part of one PAF record needed for deterministic slot reuse.
// HasTail is false for the retail file's final record, whose final 0x10 bytes
// are omitted at EOF.
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
	HasTail  bool
}

// PAF is the validated retail font metadata view.
type PAF struct {
	Version uint32
	Glyphs  []Glyph
}

// ParsePAF parses the exact retail PAF shape confirmed for ULJM-05410 v1.03.
// The parser deliberately preserves the final truncated-record anomaly.
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

	glyphs := make([]Glyph, 0, GlyphCount)
	var previous cp932.GlyphKey
	for index := 0; index < GlyphCount; index++ {
		offset := RecordOffset + index*RecordStride
		coreEnd := offset + LastCoreSize
		if coreEnd > len(data) {
			return nil, fmt.Errorf("glyph %d core extends past PAF", index)
		}
		key := cp932.GlyphKey(binary.LittleEndian.Uint16(data[offset : offset+2]))
		if index > 0 && key <= previous {
			return nil, fmt.Errorf("glyph keys are not strictly ascending at %d: %#04x <= %#04x", index, uint16(key), uint16(previous))
		}
		previous = key
		glyph := Glyph{
			Index:    index,
			Key:      key,
			Width:    data[offset+2],
			Height:   data[offset+3],
			X:        binary.LittleEndian.Uint16(data[offset+4 : offset+6]),
			Y:        binary.LittleEndian.Uint16(data[offset+6 : offset+8]),
			BearingX: int16(binary.LittleEndian.Uint16(data[offset+8 : offset+10])),
			BearingY: int16(binary.LittleEndian.Uint16(data[offset+10 : offset+12])),
			Advance:  binary.LittleEndian.Uint32(data[offset+12 : offset+16]),
		}
		tailEnd := offset + RecordStride
		if tailEnd <= len(data) {
			glyph.Left = int32(binary.LittleEndian.Uint32(data[offset+0x10 : offset+0x14]))
			glyph.Right = int32(binary.LittleEndian.Uint32(data[offset+0x14 : offset+0x18]))
			glyph.Page = binary.LittleEndian.Uint32(data[offset+0x18 : offset+0x1c])
			if binary.LittleEndian.Uint32(data[offset+0x1c:tailEnd]) != 0 {
				return nil, fmt.Errorf("glyph %d reserved tail is nonzero", index)
			}
			glyph.HasTail = true
		} else if index != GlyphCount-1 || coreEnd != len(data) {
			return nil, fmt.Errorf("unexpected truncated PAF record %d", index)
		}
		glyphs = append(glyphs, glyph)
	}
	if glyphs[len(glyphs)-1].HasTail {
		return nil, fmt.Errorf("expected final PAF glyph tail to be omitted")
	}
	return &PAF{Version: version, Glyphs: glyphs}, nil
}

// DoubleByteKeys returns the installed valid two-byte CP932 renderer keys in
// PAF order. These are renderer slots, not Unicode code points.
func (p *PAF) DoubleByteKeys() []cp932.GlyphKey {
	out := make([]cp932.GlyphKey, 0, len(p.Glyphs))
	for _, glyph := range p.Glyphs {
		if glyph.Key.IsDoubleByte() {
			out = append(out, glyph.Key)
		}
	}
	return out
}

// PageCounts reports only complete PAF records because the retail final glyph
// omits the tail containing its page field.
func (p *PAF) PageCounts() map[uint32]int {
	out := make(map[uint32]int)
	for _, glyph := range p.Glyphs {
		if glyph.HasTail {
			out[glyph.Page]++
		}
	}
	return out
}
