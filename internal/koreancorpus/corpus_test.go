// SPDX-License-Identifier: GPL-3.0-or-later

package koreancorpus

import (
	"strings"
	"testing"
)

func TestParseSectionAcceptsSparseCanonicalRows(t *testing.T) {
	data := []byte(licenseLine + `

["10007"]
japanese = "汝、無限のソウルを持つ者よ"
korean = "테스트 성공"
layout = "테스트<line-break>성공"
`)
	records, err := parseSection(data, "msgsec001.toml", 1, map[int]string{
		10007: "汝、無限のソウルを持つ者よ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1", len(records))
	}
	if records[0].ID != 10007 || records[0].Text != "테스트 성공" || records[0].Layout != "테스트<line-break>성공" {
		t.Fatalf("unexpected record: %#v", records[0])
	}
}

func TestParseSectionRejectsStaleJapaneseSource(t *testing.T) {
	data := []byte(licenseLine + `

["10007"]
japanese = "다른 원문"
korean = "테스트 성공"
`)
	_, err := parseSection(data, "msgsec001.toml", 1, map[int]string{10007: "원문"})
	if err == nil || !strings.Contains(err.Error(), "japanese differs from canonical source") {
		t.Fatalf("got %v, want stale-source rejection", err)
	}
}

func TestParseSectionRejectsWrongSectionAndUnknownFields(t *testing.T) {
	wrongSection := []byte(licenseLine + `

["20001"]
japanese = "원문"
korean = "번역"
`)
	if _, err := parseSection(wrongSection, "msgsec001.toml", 1, map[int]string{20001: "원문"}); err == nil || !strings.Contains(err.Error(), "belongs to section") {
		t.Fatalf("got %v, want section rejection", err)
	}

	unknown := []byte(licenseLine + `

["10007"]
japanese = "원문"
korean = "번역"
extra = "drift"
`)
	if _, err := parseSection(unknown, "msgsec001.toml", 1, map[int]string{10007: "원문"}); err == nil || !strings.Contains(err.Error(), "invalid TOML") {
		t.Fatalf("got %v, want unknown-field rejection", err)
	}
}

func TestParseSectionRejectsRawControls(t *testing.T) {
	data := []byte(licenseLine + "\n\n[\"10007\"]\njapanese = \"원문\"\nkorean = \"테스트\\n성공\"\n")
	_, err := parseSection(data, "msgsec001.toml", 1, map[int]string{10007: "원문"})
	if err == nil || !strings.Contains(err.Error(), "raw Unicode control") {
		t.Fatalf("got %v, want raw-control rejection", err)
	}
}

func TestParseSectionEnforcesFixedControlContract(t *testing.T) {
	source := "<value:$15>よ<line-break>答えよ<end>"
	accepted := []byte(licenseLine + `

["10010"]
japanese = "<value:$15>よ<line-break>答えよ<end>"
korean = "<value:$15>여 답하라<end>"
layout = "<value:$15>여<line-break>답하라<end>"
`)
	if _, err := parseSection(accepted, "msgsec001.toml", 1, map[int]string{10010: source}); err != nil {
		t.Fatalf("valid control contract rejected: %v", err)
	}

	changedValue := []byte(licenseLine + `

["10010"]
japanese = "<value:$15>よ<line-break>答えよ<end>"
korean = "<value:$16>여 답하라<end>"
`)
	if _, err := parseSection(changedValue, "msgsec001.toml", 1, map[int]string{10010: source}); err == nil || !strings.Contains(err.Error(), "changes fixed control sequence") {
		t.Fatalf("got %v, want changed-value rejection", err)
	}

	missingEnd := []byte(licenseLine + `

["10010"]
japanese = "<value:$15>よ<line-break>答えよ<end>"
korean = "<value:$15>여 답하라"
`)
	if _, err := parseSection(missingEnd, "msgsec001.toml", 1, map[int]string{10010: source}); err == nil || !strings.Contains(err.Error(), "changes fixed control sequence") {
		t.Fatalf("got %v, want missing-end rejection", err)
	}
}

func TestSectionFromNameAcceptsCanonicalShardsAndRejectsDrift(t *testing.T) {
	for _, name := range []string{"msgsec001.toml", "msgsec278.toml", "msgsec001-part01.toml", "msgsec001-part99.toml", "msgsec001b.toml"} {
		section, ok := sectionFromName(name)
		if !ok || section != map[string]int{"msgsec278.toml": 278}[name] && name == "msgsec278.toml" {
			t.Fatalf("%s rejected or misparsed: section=%d ok=%v", name, section, ok)
		}
		if name != "msgsec278.toml" && section != 1 {
			t.Fatalf("%s section=%d, want 1", name, section)
		}
	}
	for _, name := range []string{"msgsec1.toml", "msgsec279.toml", "msgsec001-part00.toml", "msgsec001-part1.toml", "msgsec001.txt", "notes.toml"} {
		if _, ok := sectionFromName(name); ok {
			t.Fatalf("%s accepted", name)
		}
	}
}
