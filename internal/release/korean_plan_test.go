// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"encoding/binary"
	"testing"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
)

func TestCompileKoreanBanksWithPlanUsesPlanMapping(t *testing.T) {
	record := corpus.Record{
		ID: 10000, Index: 0, Raw: []byte{'o', 'l', 'd', 0},
		Tokens: []corpus.Token{{Kind: "text", Raw: []byte("old"), Text: "old"}, {Kind: "suffix", Raw: []byte{0}}},
	}
	bank := corpus.Bank{Name: "msgsec001.dat", Section: 1, Records: []corpus.Record{record}}
	source := &corpus.Project{Items: []corpus.Item{{
		Record: record, Translation: corpus.Translation{ID: record.ID, Japanese: "old"},
	}}}
	korean := &corpus.KoreanProject{Entries: []corpus.KoreanEntry{{ID: record.ID, Japanese: "old", Korean: "가"}}}
	plan := koreanslots.Plan{
		CustomRunes: []rune{'가'},
		Mapping:     koreanslots.Mapping{'가': cp932.GlyphKey(0xAC82)},
	}

	compiled, err := compileKoreanBanksWithPlan(source, korean, []corpus.Bank{bank}, plan, nil)
	if err != nil {
		t.Fatal(err)
	}
	data := compiled[bank.Name]
	offset := int(binary.LittleEndian.Uint32(data[4:8]))
	if got := data[offset:]; string(got) != string([]byte{0x82, 0xAC, 0}) {
		t.Fatalf("Korean record bytes = % X", got)
	}
}

func TestCompileKoreanBanksWithPlanRejectsStalePlan(t *testing.T) {
	record := corpus.Record{ID: 10000, Index: 0}
	source := &corpus.Project{Items: []corpus.Item{{
		Record: record, Translation: corpus.Translation{ID: record.ID, Japanese: "old"},
	}}}
	korean := &corpus.KoreanProject{Entries: []corpus.KoreanEntry{{ID: record.ID, Japanese: "old", Korean: "가"}}}
	plan := koreanslots.Plan{CustomRunes: []rune{'나'}, Mapping: koreanslots.Mapping{'나': cp932.GlyphKey(0xAC82)}}
	if _, err := compileKoreanBanksWithPlan(source, korean, nil, plan, nil); err == nil {
		t.Fatal("expected stale slot plan error")
	}
}
