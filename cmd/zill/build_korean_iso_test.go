// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

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
