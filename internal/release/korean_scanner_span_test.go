// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/message"
)

// TestCurrentKoreanCorpusRetailScannerMaxSpanBelowInlineBoundary is the static
// discriminant requested after A-054/A-055. It does not emulate the whole
// z_un_0886C84C object builder; it proves the narrower property that the exact
// current compiler materialization cannot hand z_un_089661DC an ordinary-byte
// line span >= the 0x100 inline-region boundary for any accepted Korean record.
//
// The custom-rune mapping below deliberately maps every custom rune to one valid
// two-byte key. For this metric the identity of a renderer key is irrelevant:
// every production custom slot is exactly two non-control bytes, so the scanner
// span length is invariant under which valid two-byte key is selected.
func TestCurrentKoreanCorpusRetailScannerMaxSpanBelowInlineBoundary(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	source, _, err := corpus.LoadProject(root)
	if err != nil {
		t.Fatal(err)
	}
	canonical, _, err := corpus.LoadKoreanProject(root, source)
	if err != nil {
		t.Fatal(err)
	}
	korean, skippedStructural, err := BuildKoreanBetaProject(source, canonical)
	if err != nil {
		t.Fatal(err)
	}

	texts, err := korean.RuntimeTexts(source)
	if err != nil {
		t.Fatal(err)
	}
	mapping := make(koreanslots.Mapping)
	for _, r := range koreanslots.RequiredCustomRunes(texts) {
		mapping[r] = cp932.GlyphKey(0x8140)
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
	layouts, derived, err := engine.DeriveKoreanC5StorageLayouts(source, korean, layouts, mapping)
	if err != nil {
		t.Fatal(err)
	}

	type finding struct {
		id   int
		span int
	}
	var offenders []finding
	maxID, maxSpan := 0, 0
	checked := 0
	for _, row := range korean.Entries {
		item, ok := source.Find(row.ID)
		if !ok {
			t.Fatalf("Korean message %d lacks canonical source", row.ID)
		}
		replacement := message.KoreanRecord{Text: row.Korean, Layout: row.Layout}
		if layout := layouts[row.ID]; layout != "" {
			replacement.Layout = layout
		}
		raw, err := message.MaterializeKoreanRecordForScannerAudit(item.Record, replacement, mapping)
		if err != nil {
			t.Fatalf("message %d materialization: %v", row.ID, err)
		}
		metrics, err := message.AnalyzeRetailStringScanner(raw)
		if err != nil {
			t.Fatalf("message %d scanner analysis: %v", row.ID, err)
		}
		checked++
		if metrics.MaxSpan > maxSpan {
			maxID, maxSpan = row.ID, metrics.MaxSpan
		}
		if metrics.MaxSpan >= 0x100 {
			offenders = append(offenders, finding{id: row.ID, span: metrics.MaxSpan})
		}
	}

	sort.Slice(offenders, func(i, j int) bool {
		if offenders[i].span != offenders[j].span {
			return offenders[i].span > offenders[j].span
		}
		return offenders[i].id < offenders[j].id
	})
	t.Logf("FORENSIC KOREAN_SCANNER_SPAN_SUMMARY canonical=%d materializable=%d structural_skipped=%d checked=%d derived_c5_layouts=%d max_id=%d max_span=%d offenders_ge_0x100=%d",
		len(canonical.Entries), len(korean.Entries), skippedStructural, checked, derived, maxID, maxSpan, len(offenders))
	if len(offenders) == 0 {
		return
	}
	limit := len(offenders)
	if limit > 50 {
		limit = 50
	}
	lines := make([]string, 0, limit)
	for _, f := range offenders[:limit] {
		lines = append(lines, fmt.Sprintf("id=%d max_span=%d (0x%X)", f.id, f.span, f.span))
	}
	t.Fatalf("current compiler materialization has %d record(s) with retail-scanner max span >= 0x100; top findings:\n%s",
		len(offenders), strings.Join(lines, "\n"))
}
