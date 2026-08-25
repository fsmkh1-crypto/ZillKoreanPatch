// SPDX-License-Identifier: GPL-3.0-or-later

// Package translationexchange defines the narrow JSON Lines contract used to
// exchange Korean translation batches with external LLMs. It deliberately
// keeps repository/TOML mutation out of the exchange layer: a returned batch
// must first pass structural validation before any later import step can use it.
package translationexchange

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"
)

const Schema = "zill-gemini-v1"

// ExportRow is one source record sent to Gemini.
type ExportRow struct {
	Schema           string  `json:"schema"`
	ID               int     `json:"id"`
	Section          int     `json:"section"`
	RecordIndex      int     `json:"record_index"`
	SourceFile       string  `json:"source_file"`
	Japanese         string  `json:"japanese"`
	EnglishReference string  `json:"english_reference"`
	Speaker          *string `json:"speaker"`
	Context          *string `json:"context"`
}

// GlossaryCandidate is an optional terminology suggestion returned by Gemini.
// Candidates are advisory only; callers must not automatically promote them to
// the canonical glossary.
type GlossaryCandidate struct {
	Source string `json:"source"`
	Korean string `json:"korean"`
	Type   string `json:"type"`
}

// ResultRow is the only accepted Gemini response shape.
type ResultRow struct {
	ID                 int                 `json:"id"`
	Korean             string              `json:"korean"`
	Uncertain          bool                `json:"uncertain"`
	Note               string              `json:"note"`
	GlossaryCandidates []GlossaryCandidate `json:"glossary_candidates"`
}

// Validation summarizes a successfully checked result batch.
type Validation struct {
	Rows      []ResultRow
	Warnings  []string
	Uncertain int
}

var (
	angleTokenRE = regexp.MustCompile(`<[^<>\r\n]+>`)
	braceTokenRE = regexp.MustCompile(`\{\{[^{}\r\n]+\}\}`)
	printfRE     = regexp.MustCompile(`%[-+ #0]*[0-9]*(\.[0-9]+)?[A-Za-z%]`)
)

// WriteExport writes one compact JSON object per line with HTML escaping
// disabled so Japanese text stays readable in ordinary text editors.
func WriteExport(w io.Writer, rows []ExportRow) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	for index, row := range rows {
		if row.Schema != Schema {
			return fmt.Errorf("export row %d: schema %q, want %q", index+1, row.Schema, Schema)
		}
		if row.ID < 0 {
			return fmt.Errorf("export row %d: negative ID %d", index+1, row.ID)
		}
		if row.SourceFile == "" {
			return fmt.Errorf("export row %d ID %d: source_file is empty", index+1, row.ID)
		}
		if err := encoder.Encode(row); err != nil {
			return fmt.Errorf("encode export row %d ID %d: %w", index+1, row.ID, err)
		}
	}
	return nil
}

// ReadExport decodes a strict JSONL source batch. Blank lines, unknown fields,
// trailing JSON values, duplicate IDs, and schema mismatches are rejected.
func ReadExport(r io.Reader) ([]ExportRow, error) {
	var rows []ExportRow
	seen := make(map[int]struct{})
	err := scanJSONL(r, func(line int, data []byte) error {
		var row ExportRow
		if err := decodeStrictLine(data, &row); err != nil {
			return fmt.Errorf("source line %d: %w", line, err)
		}
		if row.Schema != Schema {
			return fmt.Errorf("source line %d ID %d: schema %q, want %q", line, row.ID, row.Schema, Schema)
		}
		if row.ID < 0 {
			return fmt.Errorf("source line %d: negative ID %d", line, row.ID)
		}
		if row.SourceFile == "" {
			return fmt.Errorf("source line %d ID %d: source_file is empty", line, row.ID)
		}
		if _, exists := seen[row.ID]; exists {
			return fmt.Errorf("source line %d: duplicate ID %d", line, row.ID)
		}
		seen[row.ID] = struct{}{}
		rows = append(rows, row)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("source batch is empty")
	}
	return rows, nil
}

// ReadResults decodes Gemini's strict response shape. No markdown fences or
// explanatory prose are accepted because every nonempty line must be exactly
// one JSON object.
func ReadResults(r io.Reader) ([]ResultRow, error) {
	var rows []ResultRow
	seen := make(map[int]struct{})
	err := scanJSONL(r, func(line int, data []byte) error {
		var row ResultRow
		if err := decodeStrictLine(data, &row); err != nil {
			return fmt.Errorf("result line %d: %w", line, err)
		}
		if row.ID < 0 {
			return fmt.Errorf("result line %d: negative ID %d", line, row.ID)
		}
		if _, exists := seen[row.ID]; exists {
			return fmt.Errorf("result line %d: duplicate ID %d", line, row.ID)
		}
		seen[row.ID] = struct{}{}
		rows = append(rows, row)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("result batch is empty")
	}
	return rows, nil
}

// Validate compares one result batch against the exact exported source batch.
// It is intentionally fail-closed on IDs and machine-readable control tokens.
func Validate(source []ExportRow, results []ResultRow) (Validation, error) {
	if len(source) == 0 {
		return Validation{}, fmt.Errorf("source batch is empty")
	}
	if len(source) != len(results) {
		return Validation{}, fmt.Errorf("result count %d, want exactly %d", len(results), len(source))
	}

	validation := Validation{Rows: append([]ResultRow(nil), results...)}
	for index := range source {
		src := source[index]
		got := results[index]
		if got.ID != src.ID {
			return Validation{}, fmt.Errorf("result position %d has ID %d, want %d; order and IDs must be preserved", index+1, got.ID, src.ID)
		}
		if strings.TrimSpace(got.Korean) == "" {
			return Validation{}, fmt.Errorf("ID %d: korean is empty", got.ID)
		}
		if err := validateProtectedTokens(src.Japanese, got.Korean); err != nil {
			return Validation{}, fmt.Errorf("ID %d: %w", got.ID, err)
		}
		if got.Uncertain {
			validation.Uncertain++
			validation.Warnings = append(validation.Warnings, fmt.Sprintf("ID %d: Gemini marked translation uncertain", got.ID))
		}
		if containsJapaneseKana(got.Korean) {
			validation.Warnings = append(validation.Warnings, fmt.Sprintf("ID %d: Korean result still contains Japanese kana", got.ID))
		}
		for candidateIndex, candidate := range got.GlossaryCandidates {
			if strings.TrimSpace(candidate.Source) == "" || strings.TrimSpace(candidate.Korean) == "" {
				return Validation{}, fmt.Errorf("ID %d: glossary_candidates[%d] requires nonempty source and korean", got.ID, candidateIndex)
			}
		}
	}
	return validation, nil
}

// WriteResults writes a normalized accepted result batch.
func WriteResults(w io.Writer, rows []ResultRow) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	for index, row := range rows {
		if err := encoder.Encode(row); err != nil {
			return fmt.Errorf("encode result row %d ID %d: %w", index+1, row.ID, err)
		}
	}
	return nil
}

func validateProtectedTokens(source, translated string) error {
	checks := []struct {
		name string
		re   *regexp.Regexp
	}{
		{name: "angle-bracket control tokens", re: angleTokenRE},
		{name: "double-brace placeholders", re: braceTokenRE},
		{name: "printf tokens", re: printfRE},
	}
	for _, check := range checks {
		want := check.re.FindAllString(source, -1)
		got := check.re.FindAllString(translated, -1)
		if !equalStrings(want, got) {
			return fmt.Errorf("%s changed: source=%q result=%q", check.name, want, got)
		}
	}
	return nil
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for index := range a {
		if a[index] != b[index] {
			return false
		}
	}
	return true
}

func containsJapaneseKana(text string) bool {
	for _, r := range text {
		if unicode.In(r, unicode.Hiragana, unicode.Katakana) {
			return true
		}
	}
	return false
}

func scanJSONL(r io.Reader, visit func(line int, data []byte) error) error {
	scanner := bufio.NewScanner(r)
	// Some game records can be unusually large; do not inherit Scanner's 64 KiB
	// default and accidentally reject a valid batch line.
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	line := 0
	for scanner.Scan() {
		line++
		data := scanner.Bytes()
		if len(strings.TrimSpace(string(data))) == 0 {
			return fmt.Errorf("line %d is blank; JSONL batches must contain exactly one object per line", line)
		}
		if err := visit(line, data); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan JSONL: %w", err)
	}
	return nil
}

func decodeStrictLine(data []byte, value any) error {
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return fmt.Errorf("invalid JSON object: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("line contains more than one JSON value")
		}
		return fmt.Errorf("trailing JSON content: %w", err)
	}
	return nil
}
