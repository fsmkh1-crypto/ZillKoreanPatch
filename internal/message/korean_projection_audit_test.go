// SPDX-License-Identifier: GPL-3.0-or-later

package message

import (
	"encoding/binary"
	"testing"

	"github.com/HK47196/zill/internal/corpus"
)

func TestVerifyKoreanProjectionCompatibilityFixedBreakBeforeControl(t *testing.T) {
	bank := auditBank(t, 1, []byte{'A', 10, 5, 'B', 0})
	checked, err := VerifyKoreanProjectionCompatibility(bank)
	if err != nil {
		t.Fatal(err)
	}
	if checked != 1 {
		t.Fatalf("checked %d records, want 1", checked)
	}
}

func TestVerifyKoreanProjectionCompatibilityFixedBreakAfterControl(t *testing.T) {
	bank := auditBank(t, 1, []byte{'A', 5, 10, 'B', 0})
	checked, err := VerifyKoreanProjectionCompatibility(bank)
	if err != nil {
		t.Fatal(err)
	}
	if checked != 1 {
		t.Fatalf("checked %d records, want 1", checked)
	}
}

func TestVerifyKoreanProjectionCompatibilityMovableBreak(t *testing.T) {
	bank := auditBank(t, 1, []byte{'A', 10, 'B', 0})
	checked, err := VerifyKoreanProjectionCompatibility(bank)
	if err != nil {
		t.Fatal(err)
	}
	if checked != 1 {
		t.Fatalf("checked %d records, want 1", checked)
	}
}

func TestVerifyKoreanProjectionCompatibilityMovableSubstitution(t *testing.T) {
	bank := auditBank(t, 1, []byte{'A', 2, 0x15, 'B', 0})
	checked, err := VerifyKoreanProjectionCompatibility(bank)
	if err != nil {
		t.Fatal(err)
	}
	if checked != 1 {
		t.Fatalf("checked %d records, want 1", checked)
	}
}

func TestVerifyKoreanProjectionCompatibilityRetailHalfWidthKana(t *testing.T) {
	bank := auditBank(t, 1, []byte{0xB1, 0}) // CP932 half-width ア
	checked, err := VerifyKoreanProjectionCompatibility(bank)
	if err != nil {
		t.Fatal(err)
	}
	if checked != 1 {
		t.Fatalf("checked %d records, want 1", checked)
	}
}

func TestVerifyKoreanProjectionCompatibilityRetailLiteralAngles(t *testing.T) {
	bank := auditBank(t, 1, []byte{'<', 'X', '>', 0})
	checked, err := VerifyKoreanProjectionCompatibility(bank)
	if err != nil {
		t.Fatal(err)
	}
	if checked != 1 {
		t.Fatalf("checked %d records, want 1", checked)
	}
}

func auditBank(t *testing.T, section int, raw []byte) corpus.Bank {
	t.Helper()
	data := make([]byte, 4+len(raw))
	binary.LittleEndian.PutUint16(data[:2], 1)
	binary.LittleEndian.PutUint16(data[2:4], 4)
	copy(data[4:], raw)
	name := "msgsec001.dat"
	bank, err := corpus.ParseBank(name, data)
	if err != nil {
		t.Fatal(err)
	}
	return bank
}
