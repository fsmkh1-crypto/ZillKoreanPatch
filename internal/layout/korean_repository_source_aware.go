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
// but no token projection. The Japanese source establishes fixed-control
// boundaries; Korean is projected against those exact boundaries instead of
// reparsing translated text independently. This matters when a fixed numeric
// expression is immediately followed by visible ASCII digits in Korean, e.g.
// source "%4８世紀" -> Korean "%48세기": retail bytecode knows the operand is
// %4, and repository fallback must not misread it as %48.
//
// Movable substitutions remain inside semantic fragments, source line breaks are
// retained only as SourceLayout hints, and the fragments run through the same
// preferred -> greedy scorer used by the authenticated retail path.
func (e *Engine) koreanRepositorySourceAware(semantic, source string, limit, id int, mapping koreanslots.Mapping) (string, error) {
	sourceFragments, sourceControls := splitRepositoryReflowFragments(source, true)
	semanticFragments, err := splitRepositorySemanticAgainstSourceControls(semantic, sourceControls)
	if err != nil {
		return "", fmt.Errorf("message %d repository Korean dialogue control projection: %w", id, err)
	}
	if len(semanticFragments) != len(sourceFragments) {
		return "", fmt.Errorf("message %d repository Korean dialogue fragment count differs from Japanese source: Korean=%d source=%d", id, len(semanticFragments), len(sourceFragments))
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
		if i < len(sourceControls) {
			result.WriteString(sourceControls[i])
		}
	}
	return result.String(), nil
}

// splitRepositorySemanticAgainstSourceControls mirrors Projection.SplitSemantic
// at repository scope. Authenticated retail projection owns the exact fixed
// control bytes; when those tokens are unavailable, the canonical Japanese
// annotation is the authoritative boundary map. Exact source controls must
// appear in Korean in the same order. Anything between them is the editable
// semantic fragment, including movable substitutions and derived line breaks.
func splitRepositorySemanticAgainstSourceControls(text string, sourceControls []string) ([]string, error) {
	fragments := make([]string, 0, len(sourceControls)+1)
	cursor := 0
	for _, control := range sourceControls {
		relative := strings.Index(text[cursor:], control)
		if relative < 0 {
			return nil, fmt.Errorf("missing fixed source control %q", control)
		}
		at := cursor + relative
		fragment := text[cursor:at]
		if unexpected := repositoryUnexpectedFixedControl(fragment); unexpected != "" {
			return nil, fmt.Errorf("unexpected fixed control %q before source control %q", unexpected, control)
		}
		fragments = append(fragments, fragment)
		cursor = at + len(control)
	}
	trailing := text[cursor:]
	if unexpected := repositoryUnexpectedFixedControl(trailing); unexpected != "" {
		return nil, fmt.Errorf("unexpected trailing fixed control %q", unexpected)
	}
	fragments = append(fragments, trailing)
	return fragments, nil
}

func repositoryUnexpectedFixedControl(text string) string {
	for len(text) > 0 {
		loc := repositoryRuntimeControlTag.FindStringIndex(text)
		if loc == nil {
			return ""
		}
		part := text[loc[0]:loc[1]]
		upper := strings.ToUpper(part)
		if part != lineBreak && !koreanDialogueMovableValueTags[upper] {
			return part
		}
		text = text[loc[1]:]
	}
	return ""
}

// splitRepositoryReflowFragments parses the canonical Japanese annotated source
// into semantic fragments and fixed controls. Source <line-break> is a movable
// layout hint. Movable runtime substitutions stay inside the surrounding text
// fragment, matching message.Project's pureMovable/callerMovable semantics.
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
				continueText := text[loc[1]:]
				text = continueText
				continue
			}
			flushControl(part)
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
