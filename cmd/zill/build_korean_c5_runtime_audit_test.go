// SPDX-License-Identifier: GPL-3.0-or-later

package main

import "testing"

func TestForensicInstructionKind(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{"jal 0x8801234", "call"},
		{"jalr r31,r25", "call"},
		{"j 0x8801234", "jump"},
		{"jr r31", "jump"},
		{"beq r2,r3,0x8800100", "branch"},
		{"bne r2,r3,0x8800100", "branch"},
		{"lbu r2,0(r4)", "load"},
		{"lw r5,16(r4)", "load"},
		{"sb r2,0(r4)", "store"},
		{"sw r5,16(r4)", "store"},
		{"lui r4,0x880", "address-or-immediate"},
		{"ori r4,r4,0x1234", "address-or-immediate"},
		{"addiu r4,r4,16", "address-or-immediate"},
		{"sll r7,r5,2", "other"},
	}
	for _, tt := range tests {
		if got := forensicInstructionKind(tt.text); got != tt.want {
			t.Fatalf("forensicInstructionKind(%q)=%q, want %q", tt.text, got, tt.want)
		}
	}
}
