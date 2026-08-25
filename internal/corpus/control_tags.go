// SPDX-License-Identifier: GPL-3.0-or-later

package corpus

import "regexp"

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
