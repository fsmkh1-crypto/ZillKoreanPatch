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
func validateControlContract(path string, id int, source, translated, field string) error {
	want := corpus.FixedRuntimeControlTags(source)
	got := corpus.FixedRuntimeControlTags(translated)
	if !slices.Equal(got, want) {
		return fmt.Errorf("%s: ID %d: %s changes fixed control sequence: got %v, want %v", path, id, field, got, want)
	}
	return nil
}
