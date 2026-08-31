// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestPR14HistoricalDiagnosticFailureIsNonBlocking(t *testing.T) {
	var stdout bytes.Buffer
	reportPR14HistoricalDiagnostic(&stdout, errors.New("missing historical fixture"))
	got := stdout.String()
	if !strings.Contains(got, "PR14_POLICY_AUDIT_UNAVAILABLE") {
		t.Fatalf("missing diagnostic marker: %q", got)
	}
	if !strings.Contains(got, "diagnostic_only=true") || !strings.Contains(got, "build_blocked=false") {
		t.Fatalf("diagnostic failure did not declare non-blocking semantics: %q", got)
	}
}

func TestPR14HistoricalDiagnosticSuccessIsSilent(t *testing.T) {
	var stdout bytes.Buffer
	reportPR14HistoricalDiagnostic(&stdout, nil)
	if stdout.Len() != 0 {
		t.Fatalf("successful historical diagnostic wrote unexpected output: %q", stdout.String())
	}
}

func TestBuildKoreanISOPreflightDoesNotRequireOutput(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := runBuildKoreanISO(root, []string{
		"--iso", filepath.Join(root, "missing.iso"),
		"--work-dir", filepath.Join(root, "work"),
		"--version", "test-preflight",
		"--preflight-only",
	}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("preflight exit=%d, want 1 after argument acceptance and missing ISO open; stderr=%q", code, stderr.String())
	}
	if strings.Contains(stderr.String(), "usage:") {
		t.Fatalf("preflight unexpectedly required --out: %q", stderr.String())
	}
}

func TestBuildKoreanISORegularModeStillRequiresOutput(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := runBuildKoreanISO(root, []string{
		"--iso", filepath.Join(root, "missing.iso"),
		"--work-dir", filepath.Join(root, "work"),
		"--version", "test-build",
	}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("regular build exit=%d, want 2 without --out; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "usage:") {
		t.Fatalf("regular build missing usage error: %q", stderr.String())
	}
}

func TestBuildKoreanISOPreflightRejectsOutputArgument(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := runBuildKoreanISO(root, []string{
		"--iso", filepath.Join(root, "missing.iso"),
		"--out", filepath.Join(root, "should-not-exist.iso"),
		"--work-dir", filepath.Join(root, "work"),
		"--version", "test-preflight",
		"--preflight-only",
	}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("preflight with --out exit=%d, want 2; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "--out is not used with --preflight-only") {
		t.Fatalf("unexpected preflight --out error: %q", stderr.String())
	}
}
