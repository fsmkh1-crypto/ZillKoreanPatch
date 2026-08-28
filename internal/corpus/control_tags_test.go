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

func TestFixedRuntimeLiteralOccupancyIgnoresLineBreaksAndWhitespace(t *testing.T) {
	text := "  <if><value:$01><equal>19本文<line-break>続き<end>  <line-break>…<value:$28>、<line-break><if>次<end>"
	want := []bool{false, false, false, true, true, true, true, false}
	if got := FixedRuntimeLiteralOccupancy(text); !slices.Equal(got, want) {
		t.Fatalf("FixedRuntimeLiteralOccupancy = %#v, want %#v", got, want)
	}
}

func TestFixedRuntimeLiteralOccupancyExposesDroppedBranchLiteral(t *testing.T) {
	source := "<if><value:$01><equal>14前<end>…<value:$28>、<line-break><if>後<end>"
	translated := "<if><value:$01><equal>14앞<end><value:$28>, <if>뒤<end>"
	want := FixedRuntimeLiteralOccupancy(source)
	got := FixedRuntimeLiteralOccupancy(translated)
	if len(got) != len(want) {
		t.Fatalf("occupancy lengths differ: got %d want %d", len(got), len(want))
	}
	missing := false
	for i := range want {
		if want[i] && !got[i] {
			missing = true
			break
		}
	}
	if !missing {
		t.Fatalf("expected dropped fixed literal to be detectable: source=%#v translated=%#v", want, got)
	}
}
