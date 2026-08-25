// SPDX-License-Identifier: GPL-3.0-or-later

package message

import (
	"fmt"
	"unicode"

	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
)

// SplitSemanticKorean applies the same source-control and fragment rules as
// SplitSemantic, but validates natural text with the exact Korean renderer-slot
// mapping that will later be used to encode it. The shared traversal prevents
// stock and Korean control/substitution rules from drifting apart.
func (p *Projection) SplitSemanticKorean(text string, mapping koreanslots.Mapping) ([]string, error) {
	return p.splitSemanticWith(text, func(id int, key, plain string) error {
		return validateKoreanText(id, key, plain, mapping)
	})
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
	return p.materializeValues(values, layout, func(value string) ([]byte, error) {
		return koreanslots.Encode(value, mapping)
	})
}
