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
		if !engine.koreanEnglishDialogueVisualConsumer(id, row.Korean) {
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

	const fixedControlID = 280181
	fixedControl, ok := korean.Find(fixedControlID)
	if !ok {
		t.Fatalf("missing Korean row %d", fixedControlID)
	}
	if !strings.Contains(strings.ToUpper(fixedControl.Korean), "<VALUE:$20>") {
		t.Fatalf("message %d fixture no longer contains fixed select value control: %q", fixedControlID, fixedControl.Korean)
	}
	if koreanDialogueRuntimeSubstitution(fixedControlID, fixedControl.Korean) {
		t.Fatalf("message %d fixed $20 select control was misclassified as a runtime-width substitution", fixedControlID)
	}
	if !engine.koreanEnglishDialogueVisualConsumer(fixedControlID, fixedControl.Korean) {
		t.Fatalf("message %d fixed-control C5 dialogue must be eligible for static reflow", fixedControlID)
	}
	for _, r := range fixedControl.Korean {
		if r > 0x7f {
			mapping[r] = cp932.GlyphKey(0xAC82)
		}
	}
	fixedMini := &corpus.KoreanProject{Entries: []corpus.KoreanEntry{fixedControl}}
	fixedLayouts, fixedDerived, err := engine.DeriveKoreanEnglishDialogueLayouts(source, fixedMini, nil, mapping)
	if err != nil {
		t.Fatalf("derive fixed-control C5 message %d: %v", fixedControlID, err)
	}
	if fixedDerived != 1 {
		t.Fatalf("fixed-control C5 message %d derived=%d, want 1", fixedControlID, fixedDerived)
	}
	fixedLayout := fixedLayouts[fixedControlID]
	if fixedLayout == "" || !strings.Contains(fixedLayout, lineBreak) {
		t.Fatalf("fixed-control C5 message %d did not receive a line break: %q", fixedControlID, fixedLayout)
	}
	if !message.PreservesLayoutSemantics(fixedControl.Korean, fixedLayout) {
		t.Fatalf("fixed-control C5 message %d derived layout changes canonical semantics: %q", fixedControlID, fixedLayout)
	}
	fixedItem, _ := source.Find(fixedControlID)
	fixedWidth, _, err := engine.koreanWarningMetrics(fixedItem.Record, fixedLayout, fixedControlID, mapping)
	if err != nil {
		t.Fatalf("fixed-control C5 message %d width check: %v", fixedControlID, err)
	}
	if fixedWidth > engine.advanceLimit(fixedControlID) {
		t.Fatalf("fixed-control C5 message %d remains over width after reflow: %d > %d", fixedControlID, fixedWidth, engine.advanceLimit(fixedControlID))
	}

	const dynamicID = 560650
	dynamic, ok := korean.Find(dynamicID)
	if !ok {
		t.Fatalf("missing Korean row %d", dynamicID)
	}
	if engine.narrowText(dynamicID) {
		t.Fatalf("message %d unexpectedly belongs to narrow_text; regression must cover newly admitted C5 scope", dynamicID)
	}
	if !engine.has(engine.consumers.C5IDs, dynamicID) && !engine.has(engine.consumers.C5PortraitIDs, dynamicID) {
		t.Fatalf("message %d lacks authenticated C5 consumer classification", dynamicID)
	}
	if !valueTag.MatchString(dynamic.Korean) {
		t.Fatalf("message %d fixture no longer contains runtime value substitution: %q", dynamicID, dynamic.Korean)
	}
	if engine.koreanEnglishDialogueVisualConsumer(dynamicID, dynamic.Korean) {
		t.Fatalf("message %d dynamic C5 dialogue must stay outside static reflow", dynamicID)
	}

	dynamicMini := &corpus.KoreanProject{Entries: []corpus.KoreanEntry{dynamic}}
	dynamicLayouts, dynamicDerived, err := engine.DeriveKoreanEnglishDialogueLayouts(source, dynamicMini, nil, mapping)
	if err != nil {
		t.Fatalf("derive dynamic C5 message %d: %v", dynamicID, err)
	}
	if dynamicDerived != 0 {
		t.Fatalf("dynamic C5 message %d derived=%d, want 0", dynamicID, dynamicDerived)
	}
	if _, exists := dynamicLayouts[dynamicID]; exists {
		t.Fatalf("dynamic C5 message %d unexpectedly received a static layout", dynamicID)
	}
	checked, overflowIDs, err := engine.AuditKoreanEnglishDialogueResiduals(source, dynamicMini, dynamicLayouts, mapping)
	if err != nil {
		t.Fatalf("audit dynamic C5 message %d: %v", dynamicID, err)
	}
	if checked != 0 || len(overflowIDs) != 0 {
		t.Fatalf("dynamic C5 message %d residual audit checked=%d overflows=%v, want excluded", dynamicID, checked, overflowIDs)
	}
}
