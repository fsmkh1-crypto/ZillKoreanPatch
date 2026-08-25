// SPDX-License-Identifier: GPL-3.0-or-later

package message

import "testing"

func TestPreservesSemanticsAllowsHangulSyllableBoundary(t *testing.T) {
	if !preservesSemantics("긴복합단어<end>", "긴복합<line-break>단어<end>") {
		t.Fatal("Hangul syllable-boundary reflow was rejected")
	}
}

func TestPreservesSemanticsAllowsWhitespaceSpanBoundary(t *testing.T) {
	if !preservesSemantics("긴복합   단어<end>", "긴복합<line-break>단어<end>") {
		t.Fatal("whitespace-span reflow was rejected")
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

func TestPreservesSemanticsRejectsLeadingTrailingOrRepeatedBoundary(t *testing.T) {
	for _, layout := range []string{
		"<line-break>긴복합단어<end>",
		"긴복합단어<line-break><end>",
		"긴복합<line-break><line-break>단어<end>",
	} {
		if preservesSemantics("긴복합단어<end>", layout) {
			t.Fatalf("invalid boundary placement accepted: %q", layout)
		}
	}
}
