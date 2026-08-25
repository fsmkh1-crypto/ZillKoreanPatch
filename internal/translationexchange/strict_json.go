// SPDX-License-Identifier: GPL-3.0-or-later

package translationexchange

import (
	"encoding/json"
	"fmt"
)

// UnmarshalJSON keeps the external response contract fail-closed. The standard
// decoder's zero values cannot distinguish a missing `uncertain`/`note` field
// from an explicitly supplied false/empty value, so presence is checked here.
func (row *ResultRow) UnmarshalJSON(data []byte) error {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return err
	}
	required := []string{"id", "korean", "uncertain", "note", "glossary_candidates"}
	allowed := make(map[string]struct{}, len(required))
	for _, key := range required {
		allowed[key] = struct{}{}
		if _, exists := object[key]; !exists {
			return fmt.Errorf("missing required field %q", key)
		}
	}
	for key := range object {
		if _, exists := allowed[key]; !exists {
			return fmt.Errorf("unknown field %q", key)
		}
	}
	type plain ResultRow
	var decoded plain
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	if decoded.GlossaryCandidates == nil {
		return fmt.Errorf("glossary_candidates must be an array, not null")
	}
	*row = ResultRow(decoded)
	return nil
}
