// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/message"
)

// ValidateKoreanC5 applies the retail C5 branch-local storage contract to the
// actual Korean renderer bytes. The stock validator materializes through CP932;
// Korean must instead use the exact authenticated slot mapping used by the
// release compiler or byte counts are not meaningful.
func (e *Engine) ValidateKoreanC5(source *corpus.Project, korean *corpus.KoreanProject, layouts map[int]string, mapping koreanslots.Mapping) error {
	if source == nil {
		return fmt.Errorf("Korean C5 validation: nil source project")
	}
	if korean == nil {
		return fmt.Errorf("Korean C5 validation: nil Korean project")
	}
	if len(mapping) == 0 && len(korean.Entries) != 0 {
		return fmt.Errorf("Korean C5 validation: empty renderer mapping")
	}

	c5 := make(map[int]struct{}, len(e.consumers.C5IDs)+len(e.consumers.SinglePageC5IDs))
	for _, id := range e.consumers.C5IDs {
		c5[id] = struct{}{}
	}
	for _, id := range e.consumers.SinglePageC5IDs {
		c5[id] = struct{}{}
	}

	var failures []string
	checked := 0
	for _, row := range korean.Entries {
		if _, ok := c5[row.ID]; !ok {
			continue
		}
		item, ok := source.Find(row.ID)
		if !ok {
			failures = append(failures, fmt.Sprintf("C5 message %d lacks source", row.ID))
			continue
		}
		text := row.Korean
		if row.Layout != "" {
			text = row.Layout
		}
		if layout, ok := layouts[row.ID]; ok && layout != "" {
			text = layout
		}
		checked++
		why, err := e.c5ViolationKorean(item, text, mapping)
		if err != nil {
			failures = append(failures, err.Error())
		} else if why != "" {
			failures = append(failures, "C5 "+why)
		}
	}
	if len(failures) != 0 {
		return fmt.Errorf("Korean C5 storage validation failed (%d records checked):\n- %s", checked, strings.Join(failures, "\n- "))
	}
	return nil
}

func (e *Engine) c5ViolationKorean(item corpus.Item, text string, mapping koreanslots.Mapping) (string, error) {
	p, err := message.Project(item.Record)
	if err != nil {
		return "", err
	}
	raw, err := p.MaterializeKorean(text, true, mapping)
	if err != nil {
		return "", fmt.Errorf("message %d C5 Korean lowering: %w", item.Record.ID, err)
	}
	bankData := make([]byte, 4+len(raw))
	binary.LittleEndian.PutUint16(bankData, 1)
	binary.LittleEndian.PutUint16(bankData[2:], 4)
	copy(bankData[4:], raw)
	bank, err := corpus.ParseBank("msgsec000.dat", bankData)
	if err != nil {
		return "", fmt.Errorf("message %d C5 Korean parse: %w", item.Record.ID, err)
	}
	leaves, err := walkC5(bank.Records[0].Tokens, 0, nil, false)
	if err != nil {
		return "", fmt.Errorf("message %d C5 Korean analysis: %w", item.Record.ID, err)
	}

	var violations []string
	maxPages := c5MaxPages
	if e.has(e.consumers.SinglePageC5IDs, item.Record.ID) {
		maxPages = 1
	}
	for branch, leaf := range leaves {
		pages := []int{0}
		breaks := 0
		for _, b := range leaf.data {
			if b == 10 {
				breaks++
				if breaks == c5LinesPerPage {
					pages = append(pages, 0)
					breaks = 0
					continue
				}
			}
			pages[len(pages)-1]++
		}
		if len(pages) > maxPages {
			violations = append(violations, fmt.Sprintf("message %d branch %d has %d pages (maximum %d)", item.Record.ID, branch+1, len(pages), maxPages))
		}
		for page, size := range pages {
			if size >= c5PageBufferCapacityBytes {
				violations = append(violations, fmt.Sprintf("message %d branch %d page %d uses %d bytes (maximum %d)", item.Record.ID, branch+1, page+1, size, c5PageBufferCapacityBytes-1))
			}
		}
	}
	return strings.Join(violations, ", "), nil
}
