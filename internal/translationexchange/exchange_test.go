// SPDX-License-Identifier: GPL-3.0-or-later

package translationexchange

import (
	"bytes"
	"strings"
	"testing"
)

func sampleSource() []ExportRow {
	return []ExportRow{
		{
			Schema: Schema, ID: 100, Section: 0, RecordIndex: 100,
			SourceFile: "translations/messages/msgsec000.toml",
			Japanese: "名は<value:$28>だ。<end>", EnglishReference: "The name is <value:$28>.<end>",
		},
		{
			Schema: Schema, ID: 101, Section: 0, RecordIndex: 101,
			SourceFile: "translations/messages/msgsec000.toml",
			Japanese: "%sを選べ<select>", EnglishReference: "Choose %s<select>",
		},
	}
}

func sampleResults() []ResultRow {
	return []ResultRow{
		{ID: 100, Korean: "이름은 <value:$28>이다.<end>", GlossaryCandidates: []GlossaryCandidate{}},
		{ID: 101, Korean: "%s을 선택해<select>", GlossaryCandidates: []GlossaryCandidate{}},
	}
}

func TestExportAndResultRoundTrip(t *testing.T) {
	var sourceBuffer bytes.Buffer
	if err := WriteExport(&sourceBuffer, sampleSource()); err != nil {
		t.Fatal(err)
	}
	loadedSource, err := ReadExport(strings.NewReader(sourceBuffer.String()))
	if err != nil {
		t.Fatal(err)
	}
	if len(loadedSource) != 2 || loadedSource[0].ID != 100 || loadedSource[1].ID != 101 {
		t.Fatalf("loaded source = %#v", loadedSource)
	}

	var resultBuffer bytes.Buffer
	if err := WriteResults(&resultBuffer, sampleResults()); err != nil {
		t.Fatal(err)
	}
	loadedResults, err := ReadResults(strings.NewReader(resultBuffer.String()))
	if err != nil {
		t.Fatal(err)
	}
	validation, err := Validate(loadedSource, loadedResults)
	if err != nil {
		t.Fatal(err)
	}
	if len(validation.Rows) != 2 || validation.Uncertain != 0 || len(validation.Warnings) != 0 {
		t.Fatalf("validation = %#v", validation)
	}
}

func TestValidateRejectsChangedControls(t *testing.T) {
	results := sampleResults()
	results[0].Korean = "이름은 <value:$15>이다.<end>"
	if _, err := Validate(sampleSource(), results); err == nil || !strings.Contains(err.Error(), "control tokens changed") {
		t.Fatalf("changed control returned %v", err)
	}

	results = sampleResults()
	results[1].Korean = "%d을 선택해<select>"
	if _, err := Validate(sampleSource(), results); err == nil || !strings.Contains(err.Error(), "printf tokens changed") {
		t.Fatalf("changed printf returned %v", err)
	}
}

func TestValidateRejectsCountOrderAndEmptyTranslation(t *testing.T) {
	if _, err := Validate(sampleSource(), sampleResults()[:1]); err == nil || !strings.Contains(err.Error(), "result count") {
		t.Fatalf("short result returned %v", err)
	}

	results := sampleResults()
	results[0], results[1] = results[1], results[0]
	if _, err := Validate(sampleSource(), results); err == nil || !strings.Contains(err.Error(), "order and IDs") {
		t.Fatalf("reordered result returned %v", err)
	}

	results = sampleResults()
	results[0].Korean = "   "
	if _, err := Validate(sampleSource(), results); err == nil || !strings.Contains(err.Error(), "korean is empty") {
		t.Fatalf("empty result returned %v", err)
	}
}

func TestReadResultsRejectsMarkdownUnknownFieldsAndBlankLines(t *testing.T) {
	for _, input := range []string{
		"```json\n{\"id\":100,\"korean\":\"가\",\"uncertain\":false,\"note\":\"\",\"glossary_candidates\":[]}\n```\n",
		"{\"id\":100,\"korean\":\"가\",\"uncertain\":false,\"note\":\"\",\"glossary_candidates\":[],\"extra\":1}\n",
		"{\"id\":100,\"korean\":\"가\",\"uncertain\":false,\"note\":\"\",\"glossary_candidates\":[]}\n\n",
	} {
		if _, err := ReadResults(strings.NewReader(input)); err == nil {
			t.Fatalf("invalid JSONL was accepted: %q", input)
		}
	}
}

func TestValidateReturnsWarningsInsteadOfFailingOnReviewFlags(t *testing.T) {
	results := sampleResults()
	results[0].Uncertain = true
	results[0].Note = "말투 확인 필요"
	results[1].Korean = "%s를 カナ로 적음<select>"
	validation, err := Validate(sampleSource(), results)
	if err != nil {
		t.Fatal(err)
	}
	if validation.Uncertain != 1 {
		t.Fatalf("uncertain = %d, want 1", validation.Uncertain)
	}
	if len(validation.Warnings) != 2 {
		t.Fatalf("warnings = %#v, want 2", validation.Warnings)
	}
}
