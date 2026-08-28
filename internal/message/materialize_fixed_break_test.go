// SPDX-License-Identifier: GPL-3.0-or-later

package message

import (
	"bytes"
	"testing"
)

func TestSplitSemanticTreatsFixedLineBreakAsSourceOwned(t *testing.T) {
	p := &Projection{
		RecordID: 950059,
		Fragments: []Fragment{
			{Key: "case_194"},
			{Key: "case_191"},
		},
		nodes: []projectionNode{
			{fragment: 0},
			{fixed: true, kind: "line_break", display: "<line-break>", raw: []byte{10}},
			{fixed: true, kind: "if", display: "<if>", raw: []byte{3}},
			{fragment: 1},
		},
	}

	values, err := p.splitSemanticWith("사룡 샨마!<if>아직 죽지 않았구나", func(int, string, string) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 2 || values[0] != "사룡 샨마!" || values[1] != "아직 죽지 않았구나" {
		t.Fatalf("split values = %#v", values)
	}

	got, err := p.materializeValues(values, false, func(text string) ([]byte, error) { return []byte(text), nil })
	if err != nil {
		t.Fatal(err)
	}
	want := append([]byte("사룡 샨마!"), 10, 3)
	want = append(want, []byte("아직 죽지 않았구나")...)
	if !bytes.Equal(got, want) {
		t.Fatalf("materialized bytes = % X, want % X", got, want)
	}
}
