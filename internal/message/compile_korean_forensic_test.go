// SPDX-License-Identifier: GPL-3.0-or-later

package message

import (
	"strings"
	"testing"
)

func TestKoreanForensicWatchlist(t *testing.T) {
	for _, id := range []int{640001, 640003, 640012} {
		if !koreanForensicWatchID(id) {
			t.Fatalf("ID %d should be watched", id)
		}
	}
	for _, id := range []int{640000, 640013, 10010} {
		if koreanForensicWatchID(id) {
			t.Fatalf("ID %d should not be watched", id)
		}
	}
}

func TestKoreanForensicLineReportsDerivedAndMaterializedBreaks(t *testing.T) {
	replacement := KoreanRecord{
		Text:   "네놈, 하녀 주제에 내게 명령할 셈이냐! 이 내가 티아나를 만나러 왔다. 비켜!<end>",
		Layout: "네놈, 하녀 주제에 내게 명령할 셈이냐! 이 내가 티아나를<line-break>만나러 왔다. 비켜!<end>",
	}
	line := koreanForensicLine(640003, replacement, []byte{0x82, 0xAC, 0x0A, 0x82, 0xAD, 0x00})
	for _, want := range []string{
		"FORENSIC_DIALOGUE id=640003",
		"canonical=\"네놈, 하녀 주제에 내게 명령할 셈이냐! 이 내가 티아나를 만나러 왔다. 비켜!<end>\"",
		"selected_layout=\"네놈, 하녀 주제에 내게 명령할 셈이냐! 이 내가 티아나를<line-break>만나러 왔다. 비켜!<end>\"",
		"derived_breaks=1",
		"materialized_0A=1",
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("forensic line %q does not contain %q", line, want)
		}
	}
}

func TestKoreanForensicLineUsesCanonicalWhenNoDerivedLayout(t *testing.T) {
	replacement := KoreanRecord{Text: "짧은 대사<end>"}
	line := koreanForensicLine(640001, replacement, []byte{0x82, 0xAC, 0x00})
	for _, want := range []string{
		"canonical=\"짧은 대사<end>\"",
		"selected_layout=\"짧은 대사<end>\"",
		"derived_breaks=0",
		"materialized_0A=0",
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("forensic line %q does not contain %q", line, want)
		}
	}
}
