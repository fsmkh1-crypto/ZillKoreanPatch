// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"fmt"
	"sort"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/message"
)

// AuditKoreanEnglishVisualWarnings mirrors the non-blocking Warning classes
// emitted by upstream English Engine.Reflow. These conditions are deliberately
// not promoted to Korean hard failures: upstream treats them as authoring/QA
// signals. Korean differs only in measurement, using the renderer mapping and
// advances that the Korean font path will actually install.
func (e *Engine) AuditKoreanEnglishVisualWarnings(source *corpus.Project, korean *corpus.KoreanProject, layouts map[int]string, mapping koreanslots.Mapping) ([]Warning, error) {
	if source == nil || korean == nil {
		return nil, fmt.Errorf("Korean English visual-warning audit: nil project")
	}
	if len(mapping) == 0 && len(korean.Entries) != 0 {
		return nil, fmt.Errorf("Korean English visual-warning audit: empty renderer mapping")
	}

	warnings := make([]Warning, 0)
	for _, row := range korean.Entries {
		item, ok := source.Find(row.ID)
		if !ok {
			return nil, fmt.Errorf("Korean English visual-warning audit: message %d lacks source", row.ID)
		}
		semantic := row.Korean
		text := effectiveKoreanText(row, layouts)

		width, lines, err := e.koreanWarningMetrics(item.Record, text, row.ID, mapping)
		if err != nil {
			return nil, err
		}
		limit := e.advanceLimit(row.ID)
		if !e.category(row.ID, "character-profile") && width > limit {
			code := "line_exceeds_authoring_ceiling"
			if e.itemDescription(row.ID) {
				code = "item_description_single_line_overflow"
			}
			warnings = append(warnings, Warning{Code: code, MessageID: row.ID})
		}
		if e.category(row.ID, "chronicle-entry") && lines > chronicleMaxLines {
			warnings = append(warnings, Warning{Code: "chronicle_vertical_overflow", MessageID: row.ID})
		}
		if valueTag.MatchString(semantic) || (formatSignatureID(row.ID) && printfConversion.MatchString(visible(semantic))) {
			warnings = append(warnings, Warning{Code: "runtime_substitution_unbounded", MessageID: row.ID})
		}
		if e.has(e.consumers.GuildClientIDs, row.ID) {
			clientWidth, err := e.measureKoreanRenderer(semantic, row.ID, mapping)
			if err != nil {
				return nil, err
			}
			if clientWidth > 104 {
				warnings = append(warnings, Warning{Code: "guild_job_client_overflow", MessageID: row.ID})
			}
		}
	}

	sort.Slice(warnings, func(i, j int) bool {
		if warnings[i].Code == warnings[j].Code {
			return warnings[i].MessageID < warnings[j].MessageID
		}
		return warnings[i].Code < warnings[j].Code
	})
	return warnings, nil
}

func (e *Engine) koreanWarningMetrics(record corpus.Record, text string, id int, mapping koreanslots.Mapping) (int, int, error) {
	projection, err := message.Project(record)
	if err == nil {
		width, err := e.maxProjectedKoreanWidth(projection, text, id, mapping)
		if err != nil {
			return 0, 0, err
		}
		lines, err := maxProjectedKoreanFragmentLines(projection, text, id, mapping)
		if err != nil {
			return 0, 0, err
		}
		return width, lines, nil
	}
	if len(record.Raw) == 0 {
		// Repository-only CI has no retail token stream. The annotated Korean text
		// is still sufficient for the warning-level width/vertical census; the
		// authenticated asset-bound build repeats this audit through projection.
		return e.measureKoreanProfileText(text, id, mapping)
	}
	return 0, 0, fmt.Errorf("message %d Korean visual-warning projection: %w", id, err)
}
