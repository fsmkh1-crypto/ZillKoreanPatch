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
	fragments, controls := splitRepositoryReflowFragments("앞<value:$21>뒤<end>", false)
	wantFragments := []string{"앞", "뒤", ""}
	wantControls := []string{"<value:$21>", "<end>"}
	if !reflect.DeepEqual(fragments, wantFragments) {
		t.Fatalf("fragments=%q, want %q", fragments, wantFragments)
	}
	if !reflect.DeepEqual(controls, wantControls) {
		t.Fatalf("controls=%q, want %q", controls, wantControls)
	}
}

func TestSplitRepositoryReflowFragmentsKeepsIfExpressionOperandFixed(t *testing.T) {
	fragments, controls := splitRepositoryReflowFragments("<if><value:$01><equal>190간다, <value:$28>.<end>", false)
	wantFragments := []string{"", "간다, <value:$28>.", ""}
	wantControls := []string{"<if><value:$01><equal>190", "<end>"}
	if !reflect.DeepEqual(fragments, wantFragments) {
		t.Fatalf("fragments=%q, want %q", fragments, wantFragments)
	}
	if !reflect.DeepEqual(controls, wantControls) {
		t.Fatalf("controls=%q, want %q", controls, wantControls)
	}
}

func TestSplitRepositoryReflowFragmentsKeepsPercentExpressionOperandFixed(t *testing.T) {
	fragments, controls := splitRepositoryReflowFragments("<if><value:$29><equal>%0<value:$28> 님<end>", false)
	wantFragments := []string{"", "<value:$28> 님", ""}
	wantControls := []string{"<if><value:$29><equal>%0", "<end>"}
	if !reflect.DeepEqual(fragments, wantFragments) {
		t.Fatalf("fragments=%q, want %q", fragments, wantFragments)
	}
	if !reflect.DeepEqual(controls, wantControls) {
		t.Fatalf("controls=%q, want %q", controls, wantControls)
	}
}

func TestSplitRepositoryReflowFragmentsKeepsMovableValueFixedInsideExpression(t *testing.T) {
	fragments, controls := splitRepositoryReflowFragments("<if><value:$15><equal>%0표시문<end>", false)
	wantFragments := []string{"", "표시문", ""}
	wantControls := []string{"<if><value:$15><equal>%0", "<end>"}
	if !reflect.DeepEqual(fragments, wantFragments) {
		t.Fatalf("fragments=%q, want %q", fragments, wantFragments)
	}
	if !reflect.DeepEqual(controls, wantControls) {
		t.Fatalf("controls=%q, want %q", controls, wantControls)
	}
}

func TestSplitRepositoryReflowFragmentsKeepsSelectArithmeticOperandFixed(t *testing.T) {
	fragments, controls := splitRepositoryReflowFragments("<select><value:$20>%1선택문<end>", false)
	wantFragments := []string{"", "선택문", ""}
	wantControls := []string{"<select><value:$20>%1", "<end>"}
	if !reflect.DeepEqual(fragments, wantFragments) {
		t.Fatalf("fragments=%q, want %q", fragments, wantFragments)
	}
	if !reflect.DeepEqual(controls, wantControls) {
		t.Fatalf("controls=%q, want %q", controls, wantControls)
	}
}

func TestRepositorySemanticProjectionUsesSourceNumericBoundary(t *testing.T) {
	sourceFragments, sourceControls := splitRepositoryReflowFragments("<select><value:$20>%4８世紀末<end>", true)
	if !reflect.DeepEqual(sourceFragments, []string{"", "８世紀末", ""}) {
		t.Fatalf("source fragments=%q", sourceFragments)
	}
	if !reflect.DeepEqual(sourceControls, []string{"<select><value:$20>%4", "<end>"}) {
		t.Fatalf("source controls=%q", sourceControls)
	}

	fragments, err := splitRepositorySemanticAgainstSourceControls("<select><value:$20>%48세기 말<end>", sourceControls)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"", "8세기 말", ""}
	if !reflect.DeepEqual(fragments, want) {
		t.Fatalf("Korean fragments=%q, want %q", fragments, want)
	}
}
