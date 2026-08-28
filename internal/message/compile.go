// SPDX-License-Identifier: GPL-3.0-or-later

package message

import (
	"encoding/binary"
	"fmt"
	"math"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/HK47196/zill/internal/corpus"
)

// RuntimeBankCapacity returns the relocated runtime slot size for a section.
func RuntimeBankCapacity(section int) int {
	if section == 0 {
		return 0x9000
	}
	if section >= 165 && section <= 167 {
		return 0x4000
	}
	return 0x17000
}

// CompileBank lowers a complete bank's joined corpus items into the patched
// runtime's aligned absolute uint32-offset representation. Todo and
// keep_japanese records remain byte-identical to their retail source.
func CompileBank(bank corpus.Bank, items []corpus.Item) ([]byte, error) {
	if len(items) != len(bank.Records) {
		return nil, fmt.Errorf("%s: compilation has %d items for %d source records", bank.Name, len(items), len(bank.Records))
	}
	if len(items) > math.MaxUint16 {
		return nil, fmt.Errorf("%s: message count exceeds uint16", bank.Name)
	}
	records := make([][]byte, len(items))
	for index, item := range items {
		source := bank.Records[index]
		if item.Record.ID != source.ID || item.Translation.ID != source.ID {
			return nil, fmt.Errorf("%s: item %d does not match source ID %d", bank.Name, index, source.ID)
		}
		switch item.Translation.State {
		case corpus.Todo, corpus.KeepJapanese:
			records[index] = append([]byte(nil), source.Raw...)
		case corpus.Translated:
			projection, err := Project(source)
			if err != nil {
				return nil, err
			}
			authored := strings.Contains(item.Translation.Text, lineBreak)
			text, layout := item.Translation.Text, authored
			if item.Layout != "" {
				if authored {
					if item.Layout != item.Translation.Text {
						return nil, fmt.Errorf("%s: ID %d: generated layout changes explicitly authored line breaks", bank.Name, source.ID)
					}
				} else {
					if _, err := projection.Materialize(item.Translation.Text, false); err != nil {
						return nil, fmt.Errorf("%s: ID %d semantic text: %w", bank.Name, source.ID, err)
					}
					if !preservesSemantics(item.Translation.Text, item.Layout) {
						return nil, fmt.Errorf("%s: ID %d: layout changes semantic/control text; line breaks may replace complete whitespace spans or split adjacent Hangul syllables", bank.Name, source.ID)
					}
				}
				text, layout = item.Layout, true
			}
			records[index], err = projection.Materialize(text, layout)
			if err != nil {
				return nil, fmt.Errorf("%s: ID %d: %w", bank.Name, source.ID, err)
			}
		default:
			return nil, fmt.Errorf("%s: ID %d: unsupported translation state %q", bank.Name, source.ID, item.Translation.State)
		}
	}
	tableEnd := uint64(4 + len(records)*4)
	position := tableEnd
	for _, record := range records {
		position += uint64(len(record))
	}
	if position > math.MaxUint32 {
		return nil, fmt.Errorf("%s: compiled bank is %d bytes; uint32 maximum is %d", bank.Name, position, uint64(math.MaxUint32))
	}
	if position > uint64(RuntimeBankCapacity(bank.Section)) {
		return nil, fmt.Errorf("%s: compiled bank is %d bytes; runtime slot capacity is %d (shorten translations by at least %d encoded bytes)", bank.Name, position, RuntimeBankCapacity(bank.Section), position-uint64(RuntimeBankCapacity(bank.Section)))
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

type semanticUnit struct{ kind, value string }

var annotatedControl = regexp.MustCompile(`<[^<>]+>`)

// semanticUnits tokenizes ordinary text at Unicode-rune granularity while
// keeping annotated controls atomic. Rune granularity is deliberate: Korean
// layout may need to wrap inside an unspaced Hangul word without changing the
// translator-owned semantic string.
func semanticUnits(text string) []semanticUnit {
	var units []semanticUnit
	for len(text) > 0 {
		if location := annotatedControl.FindStringIndex(text); location != nil && location[0] == 0 {
			value := text[:location[1]]
			kind := "literal"
			if value == lineBreak {
				kind = "boundary"
			}
			units = append(units, semanticUnit{kind, value})
			text = text[location[1]:]
			continue
		}
		end := len(text)
		if location := annotatedControl.FindStringIndex(text); location != nil {
			end = location[0]
		}
		plain := text[:end]
		for len(plain) > 0 {
			r, width := utf8.DecodeRuneInString(plain)
			kind := "literal"
			if unicode.IsSpace(r) {
				kind = "whitespace"
			}
			units = append(units, semanticUnit{kind, plain[:width]})
			plain = plain[width:]
		}
		text = text[end:]
	}
	return units
}

func isHangulSyllableUnit(unit semanticUnit) bool {
	if unit.kind != "literal" {
		return false
	}
	r, width := utf8.DecodeRuneInString(unit.value)
	return width == len(unit.value) && r >= 0xAC00 && r <= 0xD7A3
}

func preservesSemantics(semantic, layout string) bool {
	want, got := semanticUnits(semantic), semanticUnits(layout)
	wantIndex, gotIndex := 0, 0
	for gotIndex < len(got) {
		if wantIndex < len(want) && want[wantIndex] == got[gotIndex] {
			wantIndex++
			gotIndex++
			continue
		}

		// Preserve the existing layout contract that one complete whitespace span
		// may normalize to another whitespace span (for example full-width space to
		// ASCII space) without changing wording.
		if wantIndex < len(want) && want[wantIndex].kind == "whitespace" && got[gotIndex].kind == "whitespace" {
			for wantIndex < len(want) && want[wantIndex].kind == "whitespace" {
				wantIndex++
			}
			for gotIndex < len(got) && got[gotIndex].kind == "whitespace" {
				gotIndex++
			}
			continue
		}
		if got[gotIndex].kind != "boundary" {
			return false
		}

		// A generated line break may replace a semantic whitespace span. A single
		// boundary may collapse any non-empty span; multiple consecutive boundaries
		// are accepted only when the semantic text owns at least that many whitespace
		// runes. This preserves intentional blank lines without permitting layout to
		// invent extra vertical structure at a zero-width semantic position.
		boundaryEnd := gotIndex
		for boundaryEnd < len(got) && got[boundaryEnd].kind == "boundary" {
			boundaryEnd++
		}
		boundaryCount := boundaryEnd - gotIndex
		if wantIndex < len(want) && want[wantIndex].kind == "whitespace" {
			whitespaceEnd := wantIndex
			for whitespaceEnd < len(want) && want[whitespaceEnd].kind == "whitespace" {
				whitespaceEnd++
			}
			whitespaceCount := whitespaceEnd - wantIndex
			if boundaryCount > 1 && boundaryCount > whitespaceCount {
				return false
			}
			wantIndex = whitespaceEnd
			gotIndex = boundaryEnd
			continue
		}

		// Korean-specific reflow: allow exactly one zero-width layout boundary
		// between two adjacent precomposed Hangul syllables. Repeated zero-width
		// boundaries remain invalid unless backed by semantic whitespace above.
		if boundaryCount == 1 && wantIndex > 0 && wantIndex < len(want) &&
			isHangulSyllableUnit(want[wantIndex-1]) && isHangulSyllableUnit(want[wantIndex]) {
			gotIndex++
			continue
		}
		return false
	}
	return wantIndex == len(want)
}

// PreservesLayoutSemantics reports whether generated Korean layout differs
// from semantic text only by the layout boundaries and whitespace normalization
// accepted by the runtime compiler. Keep preflight validation and compilation
// on this single authoritative contract.
func PreservesLayoutSemantics(semantic, layout string) bool {
	return preservesSemantics(semantic, layout)
}
