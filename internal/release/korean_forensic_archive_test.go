// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestRuntimeCompiledRecordReadsUint32OffsetBank(t *testing.T) {
	const count = 4
	tableEnd := 4 + count*4
	records := [][]byte{{0x41, 0}, {0x42, 0}, {0x43, 0}, {0x44, 0x0A, 0x45, 0}}
	size := tableEnd
	for _, record := range records {
		size += len(record)
	}
	data := make([]byte, size)
	binary.LittleEndian.PutUint16(data[:2], count)
	position := tableEnd
	for i, record := range records {
		binary.LittleEndian.PutUint32(data[4+i*4:], uint32(position))
		copy(data[position:], record)
		position += len(record)
	}

	raw, index, err := runtimeCompiledRecord(data, 64, 640003)
	if err != nil {
		t.Fatal(err)
	}
	if index != 3 {
		t.Fatalf("record index=%d, want 3", index)
	}
	if !bytes.Equal(raw, records[3]) {
		t.Fatalf("raw=%v, want %v", raw, records[3])
	}
	if got := bytes.Count(raw, []byte{0x0A}); got != 1 {
		t.Fatalf("materialized_0A=%d, want 1", got)
	}
}

func TestRuntimeCompiledRecordRejectsRetailUint16Layout(t *testing.T) {
	// Retail banks place uint16 offsets at byte 2. Feeding that representation to
	// the runtime parser must fail rather than silently proving the wrong format.
	data := []byte{1, 0, 4, 0, 0x41, 0}
	if _, _, err := runtimeCompiledRecord(data, 64, 640000); err == nil {
		t.Fatal("runtime parser accepted retail uint16-offset bank")
	}
}
