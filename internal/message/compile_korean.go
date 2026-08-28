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
// copies every other retail record unchanged. Semantic Korean must be layout-free;
// optional generated Layout may insert line breaks only while preserving all
// semantic/control text.
//
// Record-local materialization failures are accumulated across the whole bank
// before returning. This keeps device beta testing from degenerating into a
// one-record-per-build error chase while preserving fail-closed output: no bank
// bytes are emitted unless every selected replacement materializes cleanly.
func CompileBankKorean(bank corpus.Bank, items []corpus.Item, replacements map[int]KoreanRecord, mapping koreanslots.Mapping) ([]byte, error) {
	if len(items) != len(bank.Records) {
		return nil, fmt.Errorf("%s: Korean compilation has %d items for %d source records", bank.Name, len(items), len(bank.Records))
	}
	if len(items) > math.MaxUint16 {
		return nil, fmt.Errorf("%s: message count exceeds uint16", bank.Name)
	}
	records := make([][]byte, len(items))
	matched := make(map[int]struct{}, len(replacements))
	var failures []string
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
			failures = append(failures, fmt.Sprintf("%s: ID %d projection: %v", bank.Name, source.ID, err))
			continue
		}
		semantic, err := projection.MaterializeKorean(replacement.Text, false, mapping)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: ID %d Korean semantic text: %v", bank.Name, source.ID, err))
			continue
		}
		if replacement.Layout == "" {
			records[index] = semantic
			continue
		}
		if !preservesSemantics(replacement.Text, replacement.Layout) {
			failures = append(failures, fmt.Sprintf("%s: ID %d: Korean layout changes semantic/control text; only layout boundaries may replace semantic whitespace", bank.Name, source.ID))
			continue
		}
		records[index], err = projection.MaterializeKorean(replacement.Layout, true, mapping)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: ID %d Korean layout: %v", bank.Name, source.ID, err))
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
		failures = append(failures, fmt.Sprintf("%s: Korean replacements reference IDs not present in this bank: %v", bank.Name, unmatched))
	}
	if len(failures) > 0 {
		return nil, fmt.Errorf("Korean materialization failed:\n- %s", strings.Join(failures, "\n- "))
	}

	// Diagnostic-only runtime experiment: keep ID 10010 at exactly 72 bytes,
	// preserve <value:$15> and the complete renderer-key set, but relocate one
	// ASCII space. The first space (between '는' and '그대가') is removed and an
	// extra space is inserted immediately after the next existing space. This
	// separates a pure 72-byte threshold from a location-specific whitespace
	// interaction. Canonical Korean data is not modified.
	for index, item := range items {
		if item.Record.ID != 10010 {
			continue
		}
		record := records[index]
		if len(record) != 72 {
			return nil, fmt.Errorf("%s: ID 10010 relocated-space probe expected 72-byte materialized record, got %d", bank.Name, len(record))
		}
		if len(record) < 2 || record[0] != 0x02 || record[1] != 0x15 {
			return nil, fmt.Errorf("%s: ID 10010 relocated-space probe expected leading <value:$15> bytes 02 15, got % X", bank.Name, record[:minInt(len(record), 2)])
		}
		firstSpace := -1
		for i := 2; i < len(record); i++ {
			if record[i] == 0x20 {
				firstSpace = i
				break
			}
		}
		if firstSpace < 0 {
			return nil, fmt.Errorf("%s: ID 10010 relocated-space probe found no first ASCII space", bank.Name)
		}

		shorter := append(append([]byte(nil), record[:firstSpace]...), record[firstSpace+1:]...)
		if len(shorter) != 71 {
			return nil, fmt.Errorf("%s: ID 10010 relocated-space probe intermediate length=%d, want 71", bank.Name, len(shorter))
		}
		insertAfter := -1
		for i := firstSpace; i < len(shorter); i++ {
			if shorter[i] == 0x20 {
				insertAfter = i + 1
				break
			}
		}
		if insertAfter < 0 {
			return nil, fmt.Errorf("%s: ID 10010 relocated-space probe found no later ASCII space", bank.Name)
		}
		relocated := make([]byte, 0, 72)
		relocated = append(relocated, shorter[:insertAfter]...)
		relocated = append(relocated, 0x20)
		relocated = append(relocated, shorter[insertAfter:]...)
		if len(relocated) != 72 {
			return nil, fmt.Errorf("%s: ID 10010 relocated-space probe produced %d bytes, want 72", bank.Name, len(relocated))
		}
		if relocated[firstSpace] == 0x20 {
			return nil, fmt.Errorf("%s: ID 10010 relocated-space probe failed to remove the original first space", bank.Name)
		}
		if insertAfter < 2 || relocated[insertAfter-1] != 0x20 || relocated[insertAfter] != 0x20 {
			return nil, fmt.Errorf("%s: ID 10010 relocated-space probe failed to create relocated double space", bank.Name)
		}
		records[index] = relocated
		break
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

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
