// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"debug/elf"
	"encoding/binary"
	"testing"
)

func frI(op, rs, rt uint32, imm uint16) uint32 { return op<<26 | rs<<21 | rt<<16 | uint32(imm) }
func frR(rs, rt, rd, sh, fn uint32) uint32 { return rs<<21 | rt<<16 | rd<<11 | sh<<6 | fn }

func TestRendererScanFindsLinkedStrideRecordLoads(t *testing.T) {
	words := []uint32{
		frR(0, 5, 6, 5, 0),       // sll r6,r5,5
		frR(4, 6, 7, 0, 0x21),    // addu r7,r4,r6
		frI(0x25, 7, 8, 0),        // lhu r8,0(r7)
		frI(0x24, 7, 9, 14),       // lbu r9,14(r7)
	}
	got, err := scanRendererCandidates(frSyntheticELF(frEncode(words), uint32(elf.PF_R|elf.PF_X)), 8)
	if err != nil { t.Fatal(err) }
	if len(got) != 1 { t.Fatalf("candidates=%d, want 1: %#v", len(got), got) }
	if got[0].FieldLoads != 2 { t.Fatalf("field loads=%d, want 2", got[0].FieldLoads) }
	if got[0].FileOffset != 0x100 || got[0].Vaddr != 0x08804000 {
		t.Fatalf("mapping offset=%#x vaddr=%#x", got[0].FileOffset, got[0].Vaddr)
	}
}

func TestRendererScanRejectsUnlinkedNearbyLoads(t *testing.T) {
	words := []uint32{
		frR(0, 5, 6, 5, 0),       // stride exists
		frR(4, 10, 7, 0, 0x21),   // address add does not use stride result r6
		frI(0x25, 7, 8, 0),
	}
	got, err := scanRendererCandidates(frSyntheticELF(frEncode(words), uint32(elf.PF_R|elf.PF_X)), 8)
	if err != nil { t.Fatal(err) }
	if len(got) != 0 { t.Fatalf("candidates=%d, want 0 for unlinked load proximity: %#v", len(got), got) }
}

func TestRendererScanRejectsLoadsOutsideRecordExtent(t *testing.T) {
	words := []uint32{
		frR(0, 5, 6, 5, 0),
		frR(4, 6, 7, 0, 0x21),
		frI(0x25, 7, 8, 0x20), // exactly outside 0x20-byte record
	}
	got, err := scanRendererCandidates(frSyntheticELF(frEncode(words), uint32(elf.PF_R|elf.PF_X)), 8)
	if err != nil { t.Fatal(err) }
	if len(got) != 0 { t.Fatalf("candidates=%d, want 0 for out-of-record load", len(got)) }
}

func TestRendererScanIgnoresNonExecutableSegment(t *testing.T) {
	words := []uint32{
		frR(0, 5, 6, 5, 0),
		frR(4, 6, 7, 0, 0x21),
		frI(0x25, 7, 8, 0),
	}
	got, err := scanRendererCandidates(frSyntheticELF(frEncode(words), uint32(elf.PF_R)), 8)
	if err != nil { t.Fatal(err) }
	if len(got) != 0 { t.Fatalf("candidates=%d, want 0 in non-executable PT_LOAD", len(got)) }
}

func frEncode(words []uint32) []byte {
	data := make([]byte, len(words)*4)
	for i, w := range words { binary.LittleEndian.PutUint32(data[i*4:], w) }
	return data
}

func frSyntheticELF(segment []byte, flags uint32) []byte {
	const eh = 52
	const phs = 32
	const off = 0x100
	data := make([]byte, off+len(segment))
	copy(data[:4], []byte{0x7f, 'E', 'L', 'F'})
	data[4] = byte(elf.ELFCLASS32)
	data[5] = byte(elf.ELFDATA2LSB)
	data[6] = byte(elf.EV_CURRENT)
	binary.LittleEndian.PutUint16(data[16:], uint16(elf.ET_EXEC))
	binary.LittleEndian.PutUint16(data[18:], uint16(elf.EM_MIPS))
	binary.LittleEndian.PutUint32(data[20:], uint32(elf.EV_CURRENT))
	binary.LittleEndian.PutUint32(data[28:], eh)
	binary.LittleEndian.PutUint16(data[40:], eh)
	binary.LittleEndian.PutUint16(data[42:], phs)
	binary.LittleEndian.PutUint16(data[44:], 1)
	ph := data[eh : eh+phs]
	binary.LittleEndian.PutUint32(ph[0:], uint32(elf.PT_LOAD))
	binary.LittleEndian.PutUint32(ph[4:], off)
	binary.LittleEndian.PutUint32(ph[8:], 0x08804000)
	binary.LittleEndian.PutUint32(ph[12:], 0x08804000)
	binary.LittleEndian.PutUint32(ph[16:], uint32(len(segment)))
	binary.LittleEndian.PutUint32(ph[20:], uint32(len(segment)))
	binary.LittleEndian.PutUint32(ph[24:], flags)
	binary.LittleEndian.PutUint32(ph[28:], 0x1000)
	copy(data[off:], segment)
	return data
}
