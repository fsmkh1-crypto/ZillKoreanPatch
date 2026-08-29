// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"debug/elf"
	"encoding/binary"
	"testing"
)

func TestDirectJALTargetRetainsMIPSPCHighBits(t *testing.T) {
	const pc = uint64(0x08801234)
	const target = uint64(0x08812340)
	word := uint32(0x03<<26) | (uint32(target>>2) & 0x03ffffff)
	got, ok := directJALTarget(word, pc)
	if !ok {
		t.Fatal("directJALTarget rejected jal")
	}
	if got != target {
		t.Fatalf("target=%#x, want %#x", got, target)
	}
	if _, ok := directJALTarget(0, pc); ok {
		t.Fatal("directJALTarget accepted non-jal")
	}
}

func TestVirtualAddressToFileOffsetMapsPTLOADOnly(t *testing.T) {
	data := syntheticELF([]byte{1, 2, 3, 4, 5, 6, 7, 8})
	if got, ok := virtualAddressToFileOffset(data, 0x08804004); !ok || got != 0x104 {
		t.Fatalf("mapped offset=%#x ok=%v, want %#x true", got, ok, uint64(0x104))
	}
	if got, ok := virtualAddressToFileOffset(data, 0x08805000); ok {
		t.Fatalf("unexpected mapping=%#x outside file-backed PT_LOAD", got)
	}
}

func syntheticELF(segment []byte) []byte {
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
	binary.LittleEndian.PutUint32(ph[24:], uint32(elf.PF_R|elf.PF_X))
	binary.LittleEndian.PutUint32(ph[28:], 0x1000)
	copy(data[off:], segment)
	return data
}
