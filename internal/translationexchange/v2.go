// SPDX-License-Identifier: GPL-3.0-or-later

package translationexchange

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

const SchemaV2 = "zill-gemini-v2"

// Segment is one translatable source-text fragment. Index is stable inside one
// record and must be returned unchanged by the external translator.
type Segment struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
}

// LockedPart is repository-owned non-translatable text removed before a record
// is sent to an external translator. It includes game controls/placeholders and
// exact whitespace-only runs. It is never part of the Gemini payload.
type LockedPart struct {
	AfterSegment int    `json:"after_segment"`
	Text         string `json:"text"`
}

// ExportRowV2 is the external Gemini payload. It deliberately contains no game
// controls. FullText is a control-free context aid and Segments are the only
// fields Gemini translates.
type ExportRowV2 struct {
	Schema           string            `json:"schema"`
	ID               int               `json:"id"`
	Section          int               `json:"section"`
	RecordIndex      int               `json:"record_index"`
	SourceFile       string            `json:"source_file"`
	FullText         string            `json:"full_text"`
	EnglishReference string            `json:"english_reference"`
	Glossary         map[string]string `json:"glossary"`
	Segments         []Segment         `json:"segments"`
}

// SourceRowV2 is the maintainer-side representation. Locked is intentionally
// omitted from ExportRowV2 so the external model cannot alter controls.
type SourceRowV2 struct {
	Export ExportRowV2  `json:"export"`
	Locked []LockedPart `json:"locked"`
}

// ResultRowV2 is the only accepted external response shape.
type ResultRowV2 struct {
	ID                 int                 `json:"id"`
	KoreanSegments     []Segment           `json:"korean_segments"`
	Uncertain          bool                `json:"uncertain"`
	Note               string              `json:"note"`
	GlossaryCandidates []GlossaryCandidate `json:"glossary_candidates"`
}

// AcceptedRowV2 is a verified result plus the reconstructed Korean canonical
// text. Reconstructed contains the exact locked source controls/whitespace.
type AcceptedRowV2 struct {
	ID                 int                 `json:"id"`
	Korean             string              `json:"korean"`
	Uncertain          bool                `json:"uncertain"`
	Note               string              `json:"note"`
	GlossaryCandidates []GlossaryCandidate `json:"glossary_candidates"`
}

// ValidationV2 summarizes a successfully checked v2 result batch.
type ValidationV2 struct {
	Rows      []AcceptedRowV2
	Warnings  []string
	Uncertain int
}

var protectedV2RE = regexp.MustCompile(`(<[^<>\r\n]+>|\{\{[^{}\r\n]+\}\}|%[-+ #0]*[0-9]*(?:\.[0-9]+)?[A-Za-z%])`)

// SplitV2 removes every protected token and whitespace-only run from the
// external payload while retaining enough information for exact reconstruction.
// Nonempty natural-language chunks become indexed segments.
func SplitV2(text string) ([]Segment, []LockedPart) {
	var segments []Segment
	var locked []LockedPart
	position := 0
	for _, bounds := range protectedV2RE.FindAllStringIndex(text, -1) {
		appendV2Natural(text[position:bounds[0]], &segments, &locked)
		locked = append(locked, LockedPart{AfterSegment: len(segments) - 1, Text: text[bounds[0]:bounds[1]]})
		position = bounds[1]
	}
	appendV2Natural(text[position:], &segments, &locked)
	return segments, locked
}

func appendV2Natural(text string, segments *[]Segment, locked *[]LockedPart) {
	if text == "" {
		return
	}
	// Keep leading/trailing whitespace repository-owned. This prevents an LLM
	// from collapsing alignment spaces while still exposing punctuation and
	// natural language as one coherent translatable fragment.
	start := 0
	for start < len(text) {
		r, size := runeAt(text, start)
		if !unicode.IsSpace(r) {
			break
		}
		start += size
	}
	end := len(text)
	for end > start {
		r, size := runeBefore(text, end)
		if !unicode.IsSpace(r) {
			break
		}
		end -= size
	}
	if start > 0 {
		*locked = append(*locked, LockedPart{AfterSegment: len(*segments) - 1, Text: text[:start]})
	}
	if end > start {
		*segments = append(*segments, Segment{Index: len(*segments), Text: text[start:end]})
	}
	if end < len(text) {
		*locked = append(*locked, LockedPart{AfterSegment: len(*segments) - 1, Text: text[end:]})
	}
}

func runeAt(s string, offset int) (rune, int) {
	for _, r := range s[offset:] {
		return r, len(string(r))
	}
	return 0, 0
}

func runeBefore(s string, end int) (rune, int) {
	lastStart := 0
	var last rune
	for offset, r := range s[:end] {
		lastStart = offset
		last = r
	}
	return last, end - lastStart
}

// ControlFreeContext produces the read-only full-text context sent to Gemini.
// Locked machine tokens are removed; line-break controls become spaces so
// adjacent clauses remain readable. Other controls are omitted.
func ControlFreeContext(text string) string {
	context := protectedV2RE.ReplaceAllStringFunc(text, func(token string) string {
		if token == "<line-break>" {
			return " "
		}
		return ""
	})
	return strings.Join(strings.Fields(context), " ")
}

// ReconstructV2 combines verified translated segments with repository-owned
// locked parts. Locked parts are ordered exactly as they appeared in source.
func ReconstructV2(source SourceRowV2, translated []Segment) (string, error) {
	if err := validateSegmentShape(source.Export.Segments, translated); err != nil {
		return "", err
	}
	lockedByAfter := make(map[int][]string)
	for _, part := range source.Locked {
		if part.AfterSegment < -1 || part.AfterSegment >= len(source.Export.Segments) {
			return "", fmt.Errorf("locked part after_segment %d outside segment range", part.AfterSegment)
		}
		lockedByAfter[part.AfterSegment] = append(lockedByAfter[part.AfterSegment], part.Text)
	}
	var output strings.Builder
	for _, text := range lockedByAfter[-1] {
		output.WriteString(text)
	}
	for index, segment := range translated {
		output.WriteString(segment.Text)
		for _, text := range lockedByAfter[index] {
			output.WriteString(text)
		}
	}
	return output.String(), nil
}

// ValidateV2 validates exact record/segment identity, reconstructs controls from
// source-only state, and returns accepted staging rows. It never trusts an LLM
// to return or preserve game controls.
func ValidateV2(source []SourceRowV2, results []ResultRowV2) (ValidationV2, error) {
	if len(source) == 0 {
		return ValidationV2{}, fmt.Errorf("source batch is empty")
	}
	if len(source) != len(results) {
		return ValidationV2{}, fmt.Errorf("result count %d, want exactly %d", len(results), len(source))
	}
	validation := ValidationV2{Rows: make([]AcceptedRowV2, 0, len(source))}
	for position := range source {
		src := source[position]
		got := results[position]
		if got.ID != src.Export.ID {
			return ValidationV2{}, fmt.Errorf("result position %d has ID %d, want %d; order and IDs must be preserved", position+1, got.ID, src.Export.ID)
		}
		if err := validateSegmentShape(src.Export.Segments, got.KoreanSegments); err != nil {
			return ValidationV2{}, fmt.Errorf("ID %d: %w", got.ID, err)
		}
		for index, segment := range got.KoreanSegments {
			if strings.TrimSpace(src.Export.Segments[index].Text) != "" && strings.TrimSpace(segment.Text) == "" {
				return ValidationV2{}, fmt.Errorf("ID %d segment %d: translated text is empty", got.ID, index)
			}
			if protectedV2RE.MatchString(segment.Text) {
				return ValidationV2{}, fmt.Errorf("ID %d segment %d: translated text contains a protected token", got.ID, index)
			}
			if containsJapaneseKana(segment.Text) {
				validation.Warnings = append(validation.Warnings, fmt.Sprintf("ID %d segment %d: Korean result still contains Japanese kana", got.ID, index))
			}
		}
		korean, err := ReconstructV2(src, got.KoreanSegments)
		if err != nil {
			return ValidationV2{}, fmt.Errorf("ID %d: reconstruct: %w", got.ID, err)
		}
		if got.Uncertain {
			validation.Uncertain++
			validation.Warnings = append(validation.Warnings, fmt.Sprintf("ID %d: Gemini marked translation uncertain", got.ID))
		}
		for candidateIndex, candidate := range got.GlossaryCandidates {
			if strings.TrimSpace(candidate.Source) == "" || strings.TrimSpace(candidate.Korean) == "" || strings.TrimSpace(candidate.Type) == "" {
				return ValidationV2{}, fmt.Errorf("ID %d: glossary_candidates[%d] requires nonempty source, korean, and type", got.ID, candidateIndex)
			}
		}
		validation.Rows = append(validation.Rows, AcceptedRowV2{
			ID: got.ID, Korean: korean, Uncertain: got.Uncertain, Note: got.Note,
			GlossaryCandidates: got.GlossaryCandidates,
		})
	}
	return validation, nil
}

func validateSegmentShape(source, translated []Segment) error {
	if len(source) != len(translated) {
		return fmt.Errorf("korean_segments length %d, want exactly %d", len(translated), len(source))
	}
	seen := make(map[int]struct{}, len(translated))
	for position := range source {
		want := source[position].Index
		got := translated[position].Index
		if got != want {
			return fmt.Errorf("segment position %d has index %d, want %d", position, got, want)
		}
		if _, exists := seen[got]; exists {
			return fmt.Errorf("duplicate segment index %d", got)
		}
		seen[got] = struct{}{}
	}
	return nil
}

// RetryRowsV2 returns only source rows whose IDs are present in failedIDs, in
// original batch order. This is used for manual re-send workflows.
func RetryRowsV2(source []SourceRowV2, failedIDs []int) []ExportRowV2 {
	wanted := make(map[int]struct{}, len(failedIDs))
	for _, id := range failedIDs {
		wanted[id] = struct{}{}
	}
	rows := make([]ExportRowV2, 0, len(wanted))
	for _, row := range source {
		if _, ok := wanted[row.Export.ID]; ok {
			rows = append(rows, row.Export)
		}
	}
	return rows
}

// SortedGlossaryCopy returns a stable copy for tests/callers that need
// deterministic iteration independent of map order.
func SortedGlossaryCopy(glossary map[string]string) [][2]string {
	keys := make([]string, 0, len(glossary))
	for key := range glossary {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([][2]string, 0, len(keys))
	for _, key := range keys {
		out = append(out, [2]string{key, glossary[key]})
	}
	return out
}
