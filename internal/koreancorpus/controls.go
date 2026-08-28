// SPDX-License-Identifier: GPL-3.0-or-later

package koreancorpus

import (
	"fmt"
	"slices"

	"github.com/HK47196/zill/internal/corpus"
)

// validateControlContract catches translator-facing control corruption before
// retail banks are available. Line breaks are layout-authorable, but every
// other canonical runtime control must remain byte-for-byte identical and in
// the same order as the Japanese display projection. Angle-bracketed natural
// text such as <未使用> is intentionally not classified as runtime bytecode.
//
// It also protects internal literal slots between fixed controls. A translation
// may change the wording in such a slot, but if the Japanese source owns visible
// text there, that slot must not become completely empty. Leading/trailing
// literals are deliberately excluded because natural Korean can legitimately
// omit discourse fillers such as Japanese "さ、" while preserving the message.
// The internal-slot check catches dropped branch-local literals that preserve
// every control tag and therefore evade a pure control-sequence comparison.
func validateControlContract(path string, id int, source, translated, field string) error {
	want := corpus.FixedRuntimeControlTags(source)
	got := corpus.FixedRuntimeControlTags(translated)
	if !slices.Equal(got, want) {
		return fmt.Errorf("%s: ID %d: %s changes fixed control sequence: got %v, want %v", path, id, field, got, want)
	}

	wantLiteral := corpus.FixedRuntimeLiteralOccupancy(source)
	gotLiteral := corpus.FixedRuntimeLiteralOccupancy(translated)
	if len(gotLiteral) != len(wantLiteral) {
		return fmt.Errorf("%s: ID %d: %s changes fixed literal slot count: got %d, want %d", path, id, field, len(gotLiteral), len(wantLiteral))
	}
	for i := 1; i+1 < len(wantLiteral); i++ {
		if wantLiteral[i] && !gotLiteral[i] {
			return fmt.Errorf("%s: ID %d: %s drops fixed literal slot %d between runtime controls", path, id, field, i)
		}
	}
	return nil
}
