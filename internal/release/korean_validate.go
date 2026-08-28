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
// bank is emitted.
func validateKoreanRuntimeStorage(root string, source *corpus.Project, korean *corpus.KoreanProject, layouts map[int]string, mapping koreanslots.Mapping) error {
	engine, err := loadLayout(root)
	if err != nil {
		return fmt.Errorf("Korean runtime storage validation: %w", err)
	}
	if err := engine.ValidateKoreanC5(source, korean, layouts, mapping); err != nil {
		return err
	}
	return nil
}
