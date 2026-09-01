// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import "strings"

const (
	koreanStoragePreferredWrapRunes = 14
	koreanStorageHardWrapRunes      = 18
)

type koreanStorageWrapMode struct {
	protectInitialPlainRune       bool
	protectAfterLineBreak         bool
	preserveProtectedWhitespace   bool
}

// wrapKoreanStoragePreservingControlAdjacency inserts conservative display
// boundaries without ever creating a new boundary immediately before or after a
// runtime substitution control. It also repairs an unsafe derived boundary that
// an earlier projection stage may already have inserted around a <value:...>
// token. Authored semantic/control topology is still protected by the caller's
// message.PreservesLayoutSemantics postcondition.
func wrapKoreanStoragePreservingControlAdjacency(text string) string {
	text = stripDerivedValueAdjacencyBreaks(text)
	return wrapKoreanRuneStorage(text, koreanStorageWrapMode{
		protectInitialPlainRune:     true,
		protectAfterLineBreak:       true,
		preserveProtectedWhitespace: true,
	})
}

// wrapKoreanRuneStorage is the single conservative 14/18-rune storage wrapper
// used by the Korean storage projections. Consumer-specific wrappers only select
// the legacy whitespace/adjacency mode; the wrapping state machine itself must
// not diverge between C5/scanner and C22 paths.
func wrapKoreanRuneStorage(text string, mode koreanStorageWrapMode) string {
	var out strings.Builder
	out.Grow(len(text) + len(text)/8)
	lineRunes := 0
	cursor := 0
	protectNextPlainRune := mode.protectInitialPlainRune
	for _, loc := range controlTag.FindAllStringIndex(text, -1) {
		appendKoreanStoragePlain(&out, text[cursor:loc[0]], &lineRunes, protectNextPlainRune, mode.preserveProtectedWhitespace)
		protectNextPlainRune = false
		tag := text[loc[0]:loc[1]]
		out.WriteString(tag)
		switch {
		case tag == lineBreak:
			lineRunes = 0
			protectNextPlainRune = mode.protectAfterLineBreak
		case strings.HasPrefix(tag, "<value:"):
			// A semantic string such as <value:$15>여... owns direct control→text
			// adjacency. If the line is already at the wrap threshold, emit one
			// following rune before wrapping rather than inventing a boundary here.
			// The same protection keeps canonical whitespace immediately after a
			// value token instead of silently dropping it at a line start.
			protectNextPlainRune = true
		}
		cursor = loc[1]
	}
	appendKoreanStoragePlain(&out, text[cursor:], &lineRunes, protectNextPlainRune, mode.preserveProtectedWhitespace)
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

func appendKoreanStoragePlain(out *strings.Builder, text string, lineRunes *int, protectLeading, preserveProtectedWhitespace bool) {
	firstEmitted := true
	for _, r := range text {
		space := r == ' ' || r == '\t' || r == '\r' || r == '\n'
		if space {
			if *lineRunes == 0 {
				if protectLeading && firstEmitted {
					if preserveProtectedWhitespace {
						out.WriteRune(r)
					} else {
						out.WriteRune(' ')
					}
					*lineRunes++
					firstEmitted = false
				}
				continue
			}
			if *lineRunes >= koreanStoragePreferredWrapRunes && !(protectLeading && firstEmitted) {
				out.WriteString(lineBreak)
				*lineRunes = 0
				continue
			}
			out.WriteRune(' ')
			*lineRunes++
			firstEmitted = false
			continue
		}
		if *lineRunes >= koreanStorageHardWrapRunes && !(protectLeading && firstEmitted) {
			out.WriteString(lineBreak)
			*lineRunes = 0
		}
		out.WriteRune(r)
		*lineRunes++
		firstEmitted = false
	}
}
