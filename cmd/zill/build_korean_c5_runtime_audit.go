// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"strings"

	"github.com/HK47196/zill/internal/forensics/c5scan"
)

// auditC5RuntimeCandidates runs the heuristic executable scanner against the
// authenticated retail EBOOT used by the real ISO build. Candidate count zero
// is not an error: the heuristic is intentionally incomplete. A parse/input
// error is fatal because it means the audit did not inspect the expected ELF.
func auditC5RuntimeCandidates(gameDir string) error {
	retailEBOOT, err := loadAuthenticatedRetailEBOOT(gameDir)
	if err != nil {
		return err
	}
	candidates, err := c5scan.Scan(retailEBOOT)
	if err != nil {
		return fmt.Errorf("C5 runtime candidate scan: %w", err)
	}
	fmt.Printf("FORENSIC C5_RUNTIME_SCAN candidates=%d heuristic_only=true\n", len(candidates))
	const summaryLimit = 10
	limit := len(candidates)
	if limit > summaryLimit {
		limit = summaryLimit
	}
	for i := 0; i < limit; i++ {
		candidate := candidates[i]
		fmt.Printf("FORENSIC C5_RUNTIME_CANDIDATE rank=%d file_offset=0x%X score=%d reasons=%q heuristic_only=true\n",
			i+1, candidate.FileOffset, candidate.Score, strings.Join(candidate.Reasons, "; "))
	}
	if len(candidates) == 0 {
		fmt.Println("FORENSIC C5_RUNTIME_SCAN_NOTE no heuristic candidate found; this does not disprove the retained C5 runtime contract")
	}
	return nil
}
