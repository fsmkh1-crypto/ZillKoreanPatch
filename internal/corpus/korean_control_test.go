// SPDX-License-Identifier: GPL-3.0-or-later

package corpus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadKoreanProjectRejectsFixedControlTokenDrift(t *testing.T) {
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
korean = "그대여 응답하라<end>"
`)
	_, _, err := LoadKoreanProject(root, source)
	if err == nil || !strings.Contains(err.Error(), "fixed control token sequence differs") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadKoreanProjectRejectsDroppedInternalFixedLiteral(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "translations", "korean", "messages")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	sourceText := "<if><value:$01><equal>14前<end>…<value:$28>、<line-break><if>後<end>"
	source := &Project{
		Items: []Item{{Record: Record{ID: 10007, Index: 7, Display: sourceText}, Translation: Translation{ID: 10007, Japanese: sourceText}}},
		byID:  map[int]int{10007: 0},
	}
	writeKoreanTestFile(t, dir, "msgsec001.toml", `# SPDX-License-Identifier: CC-BY-SA-4.0

["10007"]
japanese = "<if><value:$01><equal>14前<end>…<value:$28>、<line-break><if>後<end>"
korean = "<if><value:$01><equal>14앞<end><value:$28>, <if>뒤<end>"
`)
	_, _, err := LoadKoreanProject(root, source)
	if err == nil || !strings.Contains(err.Error(), "drops fixed literal slot") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadKoreanProjectAllowsLeadingSourceFillerOmission(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "translations", "korean", "messages")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	sourceText := "さ、<value:$28>さま、<line-break>神殿へ<end>"
	source := &Project{
		Items: []Item{{Record: Record{ID: 10007, Index: 7, Display: sourceText}, Translation: Translation{ID: 10007, Japanese: sourceText}}},
		byID:  map[int]int{10007: 0},
	}
	writeKoreanTestFile(t, dir, "msgsec001.toml", `# SPDX-License-Identifier: CC-BY-SA-4.0

["10007"]
japanese = "さ、<value:$28>さま、<line-break>神殿へ<end>"
korean = "<value:$28> 님, 신전으로<end>"
`)
	if _, _, err := LoadKoreanProject(root, source); err != nil {
		t.Fatalf("valid leading-filler omission rejected: %v", err)
	}
}

func TestLoadKoreanProjectAllowsLineBreakReflowAndSeparateLayout(t *testing.T) {
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
	writeKoreanTestFile(t, dir, "msgsec001-part01.toml", `# SPDX-License-Identifier: CC-BY-SA-4.0

["10007"]
japanese = "<value:$15>よ<line-break>我に応ぜよ<end>"
korean = "<value:$15>그대여 응답하라<end>"
layout = "<value:$15>그대여<line-break>응답하라<end>"
`)
	project, _, err := LoadKoreanProject(root, source)
	if err != nil {
		t.Fatal(err)
	}
	row, ok := project.Find(10007)
	if !ok || row.Korean != "<value:$15>그대여 응답하라<end>" || row.Layout != "<value:$15>그대여<line-break>응답하라<end>" {
		t.Fatalf("unexpected row: %#v, ok=%v", row, ok)
	}
	layouts := project.Layouts()
	if layouts[10007] != row.Layout {
		t.Fatalf("layouts = %#v", layouts)
	}
}

func TestLoadKoreanProjectRejectsLineBreakInSemanticText(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "translations", "korean", "messages")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	sourceText := "本文<line-break>続き<end>"
	source := &Project{
		Items: []Item{{Record: Record{ID: 10007, Index: 7, Display: sourceText}, Translation: Translation{ID: 10007, Japanese: sourceText}}},
		byID:  map[int]int{10007: 0},
	}
	writeKoreanTestFile(t, dir, "msgsec001.toml", `# SPDX-License-Identifier: CC-BY-SA-4.0

["10007"]
japanese = "本文<line-break>続き<end>"
korean = "본문<line-break>계속<end>"
`)
	_, _, err := LoadKoreanProject(root, source)
	if err == nil || !strings.Contains(err.Error(), "wrapping belongs in layout") {
		t.Fatalf("error = %v", err)
	}
}

func TestWithKoreanRejectsFixedControlTokenDriftAndClearsLayout(t *testing.T) {
	sourceText := "<value:$15>よ<end>"
	source := &Project{
		Items: []Item{{Record: Record{ID: 10007, Index: 7, Display: sourceText}, Translation: Translation{ID: 10007, Japanese: sourceText}}},
		byID:  map[int]int{10007: 0},
	}
	project := &KoreanProject{
		Entries: []KoreanEntry{{ID: 10007, Japanese: sourceText, Korean: "<value:$15>기존<end>", Layout: "<value:$15>기<line-break>존<end>"}},
		byID:    map[int]int{10007: 0},
	}
	if _, err := project.WithKorean(source, 10007, "그대여<end>"); err == nil || !strings.Contains(err.Error(), "fixed control token sequence differs") {
		t.Fatalf("error = %v", err)
	}
	if _, err := project.WithKorean(source, 10007, "<value:$15>그대<line-break>여<end>"); err == nil || !strings.Contains(err.Error(), "wrapping belongs in layout") {
		t.Fatalf("semantic line break error = %v", err)
	}
	updated, err := project.WithKorean(source, 10007, "<value:$15>그대여<end>")
	if err != nil {
		t.Fatal(err)
	}
	row, _ := updated.Find(10007)
	if row.Layout != "" {
		t.Fatalf("translation edit retained stale layout %q", row.Layout)
	}
}
