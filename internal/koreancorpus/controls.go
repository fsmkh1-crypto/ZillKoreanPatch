// SPDX-License-Identifier: GPL-3.0-or-later

package koreancorpus

import "github.com/HK47196/zill/internal/corpus"

// validateControlContract delegates to the production corpus validator so the
// strict checker and every release/font/mobile caller enforce one authoritative
// Korean runtime-control contract.
func validateControlContract(path string, id int, source, translated, field string) error {
	return corpus.ValidateKoreanControlContract(path, id, source, translated, field)
}
