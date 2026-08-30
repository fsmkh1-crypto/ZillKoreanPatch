// SPDX-License-Identifier: GPL-3.0-or-later

package message

import "testing"

func TestAnalyzeRetailStringScannerMatchesCapturedMaxSpanSemantics(t *testing.T) {
	// Ordinary encoded bytes count toward the current span; CR/LF terminate a
	// line; CRLF is one boundary; a three-byte ESC control contributes no width.
	raw := []byte{
		'A', 'B', 'C', '\n',
		'D', 'E', 'F', 'G', '\r', '\n',
		'H', 0x1B, 0x44, 0x55, 'I', 'J', 'K', 'L', 'M', 0,
	}
	got, err := AnalyzeRetailStringScanner(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Terminated {
		t.Fatal("scanner did not report NUL termination")
	}
	if got.LineBreaks != 2 {
		t.Fatalf("line breaks = %d, want 2", got.LineBreaks)
	}
	if got.MaxSpan != 6 { // H + IJKLM; ESC control is skipped.
		t.Fatalf("max span = %d, want 6", got.MaxSpan)
	}
}

func TestAnalyzeRetailStringScannerRejectsUnterminatedOrTruncatedControl(t *testing.T) {
	if _, err := AnalyzeRetailStringScanner([]byte("abc")); err == nil {
		t.Fatal("expected unterminated record error")
	}
	if _, err := AnalyzeRetailStringScanner([]byte{'a', 0x1B, 1, 0}); err == nil {
		t.Fatal("expected truncated ESC control error")
	}
}
