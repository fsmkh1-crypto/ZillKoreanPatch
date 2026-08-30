// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/fixeddata"
	"github.com/HK47196/zill/internal/koreanslots"
)

func TestPR14EBOOTFixtureIsSeparatedFromCurrentProductionInput(t *testing.T) {
	root := filepath.Join("..", "..")
	currentData, err := os.ReadFile(filepath.Join(root, "release", "korean", "strings", "eboot.toml"))
	if err != nil { t.Fatal(err) }
	current, err := fixeddata.ParseKoreanEBOOT(currentData)
	if err != nil { t.Fatal(err) }

	h0Data, err := os.ReadFile(filepath.Join(root, "docs", "audit", "fixtures", "pr14-eboot-h0.toml"))
	if err != nil { t.Fatal(err) }
	h0, err := fixeddata.ParseKoreanEBOOT(h0Data)
	if err != nil { t.Fatal(err) }

	fixtureData, err := os.ReadFile(filepath.Join(root, "docs", "audit", "fixtures", "pr14-eboot-full.toml"))
	if err != nil { t.Fatal(err) }
	fixture, err := fixeddata.ParseKoreanEBOOT(fixtureData)
	if err != nil { t.Fatal(err) }

	if got, want := len(h0), 2; got != want {
		t.Fatalf("historical H0 EBOOT fixture fields=%d, want %d", got, want)
	}
	if got, want := len(fixture), 46; got != want {
		t.Fatalf("PR14 expanded EBOOT fixture fields=%d, want historical %d", got, want)
	}
	if len(current) <= len(h0) {
		t.Fatalf("current production EBOOT fields=%d must be allowed to grow beyond historical H0=%d", len(current), len(h0))
	}
	for offset, h0Field := range h0 {
		fullField, ok := fixture[offset]
		if !ok {
			t.Fatalf("historical expanded fixture is missing H0 field %#x", offset)
		}
		if fullField.Source != h0Field.Source || fullField.Replacement != h0Field.Replacement {
			t.Fatalf("historical expanded fixture field %#x diverges from H0: h0=%+v expanded=%+v", offset, h0Field, fullField)
		}
		currentField, ok := current[offset]
		if !ok {
			t.Fatalf("current production EBOOT lost established H0 field %#x", offset)
		}
		if currentField.Source != h0Field.Source || currentField.Replacement != h0Field.Replacement {
			t.Fatalf("current production EBOOT field %#x diverges from established H0 field", offset)
		}
	}
}

func TestMappingDifferenceCount(t *testing.T) {
	a := koreanslots.Mapping{
		'가': cp932.GlyphKey(0x8140),
		'나': cp932.GlyphKey(0x8141),
	}
	b := koreanslots.Mapping{
		'가': cp932.GlyphKey(0x8140),
		'나': cp932.GlyphKey(0x8142),
		'다': cp932.GlyphKey(0x8143),
	}
	if got, want := mappingDifferenceCount(a, b), 2; got != want {
		t.Fatalf("mappingDifferenceCount=%d, want %d", got, want)
	}
	if got := mappingDifferenceCount(a, cloneKoreanMapping(a)); got != 0 {
		t.Fatalf("mappingDifferenceCount identical mapping=%d, want 0", got)
	}
}
