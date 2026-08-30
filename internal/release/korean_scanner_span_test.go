// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
)

var scannerAuditControl = regexp.MustCompile(`<(?:if|select|call:[0-9]+|jump:[0-9]+|value:\$[0-9A-F]{2}|add|subtract|multiply|divide|modulo|equal|not-equal|less|greater|less-equal|greater-equal|and|or|operator:\$[0-9A-F]{2}:\$[0-9A-F]{2}|color:[^<>]|discard:[^<>]:\$[0-9A-F]{2}|escape:\$[0-9A-F]{2}|end|separator|backspace|tab|line-break|\$[0-9A-F]{2})>`)

const scannerAudit210065Layout = "광대한 대지 바이아시온 대륙.<line-break>너무나 넓어 지도에도 기록되지<line-break>않고 여행자에게조차 알려지지 않은<line-break>작은 마을이 있다…. 마을의 이름은<line-break>미이스. 그곳에는 작은 신전과 숲,<line-break>그리고 평온한 일상 정도뿐이었다.<line-break>위대한 혼의 이야기는<line-break>여기서 시작된다…….<end>"

// TestCurrentKoreanCorpusRetailScannerMaxSpanBelowInlineBoundary is a non-vacuous,
// repository-only census of every accepted Korean record. Retail Record.Tokens
// do not exist until authenticated banks are bound, so this test deliberately
// computes a conservative upper bound from canonical annotated Korean instead
// of pretending to perform exact retail materialization without retail assets.
//
// Literal text is encoded exactly with the Korean renderer width (custom glyphs
// are two bytes). Source-owned implicit line breaks/kana ESC controls are omitted;
// both omissions can only make this upper bound larger because line breaks reset
// the scanner span and ESC controls make z_un_089661DC skip bytes. Visible fixed
// controls are charged at their exact encoded widths where known. The production
// compiler independently performs the exact post-materialization scanner check
// once authenticated retail banks are available on the user's ISO.
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
	if len(canonical.Entries) == 0 {
		t.Fatal("canonical Korean corpus is empty")
	}

	texts, err := canonical.RuntimeTexts(source)
	if err != nil {
		t.Fatal(err)
	}
	mapping := make(koreanslots.Mapping)
	for _, r := range koreanslots.RequiredCustomRunes(texts) {
		mapping[r] = cp932.GlyphKey(0x8140)
	}

	type finding struct {
		id   int
		span int
	}
	var offenders []finding
	maxID, maxSpan := 0, 0
	checked := 0
	for _, row := range canonical.Entries {
		text := row.Korean
		if row.Layout != "" {
			text = row.Layout
		}
		// This is a compiler-owned diagnostic layout and therefore must be part of
		// the repository census even though it is not translator-owned canonical layout.
		if row.ID == 210065 {
			text = scannerAudit210065Layout
		}
		// The production diagnostic inserts one literal space after value:$15.
		if row.ID == 10010 {
			text = strings.Replace(text, "<value:$15>", "<value:$15> ", 1)
		}
		span, err := conservativeAnnotatedScannerSpan(text, mapping)
		if err != nil {
			t.Fatalf("message %d scanner upper-bound analysis: %v", row.ID, err)
		}
		checked++
		if span > maxSpan {
			maxID, maxSpan = row.ID, span
		}
		if span >= 0x100 {
			offenders = append(offenders, finding{id: row.ID, span: span})
		}
	}

	if checked != len(canonical.Entries) {
		t.Fatalf("scanner-span census incomplete: checked=%d canonical=%d", checked, len(canonical.Entries))
	}

	sort.Slice(offenders, func(i, j int) bool {
		if offenders[i].span != offenders[j].span {
			return offenders[i].span > offenders[j].span
		}
		return offenders[i].id < offenders[j].id
	})
	t.Logf("FORENSIC KOREAN_SCANNER_SPAN_SUMMARY canonical=%d checked=%d method=conservative_source_independent_upper_bound max_id=%d max_span=%d offenders_ge_0x100=%d",
		len(canonical.Entries), checked, maxID, maxSpan, len(offenders))
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
	t.Fatalf("Korean corpus has %d record(s) whose conservative scanner-span upper bound reaches 0x100; top findings:\n%s",
		len(offenders), strings.Join(lines, "\n"))
}

func conservativeAnnotatedScannerSpan(text string, mapping koreanslots.Mapping) (int, error) {
	maxSpan, span := 0, 0
	finish := func() {
		if span > maxSpan {
			maxSpan = span
		}
		span = 0
	}
	addLiteral := func(s string) error {
		if s == "" {
			return nil
		}
		raw, err := koreanslots.Encode(s, mapping)
		if err != nil {
			return err
		}
		span += len(raw)
		return nil
	}

	matches := scannerAuditControl.FindAllStringIndex(text, -1)
	cursor := 0
	for _, m := range matches {
		if err := addLiteral(text[cursor:m[0]]); err != nil {
			return 0, err
		}
		tag := text[m[0]:m[1]]
		switch {
		case tag == "<line-break>":
			finish()
		case tag == "<end>":
			// Retail block terminator is 05 05 05. Counting it is conservative;
			// the later NUL/suffix terminates exact materialization.
			span += 3
			finish()
		case strings.HasPrefix(tag, "<color:"), strings.HasPrefix(tag, "<discard:"), strings.HasPrefix(tag, "<escape:"):
			// These lower to ESC controls. The captured scanner skips the ESC and
			// following bytes, so charging zero while not skipping any adjacent
			// literal bytes is an upper bound on ordinary-byte span.
		case strings.HasPrefix(tag, "<call:"), strings.HasPrefix(tag, "<jump:"):
			span += 4
		case tag == "<if>", tag == "<select>", strings.HasPrefix(tag, "<value:$"),
			tag == "<add>", tag == "<subtract>", tag == "<multiply>", tag == "<divide>", tag == "<modulo>",
			tag == "<equal>", tag == "<not-equal>", tag == "<less>", tag == "<greater>", tag == "<less-equal>",
			tag == "<greater-equal>", tag == "<and>", tag == "<or>", strings.HasPrefix(tag, "<operator:$"):
			span += 2
		case tag == "<separator>", tag == "<backspace>", tag == "<tab>", regexp.MustCompile(`^<\$[0-9A-F]{2}>$`).MatchString(tag):
			span++
		default:
			// Fail closed if the runtime tag vocabulary grows without updating this audit.
			return 0, fmt.Errorf("unhandled runtime control %q", tag)
		}
		cursor = m[1]
	}
	if err := addLiteral(text[cursor:]); err != nil {
		return 0, err
	}
	finish()
	return maxSpan, nil
}
