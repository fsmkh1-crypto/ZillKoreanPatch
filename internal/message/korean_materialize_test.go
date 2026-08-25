// SPDX-License-Identifier: GPL-3.0-or-later

package message_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/message"
)

func TestMaterializeKoreanUsesCustomRendererBytes(t *testing.T) {
	record := corpus.Record{ID: 42, Tokens: []corpus.Token{
		{Kind: "text", Raw: []byte("source")},
		{Kind: "block_terminator", Raw: []byte{5, 5, 5}},
	}}
	projection, err := message.Project(record)
	if err != nil {
		t.Fatal(err)
	}
	mapping := koreanslots.Mapping{'가': cp932.GlyphKey(0xAC82)}
	got, err := projection.MaterializeKorean("가<end>", false, mapping)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0x82, 0xAC, 5, 5, 5}
	if !bytes.Equal(got, want) {
		t.Fatalf("materialized Korean bytes = % X, want % X", got, want)
	}
}

func TestMaterializeKoreanPreservesRuntimeSubstitution(t *testing.T) {
	record := corpus.Record{ID: 43, Tokens: []corpus.Token{
		{Kind: "text", Raw: []byte("name=")},
		{Kind: "substitution", Raw: []byte{2, 0x28}},
		{Kind: "text", Raw: []byte("!")},
		{Kind: "block_terminator", Raw: []byte{5, 5, 5}},
	}}
	projection, err := message.Project(record)
	if err != nil {
		t.Fatal(err)
	}
	mapping := koreanslots.Mapping{
		'가': cp932.GlyphKey(0xAC82),
		'나': cp932.GlyphKey(0xAD82),
	}
	got, err := projection.MaterializeKorean("가<value:$28>나<end>", false, mapping)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0x82, 0xAC, 2, 0x28, 0x82, 0xAD, 5, 5, 5}
	if !bytes.Equal(got, want) {
		t.Fatalf("materialized Korean substitution bytes = % X, want % X", got, want)
	}
}

func TestMaterializeKoreanFailsClosedWithoutMapping(t *testing.T) {
	record := corpus.Record{ID: 44, Tokens: []corpus.Token{
		{Kind: "text", Raw: []byte("source")},
		{Kind: "block_terminator", Raw: []byte{5, 5, 5}},
	}}
	projection, err := message.Project(record)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := projection.MaterializeKorean("가<end>", false, nil); err == nil || !strings.Contains(err.Error(), "Korean renderer slots") {
		t.Fatalf("missing custom mapping returned %v", err)
	}
}

func TestMaterializeKoreanRejectsMappingThatOverridesStockCP932(t *testing.T) {
	record := corpus.Record{ID: 45, Tokens: []corpus.Token{
		{Kind: "text", Raw: []byte("source")},
		{Kind: "block_terminator", Raw: []byte{5, 5, 5}},
	}}
	projection, err := message.Project(record)
	if err != nil {
		t.Fatal(err)
	}
	mapping := koreanslots.Mapping{'A': cp932.GlyphKey(0xAC82)}
	if _, err := projection.MaterializeKorean("A<end>", false, mapping); err == nil || !strings.Contains(err.Error(), "stock CP932 rune") {
		t.Fatalf("stock CP932 override returned %v", err)
	}
}
