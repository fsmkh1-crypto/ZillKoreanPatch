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

func reportPR14HistoricalDiagnostic(stdout io.Writer, err error) {
	if err == nil {
		return
	}
	fmt.Fprintf(stdout, "FORENSIC PR14_POLICY_AUDIT_UNAVAILABLE error=%q diagnostic_only=true build_blocked=false\n", err.Error())
}

func runBuildKoreanISO(root string, args []string, stdout, stderr io.Writer) int {
	isoPath, outputPath, workDir, version := "", "", "", ""
	preflightOnly := false
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--iso" && i+1 < len(args):
			i++
			isoPath = args[i]
		case strings.HasPrefix(args[i], "--iso="):
			isoPath = strings.TrimPrefix(args[i], "--iso=")
		case args[i] == "--out" && i+1 < len(args):
			i++
			outputPath = args[i]
		case strings.HasPrefix(args[i], "--out="):
			outputPath = strings.TrimPrefix(args[i], "--out=")
		case args[i] == "--work-dir" && i+1 < len(args):
			i++
			workDir = args[i]
		case strings.HasPrefix(args[i], "--work-dir="):
			workDir = strings.TrimPrefix(args[i], "--work-dir=")
		case args[i] == "--version" && i+1 < len(args):
			i++
			version = args[i]
		case strings.HasPrefix(args[i], "--version="):
			version = strings.TrimPrefix(args[i], "--version=")
		case args[i] == "--preflight-only":
			preflightOnly = true
		default:
			fmt.Fprintf(stderr, "zill: build-korean-iso: unknown or incomplete argument %q\n", args[i])
			return 2
		}
	}
	if isoPath == "" || workDir == "" || (!preflightOnly && outputPath == "") {
		fmt.Fprintln(stderr, "zill: usage: zill build-korean-iso --iso RETAIL_ISO [--out OUTPUT_ISO] --work-dir DIR [--version VERSION] [--preflight-only]")
		return 2
	}
	if preflightOnly && outputPath != "" {
		fmt.Fprintln(stderr, "zill: build-korean-iso: --out is not used with --preflight-only")
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
	fmt.Fprintln(stdout, "Extracting authenticated retail ISO for Korean beta build...")
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
	fmt.Fprintln(stdout, "FORENSIC: recovering authenticated retail CDC context for message 10010...")
	if err := auditFocusRecordContext(root, gameDir, stdout); err != nil {
		fmt.Fprintf(stdout, "FORENSIC C5_FOCUS unavailable: %v\n", err)
	}
	fmt.Fprintln(stdout, "Mobile beta safety mode: retail banks are authenticated and bound before slot planning and canonical Korean compilation.")
	planner := func(source *corpus.Project, korean *corpus.KoreanProject) (koreanslots.Plan, int, int, error) {
		if err := auditC5RuntimeCandidates(gameDir); err != nil {
			return koreanslots.Plan{}, 0, 0, fmt.Errorf("C5 runtime candidate audit: %w", err)
		}
		// PR14 H0/B/A/Combined replay is historical, diagnostic-only evidence.
		// Its plans are never consumed by the production builder, so a missing or
		// stale historical fixture must not turn that diagnostic into a release
		// blocker. The current production planner and downstream contract gates
		// remain fail-closed immediately below.
		reportPR14HistoricalDiagnostic(stdout, auditPR14HistoricalPolicies(root, gameDir, source, korean))
		return buildKoreanAlphaPlanMobile(root, gameDir, source, korean)
	}
	if preflightOnly {
		fmt.Fprintln(stdout, "FORENSIC MOBILE_PREFLIGHT_BEGIN output_iso_written=false")
		if err := release.PreflightKoreanAlphaISOOnly(root, gameDir, isoPath, resolvedVersion, planner); err != nil {
			fmt.Fprintf(stdout, "FORENSIC MOBILE_PREFLIGHT_ERROR error=%q\n", err.Error())
			fmt.Fprintf(stderr, "zill: build-korean-iso preflight: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "FORENSIC MOBILE_PREFLIGHT_COMPLETE output_iso_written=false")
		return 0
	}

	fmt.Fprintln(stdout, "Building Korean beta ISO from reviewed canonical corpus...")
	if err := release.BuildKoreanAlphaISOOnly(root, gameDir, isoPath, outputPath, resolvedVersion, planner); err != nil {
		fmt.Fprintf(stdout, "FORENSIC MOBILE_BUILD_ERROR error=%q\n", err.Error())
		fmt.Fprintf(stderr, "zill: build-korean-iso: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "Built Korean beta ISO at %s\n", outputPath)
	return 0
}
