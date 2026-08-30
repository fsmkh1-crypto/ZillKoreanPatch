// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"strings"
	"testing"
)

func TestCompareExactReaders(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
		ok   bool
	}{
		{name: "equal", got: "abcdef", want: "abcdef", ok: true},
		{name: "different", got: "abcxef", want: "abcdef", ok: false},
		{name: "authored longer", got: "abcdefg", want: "abcdef", ok: false},
		{name: "authored shorter", got: "abcde", want: "abcdef", ok: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := compareExactReaders(strings.NewReader(tc.got), strings.NewReader(tc.want))
			if tc.ok && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected mismatch error")
			}
		})
	}
}
