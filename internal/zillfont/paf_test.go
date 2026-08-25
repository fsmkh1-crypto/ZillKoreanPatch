// SPDX-License-Identifier: GPL-3.0-or-later

package zillfont

import (
	"encoding/binary"
	"testing"
)

func TestParsePAFPreservesRetailFinalTailAnomaly(t *testing.T) {
	data := make([]byte, PAFSize)
	copy(data[:4], []byte{'p', 'a', 'f', 0})
	binary.LittleEndian.PutUint32(data[4:8], ExpectedVer)
	for index := 0; index < GlyphCount; index++ {
		offset := RecordOffset + index*RecordStride
		binary.LittleEndian.PutUint16(data[offset:offset+2], uint16(index+1))
		data[offset+2] = 8
		data[offset+3] = 12
		binary.LittleEndian.PutUint32(data[offset+12:offset+16], 9)
	}

	font, err := ParsePAF(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(font.Glyphs) != GlyphCount {
		t.Fatalf("glyph count = %d, want %d", len(font.Glyphs), GlyphCount)
	}
	if !font.Glyphs[GlyphCount-2].HasTail {
		t.Fatal("penultimate glyph unexpectedly lacks its tail")
	}
	if font.Glyphs[GlyphCount-1].HasTail {
		t.Fatal("final retail-shape glyph unexpectedly has a tail")
	}
	pages := font.PageCounts()
	if pages[0] != GlyphCount-1 {
		t.Fatalf("complete page-zero records = %d, want %d", pages[0], GlyphCount-1)
	}
}

func TestParsePAFRejectsNonAscendingRendererKeys(t *testing.T) {
	data := make([]byte, PAFSize)
	copy(data[:4], []byte{'p', 'a', 'f', 0})
	binary.LittleEndian.PutUint32(data[4:8], ExpectedVer)
	for index := 0; index < GlyphCount; index++ {
		offset := RecordOffset + index*RecordStride
		binary.LittleEndian.PutUint16(data[offset:offset+2], uint16(index+1))
	}
	binary.LittleEndian.PutUint16(data[RecordOffset+10*RecordStride:], 10)
	if _, err := ParsePAF(data); err == nil {
		t.Fatal("expected non-ascending renderer keys to fail")
	}
}
