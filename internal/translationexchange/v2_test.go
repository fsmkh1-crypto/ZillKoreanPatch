// SPDX-License-Identifier: GPL-3.0-or-later

package translationexchange

import (
	"bytes"
	"strings"
	"testing"
)

func TestV2NeverExportsControlsAndReconstructsExactly(t *testing.T) {
	japanese := "<value:$28>。<line-break>ノエルの求めに応じて、今すぐに<line-break>　<line-break>竜王の島へ出向く気があるか？<end>"
	source := BuildSourceV2(1870003, 3, "translations/messages/msgsec187.toml", japanese, "<value:$28>. Go now.<end>", nil)
	if got, want := len(source.Export.Segments), 3; got != want {
		t.Fatalf("segments = %d, want %d: %#v", got, want, source.Export.Segments)
	}
	for _, segment := range source.Export.Segments {
		if protectedV2RE.MatchString(segment.Text) {
			t.Fatalf("protected token leaked in segment %#v", segment)
		}
	}
	if protectedV2RE.MatchString(source.Export.FullText) || protectedV2RE.MatchString(source.Export.EnglishReference) {
		t.Fatalf("protected token leaked into context/reference: %#v", source.Export)
	}
	translated := []Segment{
		{Index: 0, Text: "."},
		{Index: 1, Text: "노엘의 요청에 응해, 지금 당장"},
		{Index: 2, Text: "용왕의 섬으로 갈 생각이 있나?"},
	}
	got, err := ReconstructV2(source, translated)
	if err != nil {
		t.Fatal(err)
	}
	want := "<value:$28>.<line-break>노엘의 요청에 응해, 지금 당장<line-break>　<line-break>용왕의 섬으로 갈 생각이 있나?<end>"
	if got != want {
		t.Fatalf("reconstructed = %q, want %q", got, want)
	}
}

func TestV2ResultStrictJSONContract(t *testing.T) {
	valid := "{\"id\":1,\"korean_segments\":[{\"index\":0,\"text\":\"가\"}],\"uncertain\":false,\"note\":\"\",\"glossary_candidates\":[]}\n"
	rows, err := ReadResultsV2(strings.NewReader(valid))
	if err != nil || len(rows) != 1 {
		t.Fatalf("valid result: rows=%#v err=%v", rows, err)
	}
	for _, input := range []string{
		"```json\n" + valid + "```\n",
		"{\"id\":1,\"korean_segments\":null,\"uncertain\":false,\"note\":\"\",\"glossary_candidates\":[]}\n",
		"{\"id\":1,\"korean_segments\":[],\"uncertain\":false,\"note\":\"\",\"glossary_candidates\":null}\n",
		"{\"id\":1,\"korean_segments\":[],\"note\":\"\",\"glossary_candidates\":[]}\n",
		"{\"id\":1,\"korean_segments\":[],\"uncertain\":false,\"note\":\"\",\"glossary_candidates\":[],\"extra\":1}\n",
		valid + "\n",
	} {
		if _, err := ReadResultsV2(strings.NewReader(input)); err == nil {
			t.Fatalf("invalid v2 JSONL was accepted: %q", input)
		}
	}
}

func TestV2ValidateRejectsSegmentDriftAndProtectedTokens(t *testing.T) {
	source := []SourceRowV2{BuildSourceV2(1, 1, "x.toml", "前<line-break>後<end>", "before after<end>", nil)}
	base := ResultRowV2{
		ID: 1,
		KoreanSegments: []Segment{{Index: 0, Text: "앞"}, {Index: 1, Text: "뒤"}},
		GlossaryCandidates: []GlossaryCandidate{},
	}
	validation, err := ValidateV2(source, []ResultRowV2{base})
	if err != nil {
		t.Fatal(err)
	}
	if got := validation.Rows[0].Korean; got != "앞<line-break>뒤<end>" {
		t.Fatalf("accepted Korean = %q", got)
	}

	bad := base
	bad.KoreanSegments = bad.KoreanSegments[:1]
	if _, err := ValidateV2(source, []ResultRowV2{bad}); err == nil || !strings.Contains(err.Error(), "length") {
		t.Fatalf("short segment array returned %v", err)
	}

	bad = base
	bad.KoreanSegments = []Segment{{Index: 1, Text: "앞"}, {Index: 0, Text: "뒤"}}
	if _, err := ValidateV2(source, []ResultRowV2{bad}); err == nil || !strings.Contains(err.Error(), "index") {
		t.Fatalf("reordered segments returned %v", err)
	}

	bad = base
	bad.KoreanSegments = []Segment{{Index: 0, Text: "앞<end>"}, {Index: 1, Text: "뒤"}}
	if _, err := ValidateV2(source, []ResultRowV2{bad}); err == nil || !strings.Contains(err.Error(), "protected token") {
		t.Fatalf("control injection returned %v", err)
	}
}

func TestV2ExportIsControlFreeJSONL(t *testing.T) {
	source := BuildSourceV2(1, 1, "x.toml", "가짜가아닌原文<end>", "Reference<end>", map[string]string{})
	var output bytes.Buffer
	if err := WriteExportV2(&output, []ExportRowV2{source.Export}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "<end>") {
		t.Fatalf("export leaked control: %s", output.String())
	}
	loaded, err := ReadExportV2(strings.NewReader(output.String()))
	if err != nil || len(loaded) != 1 || !SameExportV2(loaded[0], source.Export) {
		t.Fatalf("round trip loaded=%#v err=%v", loaded, err)
	}
}
