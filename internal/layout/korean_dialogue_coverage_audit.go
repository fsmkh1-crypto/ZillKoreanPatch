// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"fmt"
	"sort"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanslots"
)

// KoreanDialogueCoverageExclusion records a dialogue row that belongs to an
// upstream-English dialogue consumer but is not eligible for automatic Korean
// layout derivation. Exclusion is not a PASS: Reason explains why the record
// remains outside derivation and Width/Limit preserve the repository-side
// evidence available for the final visual audit.
type KoreanDialogueCoverageExclusion struct {
	ID     int
	Reason string
	Width  int
	Limit  int
}

// KoreanDialogueCoverageAudit separates derivation-subset evidence from the
// final whole-consumer visual audit. A zero residual count from
// AuditKoreanEnglishDialogueResiduals says nothing about Excluded rows; callers
// must inspect this result before claiming whole-dialogue safety.
type KoreanDialogueCoverageAudit struct {
	Relevant            int
	DerivationEligible  int
	PersistedLayout     int
	Excluded            []KoreanDialogueCoverageExclusion
	OverflowIDs         []int
	ExcludedOverflowIDs []int
}

func (e *Engine) koreanEnglishDialogueCoverageConsumer(id int) bool {
	return e.narrowText(id) || e.has(e.consumers.C5IDs, id) || e.has(e.consumers.C5PortraitIDs, id)
}

// AuditKoreanEnglishDialogueCoverage measures every Korean row owned by the
// verified narrow/C5/C5-portrait dialogue consumers after applying the supplied
// effective layouts. Unlike the derivation residual audit, this function does
// not filter out rows merely because automatic derivation excluded them.
//
// Repository mode can prove static projected width but cannot prove the runtime
// maximum of an UNBOUNDED_INLINE substitution. Such rows remain in Excluded even
// when their static width fits. If an excluded row already exceeds the consumer
// limit before the unknown runtime value is substituted, ExcludedOverflowIDs
// makes that repository-proven unsafe state impossible to hide as PENDING.
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
			continue
		}
		if e.koreanEnglishDialogueVisualConsumer(row.ID, row.Korean) {
			audit.DerivationEligible++
			continue
		}

		reason := "eligibility_mismatch"
		if koreanDialogueUnboundedRuntimeSubstitution(row.ID, row.Korean) {
			reason = "unbounded_inline_substitution"
		}
		audit.Excluded = append(audit.Excluded, KoreanDialogueCoverageExclusion{
			ID: row.ID, Reason: reason, Width: width, Limit: limit,
		})
		if width > limit {
			audit.ExcludedOverflowIDs = append(audit.ExcludedOverflowIDs, row.ID)
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
	return audit, nil
}
