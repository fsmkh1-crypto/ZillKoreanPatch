// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanfont"
	"github.com/HK47196/zill/internal/koreanslots"
)

func runKoreanFontCheck(root string, args []string, stdout, stderr io.Writer) int {
	if len(args) != 0 {
		fmt.Fprintln(stderr, "zill: usage: zill korean-font-check")
		return 2
	}
	source, _, err := corpus.LoadProject(root)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-font-check: %v\n", err)
		return 1
	}
	korean, _, err := corpus.LoadKoreanProject(root, source)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-font-check: %v\n", err)
		return 1
	}
	required := koreanslots.RequiredCustomRunes(korean.Texts())
	if len(required) == 0 {
		fmt.Fprintln(stdout, "OK: Korean overlay currently requires no custom raster glyphs")
		return 0
	}
	path := filepath.Join(root, "release", "korean", "font", "glyphs.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-font-check: read %s: %v\n", filepath.ToSlash(path), err)
		return 1
	}
	catalog, err := koreanfont.Parse(data)
	if err != nil {
		fmt.Fprintf(stderr, "zill: korean-font-check: %v\n", err)
		return 1
	}
	if err := catalog.RequireRunes(required); err != nil {
		fmt.Fprintf(stderr, "zill: korean-font-check: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "OK: Korean raster catalog covers all %d custom glyphs using %s\n", len(required), catalog.SourceFont)
	return 0
}
