// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"strings"
	"testing"

	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
)

func TestPrepareKoreanFontReplacementNoMappingNeedsNoRetailArchive(t *testing.T) {
	replacement, ok, err := prepareKoreanFontReplacement(t.TempDir(), nil, koreanslots.Plan{Mapping: koreanslots.Mapping{}})
	if err != nil {
		t.Fatal(err)
	}
	if ok || replacement.Payload != nil {
		t.Fatalf("unexpected replacement: ok=%v replacement=%+v", ok, replacement)
	}
}

func TestPrepareKoreanFontReplacementFailsClosedWithoutPAArchive(t *testing.T) {
	_, _, err := prepareKoreanFontReplacement(t.TempDir(), nil, koreanslots.Plan{
		Mapping: koreanslots.Mapping{'가': cp932.GlyphKey(0xAC82)},
	})
	if err == nil || !strings.Contains(err.Error(), "pa archive is unavailable") {
		t.Fatalf("error = %v", err)
	}
}
