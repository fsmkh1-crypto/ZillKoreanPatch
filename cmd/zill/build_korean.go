// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/fixeddata"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/release"
	"github.com/HK47196/zill/internal/slotaudit"
)

func runBuildKorean(root string, args []string, stdout, stderr io.Writer) int {
	gameDir, isoPath, version := "", "", ""
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--game-dir" && i+1 < len(args): i++; gameDir = args[i]
		case strings.HasPrefix(args[i], "--game-dir="): gameDir = strings.TrimPrefix(args[i], "--game-dir=")
		case args[i] == "--iso" && i+1 < len(args): i++; isoPath = args[i]
		case strings.HasPrefix(args[i], "--iso="): isoPath = strings.TrimPrefix(args[i], "--iso=")
		case args[i] == "--version" && i+1 < len(args): i++; version = args[i]
		case strings.HasPrefix(args[i], "--version="): version = strings.TrimPrefix(args[i], "--version=")
		default:
			fmt.Fprintf(stderr, "zill: build-korean: unknown or incomplete argument %q\n", args[i])
			return 2
		}
	}
	if gameDir == "" || isoPath == "" {
		fmt.Fprintln(stderr, "zill: usage: zill build-korean --game-dir PATH --iso RETAIL_ISO [--version VERSION]")
		return 2
	}
	resolvedVersion, err := resolveBuildVersion(root, version)
	if err != nil { fmt.Fprintf(stderr, "zill: build-korean: %v\n", err); return 1 }
	plan, coverage, total, err := buildKoreanAlphaPlan(root, gameDir)
	if err != nil { fmt.Fprintf(stderr, "zill: build-korean: %v\n", err); return 1 }
	result, err := release.BuildKoreanAlpha(root, gameDir, isoPath, resolvedVersion, plan)
	if err != nil { fmt.Fprintf(stderr, "zill: build-korean: %v\n", err); return 1 }
	fmt.Fprintf(stdout, "Built Korean beta game tree at %s\n", result.GameDirectory)
	fmt.Fprintf(stdout, "Built Korean beta ISO at %s\n", result.ISO)
	fmt.Fprintf(stdout, "Built Korean beta xdelta patch at %s\n", result.Patch)
	fmt.Fprintf(stdout, "Korean runtime coverage: %d/%d source records; custom glyphs: %d; reusable slots: %d\n", coverage, total, len(plan.CustomRunes), len(plan.Candidates))
	fmt.Fprintf(stdout, "Embedded beta version: %s\n", resolvedVersion)
	for _, warning := range result.Warnings { fmt.Fprintf(stderr, "zill: build-korean: warning: %s\n", warning) }
	return 0
}

func buildKoreanAlphaPlan(root, gameDir string) (koreanslots.Plan, int, int, error) {
	font, err := loadRetailPAF(gameDir)
	if err != nil { return koreanslots.Plan{}, 0, 0, err }
	source, sourceSummary, err := corpus.LoadProject(root)
	if err != nil { return koreanslots.Plan{}, 0, 0, err }
	canonicalKorean, _, err := corpus.LoadKoreanProject(root, source)
	if err != nil { return koreanslots.Plan{}, 0, 0, err }
	korean, skippedStructural, err := release.BuildKoreanBetaProject(source, canonicalKorean)
	if err != nil { return koreanslots.Plan{}, 0, 0, err }
	fmt.Printf("Korean beta projection: canonical=%d materializable=%d structural_retail=%d\n", len(canonicalKorean.Entries), len(korean.Entries), skippedStructural)
	usedFixed, equipment, err := loadFixedRendererKeys(root)
	if err != nil { return koreanslots.Plan{}, 0, 0, err }
	_, err = loadAuthenticatedRetailEBOOT(gameDir)
	if err != nil { return koreanslots.Plan{}, 0, 0, err }
	boot, err := loadAuthenticatedRetailBOOT(gameDir)
	if err != nil { return koreanslots.Plan{}, 0, 0, err }
	bootScan, err := slotaudit.ScanCP932Literals(boot)
	if err != nil { return koreanslots.Plan{}, 0, 0, fmt.Errorf("scan BOOT.BIN: %w", err) }
	bindata, err := loadRetailBindata(gameDir)
	if err != nil { return koreanslots.Plan{}, 0, 0, err }
	if _, err := fixeddata.ApplyEquipment(bindata, equipment); err != nil { return koreanslots.Plan{}, 0, 0, fmt.Errorf("validate bindata.dat layout: %w", err) }
	bindataScan, err := slotaudit.ScanCP932Literals(bindata)
	if err != nil { return koreanslots.Plan{}, 0, 0, fmt.Errorf("scan bindata.dat: %w", err) }

	reserved := make(map[cp932.GlyphKey]struct{})
	mergeRendererKeys(reserved, usedFixed)
	mergeRendererKeys(reserved, bootScan.Keys)
	mergeRendererKeys(reserved, bindataScan.Keys)
	keyboardReserved := koreanslots.KeyboardInputReservedKeys()
	for _, key := range keyboardReserved { reserved[key] = struct{}{} }
	texts, err := korean.RuntimeTexts(source)
	if err != nil { return koreanslots.Plan{}, 0, 0, err }
	fixedKorean, err := loadKoreanFixedEBOOT(root)
	if err != nil { return koreanslots.Plan{}, 0, 0, err }
	texts = append(texts, fixeddata.KoreanEBOOTTexts(fixedKorean)...)
	plan, err := koreanslots.BuildPlan(texts, font.KoreanCompatibleKeys(), rendererKeySetSlice(reserved))
	if err != nil { return koreanslots.Plan{}, 0, 0, fmt.Errorf("plan allocation: %w", err) }
	fmt.Printf("Korean beta slot diagnostics: boot_scan_keys=%d bindata_scan_keys=%d keyboard_keys=%d fixed_korean=%d candidates=%d custom=%d headroom=%d (structured renderer ownership enforced)\n", len(bootScan.Keys), len(bindataScan.Keys), len(keyboardReserved), len(fixedKorean), len(plan.Candidates), len(plan.CustomRunes), len(plan.Candidates)-len(plan.CustomRunes))
	return plan, len(korean.Entries), sourceSummary.Records, nil
}
