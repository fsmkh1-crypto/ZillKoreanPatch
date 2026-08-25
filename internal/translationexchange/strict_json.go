// SPDX-License-Identifier: GPL-3.0-or-later

package translationexchange

import (
	"encoding/json"
	"fmt"
)

// UnmarshalJSON keeps the external v1 response contract fail-closed. The
// standard decoder's zero values cannot distinguish a missing field from an
// explicitly supplied false/empty value, so presence is checked here.
func (row *ResultRow) UnmarshalJSON(data []byte) error {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return err
	}
	required := []string{"id", "korean", "uncertain", "note", "glossary_candidates"}
	if err := requireExactFields(object, required); err != nil {
		return err
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

// UnmarshalJSON keeps the v2 segment response contract exact. In particular,
// korean_segments and glossary_candidates must be arrays and every field must
// be explicitly present.
func (row *ResultRowV2) UnmarshalJSON(data []byte) error {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return err
	}
	required := []string{"id", "korean_segments", "uncertain", "note", "glossary_candidates"}
	if err := requireExactFields(object, required); err != nil {
		return err
	}
	type plain ResultRowV2
	var decoded plain
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	if decoded.KoreanSegments == nil {
		return fmt.Errorf("korean_segments must be an array, not null")
	}
	if decoded.GlossaryCandidates == nil {
		return fmt.Errorf("glossary_candidates must be an array, not null")
	}
	*row = ResultRowV2(decoded)
	return nil
}

func requireExactFields(object map[string]json.RawMessage, required []string) error {
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
	return nil
}
