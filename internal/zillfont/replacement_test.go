// SPDX-License-Identifier: GPL-3.0-or-later

package zillfont

import (
	"testing"

	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
)

func TestReplacementPlanResolvesInstalledCellsDeterministically(t *testing.T) {
	paf := &PAF{Glyphs: []Glyph{
		{Index: 10, Key: cp932.GlyphKey(0xAC82), X: 100, Y: 20, Page: 1},
		{Index: 11, Key: cp932.GlyphKey(0xAD82), X: 110, Y: 20, Page: 1},
	}}
	mapping := koreanslots.Mapping{'나': cp932.GlyphKey(0xAD82), '가': cp932.GlyphKey(0xAC82)}
	got, err := paf.ReplacementPlan(mapping)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Rune != '가' || got[1].Rune != '나' {
		t.Fatalf("replacement order = %+v", got)
	}
	if got[0].Glyph.Index != 10 || got[1].Glyph.Index != 11 {
		t.Fatalf("resolved glyphs = %+v", got)
	}
}

func TestReplacementPlanRejectsMissingOrSharedKeys(t *testing.T) {
	paf := &PAF{Glyphs: []Glyph{{Index: 10, Key: cp932.GlyphKey(0xAC82)}}}
	if _, err := paf.ReplacementPlan(koreanslots.Mapping{'가': cp932.GlyphKey(0xAD82)}); err == nil {
		t.Fatal("expected missing installed key error")
	}
	if _, err := paf.ReplacementPlan(koreanslots.Mapping{'가': cp932.GlyphKey(0xAC82), '나': cp932.GlyphKey(0xAC82)}); err == nil {
		t.Fatal("expected shared key error")
	}
}
