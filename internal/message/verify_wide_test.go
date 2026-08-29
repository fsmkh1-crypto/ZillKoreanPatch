// SPDX-License-Identifier: GPL-3.0-or-later

package message

import (
	"encoding/binary"
	"testing"
)

func TestVerifyWideBank(t *testing.T) {
	data := make([]byte, 18)
	binary.LittleEndian.PutUint16(data[:2], 2)
	binary.LittleEndian.PutUint32(data[4:8], 12)
	binary.LittleEndian.PutUint32(data[8:12], 15)
	copy(data[12:], []byte{1, 2, 0, 3, 4, 0})
	if err := VerifyWideBank("msgsec000.dat", data, 2); err != nil {
		t.Fatalf("valid wide bank rejected: %v", err)
	}
}

func TestVerifyWideBankRejectsRetailLikeTable(t *testing.T) {
	data := make([]byte, 12)
	binary.LittleEndian.PutUint16(data[:2], 2)
	binary.LittleEndian.PutUint16(data[2:4], 6) // retail first uint16 offset, not reserved zero
	binary.LittleEndian.PutUint16(data[4:6], 9)
	if err := VerifyWideBank("msgsec000.dat", data, 2); err == nil {
		t.Fatal("retail-like uint16 offset table was accepted as wide format")
	}
}

func TestVerifyWideBankRejectsBackwardsOffset(t *testing.T) {
	data := make([]byte, 18)
	binary.LittleEndian.PutUint16(data[:2], 2)
	binary.LittleEndian.PutUint32(data[4:8], 15)
	binary.LittleEndian.PutUint32(data[8:12], 12)
	if err := VerifyWideBank("msgsec000.dat", data, 2); err == nil {
		t.Fatal("backwards wide offset was accepted")
	}
}
