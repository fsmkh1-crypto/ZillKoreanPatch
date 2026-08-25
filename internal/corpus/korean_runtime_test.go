// SPDX-License-Identifier: GPL-3.0-or-later

package corpus

import "testing"

func TestKoreanRuntimeTextsUsesOverlayAndFallsBackToJapanese(t *testing.T) {
	source := testKoreanSourceProject()
	korean := &KoreanProject{
		Entries: []KoreanEntry{{ID: 10007, Japanese: "汝", Korean: "그대"}},
		byID:    map[int]int{10007: 0},
	}
	got, err := korean.RuntimeTexts(source)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "그대" || got[1] != "次" {
		t.Fatalf("runtime texts = %#v", got)
	}
}

func TestKoreanRuntimeTextsRejectsNilInputs(t *testing.T) {
	if _, err := (*KoreanProject)(nil).RuntimeTexts(testKoreanSourceProject()); err == nil {
		t.Fatal("expected nil Korean project error")
	}
	project := &KoreanProject{byID: map[int]int{}}
	if _, err := project.RuntimeTexts(nil); err == nil {
		t.Fatal("expected nil source project error")
	}
}
