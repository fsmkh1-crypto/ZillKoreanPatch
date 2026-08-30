// SPDX-License-Identifier: GPL-3.0-or-later

package message

import "testing"

func TestPreservesSemanticsAllowsHangulSyllableBoundary(t *testing.T) {
	if !preservesSemantics("긴복합단어<end>", "긴복합<line-break>단어<end>") {
		t.Fatal("Hangul syllable-boundary reflow was rejected")
	}
}

func TestPreservesSemanticsAllowsPunctuationBoundary(t *testing.T) {
	for _, tc := range []struct {
		semantic string
		layout   string
	}{
		{"문장이야….다음<end>", "문장이야<line-break>….다음<end>"},
		{"문장이야….다음<end>", "문장이야….<line-break>다음<end>"},
	} {
		if !preservesSemantics(tc.semantic, tc.layout) {
			t.Fatalf("punctuation-adjacent reflow rejected: semantic=%q layout=%q", tc.semantic, tc.layout)
		}
	}
}

func TestPreservesSemanticsAllowsWhitespaceSpanBoundary(t *testing.T) {
	if !preservesSemantics("긴복합   단어<end>", "긴복합<line-break>단어<end>") {
		t.Fatal("whitespace-span reflow was rejected")
	}
}

func TestPreservesSemanticsAllowsWhitespaceBackedBlankLines(t *testing.T) {
	for _, tc := range []struct {
		semantic string
		layout   string
	}{
		{"앞  뒤<end>", "앞<line-break><line-break>뒤<end>"},
		{"앞   뒤<end>", "앞<line-break><line-break><line-break>뒤<end>"},
	} {
		if !preservesSemantics(tc.semantic, tc.layout) {
			t.Fatalf("whitespace-backed blank lines rejected: semantic=%q layout=%q", tc.semantic, tc.layout)
		}
	}
}

func TestPreservesSemanticsRejectsCharacterChange(t *testing.T) {
	if preservesSemantics("긴복합단어<end>", "긴복합<line-break>다어<end>") {
		t.Fatal("layout changed a semantic Hangul character")
	}
}

func TestPreservesSemanticsRejectsControlReordering(t *testing.T) {
	semantic := "가<value:$15>나<value:$28>다<end>"
	layout := "가<value:$28>나<value:$15>다<end>"
	if preservesSemantics(semantic, layout) {
		t.Fatal("layout reordered runtime controls")
	}
}

func TestPreservesSemanticsRejectsBoundaryAtRuntimeControl(t *testing.T) {
	for _, layout := range []string{
		"가<line-break><value:$15>나<end>",
		"가<value:$15><line-break>나<end>",
	} {
		if preservesSemantics("가<value:$15>나<end>", layout) {
			t.Fatalf("layout changed runtime-control adjacency: %q", layout)
		}
	}
}

func TestPreservesSemanticsRejectsUnbackedBoundaries(t *testing.T) {
	for _, layout := range []string{
		"<line-break>긴복합단어<end>",
		"긴복합단어<line-break><end>",
		"긴복합<line-break><line-break>단어<end>",
	} {
		if preservesSemantics("긴복합단어<end>", layout) {
			t.Fatalf("invalid boundary placement accepted: %q", layout)
		}
	}
	if preservesSemantics("앞  뒤<end>", "앞<line-break><line-break><line-break>뒤<end>") {
		t.Fatal("layout invented more blank-line boundaries than semantic whitespace owns")
	}
}
