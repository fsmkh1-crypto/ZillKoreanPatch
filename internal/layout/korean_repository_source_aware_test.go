// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"reflect"
	"testing"
)

func TestSplitRepositoryReflowFragmentsKeepsMovableValuesInline(t *testing.T) {
	fragments, controls := splitRepositoryReflowFragments("앞 <value:$28> 뒤<end>", false)
	wantFragments := []string{"앞 <value:$28> 뒤", ""}
	wantControls := []string{"<end>"}
	if !reflect.DeepEqual(fragments, wantFragments) {
		t.Fatalf("fragments=%q, want %q", fragments, wantFragments)
	}
	if !reflect.DeepEqual(controls, wantControls) {
		t.Fatalf("controls=%q, want %q", controls, wantControls)
	}
}

func TestSplitRepositoryReflowFragmentsKeepsSourceBreakHintAcrossMovableValue(t *testing.T) {
	fragments, controls := splitRepositoryReflowFragments("앞<line-break><value:$28> 뒤<end>", true)
	wantFragments := []string{"앞\n<value:$28> 뒤", ""}
	wantControls := []string{"<end>"}
	if !reflect.DeepEqual(fragments, wantFragments) {
		t.Fatalf("fragments=%q, want %q", fragments, wantFragments)
	}
	if !reflect.DeepEqual(controls, wantControls) {
		t.Fatalf("controls=%q, want %q", controls, wantControls)
	}
}

func TestSplitRepositoryReflowFragmentsStillSeparatesFixedValueControl(t *testing.T) {
	fragments, controls := splitRepositoryReflowFragments("앞<value:$20>뒤<end>", false)
	wantFragments := []string{"앞", "뒤", ""}
	wantControls := []string{"<value:$20>", "<end>"}
	if !reflect.DeepEqual(fragments, wantFragments) {
		t.Fatalf("fragments=%q, want %q", fragments, wantFragments)
	}
	if !reflect.DeepEqual(controls, wantControls) {
		t.Fatalf("controls=%q, want %q", controls, wantControls)
	}
}
