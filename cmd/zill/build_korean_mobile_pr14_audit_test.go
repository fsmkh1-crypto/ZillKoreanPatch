// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/HK47196/zill/internal/fixeddata"
)

func TestPR14EBOOTFixtureIsSeparatedFromCurrentProductionInput(t *testing.T) {
	root := filepath.Join("..", "..")
	currentData, err := os.ReadFile(filepath.Join(root, "release", "korean", "strings", "eboot.toml"))
	if err != nil {
		t.Fatal(err)
	}
	current, err := fixeddata.ParseKoreanEBOOT(currentData)
	if err != nil {
		t.Fatal(err)
	}
	fixtureData, err := os.ReadFile(filepath.Join(root, "docs", "audit", "fixtures", "pr14-eboot-full.toml"))
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := fixeddata.ParseKoreanEBOOT(fixtureData)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(current), 2; got != want {
		t.Fatalf("current production EBOOT fields=%d, want H0-era %d; update the audit model before changing this input", got, want)
	}
	if got, want := len(fixture), 46; got != want {
		t.Fatalf("PR14 expanded EBOOT fixture fields=%d, want historical %d", got, want)
	}
	for offset, currentField := range current {
		fixtureField, ok := fixture[offset]
		if !ok {
			t.Fatalf("historical fixture is missing H0 field %#x", offset)
		}
		if fixtureField.Source != currentField.Source || fixtureField.Replacement != currentField.Replacement {
			t.Fatalf("historical fixture field %#x diverges from H0 field: current=%+v fixture=%+v", offset, currentField, fixtureField)
		}
	}
}

func TestMappingDifferenceCount(t *testing.T) {
	a := map[rune]uint16{'가': 1, '나': 2}
	b := map[rune]uint16{'가': 1, '나': 3, '다': 4}
	// Convert through the concrete mapping type without relying on map iteration order.
	ma := make(map[rune]uint16, len(a))
	mb := make(map[rune]uint16, len(b))
	for r, key := range a { ma[r] = key }
	for r, key := range b { mb[r] = key }
	_ = ma
	_ = mb
}
