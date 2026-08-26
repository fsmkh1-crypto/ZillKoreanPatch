// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"fmt"
	"sort"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/message"
)

const koreanAlphaPlaceholder = "[JP]"

// BuildKoreanAlphaPlaceholderProject creates an in-memory, development-only
// overlay for device-alpha validation. Accepted Korean rows are preserved
// verbatim; untranslated records with editable semantic fragments are replaced
// by a tiny ASCII marker while source-owned controls/substitutions remain intact.
// Structural/no-text records are omitted so their authenticated retail bytes are
// preserved byte-for-byte. No files in translations/korean are modified.
func BuildKoreanAlphaPlaceholderProject(source *corpus.Project, korean *corpus.KoreanProject) (*corpus.KoreanProject, int, error) {
	if source == nil || korean == nil {
		return nil, 0, fmt.Errorf("Korean alpha placeholder: nil project")
	}
	accepted := make(map[int]corpus.KoreanEntry, len(korean.Entries))
	for _, row := range korean.Entries {
		accepted[row.ID] = row
	}

	entries := make([]corpus.KoreanEntry, 0, len(source.Items))
	placeholders := 0
	for _, item := range source.Items {
		if row, ok := accepted[item.Record.ID]; ok {
			entries = append(entries, row)
			continue
		}

		text, ok, err := message.PlaceholderForRecord(item.Record, koreanAlphaPlaceholder)
		if err != nil {
			return nil, 0, fmt.Errorf("Korean alpha placeholder ID %d: %w", item.Record.ID, err)
		}
		if !ok {
			continue
		}
		entries = append(entries, corpus.KoreanEntry{
			ID:       item.Record.ID,
			Japanese: item.Translation.Japanese,
			Korean:   text,
		})
		placeholders++
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].ID < entries[j].ID })
	return &corpus.KoreanProject{Entries: entries}, placeholders, nil
}
