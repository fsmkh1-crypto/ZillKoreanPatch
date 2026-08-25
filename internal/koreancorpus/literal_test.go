// SPDX-License-Identifier: GPL-3.0-or-later

package koreancorpus

import "testing"

func TestParseSectionAllowsBracketedNaturalTextTranslation(t *testing.T) {
	data := []byte(licenseLine + `

["10007"]
japanese = "<未使用><end>"
korean = "<미사용><end>"
`)
	if _, err := parseSection(data, "msgsec001-part01.toml", 1, map[int]string{10007: "<未使用><end>"}); err != nil {
		t.Fatalf("bracketed natural text rejected as control bytecode: %v", err)
	}
}
