// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"os"
	"testing"

	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
)

func TestKoreanPreferredUsesUpstreamBreakScoring(t *testing.T) {
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

	mapping := koreanslots.Mapping{}
	for _, r := range "abcd" {
		mapping[r] = cp932.GlyphKey(0xAC82)
	}
	const text = "aaaa bbbb. cccc dddd"
	limit, err := engine.measureKoreanRenderer("aaaa bbbb. cccc", 0, mapping)
	if err != nil {
		t.Fatal(err)
	}

	greedy, err := engine.koreanGreedy(text, limit, 0, mapping)
	if err != nil {
		t.Fatal(err)
	}
	const wantGreedy = "aaaa bbbb. cccc<line-break>dddd"
	if greedy != wantGreedy {
		t.Fatalf("Korean greedy fixture changed: got %q, want %q", greedy, wantGreedy)
	}

	preferred, err := engine.koreanPreferred(text, "AAAA BBBB.\nCCCC DDDD", limit, 0, false, mapping)
	if err != nil {
		t.Fatal(err)
	}
	const wantPreferred = "aaaa bbbb.<line-break>cccc dddd"
	if preferred != wantPreferred {
		t.Fatalf("Korean preferred layout = %q, want upstream-style punctuation/source-aware break %q", preferred, wantPreferred)
	}
}
