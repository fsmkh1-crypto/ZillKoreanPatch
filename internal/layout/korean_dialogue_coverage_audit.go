// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"fmt"
	"sort"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/message"
)

// KoreanDialogueCoverageExclusion records a dialogue row that belongs to an
// upstream-English dialogue consumer but is not eligible for automatic Korean
// layout derivation. Exclusion is not a PASS: Reason explains why the record
// remains outside derivation and Width/Limit preserve repository-side evidence.
type KoreanDialogueCoverageExclusion struct {
	ID     int
	Reason string
	Width  int
	Limit  int
}

// KoreanDialogueRuntimePending records a row whose static layout can be derived
// and measured but whose inline runtime substitution lacks a proven maximum
// rendered width. Static reflow PASS and runtime-width PASS are deliberately
// separate evidence classes.
type KoreanDialogueRuntimePending struct {
	ID     int
	Reason string
	Width  int
	Limit  int
}

// KoreanDialogueCoverageAudit separates static layout coverage, derivation
// eligibility, and runtime-width proof. A zero static overflow count does not
// resolve RuntimePending rows; they remain PENDING until an engine/runtime bound
// is proven or asset-backed/runtime QA closes the uncertainty.
type KoreanDialogueCoverageAudit struct {
	Relevant            int
	DerivationEligible  int
	PersistedLayout     int
	Excluded            []KoreanDialogueCoverageExclusion
	RuntimePending      []KoreanDialogueRuntimePending
	OverflowIDs         []int
	ExcludedOverflowIDs []int
}

func (e *Engine) koreanEnglishDialogueCoverageConsumer(id int) bool {
	return e.narrowText(id) || e.has(e.consumers.C5IDs, id) || e.has(e.consumers.C5PortraitIDs, id)
}

// koreanDialogueRuntimeWidthUnproven classifies only INLINE semantic
// substitutions. Movable opcodes used inside <if>/<select> expressions are fixed
// control-flow operands and do not create rendered-width uncertainty. Retail
// mode gets that distinction from message.Project; repository mode gets the same
// distinction by projecting Korean against Japanese source-owned fixed controls.
func (e *Engine) koreanDialogueRuntimeWidthUnproven(item corpus.Item, row corpus.KoreanEntry, mapping koreanslots.Mapping) (bool, error) {
	var fragments []string
	projection, err := message.Project(item.Record)
	switch {
	case err == nil:
		fragments, err = projection.SplitSemanticKorean(row.Korean, mapping)
		if err != nil {
			return false, fmt.Errorf("message %d runtime-width projection: %w", row.ID, err)
		}
	case len(item.Record.Raw) == 0:
		_, controls := splitRepositoryReflowFragments(row.Japanese, true)
		fragments, err = splitRepositorySemanticAgainstSourceControls(row.Korean, controls)
		if err != nil {
			return false, fmt.Errorf("message %d repository runtime-width projection: %w", row.ID, err)
		}
	default:
		return false, fmt.Errorf("message %d runtime-width projection: %w", row.ID, err)
	}
	for _, fragment := range fragments {
		if koreanDialogueUnboundedRuntimeSubstitution(row.ID, fragment) {
			return true, nil
		}
	}
	return false, nil
}

// AuditKoreanEnglishDialogueCoverage measures every Korean row owned by the
// verified narrow/C5/C5-portrait dialogue consumers after applying effective
// layouts. It does not equate static derivation with runtime-width proof.
//
// Repository mode can prove the static projected line width. Unbounded inline
// substitutions are still source-aware reflowed, matching upstream English, but
// remain RuntimePending because the unknown runtime expansion is not included in
// a release-quality PASS. Excluded rows are reserved for a true population/
// eligibility mismatch and must never be silently hidden by a residual audit.
func (e *Engine) AuditKoreanEnglishDialogueCoverage(source *corpus.Project, korean *corpus.KoreanProject, layouts map[int]string, mapping koreanslots.Mapping) (KoreanDialogueCoverageAudit, error) {
	var audit KoreanDialogueCoverageAudit
	if source == nil || korean == nil {
		return audit, fmt.Errorf("Korean English dialogue coverage audit: nil project")
	}
	if len(mapping) == 0 && len(korean.Entries) != 0 {
		return audit, fmt.Errorf("Korean English dialogue coverage audit: empty renderer mapping")
	}

	for _, row := range korean.Entries {
		if !e.koreanEnglishDialogueCoverageConsumer(row.ID) {
			continue
		}
		audit.Relevant++
		item, ok := source.Find(row.ID)
		if !ok {
			return audit, fmt.Errorf("Korean English dialogue coverage audit: message %d lacks source", row.ID)
		}
		text := effectiveKoreanText(row, layouts)
		width, _, err := e.koreanWarningMetrics(item.Record, text, row.ID, mapping)
		if err != nil {
			return audit, err
		}
		limit := e.advanceLimit(row.ID)
		if width > limit {
			audit.OverflowIDs = append(audit.OverflowIDs, row.ID)
		}
		if row.Layout != "" {
			audit.PersistedLayout++
		}

		if e.koreanEnglishDialogueVisualConsumer(row.ID, row.Korean) {
			audit.DerivationEligible++
		} else {
			reason := "eligibility_mismatch"
			audit.Excluded = append(audit.Excluded, KoreanDialogueCoverageExclusion{
				ID: row.ID, Reason: reason, Width: width, Limit: limit,
			})
			if width > limit {
				audit.ExcludedOverflowIDs = append(audit.ExcludedOverflowIDs, row.ID)
			}
		}

		unproven, err := e.koreanDialogueRuntimeWidthUnproven(item, row, mapping)
		if err != nil {
			return audit, err
		}
		if unproven {
			audit.RuntimePending = append(audit.RuntimePending, KoreanDialogueRuntimePending{
				ID: row.ID, Reason: "unbounded_inline_substitution", Width: width, Limit: limit,
			})
		}
	}

	sort.Ints(audit.OverflowIDs)
	sort.Ints(audit.ExcludedOverflowIDs)
	sort.Slice(audit.Excluded, func(i, j int) bool {
		if audit.Excluded[i].Reason != audit.Excluded[j].Reason {
			return audit.Excluded[i].Reason < audit.Excluded[j].Reason
		}
		return audit.Excluded[i].ID < audit.Excluded[j].ID
	})
	sort.Slice(audit.RuntimePending, func(i, j int) bool {
		if audit.RuntimePending[i].Reason != audit.RuntimePending[j].Reason {
			return audit.RuntimePending[i].Reason < audit.RuntimePending[j].Reason
		}
		return audit.RuntimePending[i].ID < audit.RuntimePending[j].ID
	})
	return audit, nil
}
