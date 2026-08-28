// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"fmt"

	"github.com/HK47196/zill/internal/fixeddata"
	"github.com/HK47196/zill/internal/koreanslots"
)

func applyKoreanFixedEBOOT(root string, source []byte, mapping koreanslots.Mapping) ([]byte, error) {
	data, err := read(root, "release", "korean", "strings", "eboot.toml")
	if err != nil {
		return nil, fmt.Errorf("read Korean EBOOT fixed strings: %w", err)
	}
	translations, err := fixeddata.ParseKoreanEBOOT(data)
	if err != nil {
		return nil, err
	}
	return fixeddata.ApplyKoreanEBOOT(source, translations, mapping)
}
