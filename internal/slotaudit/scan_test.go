// SPDX-License-Identifier: GPL-3.0-or-later

package slotaudit_test

import (
	"testing"

	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/slotaudit"
)

func TestScanCP932LiteralsRecoversJapaneseCStringSuffix(t *testing.T) {
	encoded, err := cp932.Encode("設定")
	if err != nil {
		t.Fatal(err)
	}
	data := append([]byte{0x01, 0xFF, 0x03}, encoded...)
	data = append(data, 0, 0xFF, 0)
	report, err := slotaudit.ScanCP932Literals(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Literals) != 1 || report.Literals[0].Text != "設定" {
		t.Fatalf("literals = %#v", report.Literals)
	}
	for index := 0; index < len(encoded); index += 2 {
		key, err := cp932.GlyphKeyFromBytes(encoded[index : index+2])
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := report.Keys[key]; !ok {
			t.Fatalf("missing key %04X", uint16(key))
		}
	}
}

func TestScanCP932LiteralsAcceptsMixedJapaneseUIString(t *testing.T) {
	encoded, err := cp932.Encode("A日B")
	if err != nil {
		t.Fatal(err)
	}
	report, err := slotaudit.ScanCP932Literals(append(encoded, 0))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Literals) != 1 || report.Literals[0].Text != "A日B" {
		t.Fatalf("mixed literal not recovered: %#v", report.Literals)
	}
}

func TestScanCP932LiteralsRejectsIsolatedBinaryLookalike(t *testing.T) {
	encoded, err := cp932.Encode("日")
	if err != nil {
		t.Fatal(err)
	}
	report, err := slotaudit.ScanCP932Literals(append(encoded, 0))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Literals) != 0 || len(report.Keys) != 0 {
		t.Fatalf("isolated pair treated as string: %#v", report)
	}
}

func TestScanCP932LiteralsIgnoresPlainASCII(t *testing.T) {
	report, err := slotaudit.ScanCP932Literals([]byte("ordinary ascii\x00"))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Literals) != 0 || len(report.Keys) != 0 {
		t.Fatalf("ASCII-only string created renderer references: %#v", report)
	}
}
