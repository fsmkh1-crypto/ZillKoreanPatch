// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import "strings"

// wrapKoreanStoragePreservingControlAdjacency inserts conservative display
// boundaries without ever creating a new boundary immediately before or after a
// runtime substitution control. It also repairs an unsafe derived boundary that
// an earlier projection stage may already have inserted around a <value:...>
// token. Authored semantic/control topology is still protected by the caller's
// message.PreservesLayoutSemantics postcondition.
func wrapKoreanStoragePreservingControlAdjacency(text string) string {
	text = stripDerivedValueAdjacencyBreaks(text)

	var out strings.Builder
	out.Grow(len(text) + len(text)/8)
	lineRunes := 0
	cursor := 0
	// Whitespace already present at the beginning of the input is authored
	// semantic text, not a wrapping delimiter. Preserve its first rune so the
	// rest of the run is retained normally as well.
	protectNextPlainRune := true
	for _, loc := range controlTag.FindAllStringIndex(text, -1) {
		appendKoreanStoragePlain(&out, text[cursor:loc[0]], &lineRunes, protectNextPlainRune)
		protectNextPlainRune = false
		tag := text[loc[0]:loc[1]]
		out.WriteString(tag)
		if tag == lineBreak {
			lineRunes = 0
			// A line break already present in the input owns the whitespace that
			// follows it. Generated wrapping breaks never leave their delimiter
			// whitespace behind, so preserving an existing post-break run cannot
			// turn a machine-created separator into semantic text.
			protectNextPlainRune = true
		} else {
			// A semantic string such as <value:$15>여... owns direct control→text
			// adjacency. If the line is already at the wrap threshold, emit one
			// following rune before wrapping rather than inventing a boundary here.
			// The same protection keeps canonical whitespace immediately after a
			// value token instead of silently dropping it at a line start.
			protectNextPlainRune = strings.HasPrefix(tag, "<value:")
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

// stripDerivedValueAdjacencyBreaks is deliberately narrow. C5 runs before C22
// in the mobile/release projection chain, so C22 can inherit a build-local
// <line-break> immediately beside a runtime value token even though canonical
// Korean did not contain that boundary. Remove only those value-adjacent breaks
// before C22 reflows the record. If a boundary was actually authored/canonical,
// the caller's semantic postcondition rejects the changed candidate.
func stripDerivedValueAdjacencyBreaks(text string) string {
	for _, tag := range controlTag.FindAllString(text, -1) {
		if !strings.HasPrefix(tag, "<value:") {
			continue
		}
		text = strings.ReplaceAll(text, lineBreak+tag, tag)
		text = strings.ReplaceAll(text, tag+lineBreak, tag)
	}
	return text
}

func appendKoreanStoragePlain(out *strings.Builder, text string, lineRunes *int, protectLeading bool) {
	firstEmitted := true
	for _, r := range text {
		space := r == ' ' || r == '\t' || r == '\r' || r == '\n'
		if space {
			if *lineRunes == 0 {
				if protectLeading && firstEmitted {
					out.WriteRune(r)
					*lineRunes++
					firstEmitted = false
				}
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
