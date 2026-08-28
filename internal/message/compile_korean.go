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

	// Diagnostic-only runtime experiment: preserve ID 10010's control token and
	// every renderer key, but remove exactly one ASCII space after validated
	// materialization. The known failing record is 72 bytes; this produces a
	// 71-byte record. A second probe below then restores only msgsec001's total
	// payload size with one byte of recognized archive padding, leaving every
	// record offset unchanged from the successful 71-byte build.
	for index, item := range items {
		if item.Record.ID != 10010 {
			continue
		}
		record := records[index]
		if len(record) != 72 {
			return nil, fmt.Errorf("%s: ID 10010 length probe expected 72-byte materialized record, got %d", bank.Name, len(record))
		}
		if len(record) < 2 || record[0] != 0x02 || record[1] != 0x15 {
			return nil, fmt.Errorf("%s: ID 10010 length probe expected leading <value:$15> bytes 02 15, got % X", bank.Name, record[:minInt(len(record), 2)])
		}
		space := -1
		for i := 2; i < len(record); i++ {
			if record[i] == 0x20 {
				space = i
				break
			}
		}
		if space < 0 {
			return nil, fmt.Errorf("%s: ID 10010 length probe found no ASCII space to remove", bank.Name)
		}
		records[index] = append(append([]byte(nil), record[:space]...), record[space+1:]...)
		if len(records[index]) != 71 {
			return nil, fmt.Errorf("%s: ID 10010 length probe produced %d bytes, want 71", bank.Name, len(records[index]))
		}
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

	// Size-only control: L71 produced msgsec001.dat at 10917 bytes and passed on
	// device. Append exactly one 'Z' byte after the final block terminator. The
	// retail parser recognizes this as the first byte of its canonical
	// "ZillO'll " archive-padding pattern, so no record offset or display content
	// changes. This restores only the bank payload size to the failing 10918.
	if bank.Section == 1 {
		if len(output) != 10917 {
			return nil, fmt.Errorf("%s: tail-pad probe expected L71 bank size 10917, got %d", bank.Name, len(output))
		}
		output = append(output, 'Z')
		if len(output) != 10918 {
			return nil, fmt.Errorf("%s: tail-pad probe produced bank size %d, want 10918", bank.Name, len(output))
		}
		if len(output) > RuntimeBankCapacity(bank.Section) {
			return nil, fmt.Errorf("%s: tail-pad probe exceeds runtime capacity", bank.Name)
		}
	}
	return output, nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
