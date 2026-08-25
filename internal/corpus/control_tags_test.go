// SPDX-License-Identifier: GPL-3.0-or-later

package corpus

import (
	"slices"
	"testing"
)

func TestRuntimeControlTagsIgnoreBracketedNaturalText(t *testing.T) {
	text := "<未使用><value:$15>本文<line-break><미사용><end>"
	wantAll := []string{"<value:$15>", "<line-break>", "<end>"}
	if got := RuntimeControlTags(text); !slices.Equal(got, wantAll) {
		t.Fatalf("RuntimeControlTags = %#v, want %#v", got, wantAll)
	}
	wantFixed := []string{"<value:$15>", "<end>"}
	if got := FixedRuntimeControlTags(text); !slices.Equal(got, wantFixed) {
		t.Fatalf("FixedRuntimeControlTags = %#v, want %#v", got, wantFixed)
	}
}

func TestRuntimeControlTagsCoverDisplayTextControlForms(t *testing.T) {
	text := "<if><select><call:12><jump:34><value:$1F><add><not-equal><operator:$04:$7B><color:A><discard:C:$01><escape:$48><separator><backspace><tab><$01><end>"
	got := RuntimeControlTags(text)
	if len(got) != 16 {
		t.Fatalf("got %d control tags: %#v", len(got), got)
	}
}
