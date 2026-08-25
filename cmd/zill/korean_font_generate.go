// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanfont"
	"github.com/HK47196/zill/internal/koreanslots"
)

func runKoreanFontGenerate(root string, args []string, stdout, stderr io.Writer) int {
	fontPath := ""
	outputPath := filepath.Join(root, "release", "korean", "font", "glyphs.toml")
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--font":
			if i+1 == len(args) { fmt.Fprintln(stderr, "zill: korean-font-generate: --font requires a path"); return 2 }
			i++; fontPath = args[i]
		case strings.HasPrefix(args[i], "--font="):
			fontPath = strings.TrimPrefix(args[i], "--font=")
		case args[i] == "--output":
			if i+1 == len(args) { fmt.Fprintln(stderr, "zill: korean-font-generate: --output requires a path"); return 2 }
			i++; outputPath = args[i]
		case strings.HasPrefix(args[i], "--output="):
			outputPath = strings.TrimPrefix(args[i], "--output=")
		default:
			fmt.Fprintf(stderr, "zill: korean-font-generate: unknown argument %q\n", args[i]); return 2
		}
	}
	if fontPath == "" {
		fmt.Fprintln(stderr, "zill: usage: zill korean-font-generate --font FONT.ttf [--output PATH]")
		return 2
	}
	source, _, err := corpus.LoadProject(root)
	if err != nil { fmt.Fprintf(stderr, "zill: korean-font-generate: %v\n", err); return 1 }
	korean, _, err := corpus.LoadKoreanProject(root, source)
	if err != nil { fmt.Fprintf(stderr, "zill: korean-font-generate: %v\n", err); return 1 }
	required := koreanslots.RequiredCustomRunes(korean.Texts())
	if len(required) == 0 {
		fmt.Fprintln(stdout, "No custom Korean glyphs are currently required; catalog not changed")
		return 0
	}
	sort.Slice(required, func(i, j int) bool { return required[i] < required[j] })
	fontData, err := os.ReadFile(fontPath)
	if err != nil { fmt.Fprintf(stderr, "zill: korean-font-generate: read source font: %v\n", err); return 1 }
	rasters, err := koreanfont.RenderRequired(fontData, required)
	if err != nil { fmt.Fprintf(stderr, "zill: korean-font-generate: %v\n", err); return 1 }
	fontDigest := fmt.Sprintf("%x", sha256.Sum256(fontData))
	catalog, err := koreanfont.Encode(filepath.Base(fontPath)+" sha256:"+fontDigest, koreanfont.ProvenRenderRule, rasters)
	if err != nil { fmt.Fprintf(stderr, "zill: korean-font-generate: %v\n", err); return 1 }
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil { fmt.Fprintf(stderr, "zill: korean-font-generate: create output directory: %v\n", err); return 1 }
	if err := os.WriteFile(outputPath, catalog, 0o644); err != nil { fmt.Fprintf(stderr, "zill: korean-font-generate: write catalog: %v\n", err); return 1 }
	fmt.Fprintf(stdout, "Generated %d Korean glyph rasters at %s\n", len(required), filepath.ToSlash(outputPath))
	fmt.Fprintf(stdout, "Source font SHA-256: %s\n", fontDigest)
	return 0
}
