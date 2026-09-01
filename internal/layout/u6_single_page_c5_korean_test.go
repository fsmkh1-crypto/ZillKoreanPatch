// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"os"
	"strings"
	"testing"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
)

func TestU6KoreanSinglePageC5RepairsStayWithinOnePage(t *testing.T) {
	read := func(path string) []byte {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	engine, err := Load(
		read("../../release/layout/consumer-map.toml"),
		read("../../release/font/metrics.toml"),
		read("../../release/layout/categories.toml"),
	)
	if err != nil {
		t.Fatal(err)
	}
	source, _, err := corpus.LoadProject("../..")
	if err != nil {
		t.Fatal(err)
	}
	korean, _, err := corpus.LoadKoreanProject("../..", source)
	if err != nil {
		t.Fatal(err)
	}

	targets := map[int][]int{
		1280007: {7},
		1280008: {5, 7},
		1280012: {1, 2, 5, 6},
		1280017: {5},
		1280020: {2, 7},
		1280021: {3, 5},
		1280043: {4},
		1280050: {4},
		1280051: {4},
	}

	for id, branches := range targets {
		row, ok := korean.Find(id)
		if !ok {
			t.Fatalf("missing Korean row %d", id)
		}
		fragments := strings.Split(row.Korean, "<end>")
		if len(fragments) == 0 || fragments[len(fragments)-1] != "" {
			t.Fatalf("message %d does not end with <end>", id)
		}
		for _, branch := range branches {
			if branch < 1 || branch >= len(fragments) {
				t.Fatalf("message %d branch %d out of range", id, branch)
			}
			text := fragments[branch-1]
			mapping := koreanslots.Mapping{}
			for _, r := range text {
				if r > 0x7f {
					mapping[r] = cp932.GlyphKey(0xAC82)
				}
			}
			wrapped, err := engine.wrapKoreanVisualToLimit(text, id, mapping, engine.advanceLimit(id))
			if err != nil {
				t.Fatalf("message %d branch %d wrap failed: %v", id, branch, err)
			}
			lines := strings.Split(wrapped, lineBreak)
			if len(lines) > 3 {
				t.Errorf("message %d branch %d wraps to %d lines, want <= 3: %q", id, branch, len(lines), wrapped)
			}
			for lineNo, line := range lines {
				width, err := engine.measureKoreanRenderer(line, id, mapping)
				if err != nil {
					t.Fatalf("message %d branch %d line %d measure failed: %v", id, branch, lineNo+1, err)
				}
				if width > engine.advanceLimit(id) {
					t.Errorf("message %d branch %d line %d is %d units, maximum %d: %q", id, branch, lineNo+1, width, engine.advanceLimit(id), line)
				}
			}
		}
	}
}
