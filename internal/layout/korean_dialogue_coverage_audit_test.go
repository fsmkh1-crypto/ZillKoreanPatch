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

func TestKoreanDialogueCoverageSeparatesStaticReflowFromRuntimePending(t *testing.T) {
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

	// Replace only the synthetic test row. $15 deliberately lacks a proven
	// runtime-width bound, while the visible Korean prefix deliberately exceeds
	// the C5 static limit. English sourceAware still reflows such fragments; the
	// Korean parity path must therefore fix the static layout and retain a
	// separate RuntimePending result rather than excluding the row from reflow.
	base.Korean = strings.Repeat("가나다라마바사 ", 5) + "<value:$15><end>"
	base.Layout = ""
	mini := &corpus.KoreanProject{Entries: []corpus.KoreanEntry{base}}
	mapping := koreanslots.Mapping{}
	for _, r := range base.Korean {
		if r > 0x7f {
			mapping[r] = cp932.GlyphKey(0xAC82)
		}
	}

	layouts, derived, err := engine.DeriveKoreanEnglishDialogueLayouts(source, mini, nil, mapping)
	if err != nil {
		t.Fatal(err)
	}
	if derived != 1 {
		t.Fatalf("synthetic unbounded C5 row derived=%d, want 1", derived)
	}
	if got := layouts[id]; got == "" || !strings.Contains(got, lineBreak) {
		t.Fatalf("synthetic unbounded C5 row did not receive static source-aware reflow: %q", got)
	}

	checked, residual, err := engine.AuditKoreanEnglishDialogueResiduals(source, mini, layouts, mapping)
	if err != nil {
		t.Fatal(err)
	}
	if checked != 1 || len(residual) != 0 {
		t.Fatalf("static residual must include and pass synthetic unbounded C5 row: checked=%d residual=%v", checked, residual)
	}

	audit, err := engine.AuditKoreanEnglishDialogueCoverage(source, mini, layouts, mapping)
	if err != nil {
		t.Fatal(err)
	}
	if audit.Relevant != 1 || audit.DerivationEligible != 1 || len(audit.Excluded) != 0 {
		t.Fatalf("whole coverage misclassified static reflow population: %+v", audit)
	}
	if len(audit.OverflowIDs) != 0 || len(audit.ExcludedOverflowIDs) != 0 {
		t.Fatalf("static overflow remained after source-aware reflow: %+v", audit)
	}
	if len(audit.RuntimePending) != 1 || audit.RuntimePending[0].ID != id {
		t.Fatalf("runtime-width uncertainty was not preserved separately: %+v", audit)
	}
	if audit.RuntimePending[0].Reason != "unbounded_inline_substitution" {
		t.Fatalf("runtime pending reason=%q, want unbounded_inline_substitution", audit.RuntimePending[0].Reason)
	}
}
