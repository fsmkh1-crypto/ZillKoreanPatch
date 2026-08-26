// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/gamefmt/pspiso"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/release"
)

func runBuildKoreanISO(root string, args []string, stdout, stderr io.Writer) int {
	isoPath, outputPath, workDir, version := "", "", "", ""
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--iso" && i+1 < len(args):
			i++; isoPath = args[i]
		case strings.HasPrefix(args[i], "--iso="):
			isoPath = strings.TrimPrefix(args[i], "--iso=")
		case args[i] == "--out" && i+1 < len(args):
			i++; outputPath = args[i]
		case strings.HasPrefix(args[i], "--out="):
			outputPath = strings.TrimPrefix(args[i], "--out=")
		case args[i] == "--work-dir" && i+1 < len(args):
			i++; workDir = args[i]
		case strings.HasPrefix(args[i], "--work-dir="):
			workDir = strings.TrimPrefix(args[i], "--work-dir=")
		case args[i] == "--version" && i+1 < len(args):
			i++; version = args[i]
		case strings.HasPrefix(args[i], "--version="):
			version = strings.TrimPrefix(args[i], "--version=")
		default:
			fmt.Fprintf(stderr, "zill: build-korean-iso: unknown or incomplete argument %q\n", args[i])
			return 2
		}
	}
	if isoPath == "" || outputPath == "" || workDir == "" {
		fmt.Fprintln(stderr, "zill: usage: zill build-korean-iso --iso RETAIL_ISO --out OUTPUT_ISO --work-dir DIR [--version VERSION]")
		return 2
	}
	resolvedVersion, err := resolveBuildVersion(root, version)
	if err != nil {
		fmt.Fprintf(stderr, "zill: build-korean-iso: %v\n", err)
		return 1
	}
	if err := os.RemoveAll(workDir); err != nil {
		fmt.Fprintf(stderr, "zill: build-korean-iso: clean work dir: %v\n", err)
		return 1
	}
	defer os.RemoveAll(workDir)
	extracted := filepath.Join(workDir, "disc")
	image, err := pspiso.Open(isoPath)
	if err != nil {
		fmt.Fprintf(stderr, "zill: build-korean-iso: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, "Extracting retail ISO for Korean alpha build...")
	if err := image.Extract(extracted); err != nil {
		_ = image.Close()
		fmt.Fprintf(stderr, "zill: build-korean-iso: extract ISO: %v\n", err)
		return 1
	}
	if err := image.Close(); err != nil {
		fmt.Fprintf(stderr, "zill: build-korean-iso: close ISO after extraction: %v\n", err)
		return 1
	}
	gameDir := filepath.Join(extracted, "PSP_GAME")
	fmt.Fprintln(stdout, "Mobile alpha safety mode: authenticated retail banks bound before placeholder planning; BOOT/bindata CP932 scans are diagnostic only.")
	fmt.Fprintln(stdout, "Building Korean alpha ISO...")
	planner := func(source *corpus.Project, korean *corpus.KoreanProject) (koreanslots.Plan, int, int, error) {
		return buildKoreanAlphaPlanMobile(root, gameDir, source, korean)
	}
	if err := release.BuildKoreanAlphaISOOnly(root, gameDir, isoPath, outputPath, resolvedVersion, planner); err != nil {
		fmt.Fprintf(stderr, "zill: build-korean-iso: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "Built Korean alpha ISO at %s\n", outputPath)
	return 0
}
