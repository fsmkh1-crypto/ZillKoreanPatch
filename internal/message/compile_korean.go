// SPDX-License-Identifier: GPL-3.0-or-later
// Runtime A/B build trigger only; no executable behavior change.

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

var characterChoiceBufferDiagnostic = map[int]string{
	10016: "약점을 찾아 여러 방법 시도<end>",
	10017: "힘을 믿고 정면 싸운다<end>",
	10020: "초원의 아름다움 그림에 담는다<end>",
	10025: "강인한 의지의 늠름한 표정<end>",
	10026: "모든 것을 감싸는 온화한 표정<end>",
	10034: "폭발적 파괴력을 내는 체력<end>",
	10071: "물살을 거슬러 오르는 물고기<end>",
}

const opening210065SafeLayout = "광대한 대지 바이아시온 대륙.<line-break>너무나 넓어 지도에도 기록되지<line-break>않고 여행자에게조차 알려지지 않은<line-break>작은 마을이 있다…. 마을의 이름은<line-break>미이스. 그곳에는 작은 신전과 숲,<line-break>그리고 평온한 일상 정도뿐이었다.<line-break>위대한 혼의 이야기는<line-break>여기서 시작된다…….<end>"

const (
	wide32BoundaryProbeTargetID     = 10007
	wide32BoundaryProbeTargetOffset = uint64(0x10020)
)

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

		// Runtime diagnostic: these verified character-creation choices are copied
		// through a 31-byte consumer buffer (30 bytes plus terminator). The canonical
		// Korean strings exceeded that limit by 1-4 bytes. Shorten only those seven
		// records, without introducing any new Korean runes.
		if text, ok := characterChoiceBufferDiagnostic[source.ID]; ok {
			replacement.Text = text
			replacement.Layout = ""
		}

		// Runtime diagnostic inherited from H0: keep one stock ASCII separator after
		// the movable player-name substitution so this combined experiment does not
		// change the historical 10010 condition.
		if source.ID == 10010 {
			replacement.Text = strings.Replace(replacement.Text, "<value:$15>", "<value:$15> ", 1)
			if replacement.Layout != "" {
				replacement.Layout = strings.Replace(replacement.Layout, "<value:$15>", "<value:$15> ", 1)
			}
		}

		// Runtime diagnostic: ID 210065 is a verified C22 consumer. Use the authored
		// eight-line layout whose individual lines are 20-33 encoded bytes, well
		// below the 56-byte C22 line limit, while preserving semantic text exactly.
		if source.ID == 210065 {
			if replacement.Text != "광대한 대지 바이아시온 대륙. 너무나 넓어 지도에도 기록되지 않고 여행자에게조차 알려지지 않은 작은 마을이 있다…. 마을의 이름은 미이스. 그곳에는 작은 신전과 숲, 그리고 평온한 일상 정도뿐이었다. 위대한 혼의 이야기는 여기서 시작된다…….<end>" {
				failures = append(failures, fmt.Sprintf("%s: ID 210065 combined diagnostic semantic precondition failed", bank.Name))
				continue
			}
			replacement.Layout = opening210065SafeLayout
		}

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

	tableEnd := uint64(4 + len(records)*4)
	probeIndex := -1
	probePadding := uint64(0)

	// Runtime A/B probe: on the real bank 001 only, move the explicit early-path
	// record 10007 to 0x10020. The padding is an unreferenced gap immediately
	// before 10007, so no record payload bytes are changed by this probe. Do not
	// silently choose a later record: if 10007 cannot fit, fail closed because a
	// different target would answer a different runtime question.
	if bank.Section == 1 && bank.Name == "msgsec001.dat" {
		for index, source := range bank.Records {
			if source.ID == wide32BoundaryProbeTargetID {
				probeIndex = index
				break
			}
		}
		// Small synthetic unit-test banks also use the retail filename. They do not
		// contain 10007 and therefore remain ordinary compiler fixtures.
		if probeIndex >= 0 {
			baseSize := tableEnd
			for _, record := range records {
				baseSize += uint64(len(record))
			}
			probeOffset := tableEnd
			for index := 0; index < probeIndex; index++ {
				probeOffset += uint64(len(records[index]))
			}
			if probeOffset >= wide32BoundaryProbeTargetOffset {
				return nil, fmt.Errorf("%s: ID %d baseline offset is already 0x%X; boundary probe expected it below 0x%X", bank.Name, wide32BoundaryProbeTargetID, probeOffset, wide32BoundaryProbeTargetOffset)
			}
			probePadding = wide32BoundaryProbeTargetOffset - probeOffset
			capacity := uint64(RuntimeBankCapacity(bank.Section))
			finalSize := baseSize + probePadding
			if finalSize > capacity {
				return nil, fmt.Errorf("%s: forcing ID %d from 0x%X to 0x%X requires %d padding bytes and final size %d, exceeding runtime slot capacity %d by %d", bank.Name, wide32BoundaryProbeTargetID, probeOffset, wide32BoundaryProbeTargetOffset, probePadding, finalSize, capacity, finalSize-capacity)
			}
			fmt.Printf("FORENSIC WIDE32_BOUNDARY bank=%s id=%d index=%d old_offset=0x%X new_offset=0x%X padding=%d base_size=%d final_size=%d capacity=%d\n",
				bank.Name, wide32BoundaryProbeTargetID, probeIndex, probeOffset, wide32BoundaryProbeTargetOffset, probePadding, baseSize, finalSize, capacity)
		}
	}

	position := tableEnd + probePadding
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
		if index == probeIndex {
			offset += probePadding
		}
		binary.LittleEndian.PutUint32(output[4+index*4:], uint32(offset))
		copy(output[int(offset):], record)
		offset += uint64(len(record))
	}
	return output, nil
}
