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
// verbatim; untranslated records that actually contain translatable text are
// replaced by a tiny ASCII marker while source-owned controls/substitutions
// remain intact. No-text records are deliberately omitted so the retail raw
// record is preserved byte-for-byte. No files in translations/korean are
// modified.
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

		// Some retail records are structural/no-text records. They have no token
		// stream (or no visible Japanese text) and therefore cannot be projected
		// into editable semantic fragments. Leaving them out of the overlay makes
		// CompileBankKorean preserve their authenticated retail bytes unchanged;
		// RuntimeTexts likewise contributes no renderer keys for an empty display.
		if len(item.Record.Tokens) == 0 || item.Translation.Japanese == "" {
			continue
		}

		projection, err := message.Project(item.Record)
		if err != nil {
			return nil, 0, fmt.Errorf("Korean alpha placeholder ID %d: %w", item.Record.ID, err)
		}
		text, err := projection.PlaceholderAnnotated(koreanAlphaPlaceholder)
		if err != nil {
			return nil, 0, err
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
