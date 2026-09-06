// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/HK47196/zill/internal/koreanslots"
)

// Keep this matcher aligned with corpus.runtimeControlTag. Repository-only
// reflow has annotated source text but no authenticated token stream, so it
// must recognize the same projected runtime controls without reusing the
// narrower layout.controlTag matcher.
var repositoryRuntimeControlTag = regexp.MustCompile(`<(?:if|select|call:[0-9]+|jump:[0-9]+|value:\$[0-9A-F]{2}|add|subtract|multiply|divide|modulo|equal|not-equal|less|greater|less-equal|greater-equal|and|or|operator:\$[0-9A-F]{2}:\$[0-9A-F]{2}|color:[^<>]|discard:[^<>]:\$[0-9A-F]{2}|escape:\$[0-9A-F]{2}|end|separator|backspace|tab|line-break|\$[0-9A-F]{2})>`)

// koreanRepositorySourceAware is the asset-free counterpart of koreanSourceAware.
// corpus.LoadProject intentionally creates display-only records until BindBanks
// authenticates retail data, so repository checks have Japanese annotated text
// but no token projection. Split only on fixed annotated controls, keep movable
// substitutions inside their text fragments, preserve source line breaks as
// SourceLayout hints, and run the same preferred -> greedy scorer.
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
// hint and fixed annotated controls as fragment delimiters. Movable runtime
// substitutions stay inside the surrounding fragment, matching message.Project's
// pureMovable/callerMovable projection semantics instead of becoming artificial
// repository-only boundaries.
//
// The important exception is expression grammar. message.Project tokenizes an
// <if>/<select> expression, including numeric operands such as the 190 in
// <if><value:$01><equal>190, as fixed source bytecode. Those operands are not
// rendered dialogue and must never be measured as text in repository fallback.
func splitRepositoryReflowFragments(text string, source bool) (fragments, controls []string) {
	var current strings.Builder
	flushControl := func(control string) {
		fragments = append(fragments, current.String())
		current.Reset()
		controls = append(controls, control)
	}

	for len(text) > 0 {
		if strings.HasPrefix(text, lineBreak) {
			if source {
				current.WriteByte('\n')
			} else {
				current.WriteString(lineBreak)
			}
			text = text[len(lineBreak):]
			continue
		}

		if end, ok := repositoryExpressionControlEnd(text); ok {
			flushControl(text[:end])
			text = text[end:]
			continue
		}

		if loc := repositoryRuntimeControlTag.FindStringIndex(text); loc != nil && loc[0] == 0 {
			part := text[:loc[1]]
			if koreanDialogueMovableValueTags[strings.ToUpper(part)] {
				current.WriteString(part)
			} else {
				flushControl(part)
			}
			text = text[loc[1]:]
			continue
		}

		end := len(text)
		if loc := repositoryRuntimeControlTag.FindStringIndex(text); loc != nil {
			end = loc[0]
		}
		current.WriteString(text[:end])
		text = text[end:]
	}
	fragments = append(fragments, current.String())
	return fragments, controls
}

// repositoryExpressionControlEnd mirrors corpus.tokenize/expressionEnd over the
// annotated textual representation emitted by displayText. It deliberately
// consumes expression operands together with their owning <if>/<select> or the
// special $1F/$20 value control so repository-only reflow does not mistake
// bytecode literals for visible dialogue.
func repositoryExpressionControlEnd(text string) (int, bool) {
	switch {
	case strings.HasPrefix(text, "<if>"):
		start := len("<if>")
		end := repositoryAnnotatedExpressionEnd(text, start, "boolean")
		return max(start, end), true
	case strings.HasPrefix(text, "<select>"):
		start := len("<select>")
		end := repositoryAnnotatedExpressionEnd(text, start, "arithmetic")
		return max(start, end), true
	case strings.HasPrefix(text, "<value:$1F>"):
		start := len("<value:$1F>")
		end := repositoryAnnotatedExpressionEnd(text, start, "arithmetic")
		return max(start, end), true
	case strings.HasPrefix(text, "<value:$20>"):
		start := len("<value:$20>")
		end := repositoryAnnotatedExpressionEnd(text, start, "arithmetic")
		return max(start, end), true
	default:
		return 0, false
	}
}

func repositoryAnnotatedExpressionEnd(text string, index int, tier string) int {
	var atomEnd func(int) int
	var arithmeticEnd func(int) int

	atomEnd = func(position int) int {
		if position >= len(text) {
			return position
		}
		if loc := valueTag.FindStringIndex(text[position:]); loc != nil && loc[0] == 0 {
			tag := text[position : position+loc[1]]
			position += loc[1]
			upper := strings.ToUpper(tag)
			if upper == "<VALUE:$1F>" || upper == "<VALUE:$20>" {
				return arithmeticEnd(position)
			}
			return position
		}
		if text[position] == '%' {
			position++
		}
		start := position
		for position < len(text) && '0' <= text[position] && text[position] <= '9' {
			position++
		}
		if position > start {
			return position
		}
		return start
	}

	arithmeticEnd = func(position int) int {
		position = atomEnd(position)
		for {
			opEnd := repositoryAnnotatedOperatorEnd(text, position, 3)
			if opEnd == position {
				break
			}
			next := atomEnd(opEnd)
			if next == opEnd {
				break
			}
			position = next
		}
		return position
	}

	comparisonEnd := func(position int) int {
		position = arithmeticEnd(position)
		opEnd := repositoryAnnotatedOperatorEnd(text, position, 4)
		if opEnd != position {
			next := arithmeticEnd(opEnd)
			if next > opEnd {
				position = next
			}
		}
		return position
	}

	position := comparisonEnd(index)
	if tier == "boolean" {
		for {
			opEnd := repositoryAnnotatedOperatorEnd(text, position, 6)
			if opEnd == position {
				break
			}
			next := comparisonEnd(opEnd)
			if next == opEnd {
				break
			}
			position = next
		}
	}
	return position
}

func repositoryAnnotatedOperatorEnd(text string, position, tier int) int {
	if position >= len(text) {
		return position
	}
	known := map[int][]string{
		3: {"<add>", "<subtract>", "<multiply>", "<divide>", "<modulo>"},
		4: {"<equal>", "<not-equal>", "<less>", "<greater>", "<less-equal>", "<greater-equal>"},
		6: {"<and>", "<or>"},
	}
	for _, tag := range known[tier] {
		if strings.HasPrefix(text[position:], tag) {
			return position + len(tag)
		}
	}

	prefix := fmt.Sprintf("<operator:$%02X:$", tier)
	if !strings.HasPrefix(text[position:], prefix) {
		return position
	}
	end := position + len(prefix) + 3 // two hex digits plus '>'
	if end > len(text) || text[end-1] != '>' || !repositoryHex(text[end-3]) || !repositoryHex(text[end-2]) {
		return position
	}
	return end
}

func repositoryHex(value byte) bool {
	return '0' <= value && value <= '9' || 'A' <= value && value <= 'F' || 'a' <= value && value <= 'f'
}
