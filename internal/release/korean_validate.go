// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"fmt"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanslots"
)

// validateKoreanRuntimeStorage checks consumer contracts against the exact
// renderer-byte mapping that will be compiled into the Korean release. It must
// run after final authored layouts and slot planning, but before any message
// bank is emitted. The returned IDs are C5 records with runtime substitutions;
// they remain QA risks even when the static storage check finds no violation.
func validateKoreanRuntimeStorage(root string, source *corpus.Project, korean *corpus.KoreanProject, layouts map[int]string, mapping koreanslots.Mapping) ([]int, error) {
	engine, err := loadLayout(root)
	if err != nil {
		return nil, fmt.Errorf("Korean runtime storage validation: %w", err)
	}
	if err := engine.ValidateKoreanC5(source, korean, layouts, mapping); err != nil {
		return nil, err
	}
	dynamic, err := engine.KoreanC5DynamicIDs(source, korean, layouts, mapping)
	if err != nil {
		return nil, fmt.Errorf("Korean C5 dynamic-risk scan: %w", err)
	}
	return dynamic, nil
}
