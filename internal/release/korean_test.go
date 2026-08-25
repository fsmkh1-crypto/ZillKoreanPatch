// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"encoding/binary"
	"testing"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
)

func TestCompileKoreanBanksUsesCanonicalOverlayAndPreservesUnselectedRetail(t *testing.T) {
	first := corpus.Record{
		ID: 10000, Index: 0, Raw: []byte{'o', 'l', 'd', 0},
		Tokens: []corpus.Token{{Kind: "text", Raw: []byte("old"), Text: "old"}, {Kind: "suffix", Raw: []byte{0}}},
	}
	second := corpus.Record{
		ID: 10001, Index: 1, Raw: []byte{'s', 't', 'a', 'y', 0},
		Tokens: []corpus.Token{{Kind: "text", Raw: []byte("stay"), Text: "stay"}, {Kind: "suffix", Raw: []byte{0}}},
	}
	bank := corpus.Bank{Name: "msgsec001.dat", Section: 1, Records: []corpus.Record{first, second}}
	source := &corpus.Project{Items: []corpus.Item{
		{Record: first, Translation: corpus.Translation{ID: first.ID}},
		{Record: second, Translation: corpus.Translation{ID: second.ID}},
	}}
	korean := &corpus.KoreanProject{Entries: []corpus.KoreanEntry{{ID: first.ID, Japanese: "old", Korean: "가"}}}
	mapping := koreanslots.Mapping{'가': cp932.GlyphKey(0xAC82)}

	compiled, err := compileKoreanBanks(source, korean, []corpus.Bank{bank}, mapping, nil)
	if err != nil {
		t.Fatal(err)
	}
	data := compiled[bank.Name]
	if len(data) < 12 {
		t.Fatalf("compiled bank too short: %d", len(data))
	}
	firstOffset := int(binary.LittleEndian.Uint32(data[4:8]))
	secondOffset := int(binary.LittleEndian.Uint32(data[8:12]))
	if firstOffset != 12 {
		t.Fatalf("first offset = %d, want 12", firstOffset)
	}
	if got := data[firstOffset:secondOffset]; string(got) != string([]byte{0x82, 0xAC, 0}) {
		t.Fatalf("Korean record bytes = % X", got)
	}
	if got := data[secondOffset:]; string(got) != string(second.Raw) {
		t.Fatalf("unselected retail record changed: got % X want % X", got, second.Raw)
	}
}

func TestCompileKoreanBanksRejectsOverlayForMissingRetailSection(t *testing.T) {
	source := &corpus.Project{Items: []corpus.Item{}}
	korean := &corpus.KoreanProject{Entries: []corpus.KoreanEntry{{ID: 20000, Japanese: "x", Korean: "가"}}}
	mapping := koreanslots.Mapping{'가': cp932.GlyphKey(0xAC82)}
	if _, err := compileKoreanBanks(source, korean, nil, mapping, nil); err == nil {
		t.Fatal("expected missing retail section error")
	}
}
