// SPDX-License-Identifier: GPL-3.0-or-later

package message

import (
	"fmt"
	"strings"
)

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
