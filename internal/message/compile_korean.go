// SPDX-License-Identifier: GPL-3.0-or-later

package message

import (
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanslots"
)

// KoreanRecord is one explicitly selected Korean replacement. Records absent
// from the replacement map remain byte-identical to retail source, which keeps
// the first real-sentence PoC narrow and reversible.
type KoreanRecord struct {
	Text   string
	Layout string
}

// CompileBankKorean compiles only explicitly supplied Korean replacements and
// copies every other retail record unchanged. It uses the same projection,
// runtime-control, layout-semantics, and bank-capacity rules as CompileBank,
// but natural text is validated and encoded with the supplied renderer-slot
// mapping.
func CompileBankKorean(bank corpus.Bank, items []corpus.Item, replacements map[int]KoreanRecord, mapping koreanslots.Mapping) ([]byte, error) {
	if len(items) != len(bank.Records) {
		return nil, fmt.Errorf("%s: Korean compilation has %d items for %d source records", bank.Name, len(items), len(bank.Records))
	}
	if len(items) > math.MaxUint16 {
		return nil, fmt.Errorf("%s: message count exceeds uint16", bank.Name)
	}
	records := make([][]byte, len(items))
	matched := make(map[int]struct{}, len(replacements))
	for index, item := range items {
		source := bank.Records[index]
		if item.Record.ID != source.ID || item.Translation.ID != source.ID {
			return nil, fmt.Errorf("%s: item %d does not match source ID %d", bank.Name, index, source.ID)
		}
		replacement, ok := replacements[source.ID]
		if !ok {
			records[index] = append([]byte(nil), source.Raw...)
			continue
		}
		matched[source.ID] = struct{}{}
		projection, err := Project(source)
		if err != nil {
			return nil, err
		}
		authored := strings.Contains(replacement.Text, lineBreak)
		text, layout := replacement.Text, authored
		if replacement.Layout != "" {
			if authored {
				if replacement.Layout != replacement.Text {
					return nil, fmt.Errorf("%s: ID %d: generated Korean layout changes explicitly authored line breaks", bank.Name, source.ID)
				}
			} else {
				if _, err := projection.MaterializeKorean(replacement.Text, false, mapping); err != nil {
					return nil, fmt.Errorf("%s: ID %d Korean semantic text: %w", bank.Name, source.ID, err)
				}
				if !preservesSemantics(replacement.Text, replacement.Layout) {
					return nil, fmt.Errorf("%s: ID %d: Korean layout changes semantic/control text; only complete whitespace spans may become line breaks", bank.Name, source.ID)
				}
			}
			text, layout = replacement.Layout, true
		}
		records[index], err = projection.MaterializeKorean(text, layout, mapping)
		if err != nil {
			return nil, fmt.Errorf("%s: ID %d Korean replacement: %w", bank.Name, source.ID, err)
		}
	}
	if len(matched) != len(replacements) {
		unmatched := make([]int, 0, len(replacements)-len(matched))
		for id := range replacements {
			if _, ok := matched[id]; !ok {
				unmatched = append(unmatched, id)
			}
		}
		sort.Ints(unmatched)
		return nil, fmt.Errorf("%s: Korean replacements reference IDs not present in this bank: %v", bank.Name, unmatched)
	}

	tableEnd := uint64(4 + len(records)*4)
	position := tableEnd
	for _, record := range records {
		position += uint64(len(record))
	}
	if position > math.MaxUint32 {
		return nil, fmt.Errorf("%s: compiled Korean bank is %d bytes; uint32 maximum is %d", bank.Name, position, uint64(math.MaxUint32))
	}
	if position > uint64(RuntimeBankCapacity(bank.Section)) {
		return nil, fmt.Errorf("%s: compiled Korean bank is %d bytes; runtime slot capacity is %d (shorten translations by at least %d encoded bytes)", bank.Name, position, RuntimeBankCapacity(bank.Section), position-uint64(RuntimeBankCapacity(bank.Section)))
	}
	output := make([]byte, int(position))
	binary.LittleEndian.PutUint16(output, uint16(len(records)))
	offset := tableEnd
	for index, record := range records {
		binary.LittleEndian.PutUint32(output[4+index*4:], uint32(offset))
		copy(output[int(offset):], record)
		offset += uint64(len(record))
	}
	return output, nil
}
