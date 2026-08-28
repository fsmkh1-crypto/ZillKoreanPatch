// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/HK47196/zill/internal/fixeddata"
)

func loadKoreanFixedEBOOT(root string) (fixeddata.KoreanEBOOTTranslations, error) {
	data, err := os.ReadFile(filepath.Join(root, "release", "korean", "strings", "eboot.toml"))
	if err != nil {
		return nil, fmt.Errorf("read Korean EBOOT fixed strings: %w", err)
	}
	return fixeddata.ParseKoreanEBOOT(data)
}
