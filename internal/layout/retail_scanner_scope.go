// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"fmt"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/message"
)

// DeriveKoreanC22RetailScannerLayouts applies the captured z_un_089661DC
// scanner contract only to the consumer class for which that path is currently
// evidenced: C22. Do not project the scanner's NUL-string contract onto every
// message record; some retail records are structurally terminated by their bank
// boundary and are not NUL-terminated strings.
func (e *Engine) DeriveKoreanC22RetailScannerLayouts(source *corpus.Project, korean *corpus.KoreanProject, layouts map[int]string, mapping koreanslots.Mapping) (map[int]string, int, error) {
	if source == nil {
		return nil, 0, fmt.Errorf("Korean C22 scanner layout derivation: nil source project")
	}
	if korean == nil {
		return nil, 0, fmt.Errorf("Korean C22 scanner layout derivation: nil Korean project")
	}
	if len(mapping) == 0 && len(korean.Entries) != 0 {
		return nil, 0, fmt.Errorf("Korean C22 scanner layout derivation: empty renderer mapping")
	}

	derived := make(map[int]string, len(layouts))
	for id, text := range layouts {
		derived[id] = text
	}
	c22 := make(map[int]struct{}, len(e.consumers.C22IDs))
	for _, id := range e.consumers.C22IDs {
		c22[id] = struct{}{}
	}

	checked := 0
	count := 0
	for _, row := range korean.Entries {
		if _, ok := c22[row.ID]; !ok {
			continue
		}
		item, ok := source.Find(row.ID)
		if !ok {
			return nil, 0, fmt.Errorf("Korean C22 scanner layout derivation: message %d lacks source", row.ID)
		}
		checked++
		effective := effectiveKoreanText(row, derived)
		replacement := message.KoreanRecord{Text: row.Korean}
		if effective != row.Korean {
			replacement.Layout = effective
		}
		raw, err := message.MaterializeKoreanRecordForScannerAudit(item.Record, replacement, mapping)
		if err != nil {
			return nil, 0, fmt.Errorf("message %d C22 scanner preflight materialization: %w", row.ID, err)
		}
		metrics, err := message.AnalyzeRetailStringScanner(raw)
		if err != nil {
			return nil, 0, fmt.Errorf("message %d C22 scanner preflight analysis: %w", row.ID, err)
		}
		if metrics.MaxSpan < 0x100 {
			continue
		}

		candidate := wrapKoreanC5Storage(effective)
		if !message.PreservesLayoutSemantics(row.Korean, candidate) {
			return nil, 0, fmt.Errorf("message %d C22 scanner-derived layout changes semantic/control text", row.ID)
		}
		replacement.Layout = candidate
		raw, err = message.MaterializeKoreanRecordForScannerAudit(item.Record, replacement, mapping)
		if err != nil {
			return nil, 0, fmt.Errorf("message %d C22 scanner-derived materialization: %w", row.ID, err)
		}
		metrics, err = message.AnalyzeRetailStringScanner(raw)
		if err != nil {
			return nil, 0, fmt.Errorf("message %d C22 scanner-derived analysis: %w", row.ID, err)
		}
		if metrics.MaxSpan >= 0x100 {
			return nil, 0, fmt.Errorf("message %d C22 scanner-derived layout still reaches inline boundary: max span %d (0x%X)", row.ID, metrics.MaxSpan, metrics.MaxSpan)
		}
		derived[row.ID] = candidate
		count++
	}
	fmt.Printf("FORENSIC KOREAN_C22_SCANNER_SCOPE checked=%d derived=%d threshold=0x100\n", checked, count)
	return derived, count, nil
}
