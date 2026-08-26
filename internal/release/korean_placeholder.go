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
// overlay for device-alpha validation. Parsed retail structure is authoritative:
// records with no editable semantic fragment are omitted first, even if a Korean
// corpus row exists for that ID, so their authenticated retail bytes remain
// byte-identical. Accepted Korean rows are preserved only for genuinely editable
// records; untranslated editable records receive a tiny ASCII marker. No files
// in translations/korean are modified.
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
		// Classify the authenticated retail record before consulting the Korean
		// overlay. A structural/control-only record cannot safely be materialized
		// through message.Project, regardless of whether an accepted corpus row
		// happens to exist for its ID.
		placeholder, editable, err := message.PlaceholderForRecord(item.Record, koreanAlphaPlaceholder)
		if err != nil {
			return nil, 0, fmt.Errorf("Korean alpha placeholder ID %d: %w", item.Record.ID, err)
		}
		if !editable {
			continue
		}

		if row, ok := accepted[item.Record.ID]; ok {
			entries = append(entries, row)
			continue
		}

		entries = append(entries, corpus.KoreanEntry{
			ID:       item.Record.ID,
			Japanese: item.Translation.Japanese,
			Korean:   placeholder,
		})
		placeholders++
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].ID < entries[j].ID })
	return &corpus.KoreanProject{Entries: entries}, placeholders, nil
}
