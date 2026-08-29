// SPDX-License-Identifier: GPL-3.0-or-later

package elfpatch

import (
	"encoding/binary"
	"fmt"
)

const ptLoad = 1

// RuntimeAddressToFileOffset maps a PSP runtime address back into an ELF file
// offset using the executable's PT_LOAD program headers. moduleBase is the
// address at which module virtual address zero was loaded (normally supplied
// by the emulator/debugger, not guessed by this helper).
//
// The returned moduleVA is kept explicit so crash reports can distinguish the
// debugger's runtime address from the ELF's own virtual-address domain.
func RuntimeAddressToFileOffset(elf []byte, moduleBase, runtimeAddress uint32) (moduleVA uint32, fileOffset uint32, err error) {
	if runtimeAddress < moduleBase {
		return 0, 0, fmt.Errorf("runtime address %#x is below module base %#x", runtimeAddress, moduleBase)
	}
	moduleVA = runtimeAddress - moduleBase
	fileOffset, err = ModuleVAToFileOffset(elf, moduleVA)
	if err != nil {
		return moduleVA, 0, err
	}
	return moduleVA, fileOffset, nil
}

// ModuleVAToFileOffset resolves one ELF virtual address through the file-backed
// part of a PT_LOAD segment. Addresses that lie only in BSS/p_memsz are rejected
// because there is no corresponding on-disk instruction/data byte to inspect.
func ModuleVAToFileOffset(elf []byte, moduleVA uint32) (uint32, error) {
	if len(elf) < 0x34 || string(elf[:4]) != "\x7fELF" {
		return 0, fmt.Errorf("not an ELF32 executable")
	}
	if elf[4] != 1 || elf[5] != 1 {
		return 0, fmt.Errorf("unsupported ELF class/data encoding")
	}
	phoff := binary.LittleEndian.Uint32(elf[0x1c:0x20])
	phentsize := binary.LittleEndian.Uint16(elf[0x2a:0x2c])
	phnum := binary.LittleEndian.Uint16(elf[0x2c:0x2e])
	if phentsize < 0x20 {
		return 0, fmt.Errorf("ELF program-header entry is too small: %#x", phentsize)
	}
	for i := uint16(0); i < phnum; i++ {
		entry64 := uint64(phoff) + uint64(i)*uint64(phentsize)
		if entry64+0x20 > uint64(len(elf)) {
			return 0, fmt.Errorf("ELF program-header table extends past end of file")
		}
		entry := int(entry64)
		if binary.LittleEndian.Uint32(elf[entry:entry+4]) != ptLoad {
			continue
		}
		pOffset := binary.LittleEndian.Uint32(elf[entry+4 : entry+8])
		pVaddr := binary.LittleEndian.Uint32(elf[entry+8 : entry+12])
		pFilesz := binary.LittleEndian.Uint32(elf[entry+16 : entry+20])
		if moduleVA < pVaddr {
			continue
		}
		delta := uint64(moduleVA - pVaddr)
		if delta >= uint64(pFilesz) {
			continue
		}
		offset64 := uint64(pOffset) + delta
		if offset64 >= uint64(len(elf)) || offset64 > uint64(^uint32(0)) {
			return 0, fmt.Errorf("mapped ELF file offset is outside executable")
		}
		return uint32(offset64), nil
	}
	return 0, fmt.Errorf("module virtual address %#x is not file-backed by a PT_LOAD segment", moduleVA)
}
