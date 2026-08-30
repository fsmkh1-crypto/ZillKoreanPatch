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
// counterpart of the device builder's fixed-consumer and hard-visual gates.
// Unlike the A-054 scanner census, this exhausts every upstream-English
// asset-independent storage category over all 42,016 accepted Korean rows and
// then runs the release-blocking character-profile visual contracts with the
// Korean renderer mapping.
//
// Do NOT call BuildKoreanBetaProject here. That projection needs authenticated
// retail records to classify editability; on an asset-free repository load its
// Record.Raw fields are intentionally empty and every accepted row can otherwise
// be misclassified as structural. This census therefore validates the canonical
// overlay directly and hard-asserts its checked population.
//
// All custom Korean glyphs are two-byte renderer keys, so one valid two-byte key
// per required rune is sufficient for exact byte-width accounting and Korean
// semantic splitting. Asset-bound record projection, bank/table capacity and
// archive integration remain the independent CompileBankKorean gate on
// authenticated retail input.
func TestCurrentKoreanCorpusEnglishConsumerStorageContracts(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	source, _, err := corpus.LoadProject(root)
	if err != nil {
		t.Fatal(err)
	}
	korean, _, err := corpus.LoadKoreanProject(root, source)
	if err != nil {
		t.Fatal(err)
	}
	const wantCanonical = 42016
	if len(korean.Entries) != wantCanonical {
		t.Fatalf("canonical Korean corpus drift: got %d want %d", len(korean.Entries), wantCanonical)
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
	if err := engine.ValidateKoreanEnglishConsumerStorageContracts(source, korean, layouts, mapping); err != nil {
		t.Fatal(err)
	}
	if err := engine.ValidateKoreanEnglishVisualContracts(source, korean, layouts, mapping); err != nil {
		t.Fatal(err)
	}

	checked := len(korean.Entries)
	if checked != wantCanonical {
		t.Fatalf("consumer census checked %d rows, want %d", checked, wantCanonical)
	}
	t.Logf("FORENSIC KOREAN_CONSUMER_STORAGE_SUMMARY canonical=%d checked=%d english_layouts=%d contracts=PASS visual=PASS exact_asset_gate=CompileBankKorean",
		len(korean.Entries), checked, derivedEnglish)
}
