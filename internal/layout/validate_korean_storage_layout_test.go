// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"strings"
	"testing"

	"github.com/HK47196/zill/internal/message"
)

func TestWrapKoreanC5StoragePreservesControlsAndBoundsLines(t *testing.T) {
	input := "물론 용왕을 쓰러뜨리는 것만으로 인류의 혁신을 달성할 수 있다고 나도 생각하진 않는다. 하지만 <value:$28>에게도 그런 자질이 있을지도 모르겠군.<end>"
	got := wrapKoreanC5Storage(input)
	if !strings.Contains(got, "<value:$28>") || !strings.HasSuffix(got, "<end>") {
		t.Fatalf("control tags were not preserved: %q", got)
	}
	if !strings.Contains(got, lineBreak) {
		t.Fatalf("expected derived line breaks: %q", got)
	}
	visibleText := controlTag.ReplaceAllString(got, "")
	if strings.ReplaceAll(visibleText, " ", "") != strings.ReplaceAll(controlTag.ReplaceAllString(input, ""), " ", "") {
		t.Fatalf("visible text changed beyond whitespace-to-layout projection: %q", got)
	}
	for _, line := range strings.Split(got, lineBreak) {
		line = controlTag.ReplaceAllString(line, "")
		if n := len([]rune(line)); n > 18 {
			t.Fatalf("derived line has %d visible runes (>18): %q", n, line)
		}
	}
}

func TestWrapKoreanC5StorageKeepsExistingBreaks(t *testing.T) {
	input := "첫째 줄<line-break>둘째 줄은 조금 더 길지만 그대로 이어지는 문장이다.<end>"
	got := wrapKoreanC5Storage(input)
	if !strings.HasPrefix(got, "첫째 줄<line-break>") {
		t.Fatalf("existing line break was not retained: %q", got)
	}
}

func TestWrapKoreanC5StoragePreservesValueLeadingWhitespace(t *testing.T) {
	input := "<value:$28> 공께. 의뢰드리고 싶은 일이 있습니다. 지체 없이 엔샨트 정청으로 와 주십시오. 자기브 딘갈<end>"
	got := wrapKoreanC5Storage(input)
	if !strings.Contains(got, "<value:$28> 공께") {
		t.Fatalf("C5 wrapper dropped canonical whitespace after runtime value: %q", got)
	}
	if !message.PreservesLayoutSemantics(input, got) {
		t.Fatalf("C5 wrapper changed semantic/control topology: %q", got)
	}
}

func TestWrapKoreanStoragePreservesRuntimeControlAdjacency(t *testing.T) {
	input := strings.Repeat("가", 18) + "<value:$15>여기는이어지는문장입니다<end>"
	got := wrapKoreanStoragePreservingControlAdjacency(input)
	if !strings.Contains(got, "<value:$15>여") {
		t.Fatalf("wrapper inserted a boundary at runtime-control adjacency: %q", got)
	}
	if !strings.Contains(got, lineBreak) {
		t.Fatalf("expected a derived boundary after the protected leading rune: %q", got)
	}
	if !message.PreservesLayoutSemantics(input, got) {
		t.Fatalf("control-aware wrapper changed semantic/control topology: %q", got)
	}
}

func TestWrapKoreanStoragePreservesValueLeadingWhitespace(t *testing.T) {
	input := "<value:$28> 공께. 의뢰드리고 싶은 일이 있습니다. 지체 없이 엔샨트 정청으로 와 주십시오. 자기브 딘갈<end>"
	got := wrapKoreanStoragePreservingControlAdjacency(input)
	if !strings.Contains(got, "<value:$28> 공께") {
		t.Fatalf("C22 wrapper dropped canonical whitespace after runtime value: %q", got)
	}
	if !message.PreservesLayoutSemantics(input, got) {
		t.Fatalf("C22 wrapper changed semantic/control topology: %q", got)
	}
}

func TestWrapKoreanStoragePreservesCanonicalLeadingWhitespace(t *testing.T) {
	input := " 변덕스러운 인간의 아이여. 마음이 바뀌었다면 다시 나를 찾아오거라.<end> 잘 가거라.<end>"
	got := wrapKoreanStoragePreservingControlAdjacency(input)
	if !strings.HasPrefix(got, " 변덕스러운") {
		t.Fatalf("C22 wrapper dropped canonical record-leading whitespace: %q", got)
	}
	if !strings.Contains(got, "<end> 잘 가거라.<end>") {
		t.Fatalf("C22 wrapper changed the second semantic segment: %q", got)
	}
	if !message.PreservesLayoutSemantics(input, got) {
		t.Fatalf("C22 wrapper changed multi-segment semantic/control topology: %q", got)
	}
}

func TestWrapKoreanStoragePreservesWhitespaceAfterExistingBreak(t *testing.T) {
	input := "첫째 줄<line-break> 둘째 줄<end>"
	got := wrapKoreanStoragePreservingControlAdjacency(input)
	if !strings.Contains(got, "<line-break> 둘째") {
		t.Fatalf("C22 wrapper dropped authored whitespace after an existing break: %q", got)
	}
	if !message.PreservesLayoutSemantics(input, got) {
		t.Fatalf("C22 wrapper changed semantics around an existing break: %q", got)
	}
}

func TestWrapKoreanStorageRepairsInheritedC5ValueBoundary(t *testing.T) {
	semantic := strings.Repeat("가", 18) + "<value:$15>여기는이어지는문장입니다<end>"
	inherited := strings.Repeat("가", 18) + "<value:$15><line-break>여기는이어지는문장입니다<end>"
	got := wrapKoreanStoragePreservingControlAdjacency(inherited)
	if strings.Contains(got, "<value:$15><line-break>") {
		t.Fatalf("inherited C5 boundary survived beside runtime value control: %q", got)
	}
	if !strings.Contains(got, "<value:$15>여") {
		t.Fatalf("runtime control adjacency was not restored: %q", got)
	}
	if !message.PreservesLayoutSemantics(semantic, got) {
		t.Fatalf("repaired inherited projection does not preserve canonical semantics: %q", got)
	}
}
