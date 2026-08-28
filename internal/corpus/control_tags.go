// SPDX-License-Identifier: GPL-3.0-or-later

package corpus

import (
	"regexp"
	"unicode"
)

// runtimeControlTag matches only textual projections emitted by displayText for
// source-owned runtime bytecode. Angle-bracketed game text such as <未使用> is
// deliberately not a control tag and remains translatable.
var runtimeControlTag = regexp.MustCompile(`<(?:if|select|call:[0-9]+|jump:[0-9]+|value:\$[0-9A-F]{2}|add|subtract|multiply|divide|modulo|equal|not-equal|less|greater|less-equal|greater-equal|and|or|operator:\$[0-9A-F]{2}:\$[0-9A-F]{2}|color:[^<>]|discard:[^<>]:\$[0-9A-F]{2}|escape:\$[0-9A-F]{2}|end|separator|backspace|tab|line-break|\$[0-9A-F]{2})>`)

// RuntimeControlTags returns canonical projected runtime controls in display
// order. It includes line breaks.
func RuntimeControlTags(text string) []string {
	return runtimeControlTag.FindAllString(text, -1)
}

// FixedRuntimeControlTags returns source-owned runtime controls while excluding
// line breaks, which are layout-authorable for Korean text.
func FixedRuntimeControlTags(text string) []string {
	all := RuntimeControlTags(text)
	fixed := make([]string, 0, len(all))
	for _, token := range all {
		if token != "<line-break>" {
			fixed = append(fixed, token)
		}
	}
	return fixed
}

func hasVisibleLiteral(text string) bool {
	for _, r := range text {
		if !unicode.IsSpace(r) {
			return true
		}
	}
	return false
}

// FixedRuntimeLiteralOccupancy reports whether each literal slot around the
// fixed runtime controls contains visible source-owned text. Layout-authorable
// <line-break> tags and whitespace do not create slots and are ignored.
//
// The result always has len(FixedRuntimeControlTags(text))+1 elements: the
// literal before the first fixed control, each literal between adjacent fixed
// controls, and the literal after the final fixed control. This lets callers
// catch a dangerous class of translator corruption where the control sequence
// is preserved but an entire branch-local fixed literal is accidentally
// dropped.
func FixedRuntimeLiteralOccupancy(text string) []bool {
	matches := runtimeControlTag.FindAllStringIndex(text, -1)
	occupancy := make([]bool, 0, len(matches)+1)
	visible := false
	last := 0
	for _, match := range matches {
		if hasVisibleLiteral(text[last:match[0]]) {
			visible = true
		}
		token := text[match[0]:match[1]]
		last = match[1]
		if token == "<line-break>" {
			continue
		}
		occupancy = append(occupancy, visible)
		visible = false
	}
	if hasVisibleLiteral(text[last:]) {
		visible = true
	}
	occupancy = append(occupancy, visible)
	return occupancy
}
