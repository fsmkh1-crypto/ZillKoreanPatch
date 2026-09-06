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

func TestKoreanDialogueCoverageAuditSeesExcludedOverflow(t *testing.T) {
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

	const id = 560650 // authenticated C5-only fixture; actual row uses bounded $28.
	base, ok := korean.Find(id)
	if !ok {
		t.Fatalf("missing Korean fixture %d", id)
	}
	if engine.narrowText(id) || (!engine.has(engine.consumers.C5IDs, id) && !engine.has(engine.consumers.C5PortraitIDs, id)) {
		t.Fatalf("fixture %d no longer represents a C5-only consumer", id)
	}

	// Replace only the synthetic test row: $15 is deliberately unbounded for C5
	// static derivation, and the visible prefix is deliberately too wide. The
	// derivation residual audit must skip it, while whole coverage must still
	// report the unsafe static width.
	base.Korean = strings.Repeat("가나다라마바사 ", 5) + "<value:$15><end>"
	base.Layout = ""
	mini := &corpus.KoreanProject{Entries: []corpus.KoreanEntry{base}}
	mapping := koreanslots.Mapping{}
	for _, r := range base.Korean {
		if r > 0x7f {
			mapping[r] = cp932.GlyphKey(0xAC82)
		}
	}

	checked, residual, err := engine.AuditKoreanEnglishDialogueResiduals(source, mini, nil, mapping)
	if err != nil {
		t.Fatal(err)
	}
	if checked != 0 || len(residual) != 0 {
		t.Fatalf("derivation residual must skip synthetic unbounded C5 row: checked=%d residual=%v", checked, residual)
	}

	audit, err := engine.AuditKoreanEnglishDialogueCoverage(source, mini, nil, mapping)
	if err != nil {
		t.Fatal(err)
	}
	if audit.Relevant != 1 || audit.DerivationEligible != 0 || len(audit.Excluded) != 1 {
		t.Fatalf("whole coverage did not retain excluded row: %+v", audit)
	}
	if audit.Excluded[0].Reason != "unbounded_inline_substitution" {
		t.Fatalf("excluded reason=%q, want unbounded_inline_substitution", audit.Excluded[0].Reason)
	}
	if len(audit.ExcludedOverflowIDs) != 1 || audit.ExcludedOverflowIDs[0] != id {
		t.Fatalf("whole coverage failed to expose excluded overflow: %+v", audit)
	}
}
