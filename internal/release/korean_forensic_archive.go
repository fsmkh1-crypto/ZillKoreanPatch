// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"path/filepath"

	"github.com/HK47196/zill/internal/gamefmt/paa"
)

const koreanDialogueForensicRecordID = 640003

// runtimeCompiledRecord parses one record from the relocated runtime-bank
// representation emitted by assembleCompiledBank: uint16 count, two reserved
// bytes, then absolute uint32 offsets. This intentionally does not use
// corpus.ParseBank, which parses the retail uint16-offset on-disc source format.
func runtimeCompiledRecord(data []byte, section, id int) ([]byte, int, error) {
	if len(data) < 4 {
		return nil, 0, fmt.Errorf("compiled bank is too small")
	}
	count := int(binary.LittleEndian.Uint16(data[:2]))
	tableEnd := 4 + count*4
	if tableEnd > len(data) {
		return nil, 0, fmt.Errorf("compiled uint32 offset table extends past end of bank")
	}
	index := id - section*10_000
	if index < 0 || index >= count {
		return nil, 0, fmt.Errorf("dialogue ID %d is outside section %d record range (count=%d)", id, section, count)
	}
	offsets := make([]int, count)
	previous := tableEnd
	for i := 0; i < count; i++ {
		offset := int(binary.LittleEndian.Uint32(data[4+i*4:]))
		if offset < tableEnd || offset < previous || offset > len(data) {
			return nil, 0, fmt.Errorf("invalid runtime uint32 offset %#x for message %d", offset, i)
		}
		offsets[i] = offset
		previous = offset
	}
	start := offsets[index]
	end := len(data)
	if index+1 < count {
		end = offsets[index+1]
	}
	return bytes.Clone(data[start:end]), index, nil
}

// auditKoreanDialogueForensicArchive reopens the staged rebuilt PAA archives,
// locates the exact bank that owns the watched dialogue record, parses the
// relocated runtime bank again from rebuilt bytes, and proves the materialized
// 0x0A survived archive replacement. This is diagnostic-only and does not
// modify any release data.
func auditKoreanDialogueForensicArchive(staging string) error {
	const (
		memberName = "message/msgsec064.dat"
		section    = 64
	)
	usrdir := filepath.Join(staging, "USRDIR")
	for _, archiveName := range []string{"pa", "pami"} {
		pair, err := paa.Open(filepath.Join(usrdir, archiveName+".bin"), filepath.Join(usrdir, archiveName+".arc"))
		if err != nil {
			return fmt.Errorf("forensic reopen staged %s archive: %w", archiveName, err)
		}
		for _, member := range pair.Members() {
			if member.Name != memberName {
				continue
			}
			payload, err := pair.Payload(member.Index)
			if err != nil {
				_ = pair.Close()
				return fmt.Errorf("forensic read %s/%s: %w", archiveName, memberName, err)
			}
			raw, recordIndex, err := runtimeCompiledRecord(payload, section, koreanDialogueForensicRecordID)
			if err != nil {
				_ = pair.Close()
				return fmt.Errorf("forensic parse rebuilt runtime %s/%s: %w", archiveName, memberName, err)
			}
			count := bytes.Count(raw, []byte{0x0A})
			digest := sha256.Sum256(raw)
			fmt.Printf("FORENSIC_STAGED_ARCHIVE_DIALOGUE id=%d archive=%s member=%s member_index=%d record_index=%d raw_bytes=%d materialized_0A=%d raw_sha256=%x runtime_uint32_offsets=true exact_reparse=true\n",
				koreanDialogueForensicRecordID, archiveName, memberName, member.Index, recordIndex, len(raw), count, digest)
			_ = pair.Close()
			if count != 1 {
				return fmt.Errorf("forensic rebuilt dialogue %d has materialized_0A=%d, want 1", koreanDialogueForensicRecordID, count)
			}
			return nil
		}
		if err := pair.Close(); err != nil {
			return fmt.Errorf("forensic close staged %s archive: %w", archiveName, err)
		}
	}
	return fmt.Errorf("forensic staged archives lack %s", memberName)
}
