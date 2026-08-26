// SPDX-License-Identifier: GPL-3.0-or-later

package message_test

import (
	"strings"
	"testing"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/message"
)

func TestPlaceholderForRecordSkipsStructuralRecords(t *testing.T) {
	for _, record := range []corpus.Record{
		{ID: 1},
		{ID: 2, Tokens: []corpus.Token{{Kind: "renderer_control", Raw: []byte{1, 2}}}},
	} {
		text, ok, err := message.PlaceholderForRecord(record, "[JP]")
		if err != nil {
			t.Fatalf("record %d returned %v", record.ID, err)
		}
		if ok || text != "" {
			t.Fatalf("record %d = (%q, %v), want structural skip", record.ID, text, ok)
		}
	}
}

func TestPlaceholderForRecordReplacesEditableText(t *testing.T) {
	record := corpus.Record{ID: 3, Tokens: []corpus.Token{
		{Kind: "text", Raw: []byte("source")},
		{Kind: "block_terminator", Raw: []byte{5, 5, 5}},
	}}
	text, ok, err := message.PlaceholderForRecord(record, "[JP]")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("editable record was classified as structural")
	}
	if !strings.Contains(text, "[JP]") {
		t.Fatalf("placeholder text = %q", text)
	}
}
