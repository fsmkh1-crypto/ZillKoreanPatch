// SPDX-License-Identifier: GPL-3.0-or-later

package c5scan

import (
	"encoding/binary"
	"testing"
)

func iType(op, rs, rt uint32, imm uint16) uint32 {
	return op<<26 | rs<<21 | rt<<16 | uint32(imm)
}

func TestCapacityImmediateRecognition(t *testing.T) {
	cases := []struct {
		word uint32
		want bool
	}{
		{iType(0x09, 1, 2, 0x0100), true},  // addiu ...,256
		{iType(0x0b, 1, 2, 0x00ff), true},  // sltiu ...,255
		{iType(0x09, 29, 29, 0xff00), true}, // addiu sp,sp,-256
		{iType(0x0d, 1, 2, 0x0100), true},  // ori ...,0x100
		{iType(0x09, 1, 2, 0x0101), false},
	}
	for _, tc := range cases {
		if got := hasCapacityImmediate(tc.word); got != tc.want {
			t.Fatalf("hasCapacityImmediate(%08x)=%v, want %v", tc.word, got, tc.want)
		}
	}
}

func TestScoreWindowRequiresMoreThanCapacityAlone(t *testing.T) {
	words := []uint32{
		iType(0x09, 29, 29, 0xff00), // -256 capacity-like immediate
		0,
		0,
	}
	data := make([]byte, len(words)*4)
	for i, word := range words {
		binary.LittleEndian.PutUint32(data[i*4:], word)
	}
	candidate := scoreWindow(data, 0, uint64(len(data)), 0)
	if candidate.Score >= 5 {
		t.Fatalf("capacity-only window score=%d, want below reporting threshold", candidate.Score)
	}
}

func TestScoreWindowRecognizesContractLikeCluster(t *testing.T) {
	words := []uint32{
		iType(0x09, 29, 29, 0xff00), // addiu sp,sp,-256
		iType(0x09, 1, 1, 3),       // addiu r1,r1,3
		iType(0x09, 2, 2, 9),       // addiu r2,r2,9
		iType(0x24, 4, 3, 0),       // lbu r3,0(r4)
		iType(0x28, 5, 3, 0),       // sb r3,0(r5)
		iType(0x05, 3, 0, 0xfffc),  // bne r3,r0,...
	}
	data := make([]byte, len(words)*4)
	for i, word := range words {
		binary.LittleEndian.PutUint32(data[i*4:], word)
	}
	candidate := scoreWindow(data, 0, uint64(len(data)), 0)
	if candidate.Score < 5 {
		t.Fatalf("contract-like window score=%d, want >=5; reasons=%v", candidate.Score, candidate.Reasons)
	}
	if len(candidate.Window) != len(words) {
		t.Fatalf("window instructions=%d, want %d", len(candidate.Window), len(words))
	}
}
