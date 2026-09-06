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
	addMapping := func(text string) {
		for _, r := range text {
			if r > 0x7f {
				mapping[r] = cp932.GlyphKey(0xAC82)
			}
		}
	}
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
		addMapping(row.Korean)
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
	addMapping(fixedControl.Korean)
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

	for _, boundedID := range []int{560650, 1980005, 950059} {
		bounded, ok := korean.Find(boundedID)
		if !ok {
			t.Fatalf("missing Korean row %d", boundedID)
		}
		if engine.narrowText(boundedID) {
			t.Fatalf("message %d unexpectedly belongs to narrow_text; regression must cover C5-only bounded substitution", boundedID)
		}
		if !engine.has(engine.consumers.C5IDs, boundedID) && !engine.has(engine.consumers.C5PortraitIDs, boundedID) {
			t.Fatalf("message %d lacks authenticated C5 consumer classification", boundedID)
		}
		if !strings.Contains(strings.ToUpper(bounded.Korean), "<VALUE:$28>") {
			t.Fatalf("message %d fixture no longer contains bounded player-name substitution: %q", boundedID, bounded.Korean)
		}
		if !koreanDialogueRuntimeSubstitution(boundedID, bounded.Korean) {
			t.Fatalf("message %d should be recognized as containing a runtime substitution", boundedID)
		}
		if koreanDialogueUnboundedRuntimeSubstitution(boundedID, bounded.Korean) {
			t.Fatalf("message %d bounded $28 substitution was misclassified as unbounded", boundedID)
		}
		if !engine.koreanEnglishDialogueVisualConsumer(boundedID, bounded.Korean) {
			t.Fatalf("message %d bounded-substitution C5 dialogue must be eligible for source-aware reflow", boundedID)
		}
		addMapping(bounded.Korean)
		boundedMini := &corpus.KoreanProject{Entries: []corpus.KoreanEntry{bounded}}
		boundedLayouts, boundedDerived, err := engine.DeriveKoreanEnglishDialogueLayouts(source, boundedMini, nil, mapping)
		if err != nil {
			t.Fatalf("derive bounded-substitution C5 message %d: %v", boundedID, err)
		}
		if boundedDerived != 1 {
			t.Fatalf("bounded-substitution C5 message %d derived=%d, want 1", boundedID, boundedDerived)
		}
		boundedLayout := boundedLayouts[boundedID]
		if boundedLayout == "" || !strings.Contains(boundedLayout, lineBreak) {
			t.Fatalf("bounded-substitution C5 message %d did not receive a line break: %q", boundedID, boundedLayout)
		}
		if !message.PreservesLayoutSemantics(bounded.Korean, boundedLayout) {
			t.Fatalf("bounded-substitution C5 message %d derived layout changes canonical semantics: %q", boundedID, boundedLayout)
		}
		boundedItem, _ := source.Find(boundedID)
		boundedWidth, _, err := engine.koreanWarningMetrics(boundedItem.Record, boundedLayout, boundedID, mapping)
		if err != nil {
			t.Fatalf("bounded-substitution C5 message %d width check: %v", boundedID, err)
		}
		if boundedWidth > engine.advanceLimit(boundedID) {
			t.Fatalf("bounded-substitution C5 message %d remains over width after reflow: %d > %d", boundedID, boundedWidth, engine.advanceLimit(boundedID))
		}
	}

	const c5FixtureID = 560650
	if !koreanDialogueUnboundedRuntimeSubstitution(c5FixtureID, "앞 <value:$15> 뒤<end>") {
		t.Fatal("unproven $15 inline substitution must remain classified as unbounded")
	}
	if !engine.koreanEnglishDialogueVisualConsumer(c5FixtureID, "앞 <value:$15> 뒤<end>") {
		t.Fatal("C5 dialogue with unproven inline $15 substitution must remain eligible for static source-aware reflow")
	}
}
