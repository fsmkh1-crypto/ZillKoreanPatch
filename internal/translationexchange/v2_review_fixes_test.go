// SPDX-License-Identifier: GPL-3.0-or-later

package translationexchange

import (
	"bytes"
	"strings"
	"testing"
)

func TestProtectedV2REDoesNotSplitPlainPercentSentence(t *testing.T) {
	const text = "Achieved 100% Match rate"
	segments, locked := SplitV2(text)
	if len(segments) != 1 || segments[0].Index != 0 || segments[0].Text != text {
		t.Fatalf("plain percent sentence split incorrectly: segments=%#v locked=%#v", segments, locked)
	}
	if len(locked) != 0 {
		t.Fatalf("plain percent sentence created locked parts: %#v", locked)
	}
}

func TestV1ProtectedTokenCheckDoesNotTreatPlainPercentAsPrintf(t *testing.T) {
	if err := validateProtectedTokens("Achieved 100% Match rate", "100% 일치율 달성"); err != nil {
		t.Fatalf("plain percent sentence was treated as a protected printf token: %v", err)
	}
}

func TestV2StillProtectsActualRuntimePrintf(t *testing.T) {
	segments, locked := SplitV2("Choose %s now")
	if len(segments) != 2 || segments[0].Text != "Choose" || segments[1].Text != "now" {
		t.Fatalf("runtime printf split = segments %#v locked %#v", segments, locked)
	}
	found := false
	for _, part := range locked {
		if part.Text == "%s" {
			found = true
		}
	}
	if !found {
		t.Fatalf("runtime %%s was not protected: %#v", locked)
	}
}

func TestReadExportV2RejectsExternallyModifiedGlossary(t *testing.T) {
	source := BuildSourceV2(1, 1, "x.toml", "原文<end>", "Reference<end>", map[string]string{})
	var clean bytes.Buffer
	if err := WriteExportV2(&clean, []ExportRowV2{source.Export}); err != nil {
		t.Fatal(err)
	}
	modified := strings.Replace(clean.String(), `"glossary":{}`, `"glossary":{"ノエル":"노엘"}`, 1)
	if modified == clean.String() {
		t.Fatalf("test failed to modify glossary: %s", clean.String())
	}
	if _, err := ReadExportV2(strings.NewReader(modified)); err == nil || !strings.Contains(err.Error(), "canonical glossary") {
		t.Fatalf("modified glossary returned %v", err)
	}
}

func TestReadExportV2RejectsProtectedTokenInModifiedGlossary(t *testing.T) {
	source := BuildSourceV2(1, 1, "x.toml", "原文<end>", "Reference<end>", map[string]string{})
	var clean bytes.Buffer
	if err := WriteExportV2(&clean, []ExportRowV2{source.Export}); err != nil {
		t.Fatal(err)
	}
	modified := strings.Replace(clean.String(), `"glossary":{}`, `"glossary":{"ノエル":"노엘<end>"}`, 1)
	if _, err := ReadExportV2(strings.NewReader(modified)); err == nil {
		t.Fatal("protected token injected through glossary was accepted")
	}
}
