// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"path/filepath"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/gamefmt/paa"
)

const koreanDialogueForensicRecordID = 640003

// auditKoreanDialogueForensicArchive reopens the staged rebuilt PAA archives,
// locates the exact bank that owns the watched dialogue record, parses that bank
// again from rebuilt bytes, and proves the materialized 0x0A survived archive
// replacement. This is diagnostic-only and does not modify any release data.
func auditKoreanDialogueForensicArchive(staging string) error {
	const memberName = "message/msgsec064.dat"
	usrdir := filepath.Join(staging, "USRDIR")
	for _, archiveName := range []string{"pa", "pami"} {
		pair, err := paa.Open(filepath.Join(usrdir, archiveName+".bin"), filepath.Join(usrdir, archiveName+".arc"))
		if err != nil {
			return fmt.Errorf("forensic reopen staged %s archive: %w", archiveName, err)
		}
		foundMember := false
		for _, member := range pair.Members() {
			if member.Name != memberName {
				continue
			}
			foundMember = true
			payload, err := pair.Payload(member.Index)
			if err != nil {
				_ = pair.Close()
				return fmt.Errorf("forensic read %s/%s: %w", archiveName, memberName, err)
			}
			bank, err := corpus.ParseBank(filepath.Base(memberName), payload)
			if err != nil {
				_ = pair.Close()
				return fmt.Errorf("forensic parse rebuilt %s/%s: %w", archiveName, memberName, err)
			}
			for _, record := range bank.Records {
				if record.ID != koreanDialogueForensicRecordID {
					continue
				}
				count := bytes.Count(record.Raw, []byte{0x0A})
				digest := sha256.Sum256(record.Raw)
				fmt.Printf("FORENSIC_STAGED_ARCHIVE_DIALOGUE id=%d archive=%s member=%s member_index=%d raw_bytes=%d materialized_0A=%d raw_sha256=%x exact_reparse=true\n",
					record.ID, archiveName, memberName, member.Index, len(record.Raw), count, digest)
				_ = pair.Close()
				if count != 1 {
					return fmt.Errorf("forensic rebuilt dialogue %d has materialized_0A=%d, want 1", record.ID, count)
				}
				return nil
			}
			_ = pair.Close()
			return fmt.Errorf("forensic rebuilt %s/%s lacks dialogue ID %d", archiveName, memberName, koreanDialogueForensicRecordID)
		}
		if err := pair.Close(); err != nil {
			return fmt.Errorf("forensic close staged %s archive: %w", archiveName, err)
		}
		if foundMember {
			break
		}
	}
	return fmt.Errorf("forensic staged archives lack %s", memberName)
}
