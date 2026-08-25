// SPDX-License-Identifier: GPL-3.0-or-later

package message

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
)

var koreanMaterializeControl = regexp.MustCompile(`(<value:\$[0-9A-F]{2}>|<line-break>)`)

// SplitSemanticKorean applies the same source-control and fragment rules as
// SplitSemantic, but validates natural text with the exact Korean renderer-slot
// mapping that will later be used to encode it. This prevents validation and
// byte materialization from disagreeing about whether Hangul is representable.
func (p *Projection) SplitSemanticKorean(text string, mapping koreanslots.Mapping) ([]string, error) {
	if text == "" {
		return nil, fmt.Errorf("message %d: canonical text must be nonempty", p.RecordID)
	}
	values := make([]string, len(p.Fragments))
	cursor := 0
	for index, node := range p.nodes {
		if node.fixed {
			if node.display != "" && !strings.HasPrefix(text[cursor:], node.display) {
				return nil, fmt.Errorf("message %d: canonical text changes fixed %s control", p.RecordID, node.kind)
			}
			cursor += len(node.display)
			continue
		}
		next := ""
		for _, following := range p.nodes[index+1:] {
			if !following.fixed {
				break
			}
			next += following.display
		}
		end := len(text)
		if next != "" {
			relative := strings.Index(text[cursor:], next)
			if relative < 0 {
				return nil, fmt.Errorf("message %d: canonical text is missing a fixed control", p.RecordID)
			}
			end = cursor + relative
		}
		values[node.fragment] = text[cursor:end]
		cursor = end
	}
	if cursor != len(text) {
		return nil, fmt.Errorf("message %d: canonical text has trailing material outside its projection", p.RecordID)
	}
	for index, value := range values {
		if err := p.validateFragmentKorean(index, value, mapping); err != nil {
			return nil, err
		}
	}
	return values, nil
}

func (p *Projection) validateFragmentKorean(index int, value string, mapping koreanslots.Mapping) error {
	f := p.Fragments[index]
	found := valueTag.FindAllStringSubmatch(value, -1)
	available := make(map[byte]int, len(f.Substitutions))
	for _, opcode := range f.Substitutions {
		available[opcode]++
	}
	for _, match := range found {
		var opcode byte
		_, _ = fmt.Sscanf(match[1], "%02X", &opcode)
		if available[opcode] == 0 {
			return fmt.Errorf("message %d fragment %s changes runtime substitutions", p.RecordID, f.Key)
		}
		available[opcode]--
	}
	if formatSignatureIDs[p.RecordID] {
		got := printfConversion.FindAllString(value, -1)
		if strings.Count(value, "%") != len(got) || strings.Join(got, "\x00") != strings.Join(f.FormatSignature, "\x00") {
			return fmt.Errorf("message %d fragment %s changes runtime format signature", p.RecordID, f.Key)
		}
	}
	plain := valueTag.ReplaceAllString(value, "")
	plain = strings.ReplaceAll(plain, lineBreak, "")
	if reservedMarkup.MatchString(plain) || reservedAnchor.MatchString(plain) {
		return fmt.Errorf("message %d fragment %s contains reserved markup", p.RecordID, f.Key)
	}
	return validateKoreanText(p.RecordID, f.Key, plain, mapping)
}

func validateKoreanText(id int, key, text string, mapping koreanslots.Mapping) error {
	for _, r := range text {
		if unicode.IsControl(r) {
			return fmt.Errorf("message %d fragment %s contains a raw control character", id, key)
		}
		if r >= 0xff61 && r <= 0xff9f {
			return fmt.Errorf("message %d fragment %s contains half-width kana", id, key)
		}
		if _, mapped := mapping[r]; mapped {
			if _, err := cp932.Encode(string(r)); err == nil {
				return fmt.Errorf("message %d fragment %s maps stock CP932 rune %U to a custom renderer slot", id, key, r)
			}
		}
	}
	encoded, err := koreanslots.Encode(text, mapping)
	if err != nil {
		return fmt.Errorf("message %d fragment %s is not encodable with Korean renderer slots: %w", id, key, err)
	}
	for index := 0; index < len(encoded); index++ {
		if encoded[index] >= 0x81 && encoded[index] <= 0x9f || encoded[index] >= 0xe0 && encoded[index] <= 0xfc {
			if index+1 >= len(encoded) {
				return fmt.Errorf("message %d fragment %s ends with an incomplete two-byte renderer key", id, key)
			}
			index++
		} else if encoded[index] >= 0xa1 && encoded[index] <= 0xdf {
			return fmt.Errorf("message %d fragment %s contains half-width kana", id, key)
		}
	}
	return nil
}

// MaterializeKorean lowers canonical Korean annotated text while preserving all
// source-owned controls. Natural text is encoded with the same mapping already
// checked by SplitSemanticKorean; unmapped Hangul therefore fails closed.
func (p *Projection) MaterializeKorean(text string, layout bool, mapping koreanslots.Mapping) ([]byte, error) {
	values, err := p.SplitSemanticKorean(text, mapping)
	if err != nil {
		return nil, err
	}
	var output []byte
	for _, node := range p.nodes {
		if node.fixed {
			output = append(output, node.raw...)
			continue
		}
		value := values[node.fragment]
		matches := koreanMaterializeControl.FindAllStringIndex(value, -1)
		cursor := 0
		for _, match := range matches {
			encoded, err := koreanslots.Encode(value[cursor:match[0]], mapping)
			if err != nil {
				return nil, err
			}
			output = append(output, encoded...)
			piece := value[match[0]:match[1]]
			if piece == lineBreak {
				if !layout {
					return nil, fmt.Errorf("message %d: semantic text contains a layout break", p.RecordID)
				}
				output = append(output, 10)
			} else {
				var opcode byte
				_, _ = fmt.Sscanf(piece, "<value:$%02X>", &opcode)
				output = append(output, 2, opcode)
			}
			cursor = match[1]
		}
		encoded, err := koreanslots.Encode(value[cursor:], mapping)
		if err != nil {
			return nil, err
		}
		output = append(output, encoded...)
		for _, control := range node.kanaControl {
			output = append(output, control...)
		}
	}
	return output, nil
}
