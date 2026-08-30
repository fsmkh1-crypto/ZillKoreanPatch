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
