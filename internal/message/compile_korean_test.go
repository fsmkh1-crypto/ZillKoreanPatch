// SPDX-License-Identifier: GPL-3.0-or-later

package message_test

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/message"
)

func TestCompileBankKoreanReplacesOnlySelectedRecord(t *testing.T) {
	first := corpus.Record{ID: 10000, Raw: []byte("old\x00"), Tokens: []corpus.Token{
		{Kind: "text", Raw: []byte("old")}, {Kind: "suffix", Raw: []byte{0}},
	}}
	second := corpus.Record{ID: 10001, Raw: []byte("stay\x00"), Tokens: []corpus.Token{
		{Kind: "text", Raw: []byte("stay")}, {Kind: "suffix", Raw: []byte{0}},
	}}
	bank := corpus.Bank{Name: "msgsec001.dat", Section: 1, Records: []corpus.Record{first, second}}
	items := []corpus.Item{
		{Record: first, Translation: corpus.Translation{ID: first.ID, State: corpus.Todo}},
		{Record: second, Translation: corpus.Translation{ID: second.ID, State: corpus.Todo}},
	}
	mapping := koreanslots.Mapping{'가': cp932.GlyphKey(0xAC82)}
	compiled, err := message.CompileBankKorean(bank, items, map[int]message.KoreanRecord{
		first.ID: {Text: "가"},
	}, mapping)
	if err != nil {
		t.Fatal(err)
	}
	if count := binary.LittleEndian.Uint16(compiled); count != 2 {
		t.Fatalf("record count = %d, want 2", count)
	}
	firstOffset := binary.LittleEndian.Uint32(compiled[4:])
	secondOffset := binary.LittleEndian.Uint32(compiled[8:])
	if got, want := compiled[firstOffset:secondOffset], []byte{0x82, 0xAC, 0}; !bytes.Equal(got, want) {
		t.Fatalf("Korean record = % X, want % X", got, want)
	}
	if got := compiled[secondOffset:]; !bytes.Equal(got, second.Raw) {
		t.Fatalf("unselected retail record changed: % X", got)
	}
}

func TestCompileBankKoreanFailsWhenSelectedTextLacksMapping(t *testing.T) {
	source := corpus.Record{ID: 10000, Raw: []byte("old\x00"), Tokens: []corpus.Token{
		{Kind: "text", Raw: []byte("old")}, {Kind: "suffix", Raw: []byte{0}},
	}}
	bank := corpus.Bank{Name: "msgsec001.dat", Section: 1, Records: []corpus.Record{source}}
	items := []corpus.Item{{Record: source, Translation: corpus.Translation{ID: source.ID, State: corpus.Todo}}}
	_, err := message.CompileBankKorean(bank, items, map[int]message.KoreanRecord{
		source.ID: {Text: "가"},
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "Korean renderer slots") {
		t.Fatalf("missing mapping returned %v", err)
	}
}

func TestCompileBankKoreanAggregatesRecordFailures(t *testing.T) {
	first := corpus.Record{ID: 10000, Raw: []byte("old\x00"), Tokens: []corpus.Token{
		{Kind: "text", Raw: []byte("old")}, {Kind: "suffix", Raw: []byte{0}},
	}}
	second := corpus.Record{ID: 10001, Raw: []byte("old\x00"), Tokens: []corpus.Token{
		{Kind: "text", Raw: []byte("old")}, {Kind: "suffix", Raw: []byte{0}},
	}}
	bank := corpus.Bank{Name: "msgsec001.dat", Section: 1, Records: []corpus.Record{first, second}}
	items := []corpus.Item{
		{Record: first, Translation: corpus.Translation{ID: first.ID, State: corpus.Todo}},
		{Record: second, Translation: corpus.Translation{ID: second.ID, State: corpus.Todo}},
	}
	_, err := message.CompileBankKorean(bank, items, map[int]message.KoreanRecord{
		first.ID:  {Text: "가"},
		second.ID: {Text: "나"},
	}, nil)
	if err == nil {
		t.Fatal("expected aggregated materialization failure")
	}
	text := err.Error()
	for _, want := range []string{"Korean materialization failed", "ID 10000", "ID 10001"} {
		if !strings.Contains(text, want) {
			t.Fatalf("error %q does not contain %q", text, want)
		}
	}
}

func TestCompileBankKoreanRejectsUnmatchedReplacementID(t *testing.T) {
	source := corpus.Record{ID: 10000, Raw: []byte("old\x00"), Tokens: []corpus.Token{
		{Kind: "text", Raw: []byte("old")}, {Kind: "suffix", Raw: []byte{0}},
	}}
	bank := corpus.Bank{Name: "msgsec001.dat", Section: 1, Records: []corpus.Record{source}}
	items := []corpus.Item{{Record: source, Translation: corpus.Translation{ID: source.ID, State: corpus.Todo}}}
	mapping := koreanslots.Mapping{'가': cp932.GlyphKey(0xAC82)}

	_, err := message.CompileBankKorean(bank, items, map[int]message.KoreanRecord{
		99999: {Text: "가"},
	}, mapping)
	if err == nil || !strings.Contains(err.Error(), "99999") || !strings.Contains(err.Error(), "not present in this bank") {
		t.Fatalf("unmatched replacement ID returned %v", err)
	}
}

func TestCompileBankKoreanRejectsLineBreakInSemanticText(t *testing.T) {
	source := corpus.Record{ID: 10000, Raw: []byte("old\x00"), Tokens: []corpus.Token{
		{Kind: "text", Raw: []byte("old")}, {Kind: "suffix", Raw: []byte{0}},
	}}
	bank := corpus.Bank{Name: "msgsec001.dat", Section: 1, Records: []corpus.Record{source}}
	items := []corpus.Item{{Record: source, Translation: corpus.Translation{ID: source.ID, State: corpus.Todo}}}
	mapping := koreanslots.Mapping{'가': cp932.GlyphKey(0xAC82)}
	_, err := message.CompileBankKorean(bank, items, map[int]message.KoreanRecord{
		source.ID: {Text: "가<line-break>가"},
	}, mapping)
	if err == nil || !strings.Contains(err.Error(), "layout break") {
		t.Fatalf("semantic line break returned %v", err)
	}
}

func TestCompileBankKoreanAllowsGeneratedLayoutBreak(t *testing.T) {
	source := corpus.Record{ID: 10000, Raw: []byte("old\x00"), Tokens: []corpus.Token{
		{Kind: "text", Raw: []byte("old")}, {Kind: "suffix", Raw: []byte{0}},
	}}
	bank := corpus.Bank{Name: "msgsec001.dat", Section: 1, Records: []corpus.Record{source}}
	items := []corpus.Item{{Record: source, Translation: corpus.Translation{ID: source.ID, State: corpus.Todo}}}
	mapping := koreanslots.Mapping{'가': cp932.GlyphKey(0xAC82)}
	compiled, err := message.CompileBankKorean(bank, items, map[int]message.KoreanRecord{
		source.ID: {Text: "가 가", Layout: "가<line-break>가"},
	}, mapping)
	if err != nil {
		t.Fatal(err)
	}
	offset := binary.LittleEndian.Uint32(compiled[4:])
	if got, want := compiled[offset:], []byte{0x82, 0xAC, 0x0A, 0x82, 0xAC, 0}; !bytes.Equal(got, want) {
		t.Fatalf("layout record = % X, want % X", got, want)
	}
}
