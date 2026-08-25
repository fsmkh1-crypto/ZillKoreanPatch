// SPDX-License-Identifier: GPL-3.0-or-later

package corpus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadKoreanProjectAllowsMissingDirectoryAsEmptyOverlay(t *testing.T) {
	project, summary, err := LoadKoreanProject(t.TempDir(), testKoreanSourceProject())
	if err != nil {
		t.Fatal(err)
	}
	if summary.Sections != 0 || summary.Records != 0 || len(project.Entries) != 0 {
		t.Fatalf("unexpected nonempty overlay: summary=%+v entries=%+v", summary, project.Entries)
	}
}

func TestLoadKoreanProjectSparseAndStable(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "translations", "korean", "messages")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	source := testKoreanSourceProject()
	writeKoreanTestFile(t, dir, "msgsec001.toml", `# SPDX-License-Identifier: CC-BY-SA-4.0

["10008"]
japanese = "次"
korean = "다음"

["10007"]
japanese = "汝"
korean = "그대"
`)

	project, summary, err := LoadKoreanProject(root, source)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Sections != 1 || summary.Records != 2 {
		t.Fatalf("summary = %+v", summary)
	}
	if len(project.Entries) != 2 || project.Entries[0].ID != 10007 || project.Entries[1].ID != 10008 {
		t.Fatalf("entries not stable by ID: %+v", project.Entries)
	}
	if got := project.Texts(); len(got) != 2 || got[0] != "그대" || got[1] != "다음" {
		t.Fatalf("texts = %#v", got)
	}
	if row, ok := project.Find(10007); !ok || row.Korean != "그대" {
		t.Fatalf("Find(10007) = %+v, %v", row, ok)
	}
}

func TestLoadKoreanProjectRejectsJapaneseDrift(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "translations", "korean", "messages")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeKoreanTestFile(t, dir, "msgsec001.toml", `# SPDX-License-Identifier: CC-BY-SA-4.0

["10007"]
japanese = "different"
korean = "그대"
`)
	_, _, err := LoadKoreanProject(root, testKoreanSourceProject())
	if err == nil || !strings.Contains(err.Error(), "Japanese reference differs") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadKoreanProjectRejectsUnknownField(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "translations", "korean", "messages")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeKoreanTestFile(t, dir, "msgsec001.toml", `# SPDX-License-Identifier: CC-BY-SA-4.0

["10007"]
japanese = "汝"
korean = "그대"
english = "you"
`)
	_, _, err := LoadKoreanProject(root, testKoreanSourceProject())
	if err == nil || !strings.Contains(err.Error(), "invalid TOML") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadKoreanProjectRejectsWrongSection(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "translations", "korean", "messages")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeKoreanTestFile(t, dir, "msgsec002.toml", `# SPDX-License-Identifier: CC-BY-SA-4.0

["10007"]
japanese = "汝"
korean = "그대"
`)
	_, _, err := LoadKoreanProject(root, testKoreanSourceProject())
	if err == nil || !strings.Contains(err.Error(), "belongs to section") {
		t.Fatalf("error = %v", err)
	}
}

func TestWithKoreanCopiesCanonicalJapaneseAndRenders(t *testing.T) {
	source := testKoreanSourceProject()
	project := &KoreanProject{byID: map[int]int{}}
	updated, err := project.WithKorean(source, 10008, "다음")
	if err != nil {
		t.Fatal(err)
	}
	updated, err = updated.WithKorean(source, 10007, "그대")
	if err != nil {
		t.Fatal(err)
	}
	data, err := updated.RenderKoreanSection(1)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.HasPrefix(text, koreanLicenseLine+"\n") {
		t.Fatalf("missing license header: %q", text)
	}
	first := strings.Index(text, `["10007"]`)
	second := strings.Index(text, `["10008"]`)
	if first < 0 || second < 0 || first >= second {
		t.Fatalf("rows not rendered in stable ID order:\n%s", text)
	}
	if !strings.Contains(text, `japanese = "汝"`) || !strings.Contains(text, `korean = "그대"`) {
		t.Fatalf("rendered text missing canonical pair:\n%s", text)
	}
}

func TestWithKoreanRejectsRawControl(t *testing.T) {
	project := &KoreanProject{byID: map[int]int{}}
	_, err := project.WithKorean(testKoreanSourceProject(), 10007, "그\n대")
	if err == nil || !strings.Contains(err.Error(), "raw Unicode control") {
		t.Fatalf("error = %v", err)
	}
}

func testKoreanSourceProject() *Project {
	items := []Item{
		{Record: Record{ID: 10007, Index: 7, Display: "汝"}, Translation: Translation{ID: 10007, Japanese: "汝"}},
		{Record: Record{ID: 10008, Index: 8, Display: "次"}, Translation: Translation{ID: 10008, Japanese: "次"}},
	}
	return &Project{Items: items, byID: map[int]int{10007: 0, 10008: 1}}
}

func writeKoreanTestFile(t *testing.T, dir, name, data string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}
