// SPDX-License-Identifier: GPL-3.0-or-later

package elfpatch

import (
	"encoding/binary"
	"testing"
)

func syntheticELF() []byte {
	data := make([]byte, 0x400)
	copy(data[:4], []byte("\x7fELF"))
	data[4] = 1 // ELFCLASS32
	data[5] = 1 // little endian
	binary.LittleEndian.PutUint32(data[0x1c:0x20], 0x34)
	binary.LittleEndian.PutUint16(data[0x2a:0x2c], 0x20)
	binary.LittleEndian.PutUint16(data[0x2c:0x2e], 1)
	ph := 0x34
	binary.LittleEndian.PutUint32(data[ph:ph+4], ptLoad)
	binary.LittleEndian.PutUint32(data[ph+4:ph+8], 0x80)
	binary.LittleEndian.PutUint32(data[ph+8:ph+12], 0)
	binary.LittleEndian.PutUint32(data[ph+16:ph+20], 0x200)
	binary.LittleEndian.PutUint32(data[ph+20:ph+24], 0x300)
	return data
}

func TestRuntimeAddressToFileOffset(t *testing.T) {
	elf := syntheticELF()
	moduleVA, offset, err := RuntimeAddressToFileOffset(elf, 0x08804000, 0x08804120)
	if err != nil {
		t.Fatal(err)
	}
	if moduleVA != 0x120 || offset != 0x1a0 {
		t.Fatalf("got moduleVA=%#x offset=%#x; want %#x %#x", moduleVA, offset, 0x120, 0x1a0)
	}
}

func TestModuleVAToFileOffsetRejectsBSSOnlyAddress(t *testing.T) {
	elf := syntheticELF()
	if _, err := ModuleVAToFileOffset(elf, 0x280); err == nil {
		t.Fatal("expected BSS-only virtual address to be rejected")
	}
}

func TestRuntimeAddressToFileOffsetRejectsAddressBelowBase(t *testing.T) {
	elf := syntheticELF()
	if _, _, err := RuntimeAddressToFileOffset(elf, 0x08804000, 0x08803ffc); err == nil {
		t.Fatal("expected address below module base to fail")
	}
}
