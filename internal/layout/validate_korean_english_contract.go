// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"fmt"
	"sort"
	"strings"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/message"
)

// DeriveKoreanEnglishConsumerLayouts applies the same C22 storage contract used
// by the upstream English patcher, but measures the actual Korean renderer-slot
// bytes. Only layout is changed; canonical Korean semantics remain untouched.
// Contracts that cannot be repaired by inserting line boundaries remain
// fail-closed in ValidateKoreanEnglishConsumerContracts.
func (e *Engine) DeriveKoreanEnglishConsumerLayouts(source *corpus.Project, korean *corpus.KoreanProject, layouts map[int]string, mapping koreanslots.Mapping) (map[int]string, int, error) {
	if source == nil || korean == nil {
		return nil, 0, fmt.Errorf("Korean English-contract derivation: nil project")
	}
	derived := make(map[int]string, len(layouts))
	for id, text := range layouts {
		derived[id] = text
	}
	c22 := make(map[int]struct{}, len(e.consumers.C22IDs))
	for _, id := range e.consumers.C22IDs {
		c22[id] = struct{}{}
	}
	count := 0
	for _, row := range korean.Entries {
		if _, ok := c22[row.ID]; !ok {
			continue
		}
		effective := effectiveKoreanText(row, derived)
		why, err := e.c22ViolationKoreanBytes(row.ID, effective, mapping)
		if err != nil {
			return nil, 0, err
		}
		if why == "" {
			continue
		}
		candidate := wrapKoreanC5Storage(effective)
		if !message.PreservesLayoutSemantics(row.Korean, candidate) {
			return nil, 0, fmt.Errorf("message %d C22 derived layout changes semantic/control text", row.ID)
		}
		post, err := e.c22ViolationKoreanBytes(row.ID, candidate, mapping)
		if err != nil {
			return nil, 0, err
		}
		if post != "" {
			return nil, 0, fmt.Errorf("message %d C22 cannot be made English-contract safe by layout alone: %s", row.ID, post)
		}
		derived[row.ID] = candidate
		count++
	}
	return derived, count, nil
}

// ValidateKoreanEnglishConsumerContracts mirrors the upstream English patcher's
// fixed-storage validation categories and limits. The only deliberate language
// difference is byte measurement: natural Korean text is encoded with the exact
// authenticated renderer mapping instead of stock CP932.
func (e *Engine) ValidateKoreanEnglishConsumerContracts(source *corpus.Project, korean *corpus.KoreanProject, layouts map[int]string, mapping koreanslots.Mapping) error {
	if source == nil || korean == nil {
		return fmt.Errorf("Korean English-contract validation: nil project")
	}
	effective := make(map[int]string, len(korean.Entries))
	translated := make(map[int]bool, len(korean.Entries))
	for _, row := range korean.Entries {
		effective[row.ID] = effectiveKoreanText(row, layouts)
		translated[row.ID] = true
	}
	var failures []string
	checkFixed := func(label string, ids []int, capacity int) {
		for _, id := range ids {
			if !translated[id] {
				continue
			}
			size, err := koreanExpandedBytes(effective[id], id, mapping)
			if err != nil {
				failures = append(failures, err.Error())
				continue
			}
			if size >= capacity {
				failures = append(failures, fmt.Sprintf("%s message %d uses %d bytes (maximum %d)", label, id, size, capacity-1))
			}
		}
	}
	checkFixed("bounded label", e.consumers.BoundedLabelIDs, boundedLabelBufferCapacityBytes)
	checkFixed("guild client", e.consumers.GuildClientIDs, guildClientBufferCapacityBytes)
	checkFixed("guild region", e.consumers.GuildRegionIDs, guildRegionBufferCapacityBytes)

	if translated[trapID] {
		size, err := koreanExpandedBytes(effective[trapID], trapID, mapping)
		if err != nil {
			failures = append(failures, err.Error())
		} else {
			size += len(valueTag.FindAllString(effective[trapID], -1)) * trapValueMaxBytes
			if size >= trapBufferCapacityBytes {
				failures = append(failures, fmt.Sprintf("trap message %d uses up to %d bytes (maximum %d)", trapID, size, trapBufferCapacityBytes-1))
			}
		}
	}

	for id := range translated {
		if e.category(id, "character-creation-choice") {
			size, err := koreanExpandedBytes(effective[id], id, mapping)
			if err != nil {
				failures = append(failures, err.Error())
			} else if size >= characterCreationChoiceCapacityBytes {
				failures = append(failures, fmt.Sprintf("character-creation choice message %d uses %d bytes (maximum %d)", id, size, characterCreationChoiceCapacityBytes-1))
			}
		}
		if e.equipmentFeedback(id) {
			size, err := koreanExpandedBytes(effective[id], id, mapping)
			if err != nil {
				failures = append(failures, err.Error())
			} else if size >= equipmentFeedbackBufferCapacityBytes {
				failures = append(failures, fmt.Sprintf("equipment feedback message %d uses %d bytes (maximum %d)", id, size, equipmentFeedbackBufferCapacityBytes-1))
			}
		}
		if e.category(id, "chronicle-entry") {
			size, err := koreanExpandedBytes(effective[id], id, mapping)
			if err != nil {
				failures = append(failures, err.Error())
				continue
			}
			size += strings.Count(effective[id], "<value:$28>") * playerNameMaxEncodedBytes
			if size > chronicleEntryMaxPayloadBytes {
				failures = append(failures, fmt.Sprintf("chronicle entry message %d uses up to %d bytes (maximum %d)", id, size, chronicleEntryMaxPayloadBytes))
			}
		}
	}

	for _, group := range e.consumers.C20Groups {
		changed, total, missing := false, 0, false
		for _, id := range group.IDs {
			changed = changed || translated[id]
			text, ok := effective[id]
			if !ok {
				missing = true
				break
			}
			size, err := koreanMinimumBytes(text, id, mapping)
			if err != nil {
				failures = append(failures, err.Error())
				missing = true
				break
			}
			total += size + 1
		}
		if changed && missing {
			failures = append(failures, fmt.Sprintf("C20 group starting at %d lacks a message", group.IDs[0]))
		} else if changed && total >= c20GroupBufferCapacityBytes {
			failures = append(failures, fmt.Sprintf("C20 group starting at %d uses %d bytes (maximum %d)", group.IDs[0], total, c20GroupBufferCapacityBytes-1))
		}
	}

	for _, id := range e.consumers.C22IDs {
		if !translated[id] {
			continue
		}
		why, err := e.c22ViolationKoreanBytes(id, effective[id], mapping)
		if err != nil {
			failures = append(failures, err.Error())
		} else if why != "" {
			failures = append(failures, "C22 "+why)
		}
	}

	e.validateKoreanPostings(effective, translated, mapping, &failures)
	if len(failures) != 0 {
		sort.Strings(failures)
		return fmt.Errorf("Korean upstream-English consumer storage validation failed:\n- %s", strings.Join(failures, "\n- "))
	}
	return nil
}

func koreanExpandedBytes(text string, id int, mapping koreanslots.Mapping) (int, error) {
	total := 0
	for _, part := range splitParts(normalize(text)) {
		switch {
		case part == lineBreak:
			total++
		case strings.HasPrefix(part, "<color:"):
			total += 3
		case controlTag.MatchString(part) && controlTag.FindString(part) == part:
			continue
		default:
			b, err := koreanslots.Encode(part, mapping)
			if err != nil {
				return 0, fmt.Errorf("message %d Korean storage: %w", id, err)
			}
			total += len(b)
		}
	}
	return total, nil
}

func koreanMinimumBytes(text string, id int, mapping koreanslots.Mapping) (int, error) {
	if formatSignatureID(id) {
		text = printfConversion.ReplaceAllString(text, "")
	}
	return koreanExpandedBytes(text, id, mapping)
}

func (e *Engine) c22ViolationKoreanBytes(id int, text string, mapping koreanslots.Mapping) (string, error) {
	runtime := strings.Split(normalize(text), "<end>")[0]
	lines := strings.Split(runtime, lineBreak)
	total, err := koreanMinimumBytes(runtime, id, mapping)
	if err != nil {
		return "", err
	}
	var violations []string
	if (len(lines)+3)/4 > c22MaxPages {
		violations = append(violations, fmt.Sprintf("message %d has %d pages (maximum %d)", id, (len(lines)+3)/4, c22MaxPages))
	}
	if total >= c22TotalBufferCapacityBytes {
		violations = append(violations, fmt.Sprintf("message %d uses %d total bytes (maximum %d)", id, total, c22TotalBufferCapacityBytes-1))
	}
	for i, line := range lines {
		size, err := koreanMinimumBytes(line, id, mapping)
		if err != nil {
			return "", err
		}
		if size > c22MaxLineBytes {
			violations = append(violations, fmt.Sprintf("message %d line %d uses %d bytes (maximum %d)", id, i+1, size, c22MaxLineBytes))
		}
	}
	for p := 0; p < len(lines); p += 4 {
		end := min(p+4, len(lines))
		size, err := koreanMinimumBytes(strings.Join(lines[p:end], lineBreak), id, mapping)
		if err != nil {
			return "", err
		}
		if size >= c22PageBufferCapacityBytes {
			violations = append(violations, fmt.Sprintf("message %d page %d uses %d bytes (maximum %d)", id, p/4+1, size, c22PageBufferCapacityBytes-1))
		}
	}
	return strings.Join(violations, ", "), nil
}

func (e *Engine) validateKoreanPostings(effective map[int]string, translated map[int]bool, mapping koreanslots.Mapping, failures *[]string) {
	maxima := map[string]int{}
	for role, ids := range e.postingCandidateIDs() {
		for _, id := range ids {
			if text, ok := effective[id]; ok {
				if size, err := koreanExpandedBytes(text, id, mapping); err == nil && size > maxima[role] {
					maxima[role] = size
				}
			}
		}
	}
	integer := map[string]bool{}
	for _, role := range e.consumers.PostingCandidates.IntegerRoles {
		integer[role] = true
	}
	for _, p := range e.consumers.Postings {
		if !translated[p.ID] {
			continue
		}
		text, ok := effective[p.ID]
		if !ok {
			*failures = append(*failures, fmt.Sprintf("guild posting %d lacks text", p.ID))
			continue
		}
		size, err := koreanExpandedBytes(text, p.ID, mapping)
		if err != nil {
			*failures = append(*failures, err.Error())
			continue
		}
		for tag, role := range p.Bindings {
			count := strings.Count(text, tag)
			if integer[role] {
				size += count * guildPostingIntegerMaxBytes
			} else {
				size += count * maxima[role]
			}
		}
		if size >= guildPostingBufferCapacityBytes {
			*failures = append(*failures, fmt.Sprintf("guild posting message %d uses %d bytes (maximum %d)", p.ID, size, guildPostingBufferCapacityBytes-1))
		}
	}
}
