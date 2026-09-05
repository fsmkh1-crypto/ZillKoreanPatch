// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"fmt"
	"strings"

	"github.com/HK47196/zill/internal/koreanslots"
)

// koreanRepositorySourceAware is the asset-free counterpart of koreanSourceAware.
// corpus.LoadProject intentionally creates display-only records until BindBanks
// authenticates retail data, so repository checks have Japanese annotated text
// but no token projection. Split only on fixed annotated controls, keep source
// line breaks as SourceLayout hints, and run the same preferred -> greedy scorer.
// Production Korean builds bind retail banks before this derivation and therefore
// use koreanSourceAware with the authenticated message.Projection instead.
func (e *Engine) koreanRepositorySourceAware(semantic, source string, limit, id int, mapping koreanslots.Mapping) (string, error) {
	semanticFragments, semanticControls := splitRepositoryReflowFragments(semantic, false)
	sourceFragments, sourceControls := splitRepositoryReflowFragments(source, true)
	if len(semanticFragments) != len(sourceFragments) || len(semanticControls) != len(sourceControls) {
		return "", fmt.Errorf("message %d repository Korean dialogue control projection differs from Japanese source", id)
	}
	for i := range semanticControls {
		if semanticControls[i] != sourceControls[i] {
			return "", fmt.Errorf("message %d repository Korean dialogue changes fixed control %q to %q", id, sourceControls[i], semanticControls[i])
		}
	}

	c5 := e.has(e.consumers.C5IDs, id) || e.has(e.consumers.SinglePageC5IDs, id)
	var result strings.Builder
	for i, fragment := range semanticFragments {
		flow, err := e.koreanPreferred(fragment, sourceFragments[i], limit, id, c5, mapping)
		if err != nil {
			return "", err
		}
		if flow == "" && fragment != "" {
			return "", nil
		}
		result.WriteString(flow)
		if i < len(semanticControls) {
			result.WriteString(semanticControls[i])
		}
	}
	return result.String(), nil
}

// splitRepositoryReflowFragments treats source <line-break> as a movable layout
// hint and every other known annotated control as a fixed fragment delimiter.
// This is intentionally narrower than retail token projection: it exists only
// for asset-free checks where the token-level fixedLineBreaks distinction cannot
// be authenticated.
func splitRepositoryReflowFragments(text string, source bool) (fragments, controls []string) {
	var current strings.Builder
	for _, part := range splitParts(text) {
		if part == lineBreak {
			if source {
				current.WriteByte('\n')
			} else {
				current.WriteString(part)
			}
			continue
		}
		if loc := controlTag.FindStringIndex(part); loc != nil && loc[0] == 0 && loc[1] == len(part) {
			fragments = append(fragments, current.String())
			current.Reset()
			controls = append(controls, part)
			continue
		}
		current.WriteString(part)
	}
	fragments = append(fragments, current.String())
	return fragments, controls
}
