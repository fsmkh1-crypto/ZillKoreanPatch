// SPDX-License-Identifier: GPL-3.0-or-later

package corpus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadKoreanProjectRejectsControlTokenDrift(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "translations", "korean", "messages")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	sourceText := "<value:$15>よ<line-break>我に応ぜよ<end>"
	source := &Project{
		Items: []Item{{Record: Record{ID: 10007, Index: 7, Display: sourceText}, Translation: Translation{ID: 10007, Japanese: sourceText}}},
		byID:  map[int]int{10007: 0},
	}
	writeKoreanTestFile(t, dir, "msgsec001.toml", `# SPDX-License-Identifier: CC-BY-SA-4.0

["10007"]
japanese = "<value:$15>よ<line-break>我に応ぜよ<end>"
korean = "그대여<line-break>응답하라<end>"
`)
	_, _, err := LoadKoreanProject(root, source)
	if err == nil || !strings.Contains(err.Error(), "control token sequence differs") {
		t.Fatalf("error = %v", err)
	}
}

func TestWithKoreanRejectsControlTokenDrift(t *testing.T) {
	sourceText := "<value:$15>よ<end>"
	source := &Project{
		Items: []Item{{Record: Record{ID: 10007, Index: 7, Display: sourceText}, Translation: Translation{ID: 10007, Japanese: sourceText}}},
		byID:  map[int]int{10007: 0},
	}
	project := &KoreanProject{byID: map[int]int{}}
	_, err := project.WithKorean(source, 10007, "그대여<end>")
	if err == nil || !strings.Contains(err.Error(), "control token sequence differs") {
		t.Fatalf("error = %v", err)
	}
}
