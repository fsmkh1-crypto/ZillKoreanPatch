// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"bytes"
	"fmt"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/message"
)

// DeriveKoreanC22RetailScannerLayouts applies the captured z_un_089661DC
// scanner contract only to C22 records whose authenticated retail source is
// structurally compatible with that scanner. C22 is a storage-consumer class,
// not proof that every member flows through a NUL-terminated string scanner.
// Records whose retail bytes contain no NUL cannot have been consumed by the
// captured scanner as modeled here, so they are excluded instead of forcing a
// false universal NUL contract onto the C22 set.
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
	eligible := 0
	skippedNonNUL := 0
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
		if !retailScannerSourceCompatible(item.Record.Raw) {
			skippedNonNUL++
			continue
		}
		eligible++
		// Validate the authenticated retail bytes with the captured scanner model
		// before projecting the same structural contract onto Korean. A malformed
		// ESC sequence remains a hard error; only the absence of any NUL is a
		// positive exclusion criterion handled above.
		if _, err := message.AnalyzeRetailStringScanner(item.Record.Raw); err != nil {
			return nil, 0, fmt.Errorf("message %d C22 retail scanner source analysis: %w", row.ID, err)
		}

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
	fmt.Printf("FORENSIC KOREAN_C22_SCANNER_SCOPE checked=%d eligible_retail_nul=%d skipped_non_nul=%d derived=%d threshold=0x100\n", checked, eligible, skippedNonNUL, count)
	return derived, count, nil
}

func retailScannerSourceCompatible(raw []byte) bool {
	return bytes.IndexByte(raw, 0) >= 0
}
