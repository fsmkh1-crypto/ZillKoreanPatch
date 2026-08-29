// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/HK47196/zill/internal/forensics/c5scan"
	"github.com/HK47196/zill/internal/forensics/fontscan"
	"github.com/HK47196/zill/internal/forensics/valuescan"
)

// auditC5RuntimeCandidates runs heuristic executable scanners against the
// authenticated retail EBOOT used by the real ISO build. Candidate count zero
// is not an error: all heuristics are intentionally incomplete. A parse/input
// error is fatal because it means the audit did not inspect the expected ELF.
func auditC5RuntimeCandidates(gameDir string) error {
	retailEBOOT, err := loadAuthenticatedRetailEBOOT(gameDir)
	if err != nil {
		return err
	}
	retailDigest := sha256.Sum256(retailEBOOT)
	fmt.Printf("FORENSIC RETAIL_EBOOT_BINDING sha256=%x bytes=%d authenticated=true\n", retailDigest, len(retailEBOOT))

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

	valueCandidates, err := valuescan.Scan(retailEBOOT)
	if err != nil {
		return fmt.Errorf("value substitution runtime candidate scan: %w", err)
	}
	fmt.Printf("FORENSIC VALUE_RUNTIME_SCAN focus_opcode=$15 candidates=%d heuristic_only=true\n", len(valueCandidates))
	valueLimit := len(valueCandidates)
	if valueLimit > summaryLimit {
		valueLimit = summaryLimit
	}
	for i := 0; i < valueLimit; i++ {
		candidate := valueCandidates[i]
		fmt.Printf("FORENSIC VALUE_RUNTIME_CANDIDATE rank=%d file_offset=0x%X vaddr=0x%X score=%d reasons=%q heuristic_only=true\n",
			i+1, candidate.FileOffset, candidate.VirtualAddress, candidate.Score, strings.Join(candidate.Reasons, "; "))
		if i >= 3 {
			continue
		}
		for _, instruction := range candidate.Window {
			delta := int64(instruction.FileOffset) - int64(candidate.FileOffset)
			if delta < -0x20 || delta > 0x20 {
				continue
			}
			fmt.Printf("FORENSIC VALUE_RUNTIME_WINDOW rank=%d delta=%+d file_offset=0x%X vaddr=0x%X word=0x%08X kind=%s text=%q heuristic_only=true\n",
				i+1, delta, instruction.FileOffset, instruction.VirtualAddress, instruction.Word, forensicInstructionKind(instruction.Text), instruction.Text)
		}
	}
	if len(valueCandidates) == 0 {
		fmt.Println("FORENSIC VALUE_RUNTIME_SCAN_NOTE no heuristic candidate found; this does not disprove a shared 0x02 substitution dispatcher")
	}

	fontCandidates, err := fontscan.Scan(retailEBOOT, 12)
	if err != nil {
		return fmt.Errorf("font renderer runtime candidate scan: %w", err)
	}
	fmt.Printf("FORENSIC FONT_RUNTIME_SCAN candidates=%d linked_stride32=true heuristic_only=true\n", len(fontCandidates))
	fontLimit := len(fontCandidates)
	if fontLimit > summaryLimit {
		fontLimit = summaryLimit
	}
	for i := 0; i < fontLimit; i++ {
		candidate := fontCandidates[i]
		fmt.Printf("FORENSIC FONT_RUNTIME_CANDIDATE rank=%d file_offset=0x%X vaddr=0x%X linked_field_loads=%d heuristic_only=true\n",
			i+1, candidate.FileOffset, candidate.VirtualAddress, candidate.FieldLoads)
		if i >= 3 {
			continue
		}
		for _, instruction := range candidate.Window {
			delta := int64(instruction.FileOffset) - int64(candidate.FileOffset)
			if delta < -0x30 || delta > 0x30 {
				continue
			}
			fmt.Printf("FORENSIC FONT_RUNTIME_WINDOW rank=%d delta=%+d file_offset=0x%X vaddr=0x%X word=0x%08X text=%q heuristic_only=true\n",
				i+1, delta, instruction.FileOffset, instruction.VirtualAddress, instruction.Word, instruction.Text)
		}
	}
	if len(fontCandidates) == 0 {
		fmt.Println("FORENSIC FONT_RUNTIME_SCAN_NOTE no linked stride-32 candidate found; this does not disprove a renderer that computes PAF record addresses differently")
	}
	return nil
}

func forensicInstructionKind(text string) string {
	switch {
	case strings.HasPrefix(text, "jal "), strings.HasPrefix(text, "jalr "):
		return "call"
	case strings.HasPrefix(text, "j "), strings.HasPrefix(text, "jr "):
		return "jump"
	case strings.HasPrefix(text, "beq "), strings.HasPrefix(text, "bne "), strings.HasPrefix(text, "blez "), strings.HasPrefix(text, "bgtz "):
		return "branch"
	case strings.HasPrefix(text, "lb "), strings.HasPrefix(text, "lbu "), strings.HasPrefix(text, "lh "), strings.HasPrefix(text, "lhu "), strings.HasPrefix(text, "lw "):
		return "load"
	case strings.HasPrefix(text, "sb "), strings.HasPrefix(text, "sh "), strings.HasPrefix(text, "sw "):
		return "store"
	case strings.HasPrefix(text, "lui "), strings.HasPrefix(text, "ori "), strings.HasPrefix(text, "addiu "):
		return "address-or-immediate"
	default:
		return "other"
	}
}
