// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"path/filepath"
	"testing"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
)

// TestCurrentKoreanCorpusEnglishConsumerStorageContracts is the repository-only
// counterpart of the device builder's consumer gate. Unlike the A-054 scanner
// census, this exhausts every upstream-English fixed-storage category over the
// current Korean beta projection: C20/C22, bounded labels, character choices,
// guild strings/postings, trap, equipment feedback and chronicle payloads.
//
// All custom Korean glyphs are two-byte renderer keys, so one valid two-byte key
// per required rune is sufficient for exact storage-size accounting without a
// retail ISO. Asset-bound bank/table/archive checks remain in CompileBankKorean.
func TestCurrentKoreanCorpusEnglishConsumerStorageContracts(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	source, _, err := corpus.LoadProject(root)
	if err != nil {
		t.Fatal(err)
	}
	canonical, _, err := corpus.LoadKoreanProject(root, source)
	if err != nil {
		t.Fatal(err)
	}
	if len(canonical.Entries) != 42016 {
		t.Fatalf("canonical Korean corpus drift: got %d want 42016", len(canonical.Entries))
	}
	korean, skipped, err := BuildKoreanBetaProject(source, canonical)
	if err != nil {
		t.Fatal(err)
	}

	texts, err := korean.RuntimeTexts(source)
	if err != nil {
		t.Fatal(err)
	}
	mapping := make(koreanslots.Mapping)
	for _, r := range koreanslots.RequiredCustomRunes(texts) {
		mapping[r] = cp932.GlyphKey(0xAC82)
	}

	layouts := make(map[int]string)
	for _, row := range korean.Entries {
		if row.Layout != "" {
			layouts[row.ID] = row.Layout
		}
	}
	engine, err := loadLayout(root)
	if err != nil {
		t.Fatal(err)
	}
	layouts, derivedEnglish, err := engine.DeriveKoreanEnglishConsumerLayouts(source, korean, layouts, mapping)
	if err != nil {
		t.Fatal(err)
	}
	layouts, derivedC5, err := engine.DeriveKoreanC5StorageLayouts(source, korean, layouts, mapping)
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.ValidateKoreanEnglishConsumerStorageContracts(source, korean, layouts, mapping); err != nil {
		t.Fatal(err)
	}
	t.Logf("FORENSIC KOREAN_CONSUMER_STORAGE_SUMMARY canonical=%d materializable=%d structural_retail=%d english_layouts=%d c5_layouts=%d contracts=PASS exact_asset_gate=CompileBankKorean",
		len(canonical.Entries), len(korean.Entries), skipped, derivedEnglish, derivedC5)
}
