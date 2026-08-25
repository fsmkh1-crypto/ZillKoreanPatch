// SPDX-License-Identifier: GPL-3.0-or-later

package translationexchange

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// BuildSourceV2 creates both the safe external payload and the private
// reconstruction state from one canonical source record.
func BuildSourceV2(id, recordIndex int, sourceFile, japanese, englishReference string, glossary map[string]string) SourceRowV2 {
	segments, locked := SplitV2(japanese)
	if glossary == nil {
		glossary = map[string]string{}
	}
	return SourceRowV2{
		Export: ExportRowV2{
			Schema:           SchemaV2,
			ID:               id,
			Section:          id / 10_000,
			RecordIndex:      recordIndex,
			SourceFile:       sourceFile,
			FullText:         ControlFreeContext(japanese),
			EnglishReference: ControlFreeContext(englishReference),
			Glossary:         glossary,
			Segments:         segments,
		},
		Locked: locked,
	}
}

// WriteExportV2 writes only the external-safe payload. Locked controls are not
// serialized here by design.
func WriteExportV2(w io.Writer, rows []ExportRowV2) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	for position, row := range rows {
		if err := validateExportV2(row); err != nil {
			return fmt.Errorf("export row %d: %w", position+1, err)
		}
		if err := encoder.Encode(row); err != nil {
			return fmt.Errorf("encode export row %d ID %d: %w", position+1, row.ID, err)
		}
	}
	return nil
}

// ReadExportV2 decodes the exact source file that was sent to Gemini.
func ReadExportV2(r io.Reader) ([]ExportRowV2, error) {
	var rows []ExportRowV2
	seen := make(map[int]struct{})
	err := scanJSONL(r, func(line int, data []byte) error {
		var row ExportRowV2
		if err := decodeStrictLine(data, &row); err != nil {
			return fmt.Errorf("source line %d: %w", line, err)
		}
		if err := validateExportV2(row); err != nil {
			return fmt.Errorf("source line %d: %w", line, err)
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

// ReadResultsV2 rejects prose, Markdown, blank lines, unknown/missing fields,
// null arrays, duplicate IDs, and malformed segment indices before semantic QA.
func ReadResultsV2(r io.Reader) ([]ResultRowV2, error) {
	var rows []ResultRowV2
	seen := make(map[int]struct{})
	err := scanJSONL(r, func(line int, data []byte) error {
		var row ResultRowV2
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

func WriteAcceptedV2(w io.Writer, rows []AcceptedRowV2) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	for position, row := range rows {
		if err := encoder.Encode(row); err != nil {
			return fmt.Errorf("encode accepted row %d ID %d: %w", position+1, row.ID, err)
		}
	}
	return nil
}

// SameExportV2 is used by the checker to prove the user-returned source file is
// exactly what current canonical data would generate. This prevents a modified
// or stale source payload from being silently accepted.
func SameExportV2(a, b ExportRowV2) bool {
	return reflect.DeepEqual(a, b)
}

func validateExportV2(row ExportRowV2) error {
	if row.Schema != SchemaV2 {
		return fmt.Errorf("ID %d: schema %q, want %q", row.ID, row.Schema, SchemaV2)
	}
	if row.ID < 0 {
		return fmt.Errorf("negative ID %d", row.ID)
	}
	if row.Section != row.ID/10_000 {
		return fmt.Errorf("ID %d: section %d does not match ID", row.ID, row.Section)
	}
	if row.RecordIndex < 0 {
		return fmt.Errorf("ID %d: negative record_index", row.ID)
	}
	if strings.TrimSpace(row.SourceFile) == "" {
		return fmt.Errorf("ID %d: source_file is empty", row.ID)
	}
	if row.Glossary == nil {
		return fmt.Errorf("ID %d: glossary must be an object, not null", row.ID)
	}
	if row.Segments == nil {
		return fmt.Errorf("ID %d: segments must be an array, not null", row.ID)
	}
	for position, segment := range row.Segments {
		if segment.Index != position {
			return fmt.Errorf("ID %d: segment position %d has index %d", row.ID, position, segment.Index)
		}
		if strings.TrimSpace(segment.Text) == "" {
			return fmt.Errorf("ID %d segment %d: source text is empty/whitespace", row.ID, position)
		}
		if protectedV2RE.MatchString(segment.Text) {
			return fmt.Errorf("ID %d segment %d: source payload leaked a protected token", row.ID, position)
		}
	}
	if protectedV2RE.MatchString(row.FullText) || protectedV2RE.MatchString(row.EnglishReference) {
		return fmt.Errorf("ID %d: context/reference leaked a protected token", row.ID)
	}
	return nil
}
