// SPDX-License-Identifier: GPL-3.0-or-later

// Package koreanslots deterministically assigns Unicode characters that are
// not representable by stock CP932 to reusable two-byte renderer keys.
package koreanslots

import (
	"fmt"
	"sort"

	"github.com/HK47196/zill/internal/cp932"
)

// Mapping binds one Unicode rune to one existing two-byte renderer key.
type Mapping map[rune]cp932.GlyphKey

// RequiredCustomRunes returns the sorted unique characters from texts that
// stock CP932 cannot encode. This deliberately measures the actual Korean
// corpus instead of assuming that all Hangul syllables will be needed.
func RequiredCustomRunes(texts []string) []rune {
	set := make(map[rune]struct{})
	for _, text := range texts {
		for _, r := range text {
			if _, err := cp932.Encode(string(r)); err != nil {
				set[r] = struct{}{}
			}
	}
	out := make([]rune, 0, len(set))
	for r := range set {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Allocate assigns sorted unique runes to sorted unique valid two-byte keys.
// The same inputs always produce the same mapping regardless of input order.
func Allocate(runes []rune, available []cp932.GlyphKey) (Mapping, error) {
	runeSet := make(map[rune]struct{}, len(runes))
	for _, r := range runes {
		runeSet[r] = struct{}{}
	}
	sortedRunes := make([]rune, 0, len(runeSet))
	for r := range runeSet {
		sortedRunes = append(sortedRunes, r)
	}
	sort.Slice(sortedRunes, func(i, j int) bool { return sortedRunes[i] < sortedRunes[j] })

	keySet := make(map[cp932.GlyphKey]struct{}, len(available))
	for _, key := range available {
		if !key.IsDoubleByte() {
			return nil, fmt.Errorf("slot key 0x%04X is not a valid two-byte CP932 renderer key", uint16(key))
		}
		keySet[key] = struct{}{}
	}
	sortedKeys := make([]cp932.GlyphKey, 0, len(keySet))
	for key := range keySet {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Slice(sortedKeys, func(i, j int) bool { return sortedKeys[i] < sortedKeys[j] })

	if len(sortedRunes) > len(sortedKeys) {
		return nil, fmt.Errorf("need %d custom glyphs but only %d reusable two-byte slots are available",
			len(sortedRunes), len(sortedKeys))
	}

	out := make(Mapping, len(sortedRunes))
	for i, r := range sortedRunes {
		out[r] = sortedKeys[i]
	}
	return out, nil
}

// Encode emits mapped renderer bytes for custom runes and ordinary CP932 for
// all other runes. Missing mappings therefore fail closed instead of silently
// substituting or corrupting text.
func Encode(text string, mapping Mapping) ([]byte, error) {
	out := make([]byte, 0, len(text))
	for _, r := range text {
		if key, ok := mapping[r]; ok {
			encoded, err := key.Bytes()
			if err != nil {
				return nil, fmt.Errorf("encode mapped %U: %w", r, err)
			}
			out = append(out, encoded...)
			continue
		}
		encoded, err := cp932.Encode(string(r))
		if err != nil {
			return nil, fmt.Errorf("encode %U: no custom slot mapping and not CP932: %w", r, err)
		}
		out = append(out, encoded...)
	}
	return out, nil
}
