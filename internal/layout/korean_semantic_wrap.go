// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import "strings"

// wrapKoreanStoragePreservingControlAdjacency inserts conservative display
// boundaries without ever creating a new boundary immediately before or after a
// runtime control token. message.PreservesLayoutSemantics is the authoritative
// postcondition; this helper only chooses safer candidate positions.
func wrapKoreanStoragePreservingControlAdjacency(text string) string {
	var out strings.Builder
	out.Grow(len(text) + len(text)/8)
	lineRunes := 0
	cursor := 0
	protectNextPlainRune := false
	for _, loc := range controlTag.FindAllStringIndex(text, -1) {
		appendKoreanStoragePlain(&out, text[cursor:loc[0]], &lineRunes, false)
		tag := text[loc[0]:loc[1]]
		out.WriteString(tag)
		if tag == lineBreak {
			lineRunes = 0
			protectNextPlainRune = false
		} else {
			// A semantic string such as <value:$15>여... owns direct control→text
			// adjacency. If the line is already at the wrap threshold, emit one
			// following rune before wrapping rather than inventing a boundary here.
			protectNextPlainRune = true
		}
		cursor = loc[1]
		if cursor < len(text) {
			nextControl := len(text)
			if next := controlTag.FindStringIndex(text[cursor:]); next != nil {
				nextControl = cursor + next[0]
			}
			appendKoreanStoragePlain(&out, text[cursor:nextControl], &lineRunes, protectNextPlainRune)
			cursor = nextControl
			protectNextPlainRune = false
		}
	}
	if cursor < len(text) {
		appendKoreanStoragePlain(&out, text[cursor:], &lineRunes, protectNextPlainRune)
	}
	return out.String()
}

func appendKoreanStoragePlain(out *strings.Builder, text string, lineRunes *int, protectLeading bool) {
	firstEmitted := true
	for _, r := range text {
		space := r == ' ' || r == '\t' || r == '\r' || r == '\n'
		if space {
			if *lineRunes == 0 {
				continue
			}
			if *lineRunes >= 14 && !(protectLeading && firstEmitted) {
				out.WriteString(lineBreak)
				*lineRunes = 0
				continue
			}
			out.WriteRune(' ')
			*lineRunes++
			firstEmitted = false
			continue
		}
		if *lineRunes >= 18 && !(protectLeading && firstEmitted) {
			out.WriteString(lineBreak)
			*lineRunes = 0
		}
		out.WriteRune(r)
		*lineRunes++
		firstEmitted = false
	}
}
