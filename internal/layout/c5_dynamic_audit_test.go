// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"bytes"
	"testing"

	"github.com/HK47196/zill/internal/corpus"
)

func TestAuditC5KnownPlayerNameCanCrossPageCapacity(t *testing.T) {
	for _, tc := range []struct {
		name       string
		static     int
		wantKnown  int
		wantExceed bool
	}{
		{name: "last safe", static: 239, wantKnown: 255, wantExceed: false},
		{name: "first overflow", static: 240, wantKnown: 256, wantExceed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			leaf := c5AuditLeaf{events: []c5AuditEvent{
				{raw: bytes.Repeat([]byte{'A'}, tc.static)},
				{substitution: true, opcode: 0x28},
			}}
			pages := auditC5LeafPages(leaf)
			if len(pages) != 1 {
				t.Fatalf("pages = %d, want 1", len(pages))
			}
			got := pages[0]
			if got.StaticBytes != tc.static || got.KnownMaxBytes != tc.wantKnown || got.PlayerNameCount != 1 {
				t.Fatalf("page = %#v, want static=%d known=%d playerNames=1", got, tc.static, tc.wantKnown)
			}
			if got.ExceedsPageBuffer() != tc.wantExceed {
				t.Fatalf("ExceedsPageBuffer = %v, want %v", got.ExceedsPageBuffer(), tc.wantExceed)
			}
		})
	}
}

func TestAuditC5UnknownSubstitutionDoesNotInventBound(t *testing.T) {
	leaf := c5AuditLeaf{events: []c5AuditEvent{
		{raw: bytes.Repeat([]byte{'A'}, 250)},
		{substitution: true, opcode: 0x15},
	}}
	pages := auditC5LeafPages(leaf)
	if len(pages) != 1 {
		t.Fatalf("pages = %d, want 1", len(pages))
	}
	got := pages[0]
	if got.StaticBytes != 250 || got.KnownMaxBytes != 250 || got.UnknownSubstitutions != 1 {
		t.Fatalf("page = %#v, want unknown substitution without guessed bytes", got)
	}
	if got.ExceedsPageBuffer() {
		t.Fatalf("unknown substitution was incorrectly promoted into a proven overflow")
	}
}

func TestAuditC5ThirdLineBreakIsCountedBeforeNextPage(t *testing.T) {
	leaf := c5AuditLeaf{events: []c5AuditEvent{
		{raw: []byte{'a', 10, 'b', 10, 'c', 10}},
		{substitution: true, opcode: 0x28},
	}}
	pages := auditC5LeafPages(leaf)
	if len(pages) != 2 {
		t.Fatalf("pages = %d, want 2", len(pages))
	}
	if pages[0].StaticBytes != 6 || pages[0].KnownMaxBytes != 6 {
		t.Fatalf("first page = %#v, want all six static bytes including boundary newline", pages[0])
	}
	if pages[1].StaticBytes != 0 || pages[1].KnownMaxBytes != playerNameMaxEncodedBytes || pages[1].PlayerNameCount != 1 {
		t.Fatalf("second page = %#v, want player-name expansion on new page", pages[1])
	}
}

func TestC5PageCursorCountsThirdBreakThenTransitions(t *testing.T) {
	cursor := c5PageCursor{}
	for i, b := range []byte{'a', 10, 'b', 10, 'c'} {
		if cursor.addByte(b) {
			t.Fatalf("byte %d unexpectedly started next page", i)
		}
	}
	if !cursor.addByte(10) {
		t.Fatal("third line break did not start next page after being counted")
	}
}

func TestC5SelectShapeCentralizesFixed33Arms(t *testing.T) {
	arms, sink, err := c5SelectShape([]byte{2, 0x33})
	if err != nil {
		t.Fatal(err)
	}
	if arms != c5Select33Arms || arms != 8 || !sink {
		t.Fatalf("$33 select shape = arms=%d sink=%v, want 8,true", arms, sink)
	}
}

func TestWalkC5AuditKeepsInlineSubstitutionEvent(t *testing.T) {
	tokens := []corpus.Token{
		{Kind: "text", Raw: []byte("hello ")},
		{Kind: "substitution", Raw: []byte{2, 0x28}},
		{Kind: "block_terminator", Raw: []byte{5, 5, 5}},
	}
	leaves, err := walkC5Audit(tokens, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(leaves) != 1 || len(leaves[0].events) != 2 {
		t.Fatalf("leaves = %#v, want one leaf with text and substitution", leaves)
	}
	if event := leaves[0].events[1]; !event.substitution || event.opcode != 0x28 {
		t.Fatalf("substitution event = %#v, want opcode 0x28", event)
	}
}
