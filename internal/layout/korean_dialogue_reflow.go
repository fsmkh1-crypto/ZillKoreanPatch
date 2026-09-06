// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/message"
)

// Keep this set aligned with message.Project's pureMovable and callerMovable
// substitutions. Other <value:$XX> tokens are fixed source controls; notably
// $20 is the select-count control used by records such as 280181 and must not
// make an otherwise static dialogue record opt out of source-aware reflow.
var koreanDialogueMovableValueTags = map[string]bool{
	"<VALUE:$15>": true,
	"<VALUE:$16>": true,
	"<VALUE:$17>": true,
	"<VALUE:$1A>": true,
	"<VALUE:$1B>": true,
	"<VALUE:$24>": true,
	"<VALUE:$25>": true,
	"<VALUE:$28>": true,
	"<VALUE:$2B>": true,
}

// Bounded movable substitutions may be called statically width-safe because
// the layout engine reserves their proven worst-case rendered advance. Do not
// add an opcode here merely because it is movable: each entry needs an
// engine/game-contract bound that is also reflected by Korean measurement.
var koreanDialogueBoundedInlineValueTags = map[string]bool{
	"<VALUE:$28>": true, // player name: bounded by playerNameMaxCharacters/EncodedBytes
}

func koreanDialogueRuntimeSubstitution(id int, text string) bool {
	upper := strings.ToUpper(text)
	for tag := range koreanDialogueMovableValueTags {
		if strings.Contains(upper, tag) {
			return true
		}
	}
	return formatSignatureID(id) && printfConversion.MatchString(visible(text))
}

func koreanDialogueUnboundedRuntimeSubstitution(id int, text string) bool {
	upper := strings.ToUpper(text)
	for tag := range koreanDialogueMovableValueTags {
		if strings.Contains(upper, tag) && !koreanDialogueBoundedInlineValueTags[tag] {
			return true
		}
	}
	return formatSignatureID(id) && printfConversion.MatchString(visible(text))
}

// koreanEnglishDialogueVisualConsumer mirrors the upstream English visual
// reflow population. Runtime-width proof is deliberately NOT an eligibility
// predicate: upstream English sourceAware reflows movable-substitution fragments
// too. Korean therefore reflows the statically measurable portion of every
// verified narrow/C5/C5-portrait dialogue record and tracks unbounded inline
// substitutions separately as PENDING runtime-width evidence.
func (e *Engine) koreanEnglishDialogueVisualConsumer(id int, _ string) bool {
	return e.narrowText(id) || e.has(e.consumers.C5IDs, id) || e.has(e.consumers.C5PortraitIDs, id)
}

// DeriveKoreanEnglishDialogueLayouts mirrors the upstream English Reflow path
// for verified narrow dialogue/in-world-guidance and authenticated C5 dialogue
// consumers. Canonical Korean is never rewritten: derived breaks live only in
// the build-local layout map.
//
// Persisted Korean layout is build-owned output, not authored semantic text. An
// existing layout therefore does not exempt a record from current derivation.
// If canonical Korean exceeds the consumer limit, derive again from semantic
// Korean and the current English source-layout hints. This repairs stale/older
// generated layouts such as 30032/40007 without hand-authoring line breaks.
//
// Unbounded inline substitutions remain eligible for static reflow. Their
// unknown runtime maximum is a separate coverage/PENDING question and must not
// be converted into a false PASS merely because the static layout fits.
func (e *Engine) DeriveKoreanEnglishDialogueLayouts(source *corpus.Project, korean *corpus.KoreanProject, layouts map[int]string, mapping koreanslots.Mapping) (map[int]string, int, error) {
	if source == nil || korean == nil {
		return nil, 0, fmt.Errorf("Korean English dialogue derivation: nil project")
	}
	if len(mapping) == 0 && len(korean.Entries) != 0 {
		return nil, 0, fmt.Errorf("Korean English dialogue derivation: empty renderer mapping")
	}
	derived := make(map[int]string, len(layouts))
	for id, text := range layouts {
		derived[id] = text
	}
	count := 0
	for _, row := range korean.Entries {
		if !e.koreanEnglishDialogueVisualConsumer(row.ID, row.Korean) {
			continue
		}
		item, ok := source.Find(row.ID)
		if !ok {
			return nil, 0, fmt.Errorf("dialogue message %d lacks source", row.ID)
		}

		// Decide whether reflow is needed from canonical semantic Korean, not from
		// a previously generated layout that may predate the current contract.
		semantic := row.Korean
		width, _, err := e.koreanWarningMetrics(item.Record, semantic, row.ID, mapping)
		if err != nil {
			return nil, 0, err
		}
		limit := e.advanceLimit(row.ID)
		if width <= limit {
			continue
		}

		var candidate string
		projection, projectionErr := message.Project(item.Record)
		switch {
		case projectionErr == nil:
			// Authenticated retail builds take the exact same token-derived
			// SourceLayout path as upstream English Reflow.
			candidate, err = e.koreanSourceAware(projection, semantic, limit, row.ID, mapping)
		case len(item.Record.Raw) == 0:
			// Repository checks intentionally run before BindBanks and therefore
			// have only the canonical Japanese annotated source. Project Korean
			// against source-owned fixed controls and feed source break hints into
			// the same preferred -> greedy scorer used by the retail path.
			candidate, err = e.koreanRepositorySourceAware(semantic, row.Japanese, limit, row.ID, mapping)
		default:
			// Once retail bytes exist, a projection failure is real evidence of
			// contract drift and must never be hidden by the repository fallback.
			return nil, 0, fmt.Errorf("message %d Korean dialogue projection: %w", row.ID, projectionErr)
		}
		if err != nil {
			return nil, 0, err
		}
		if candidate == "" {
			// Match upstream Reflow: an impossible preferred/greedy derivation
			// falls back to semantic text and is then rejected by the width audit.
			candidate = semantic
		}
		if !message.PreservesLayoutSemantics(row.Korean, candidate) {
			return nil, 0, fmt.Errorf("message %d dialogue derived layout changes semantic/control text", row.ID)
		}
		postWidth, _, err := e.koreanWarningMetrics(item.Record, candidate, row.ID, mapping)
		if err != nil {
			return nil, 0, err
		}
		if postWidth > limit {
			return nil, 0, fmt.Errorf("message %d dialogue cannot be made English-reflow safe by layout alone: width=%d/%d", row.ID, postWidth, limit)
		}
		derived[row.ID] = candidate
		count++
	}
	return derived, count, nil
}

// AuditKoreanEnglishDialogueResiduals checks the derivation consumer population
// after layout generation. It is a STATIC layout residual audit only. Runtime
// width proof for unbounded inline substitutions is intentionally separate and
// is reported by AuditKoreanEnglishDialogueCoverage.
func (e *Engine) AuditKoreanEnglishDialogueResiduals(source *corpus.Project, korean *corpus.KoreanProject, layouts map[int]string, mapping koreanslots.Mapping) (checked int, overflowIDs []int, err error) {
	if source == nil || korean == nil {
		return 0, nil, fmt.Errorf("Korean English dialogue residual audit: nil project")
	}
	if len(mapping) == 0 && len(korean.Entries) != 0 {
		return 0, nil, fmt.Errorf("Korean English dialogue residual audit: empty renderer mapping")
	}
	for _, row := range korean.Entries {
		if !e.koreanEnglishDialogueVisualConsumer(row.ID, row.Korean) {
			continue
		}
		item, ok := source.Find(row.ID)
		if !ok {
			return checked, overflowIDs, fmt.Errorf("dialogue message %d lacks source", row.ID)
		}
		checked++
		effective := effectiveKoreanText(row, layouts)
		width, _, metricErr := e.koreanWarningMetrics(item.Record, effective, row.ID, mapping)
		if metricErr != nil {
			return checked, overflowIDs, metricErr
		}
		if width > e.advanceLimit(row.ID) {
			overflowIDs = append(overflowIDs, row.ID)
		}
	}
	return checked, overflowIDs, nil
}

func (e *Engine) wrapKoreanVisualToLimit(text string, id int, mapping koreanslots.Mapping, limit int) (string, error) {
	return wrapKoreanDelimitedParagraphs(text, func(paragraph string) (string, error) {
		return e.wrapKoreanVisualParagraphToLimit(paragraph, id, mapping, limit)
	})
}

type koreanVisualRun struct {
	text       string
	whitespace bool
}

func splitKoreanVisualRuns(text string) []koreanVisualRun {
	runes := []rune(text)
	if len(runes) == 0 {
		return nil
	}
	start := 0
	space := unicode.IsSpace(runes[0])
	runs := make([]koreanVisualRun, 0, 8)
	for i := 1; i < len(runes); i++ {
		nextSpace := unicode.IsSpace(runes[i])
		if nextSpace == space {
			continue
		}
		runs = append(runs, koreanVisualRun{text: string(runes[start:i]), whitespace: space})
		start = i
		space = nextSpace
	}
	return append(runs, koreanVisualRun{text: string(runes[start:]), whitespace: space})
}

// lastBreakableWhitespaceRun returns the final complete whitespace span that
// separates two non-whitespace runs. Replacing that whole span with
// <line-break> is accepted by PreservesLayoutSemantics; normalizing or partly
// consuming the span is not.
func lastBreakableWhitespaceRun(text string) (prefix, tail string, ok bool) {
	runes := []rune(text)
	lastStart, lastEnd := -1, -1
	for i := 0; i < len(runes); {
		if !unicode.IsSpace(runes[i]) {
			i++
			continue
		}
		start := i
		for i < len(runes) && unicode.IsSpace(runes[i]) {
			i++
		}
		if start > 0 && i < len(runes) {
			lastStart, lastEnd = start, i
		}
	}
	if lastStart < 0 {
		return "", "", false
	}
	return string(runes[:lastStart]), string(runes[lastEnd:]), true
}

func (e *Engine) wrapKoreanVisualParagraphToLimit(text string, id int, mapping koreanslots.Mapping, limit int) (string, error) {
	if strings.TrimSpace(text) == "" {
		return text, nil
	}
	runs := splitKoreanVisualRuns(text)
	lines := make([]string, 0, 4)
	current := ""
	pendingWhitespace := ""
	hasWord := false

	for _, run := range runs {
		if run.whitespace {
			pendingWhitespace += run.text
			continue
		}

		word := run.text
		candidate := current + pendingWhitespace + word
		width, err := e.measureKoreanRenderer(candidate, id, mapping)
		if err != nil {
			return "", err
		}
		if width <= limit {
			current = candidate
			pendingWhitespace = ""
			hasWord = true
			continue
		}

		if !hasWord {
			return "", fmt.Errorf("message %d dialogue word exceeds %d units and cannot be whitespace-reflowed: %q (%d units)", id, limit, word, width)
		}

		// Legacy wrapper retained for non-dialogue callers/tests. The verified
		// English-parity dialogue derivation above no longer uses this separate
		// value-adjacency heuristic; it uses sourceAware for all dialogue classes.
		if strings.HasPrefix(word, "<value:") {
			prefix, tail, ok := lastBreakableWhitespaceRun(current)
			if !ok {
				return "", fmt.Errorf("message %d dialogue cannot wrap safely before runtime substitution %q", id, word)
			}
			moved := tail + pendingWhitespace + word
			movedWidth, err := e.measureKoreanRenderer(moved, id, mapping)
			if err != nil {
				return "", err
			}
			if movedWidth > limit {
				return "", fmt.Errorf("message %d dialogue runtime-substitution group exceeds %d units: %q (%d units)", id, limit, moved, movedWidth)
			}
			lines = append(lines, prefix)
			current = moved
			pendingWhitespace = ""
			hasWord = true
			continue
		}

		wordWidth, err := e.measureKoreanRenderer(word, id, mapping)
		if err != nil {
			return "", err
		}
		if wordWidth > limit {
			return "", fmt.Errorf("message %d dialogue word exceeds %d units and cannot be whitespace-reflowed: %q (%d units)", id, limit, word, wordWidth)
		}

		// The entire pending whitespace span is replaced by a layout break. All
		// whitespace that is not selected as a break remains byte-for-byte intact.
		lines = append(lines, current)
		current = word
		pendingWhitespace = ""
		hasWord = true
	}

	current += pendingWhitespace
	if current != "" {
		lines = append(lines, current)
	}
	return strings.Join(lines, lineBreak), nil
}
