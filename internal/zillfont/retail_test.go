// SPDX-License-Identifier: GPL-3.0-or-later

package zillfont

import (
	"strings"
	"testing"

	"github.com/HK47196/zill/internal/koreanslots"
)

func TestParseAuthenticatedRetailPAFRejectsWrongSizeBeforeParsing(t *testing.T) {
	_, err := ParseAuthenticatedRetailPAF(make([]byte, RetailPAFMemberSize-1))
	if err == nil || !strings.Contains(err.Error(), "size") {
		t.Fatalf("error = %v", err)
	}
}

func TestParseAuthenticatedRetailPAFRejectsWrongFingerprint(t *testing.T) {
	_, err := ParseAuthenticatedRetailPAF(make([]byte, RetailPAFMemberSize))
	if err == nil || !strings.Contains(err.Error(), "fingerprint") {
		t.Fatalf("error = %v", err)
	}
}

func TestPatchAuthenticatedRetailAtlasRejectsUnauthenticatedAtlasFirst(t *testing.T) {
	_, err := PatchAuthenticatedRetailAtlas(
		make([]byte, RetailAtlasMemberSize),
		make([]byte, RetailPAFMemberSize),
		koreanslots.Mapping{},
		map[rune]Raster{},
	)
	if err == nil || !strings.Contains(err.Error(), "font/zillfont.par") || !strings.Contains(err.Error(), "fingerprint") {
		t.Fatalf("error = %v", err)
	}
}
