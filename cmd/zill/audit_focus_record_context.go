// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/HK47196/zill/internal/cdccontext"
)

const focusRecordID = 10010

// auditFocusRecordContext reuses the authenticated-retail context command and
// prints only the recovered entries for the freeze-adjacent focus record. This
// is forensic output: failure to recover context does not itself prove or
// refute a runtime storage bug.
func auditFocusRecordContext(root, gameDir string, stdout io.Writer) error {
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := runContext(root, []string{
		"--game-dir", gameDir,
		"--record", fmt.Sprint(focusRecordID),
		"--format", "json",
		"--verbose",
	}, &out, &errOut)
	if code != 0 {
		return fmt.Errorf("context command exited %d: %s", code, bytes.TrimSpace(errOut.Bytes()))
	}

	var result cdccontext.Result
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return fmt.Errorf("decode context JSON: %w", err)
	}

	matches := 0
	for _, scene := range result.Scenes {
		for _, entry := range scene.Entries {
			if entry.MessageID != focusRecordID {
				continue
			}
			matches++
			evidence, _ := json.Marshal(entry.ConsumerEvidence)
			controls, _ := json.Marshal(entry.SourceControls)
			addressees, _ := json.Marshal(entry.PossibleAddressees)
			fmt.Fprintf(stdout, "FORENSIC C5_FOCUS record=%d scene=%q member=%q source_archive=%q offset=%d reachability=%q raw=%q display_mode=%v association_handle=%v portrait=%v name_label=%v\n",
				focusRecordID, scene.ID, scene.Member, scene.SourceArchive, entry.Offset, entry.Reachability,
				entry.Raw, entry.DisplayMode, entry.EntityAssociationHandleRaw, entry.PortraitRequested, entry.NameLabelRequested)
			fmt.Fprintf(stdout, "FORENSIC C5_FOCUS consumer_evidence=%s\n", evidence)
			fmt.Fprintf(stdout, "FORENSIC C5_FOCUS source_controls=%s\n", controls)
			fmt.Fprintf(stdout, "FORENSIC C5_FOCUS possible_addressees=%s\n", addressees)
		}
	}
	if matches == 0 {
		return fmt.Errorf("authenticated retail context recovered no entries for record %d", focusRecordID)
	}
	fmt.Fprintf(stdout, "FORENSIC C5_FOCUS summary record=%d recovered_entries=%d\n", focusRecordID, matches)
	return nil
}
