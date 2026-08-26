// SPDX-License-Identifier: GPL-3.0-or-later

package message

import (
	"fmt"
	"strings"

	"github.com/HK47196/zill/internal/corpus"
)

// PlaceholderForRecord builds a development-only replacement for one parsed
// retail record. Records with no editable semantic fragment are reported as
// structural (ok=false) so callers can preserve their authenticated retail raw
// bytes instead of treating them as translation failures.
func PlaceholderForRecord(record corpus.Record, marker string) (text string, ok bool, err error) {
	if len(record.Tokens) == 0 {
		return "", false, nil
	}
	meaningful := false
	for _, token := range record.Tokens {
		if token.Kind == "text" || movableSubstitution(token) {
			meaningful = true
			break
		}
	}
	if !meaningful {
		return "", false, nil
	}
	projection, err := Project(record)
	if err != nil {
		return "", false, err
	}
	text, err = projection.PlaceholderAnnotated(marker)
	if err != nil {
		return "", false, err
	}
	return text, true, nil
}

// PlaceholderAnnotated builds a development-only annotated replacement that
// preserves source-owned fixed controls, movable substitutions, and printf
// signatures while replacing human-readable source text with a tiny ASCII
// marker. This is intentionally for device-alpha validation only; it must not be
// used for production localization output.
func (p *Projection) PlaceholderAnnotated(marker string) (string, error) {
	if p == nil {
		return "", fmt.Errorf("placeholder projection is nil")
	}
	if marker == "" {
		return "", fmt.Errorf("placeholder marker must be nonempty")
	}
	if strings.ContainsAny(marker, "<>{}%") {
		return "", fmt.Errorf("placeholder marker contains reserved markup")
	}

	var out strings.Builder
	for _, node := range p.nodes {
		if node.fixed {
			out.WriteString(node.display)
			continue
		}
		fragment := p.Fragments[node.fragment]
		out.WriteString(marker)
		for _, opcode := range fragment.Substitutions {
			out.WriteString(fmt.Sprintf("<value:$%02X>", opcode))
		}
		for _, conversion := range fragment.FormatSignature {
			out.WriteString(conversion)
		}
	}
	text := out.String()
	if _, err := p.SplitSemantic(text); err != nil {
		return "", fmt.Errorf("validate placeholder for message %d: %w", p.RecordID, err)
	}
	return text, nil
}
