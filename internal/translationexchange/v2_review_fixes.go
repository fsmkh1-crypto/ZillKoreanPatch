// SPDX-License-Identifier: GPL-3.0-or-later

package translationexchange

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// The game message compiler only treats %s / %u (optionally positional) as
// runtime format conversions. The earlier generic printf pattern also matched
// ordinary prose such as "100% Match" by consuming "% M". Keep the exchange
// contract aligned with message.printfConversion instead of generic printf.
func init() {
	printfRE = regexp.MustCompile(`%(?:[1-9][0-9]*)?[su]`)
	protectedV2RE = regexp.MustCompile(`(<[^<>\r\n]+>|\{\{[^{}\r\n]+\}\}|%(?:[1-9][0-9]*)?[su])`)
}

// UnmarshalJSON keeps the externally returned v2 source payload fail-closed.
// The current exporter intentionally emits an empty glossary because a
// repository-owned Korean glossary has not been wired yet. Accepting arbitrary
// glossary data from the external source file would make the stale-source check
// compare attacker-controlled glossary data against itself.
func (row *ExportRowV2) UnmarshalJSON(data []byte) error {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return err
	}
	required := []string{
		"schema", "id", "section", "record_index", "source_file",
		"full_text", "english_reference", "glossary", "segments",
	}
	if err := requireExactFields(object, required); err != nil {
		return err
	}
	type plain ExportRowV2
	var decoded plain
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	if decoded.Glossary == nil {
		return fmt.Errorf("glossary must be an object, not null")
	}
	if len(decoded.Glossary) != 0 {
		return fmt.Errorf("glossary must match the repository-owned canonical glossary; current v2 export requires an empty object")
	}
	if decoded.Segments == nil {
		return fmt.Errorf("segments must be an array, not null")
	}
	*row = ExportRowV2(decoded)
	return nil
}
