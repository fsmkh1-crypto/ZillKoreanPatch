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
	for _, target := range []struct {
		bank string
		id   int
	}{
		{"msgsec001.dat", 10007},
		{"msgsec001.dat", 10010},
		{"msgsec021.dat", 210065},
	} {
		if err := logWideRecord(compiled, target.bank, target.id); err != nil {
			return err
		}
	}
	return nil
}

func logWideRecord(compiled map[string][]byte, bankName string, id int) error {
	data, ok := compiled[bankName]
	if !ok {
		return fmt.Errorf("compiled %s is missing", bankName)
	}
	if len(data) < 4 {
		return fmt.Errorf("compiled %s is too small", bankName)
	}
	count := int(binary.LittleEndian.Uint16(data[:2]))
	reserved := binary.LittleEndian.Uint16(data[2:4])
	if reserved != 0 {
		return fmt.Errorf("compiled %s reserved halfword is %#x, want 0", bankName, reserved)
	}
	tableEnd := 4 + count*4
	if count <= 0 || tableEnd > len(data) {
		return fmt.Errorf("compiled %s has invalid wide offset table: count=%d table_end=%#x size=%#x", bankName, count, tableEnd, len(data))
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
	index := id % 10_000
	start, err := offset(index)
	if err != nil {
		return fmt.Errorf("%s ID %d: %w", bankName, id, err)
	}
	end := len(data)
	if index+1 < count {
		end, err = offset(index + 1)
		if err != nil {
			return fmt.Errorf("%s ID %d next offset: %w", bankName, id, err)
		}
	}
	if end < start {
		return fmt.Errorf("%s ID %d has descending record span %#x:%#x", bankName, id, start, end)
	}
	raw := data[start:end]
	display := raw
	if terminator := bytes.Index(raw, []byte{5, 5, 5}); terminator >= 0 {
		display = raw[:terminator+3]
	}
	fmt.Printf("FORENSIC COMPILED bank=%s id=%d format=wide32 bank_size=%d bank_mod16=%d index=%d start=%#x end=%#x record_span=%d display_size=%d display_hex=%X raw_hex=%X\n",
		bankName, id, len(data), len(data)%16, index, start, end, len(raw), len(display), display, raw)
	return nil
}
