// SPDX-License-Identifier: GPL-3.0-or-later

package fixeddata

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/pelletier/go-toml/v2"
)

// KoreanEBOOTTranslations is a sparse Korean overlay for fixed-width strings
// embedded in the patched executable. Unlike the complete English table, only
// reviewed Korean fields need to be present; all other fixed strings remain
// byte-identical to retail.
type KoreanEBOOTTranslations map[uint64]EBOOTField

func ParseKoreanEBOOT(data []byte) (KoreanEBOOTTranslations, error) {
	var raw map[string]EBOOTField
	decoder := toml.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode Korean eboot translations: %w", err)
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("Korean eboot translations are empty")
	}
	translations := make(KoreanEBOOTTranslations, len(raw))
	for key, field := range raw {
		offset, err := strconv.ParseUint(key, 0, 64)
		if err != nil || !strings.HasPrefix(key, "0x") {
			return nil, fmt.Errorf("invalid Korean eboot offset %q", key)
		}
		if field.Source == "" || field.Replacement == "" {
			return nil, fmt.Errorf("Korean eboot field %#x requires source and replacement", offset)
		}
		translations[offset] = field
	}
	return translations, nil
}

// KoreanEBOOTTexts returns replacement strings so renderer-slot planning owns
// every custom rune required by the executable overlay before font generation.
func KoreanEBOOTTexts(translations KoreanEBOOTTranslations) []string {
	offsets := make([]uint64, 0, len(translations))
	for offset := range translations { offsets = append(offsets, offset) }
	sort.Slice(offsets, func(i, j int) bool { return offsets[i] < offsets[j] })
	out := make([]string, 0, len(offsets))
	for _, offset := range offsets { out = append(out, translations[offset].Replacement) }
	return out
}

// ApplyKoreanEBOOT authenticates the manifest-patched executable, verifies each
// retail source field, encodes Korean through the exact runtime slot mapping,
// and writes only the sparse reviewed fields. No field may grow beyond its
// original fixed-width capacity.
func ApplyKoreanEBOOT(source []byte, translations KoreanEBOOTTranslations, mapping koreanslots.Mapping) ([]byte, error) {
	if sha256.Sum256(source) != patchedELFSHA256 {
		return nil, fmt.Errorf("unsupported patched ELF fingerprint")
	}
	type replacement struct { offset, capacity int; data []byte }
	replacements := make([]replacement, 0, len(translations))
	for offset, field := range translations {
		sourceText, err := executableNewlines(field.Source)
		if err != nil { return nil, fmt.Errorf("Korean eboot field %#x source: %w", offset, err) }
		replacementText, err := executableNewlines(field.Replacement)
		if err != nil { return nil, fmt.Errorf("Korean eboot field %#x replacement: %w", offset, err) }
		expected, err := cp932.Encode(sourceText)
		if err != nil { return nil, fmt.Errorf("Korean eboot field %#x source: %w", offset, err) }
		encoded, err := koreanslots.Encode(replacementText, mapping)
		if err != nil { return nil, fmt.Errorf("Korean eboot field %#x replacement: %w", offset, err) }
		if len(encoded) > len(expected) {
			return nil, fmt.Errorf("Korean eboot field %#x replacement uses %d bytes; capacity is %d", offset, len(encoded), len(expected))
		}
		if offset > uint64(len(source)) || uint64(len(expected)+1) > uint64(len(source))-offset {
			return nil, fmt.Errorf("Korean eboot field %#x is outside the ELF", offset)
		}
		start := int(offset)
		if !bytes.Equal(source[start:start+len(expected)], expected) || source[start+len(expected)] != 0 {
			return nil, fmt.Errorf("Korean eboot field %#x source guard does not match", offset)
		}
		replacements = append(replacements, replacement{start, len(expected), encoded})
	}
	sort.Slice(replacements, func(i, j int) bool { return replacements[i].offset < replacements[j].offset })
	for i := 1; i < len(replacements); i++ {
		if replacements[i].offset < replacements[i-1].offset+replacements[i-1].capacity {
			return nil, fmt.Errorf("Korean eboot translation fields overlap")
		}
	}
	result := bytes.Clone(source)
	for _, replacement := range replacements {
		clear(result[replacement.offset:replacement.offset+replacement.capacity])
		copy(result[replacement.offset:], replacement.data)
	}
	return result, nil
}
