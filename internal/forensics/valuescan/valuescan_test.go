// SPDX-License-Identifier: GPL-3.0-or-later

package valuescan

import (
	"debug/elf"
	"encoding/binary"
	"testing"
)

func iType(op, rs, rt uint32, imm uint16) uint32 { return op<<26 | rs<<21 | rt<<16 | uint32(imm) }
func rType(rs, rt, rd, sh, fn uint32) uint32 { return rs<<21 | rt<<16 | rd<<11 | sh<<6 | fn }

func TestScanFindsSyntheticSubstitutionDispatcherShape(t *testing.T) {
	words := []uint32{
		iType(0x24, 4, 2, 0),      // lbu r2,0(r4)
		iType(0x09, 0, 3, 2),      // addiu r3,r0,2
		iType(0x05, 2, 3, 4),      // bne r2,r3,...
		iType(0x09, 4, 4, 1),      // addiu r4,r4,1
		iType(0x24, 4, 5, 0),      // lbu r5,0(r4)
		iType(0x09, 0, 6, 0x15),   // addiu r6,r0,0x15
		iType(0x04, 5, 6, 2),      // beq r5,r6,...
		rType(0, 5, 7, 2, 0),      // sll r7,r5,2
	}
	data := syntheticELF32MIPS(encodeWords(words), uint32(elf.PF_R|elf.PF_X))
	got, err := Scan(data)
	if err != nil { t.Fatal(err) }
	if len(got) != 1 { t.Fatalf("candidates=%d, want 1", len(got)) }
	if got[0].Score < 8 { t.Fatalf("score=%d reasons=%v", got[0].Score, got[0].Reasons) }
}

func TestScanIgnoresNonExecutableSegment(t *testing.T) {
	words := []uint32{iType(0x09,0,3,2), iType(0x24,4,5,0), iType(0x09,0,6,0x15), iType(0x04,5,6,2)}
	data := syntheticELF32MIPS(encodeWords(words), uint32(elf.PF_R))
	got, err := Scan(data)
	if err != nil { t.Fatal(err) }
	if len(got) != 0 { t.Fatalf("candidates=%d, want 0", len(got)) }
}

func encodeWords(words []uint32) []byte {
	data := make([]byte,len(words)*4)
	for i,w := range words { binary.LittleEndian.PutUint32(data[i*4:],w) }
	return data
}

func syntheticELF32MIPS(segment []byte, flags uint32) []byte {
	const eh=52; const phs=32; const off=0x100
	data := make([]byte,off+len(segment))
	copy(data[:4],[]byte{0x7f,'E','L','F'}); data[4]=byte(elf.ELFCLASS32); data[5]=byte(elf.ELFDATA2LSB); data[6]=byte(elf.EV_CURRENT)
	binary.LittleEndian.PutUint16(data[16:],uint16(elf.ET_EXEC)); binary.LittleEndian.PutUint16(data[18:],uint16(elf.EM_MIPS)); binary.LittleEndian.PutUint32(data[20:],uint32(elf.EV_CURRENT))
	binary.LittleEndian.PutUint32(data[28:],eh); binary.LittleEndian.PutUint16(data[40:],eh); binary.LittleEndian.PutUint16(data[42:],phs); binary.LittleEndian.PutUint16(data[44:],1)
	ph:=data[eh:eh+phs]; binary.LittleEndian.PutUint32(ph[0:],uint32(elf.PT_LOAD)); binary.LittleEndian.PutUint32(ph[4:],off); binary.LittleEndian.PutUint32(ph[8:],0x08804000); binary.LittleEndian.PutUint32(ph[12:],0x08804000)
	binary.LittleEndian.PutUint32(ph[16:],uint32(len(segment))); binary.LittleEndian.PutUint32(ph[20:],uint32(len(segment))); binary.LittleEndian.PutUint32(ph[24:],flags); binary.LittleEndian.PutUint32(ph[28:],0x1000)
	copy(data[off:],segment); return data
}
