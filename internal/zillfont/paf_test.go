// SPDX-License-Identifier: GPL-3.0-or-later

package zillfont

import (
	"encoding/binary"
	"testing"
)

func makeRetailShapePAF() []byte {
	data := make([]byte, PAFSize)
	copy(data[:4], []byte{'p', 'a', 'f', 0})
	binary.LittleEndian.PutUint32(data[4:8], ExpectedVer)
	binary.LittleEndian.PutUint32(data[8:12], GlyphCount)
	binary.LittleEndian.PutUint32(data[12:16], BSTRoot)
	for index := 0; index < GlyphCount; index++ {
		offset := RecordOffset + index*RecordStride
		binary.LittleEndian.PutUint16(data[offset:offset+2], uint16(index+1))
		data[offset+2] = 8
		data[offset+3] = 12
		binary.LittleEndian.PutUint32(data[offset+12:offset+16], 9)
		binary.LittleEndian.PutUint32(data[offset+0x18:offset+0x1c], 2)
	}
	return data
}

func TestParsePAFConsumesAllRetailRecords(t *testing.T) {
	font, err := ParsePAF(makeRetailShapePAF())
	if err != nil {
		t.Fatal(err)
	}
	if len(font.Glyphs) != GlyphCount {
		t.Fatalf("glyph count = %d, want %d", len(font.Glyphs), GlyphCount)
	}
	pages := font.PageCounts()
	if pages[2] != GlyphCount {
		t.Fatalf("page-two records = %d, want %d", pages[2], GlyphCount)
	}
}

func TestParsePAFRejectsNonAscendingRendererKeys(t *testing.T) {
	data := makeRetailShapePAF()
	binary.LittleEndian.PutUint16(data[RecordOffset+10*RecordStride:], 10)
	if _, err := ParsePAF(data); err == nil {
		t.Fatal("expected non-ascending renderer keys to fail")
	}
}

func TestParsePAFRejectsWrongHeaderCount(t *testing.T) {
	data := makeRetailShapePAF()
	binary.LittleEndian.PutUint32(data[8:12], GlyphCount-1)
	if _, err := ParsePAF(data); err == nil {
		t.Fatal("expected wrong header glyph count to fail")
	}
}
