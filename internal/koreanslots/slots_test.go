// SPDX-License-Identifier: GPL-3.0-or-later

package koreanslots

import (
	"reflect"
	"testing"

	"github.com/HK47196/zill/internal/cp932"
)

func TestAllocateIsDeterministic(t *testing.T) {
	keys := []cp932.GlyphKey{0xA082, 0x9F82, 0xA182}
	a, err := Allocate([]rune{'힣', '가', '나', '가'}, keys)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Allocate([]rune{'나', '가', '힣'}, []cp932.GlyphKey{0xA182, 0xA082, 0x9F82})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("mapping depends on input order: %#v vs %#v", a, b)
	}
}

func TestEncodeUsesMappedBytes(t *testing.T) {
	mapping := Mapping{'가': cp932.GlyphKey(0xAC82)}
	got, err := Encode("A가B", mapping)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{'A', 0x82, 0xAC, 'B'}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got % X, want % X", got, want)
	}
}

func TestMissingKoreanMappingFailsClosed(t *testing.T) {
	if _, err := Encode("가", nil); err == nil {
		t.Fatal("expected unmapped Hangul to fail")
	}
}
