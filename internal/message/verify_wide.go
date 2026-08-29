// SPDX-License-Identifier: GPL-3.0-or-later

package message

import (
	"encoding/binary"
	"fmt"
)

// VerifyWideBank validates the exact on-disc table contract consumed by the
// wide-message-offset runtime patch: uint16 count, zero reserved halfword,
// followed by monotonic uint32 absolute record offsets.
//
// This deliberately does not parse record payloads. Its job is to prove that a
// compiled bank cannot silently fall back to, mix with, or resemble the retail
// uint16-offset table while the executable is running the widened consumer.
func VerifyWideBank(name string, data []byte, expectedCount int) error {
	if expectedCount < 0 || expectedCount > 0xffff {
		return fmt.Errorf("%s: invalid expected message count %d", name, expectedCount)
	}
	if len(data) < 4 {
		return fmt.Errorf("%s: wide bank is too small for header", name)
	}
	count := int(binary.LittleEndian.Uint16(data[:2]))
	if count != expectedCount {
		return fmt.Errorf("%s: wide bank count is %d, want %d", name, count, expectedCount)
	}
	if reserved := binary.LittleEndian.Uint16(data[2:4]); reserved != 0 {
		return fmt.Errorf("%s: wide bank reserved halfword is %#x, want zero", name, reserved)
	}
	tableEnd64 := uint64(4) + uint64(count)*4
	if tableEnd64 > uint64(len(data)) {
		return fmt.Errorf("%s: wide offset table extends past end of file", name)
	}
	tableEnd := uint32(tableEnd64)
	previous := tableEnd
	for index := 0; index < count; index++ {
		start := 4 + index*4
		offset := binary.LittleEndian.Uint32(data[start : start+4])
		if offset < tableEnd || offset < previous || uint64(offset) > uint64(len(data)) {
			return fmt.Errorf("%s: invalid wide offset %#x for message %d", name, offset, index)
		}
		previous = offset
	}
	return nil
}
