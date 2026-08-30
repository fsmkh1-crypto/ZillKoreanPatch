// SPDX-License-Identifier: GPL-3.0-or-later

package c5scan

import (
	"debug/elf"
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
	words := contractLikeWords()
	data := encodeWords(words)
	candidate := scoreWindow(data, 0, uint64(len(data)), 0)
	if candidate.Score < 5 {
		t.Fatalf("contract-like window score=%d, want >=5; reasons=%v", candidate.Score, candidate.Reasons)
	}
	if len(candidate.Window) != len(words) {
		t.Fatalf("window instructions=%d, want %d", len(candidate.Window), len(words))
	}
}

func TestScanSyntheticMIPSExecutableSegment(t *testing.T) {
	segment := encodeWords(contractLikeWords())
	data := syntheticELF32MIPS(segment, uint32(elf.PF_R|elf.PF_X))

	candidates, err := Scan(data)
	if err != nil {
		t.Fatalf("Scan synthetic MIPS ELF: %v", err)
	}
	if len(candidates) != 1 {
		t.Fatalf("candidate count=%d, want 1", len(candidates))
	}
	if candidates[0].FileOffset != 0x100 {
		t.Fatalf("candidate file offset=%#x, want %#x", candidates[0].FileOffset, 0x100)
	}
	if candidates[0].Score < 5 {
		t.Fatalf("candidate score=%d, want >=5; reasons=%v", candidates[0].Score, candidates[0].Reasons)
	}
}

func TestScanIgnoresNonExecutableLoadSegment(t *testing.T) {
	segment := encodeWords(contractLikeWords())
	data := syntheticELF32MIPS(segment, uint32(elf.PF_R))

	candidates, err := Scan(data)
	if err != nil {
		t.Fatalf("Scan synthetic non-executable MIPS ELF: %v", err)
	}
	if len(candidates) != 0 {
		t.Fatalf("candidate count=%d, want 0 for non-executable PT_LOAD", len(candidates))
	}
}

func contractLikeWords() []uint32 {
	return []uint32{
		iType(0x09, 29, 29, 0xff00), // addiu sp,sp,-256
		iType(0x09, 1, 1, 3),       // addiu r1,r1,3
		iType(0x09, 2, 2, 9),       // addiu r2,r2,9
		iType(0x24, 4, 3, 0),       // lbu r3,0(r4)
		iType(0x28, 5, 3, 0),       // sb r3,0(r5)
		iType(0x05, 3, 0, 0xfffc),  // bne r3,r0,...
	}
}

func encodeWords(words []uint32) []byte {
	data := make([]byte, len(words)*4)
	for i, word := range words {
		binary.LittleEndian.PutUint32(data[i*4:], word)
	}
	return data
}

func syntheticELF32MIPS(segment []byte, flags uint32) []byte {
	const (
		elfHeaderSize     = 52
		programHeaderSize = 32
		segmentOffset     = 0x100
	)
	data := make([]byte, segmentOffset+len(segment))
	copy(data[:4], []byte{0x7f, 'E', 'L', 'F'})
	data[4] = byte(elf.ELFCLASS32)
	data[5] = byte(elf.ELFDATA2LSB)
	data[6] = byte(elf.EV_CURRENT)
	binary.LittleEndian.PutUint16(data[16:], uint16(elf.ET_EXEC))
	binary.LittleEndian.PutUint16(data[18:], uint16(elf.EM_MIPS))
	binary.LittleEndian.PutUint32(data[20:], uint32(elf.EV_CURRENT))
	binary.LittleEndian.PutUint32(data[28:], elfHeaderSize)
	binary.LittleEndian.PutUint16(data[40:], elfHeaderSize)
	binary.LittleEndian.PutUint16(data[42:], programHeaderSize)
	binary.LittleEndian.PutUint16(data[44:], 1)

	ph := data[elfHeaderSize : elfHeaderSize+programHeaderSize]
	binary.LittleEndian.PutUint32(ph[0:], uint32(elf.PT_LOAD))
	binary.LittleEndian.PutUint32(ph[4:], segmentOffset)
	binary.LittleEndian.PutUint32(ph[8:], 0x08804000)
	binary.LittleEndian.PutUint32(ph[12:], 0x08804000)
	binary.LittleEndian.PutUint32(ph[16:], uint32(len(segment)))
	binary.LittleEndian.PutUint32(ph[20:], uint32(len(segment)))
	binary.LittleEndian.PutUint32(ph[24:], flags)
	binary.LittleEndian.PutUint32(ph[28:], 0x1000)
	copy(data[segmentOffset:], segment)
	return data
}
