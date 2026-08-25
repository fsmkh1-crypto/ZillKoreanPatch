// SPDX-License-Identifier: GPL-3.0-or-later

package corpus

import "fmt"

// RuntimeTexts returns the final message text for every source record in stable
// source order. Accepted Korean rows replace their Japanese source; records not
// present in the sparse Korean overlay remain Japanese. This is the canonical
// input for stock-glyph preservation and custom-glyph planning.
func (project *KoreanProject) RuntimeTexts(source *Project) ([]string, error) {
	if project == nil {
		return nil, fmt.Errorf("Korean runtime texts: nil Korean project")
	}
	if source == nil {
		return nil, fmt.Errorf("Korean runtime texts: nil source project")
	}

	overlay := make(map[int]string, len(project.Entries))
	for _, row := range project.Entries {
		if _, exists := overlay[row.ID]; exists {
			return nil, fmt.Errorf("Korean runtime texts: duplicate Korean ID %d", row.ID)
		}
		if _, ok := source.Find(row.ID); !ok {
			return nil, fmt.Errorf("Korean runtime texts: Korean ID %d does not exist in source project", row.ID)
		}
		overlay[row.ID] = row.Korean
	}

	texts := make([]string, 0, len(source.Items))
	for _, item := range source.Items {
		if korean, ok := overlay[item.Record.ID]; ok {
			texts = append(texts, korean)
			continue
		}
		texts = append(texts, item.Translation.Japanese)
	}
	return texts, nil
}
