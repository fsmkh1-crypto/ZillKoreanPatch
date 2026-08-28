// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// logCompiledKoreanForensicsWide reads the translated-bank format used by the
// Korean runtime patch: uint16 record count, reserved uint16 zero, then uint32
// absolute offsets. It is diagnostic-only and never decodes renderer-key bytes
// as retail CP932 text.
func logCompiledKoreanForensicsWide(compiled map[string][]byte) error {
	data, ok := compiled["msgsec001.dat"]
	if !ok {
		return fmt.Errorf("compiled msgsec001.dat is missing")
	}
	if len(data) < 4 {
		return fmt.Errorf("compiled msgsec001.dat is too small")
	}
	count := int(binary.LittleEndian.Uint16(data[:2]))
	reserved := binary.LittleEndian.Uint16(data[2:4])
	if reserved != 0 {
		return fmt.Errorf("compiled msgsec001.dat reserved halfword is %#x, want 0", reserved)
	}
	tableEnd := 4 + count*4
	if count <= 0 || tableEnd > len(data) {
		return fmt.Errorf("compiled msgsec001.dat has invalid wide offset table: count=%d table_end=%#x size=%#x", count, tableEnd, len(data))
	}
	offset := func(index int) (int, error) {
		if index < 0 || index >= count {
			return 0, fmt.Errorf("record index %d out of range 0..%d", index, count-1)
		}
		pos := 4 + index*4
		value := int(binary.LittleEndian.Uint32(data[pos : pos+4]))
		if value < tableEnd || value > len(data) {
			return 0, fmt.Errorf("record index %d has invalid wide offset %#x", index, value)
		}
		return value, nil
	}
	for _, id := range []int{10007, 10010} {
		index := id % 10_000
		start, err := offset(index)
		if err != nil {
			return fmt.Errorf("ID %d: %w", id, err)
		}
		end := len(data)
		if index+1 < count {
			end, err = offset(index + 1)
			if err != nil {
				return fmt.Errorf("ID %d next offset: %w", id, err)
			}
		}
		if end < start {
			return fmt.Errorf("ID %d has descending record span %#x:%#x", id, start, end)
		}
		raw := data[start:end]
		display := raw
		if terminator := bytes.Index(raw, []byte{5, 5, 5}); terminator >= 0 {
			display = raw[:terminator+3]
		}
		fmt.Printf("FORENSIC COMPILED id=%d format=wide32 index=%d start=%#x end=%#x record_span=%d display_size=%d display_hex=%X raw_hex=%X\n",
			id, index, start, end, len(raw), len(display), display, raw)
	}
	return nil
}
