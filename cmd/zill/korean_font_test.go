// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRetailBindataDoesNotRequirePamiWhenFoundInPa(t *testing.T) {
	gameDir := t.TempDir()
	usrdir := filepath.Join(gameDir, "USRDIR")
	if err := os.MkdirAll(usrdir, 0o755); err != nil {
		t.Fatal(err)
	}
	payload := []byte("authenticated bindata fixture")
	writeSingleMemberPAA(t, usrdir, "pa", "data/bindata.dat", payload)
	expected := fmt.Sprintf("%x", sha256.Sum256(payload))

	got, err := loadRetailBindataWithSHA(gameDir, expected)
	if err != nil {
		t.Fatalf("bindata in pa should not require pami: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("payload = %q, want %q", got, payload)
	}
	if _, err := os.Stat(filepath.Join(usrdir, "pami.bin")); !os.IsNotExist(err) {
		t.Fatalf("test unexpectedly has pami.bin: %v", err)
	}
}

func TestLoadRetailBindataRejectsWrongFingerprint(t *testing.T) {
	gameDir := t.TempDir()
	usrdir := filepath.Join(gameDir, "USRDIR")
	if err := os.MkdirAll(usrdir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeSingleMemberPAA(t, usrdir, "pa", "data/bindata.dat", []byte("wrong source"))

	_, err := loadRetailBindataWithSHA(gameDir, strings.Repeat("0", 64))
	if err == nil || !strings.Contains(err.Error(), "unsupported retail data/bindata.dat fingerprint") {
		t.Fatalf("wrong fingerprint error = %v", err)
	}
}

func writeSingleMemberPAA(t *testing.T, directory, archive, name string, payload []byte) {
	t.Helper()
	const (
		nameOffset  = 0x40
		offsetTable = 0x70
	)
	index := make([]byte, 0x80)
	copy(index[:4], []byte{'P', 'A', 'A', 0})
	binary.LittleEndian.PutUint32(index[8:12], 1)
	binary.LittleEndian.PutUint32(index[16:20], offsetTable)
	binary.LittleEndian.PutUint32(index[0x20:0x24], nameOffset)
	binary.LittleEndian.PutUint32(index[0x24:0x28], uint32(len(payload)))
	binary.LittleEndian.PutUint32(index[0x28:0x2c], 0xffffffff)
	binary.LittleEndian.PutUint32(index[0x2c:0x30], 0xffffffff)
	copy(index[nameOffset:], []byte(name))
	index[nameOffset+len(name)] = 0
	binary.LittleEndian.PutUint32(index[offsetTable:offsetTable+4], 0x10)

	archiveSize := (0x10 + len(payload) + 0x0f) &^ 0x0f
	arc := make([]byte, archiveSize)
	copy(arc[0x10:], payload)
	if err := os.WriteFile(filepath.Join(directory, archive+".bin"), index, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, archive+".arc"), arc, 0o644); err != nil {
		t.Fatal(err)
	}
}
