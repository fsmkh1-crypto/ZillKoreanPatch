// SPDX-License-Identifier: GPL-3.0-or-later

package valuescan

import (
	"debug/elf"
	"encoding/binary"
	"strings"
	"testing"
)

func iType(op, rs, rt uint32, imm uint16) uint32 { return op<<26 | rs<<21 | rt<<16 | uint32(imm) }
func rType(rs, rt, rd, sh, fn uint32) uint32 { return rs<<21 | rt<<16 | rd<<11 | sh<<6 | fn }
func jType(op, target uint32) uint32 { return op<<26 | ((target >> 2) & 0x03ffffff) }

func TestScanFindsSyntheticSubstitutionDispatcherShape(t *testing.T) {
	words := []uint32{
		iType(0x24, 4, 2, 0),
		iType(0x09, 0, 3, 2),
		iType(0x05, 2, 3, 4),
		iType(0x09, 4, 4, 1),
		iType(0x24, 4, 5, 0),
		iType(0x09, 0, 6, 0x15),
		iType(0x04, 5, 6, 2),
		rType(0, 5, 7, 2, 0),
	}
	data := syntheticELF32MIPS(encodeWords(words), uint32(elf.PF_R|elf.PF_X))
	got, err := Scan(data)
	if err != nil { t.Fatal(err) }
	if len(got) != 1 { t.Fatalf("candidates=%d, want 1", len(got)) }
	candidate := got[0]
	if candidate.Score < 10 { t.Fatalf("score=%d reasons=%v", candidate.Score, candidate.Reasons) }
	if candidate.FileOffset != 0x104 { t.Fatalf("file offset=%#x, want %#x", candidate.FileOffset, 0x104) }
	if candidate.VirtualAddress != 0x08804004 { t.Fatalf("vaddr=%#x, want %#x", candidate.VirtualAddress, uint64(0x08804004)) }
	if len(candidate.Window) != len(words) { t.Fatalf("window=%d instructions, want %d", len(candidate.Window), len(words)) }
	if candidate.Window[0].FileOffset != 0x100 || candidate.Window[0].VirtualAddress != 0x08804000 {
		t.Fatalf("first instruction mapping=%#v", candidate.Window[0])
	}
	joined := ""
	for _, instruction := range candidate.Window { joined += instruction.Text + "\n" }
	if !strings.Contains(joined, "addiu r6,r0,21") || !strings.Contains(joined, "sll r7,r5,2") {
		t.Fatalf("decoded window missing focus evidence:\n%s", joined)
	}
}

func TestDecodeRetainsCallsAndPointerMemoryOps(t *testing.T) {
	tests := []struct {
		word uint32
		want string
	}{
		{jType(0x03, 0x00123450), "jal 0x123450"},
		{jType(0x02, 0x000abc00), "j 0xabc00"},
		{rType(31, 0, 0, 0, 0x08), "jr r31"},
		{rType(25, 0, 31, 0, 0x09), "jalr r31,r25"},
		{iType(0x0f, 0, 14, 0x1234), "lui r14,0x1234"},
		{iType(0x23, 4, 5, 16), "lw r5,16(r4)"},
		{iType(0x2b, 6, 7, 32), "sw r7,32(r6)"},
		{iType(0x21, 8, 9, 2), "lh r9,2(r8)"},
		{iType(0x25, 10, 11, 4), "lhu r11,4(r10)"},
		{iType(0x29, 12, 13, 6), "sh r13,6(r12)"},
	}
	for _, tt := range tests {
		if got := decode(tt.word); got != tt.want {
			t.Fatalf("decode(%#08x)=%q, want %q", tt.word, got, tt.want)
		}
	}
	if got, want := decodeAt(jType(0x03, 0x00123450), 0x81201234), "jal 0x80123450"; got != want {
		t.Fatalf("decodeAt jump target=%q, want %q", got, want)
	}
	if got, want := decodeAt(iType(0x04, 2, 3, 2), 0x08804000), "beq r2,r3,0x880400c"; got != want {
		t.Fatalf("decodeAt branch target=%q, want %q", got, want)
	}
	if got, want := decodeAt(iType(0x05, 2, 3, 0xffff), 0x08804000), "bne r2,r3,0x8804000"; got != want {
		t.Fatalf("decodeAt negative branch target=%q, want %q", got, want)
	}
}

func TestScanRejectsColocatedButRegisterUnlinkedLiterals(t *testing.T) {
	words := []uint32{
		iType(0x09, 0, 3, 2),
		iType(0x24, 4, 5, 0),
		iType(0x05, 7, 8, 4),
		iType(0x09, 0, 6, 0x15),
		iType(0x24, 9, 10, 0),
		iType(0x04, 11, 12, 2),
		rType(0, 10, 13, 2, 0),
	}
	data := syntheticELF32MIPS(encodeWords(words), uint32(elf.PF_R|elf.PF_X))
	got, err := Scan(data)
	if err != nil { t.Fatal(err) }
	if len(got) != 0 { t.Fatalf("candidates=%d, want 0 for register-unlinked colocated evidence: %#v", len(got), got) }
}

func TestScanRequiresLiteralConstructionFromZeroRegister(t *testing.T) {
	words := []uint32{
		iType(0x24, 4, 2, 0),
		iType(0x09, 9, 3, 2),
		iType(0x05, 2, 3, 4),
		iType(0x09, 0, 6, 0x15),
		iType(0x04, 2, 6, 2),
	}
	data := syntheticELF32MIPS(encodeWords(words), uint32(elf.PF_R|elf.PF_X))
	got, err := Scan(data)
	if err != nil { t.Fatal(err) }
	if len(got) != 0 { t.Fatalf("candidates=%d, want 0 when immediate 2 is not constructed from r0", len(got)) }
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
