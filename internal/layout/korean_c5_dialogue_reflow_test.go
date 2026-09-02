// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"os"
	"strings"
	"testing"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/message"
)

func TestKoreanC5DialogueMirrorsEnglishVisualReflow(t *testing.T) {
	read := func(path string) []byte {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	engine, err := Load(
		read("../../release/layout/consumer-map.toml"),
		read("../../release/font/metrics.toml"),
		read("../../release/layout/categories.toml"),
	)
	if err != nil {
		t.Fatal(err)
	}
	source, _, err := corpus.LoadProject("../..")
	if err != nil {
		t.Fatal(err)
	}
	korean, _, err := corpus.LoadKoreanProject("../..", source)
	if err != nil {
		t.Fatal(err)
	}

	targets := []int{300001, 300003}
	mini := &corpus.KoreanProject{}
	mapping := koreanslots.Mapping{}
	for _, id := range targets {
		row, ok := korean.Find(id)
		if !ok {
			t.Fatalf("missing Korean row %d", id)
		}
		if engine.narrowText(id) {
			t.Fatalf("message %d unexpectedly belongs to narrow_text; regression must cover C5-only eligibility", id)
		}
		if !engine.has(engine.consumers.C5IDs, id) && !engine.has(engine.consumers.C5PortraitIDs, id) {
			t.Fatalf("message %d lacks authenticated C5 consumer classification", id)
		}
		if !engine.koreanEnglishDialogueVisualConsumer(id) {
			t.Fatalf("message %d is not eligible for Korean dialogue visual reflow", id)
		}
		mini.Entries = append(mini.Entries, row)
		for _, r := range row.Korean {
			if r > 0x7f {
				mapping[r] = cp932.GlyphKey(0xAC82)
			}
		}
	}

	layouts, derived, err := engine.DeriveKoreanEnglishDialogueLayouts(source, mini, nil, mapping)
	if err != nil {
		t.Fatal(err)
	}
	if derived != len(targets) {
		t.Fatalf("derived=%d, want %d", derived, len(targets))
	}
	for _, id := range targets {
		row, _ := korean.Find(id)
		got := layouts[id]
		if got == "" || !strings.Contains(got, lineBreak) {
			t.Fatalf("message %d did not receive a line break: %q", id, got)
		}
		if !message.PreservesLayoutSemantics(row.Korean, got) {
			t.Fatalf("message %d derived layout changes canonical semantics: %q", id, got)
		}
		item, _ := source.Find(id)
		width, _, err := engine.koreanWarningMetrics(item.Record, got, id, mapping)
		if err != nil {
			t.Fatalf("message %d width check: %v", id, err)
		}
		if width > engine.advanceLimit(id) {
			t.Fatalf("message %d remains over width after reflow: %d > %d", id, width, engine.advanceLimit(id))
		}
	}
}
