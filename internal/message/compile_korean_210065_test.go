// SPDX-License-Identifier: GPL-3.0-or-later

package message_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/message"
)

func TestCompileBankKoreanMaterializes210065EightLineDiagnostic(t *testing.T) {
	const semantic = "광대한 대지 바이아시온 대륙. 너무나 넓어 지도에도 기록되지 않고 여행자에게조차 알려지지 않은 작은 마을이 있다…. 마을의 이름은 미이스. 그곳에는 작은 신전과 숲, 그리고 평온한 일상 정도뿐이었다. 위대한 혼의 이야기는 여기서 시작된다…….<end>"

	source := corpus.Record{ID: 210065, Raw: append(append([]byte("old"), 5, 5, 5), 0), Tokens: []corpus.Token{
		{Kind: "text", Raw: []byte("old")},
		{Kind: "block_terminator", Raw: []byte{5, 5, 5}},
		{Kind: "suffix", Raw: []byte{0}},
	}}
	bank := corpus.Bank{Name: "msgsec021.dat", Section: 21, Records: []corpus.Record{source}}
	items := []corpus.Item{{Record: source, Translation: corpus.Translation{ID: source.ID, State: corpus.Todo}}}

	mapping := make(koreanslots.Mapping)
	trail := uint16(0x40)
	for _, r := range semantic {
		if r < '가' || r > '힣' {
			continue
		}
		if _, ok := mapping[r]; ok {
			continue
		}
		if trail == 0x7F {
			trail++
		}
		mapping[r] = cp932.GlyphKey(trail<<8 | 0x82)
		trail++
	}

	compiled, err := message.CompileBankKorean(bank, items, map[int]message.KoreanRecord{
		source.ID: {Text: semantic},
	}, mapping)
	if err != nil {
		t.Fatal(err)
	}
	offset := binary.LittleEndian.Uint32(compiled[4:])
	record := compiled[offset:]
	if len(record) == 0 || record[len(record)-1] != 0 {
		t.Fatalf("210065 materialized record is not NUL terminated: % X", record)
	}
	if len(record) < 4 || !bytes.Equal(record[len(record)-4:len(record)-1], []byte{5, 5, 5}) {
		t.Fatalf("210065 materialized record lost fixed <end> terminator: % X", record)
	}
	payload := record[:len(record)-4]
	if got := bytes.Count(payload, []byte{0x0A}); got != 7 {
		t.Fatalf("210065 line-break count = %d, want 7; payload=% X", got, payload)
	}
	lines := bytes.Split(payload, []byte{0x0A})
	if got := len(lines); got != 8 {
		t.Fatalf("210065 materialized lines = %d, want 8", got)
	}
	for i, line := range lines {
		if len(line) > 56 {
			t.Fatalf("210065 line %d encoded length = %d, exceeds C22 56-byte line contract", i+1, len(line))
		}
	}
	if len(record) >= 0x100 {
		t.Fatalf("210065 diagnostic record length = %d, expected < 0x100 bytes", len(record))
	}
}
