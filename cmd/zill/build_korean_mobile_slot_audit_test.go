// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"testing"

	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
)

func TestAuditMobileExactByteReuseFindsMappedAndUnmappedCandidates(t *testing.T) {
	mappedKey, err := cp932.GlyphKeyFromBytes([]byte{0x81, 0x40})
	if err != nil {
		t.Fatal(err)
	}
	unmappedKey, err := cp932.GlyphKeyFromBytes([]byte{0x81, 0x41})
	if err != nil {
		t.Fatal(err)
	}
	cleanKey, err := cp932.GlyphKeyFromBytes([]byte{0x81, 0x42})
	if err != nil {
		t.Fatal(err)
	}
	plan := koreanslots.Plan{
		Candidates: []cp932.GlyphKey{mappedKey, unmappedKey, cleanKey},
		Mapping:    koreanslots.Mapping{'가': mappedKey},
	}
	audit, err := auditMobileExactByteReuse(plan,
		slotAuditBlob{Name: "BOOT.BIN", Data: []byte{0x00, 0x81, 0x40, 0x00, 0x81, 0x41}},
		slotAuditBlob{Name: "EBOOT.BIN", Data: []byte{0x81, 0x40}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if audit.CandidateHits != 2 {
		t.Fatalf("candidate hits = %d, want 2", audit.CandidateHits)
	}
	if len(audit.MappedHits) != 2 {
		t.Fatalf("mapped hit records = %d, want 2", len(audit.MappedHits))
	}
	for _, hit := range audit.MappedHits {
		if hit.Rune != '가' || hit.Key != mappedKey {
			t.Fatalf("unexpected mapped hit: %+v", hit)
		}
	}
}

func TestAuditMobileExactByteReuseCanAppearOnlyAfterRelocation(t *testing.T) {
	originalKey, err := cp932.GlyphKeyFromBytes([]byte{0x81, 0x40})
	if err != nil {
		t.Fatal(err)
	}
	relocatedKey, err := cp932.GlyphKeyFromBytes([]byte{0x81, 0x41})
	if err != nil {
		t.Fatal(err)
	}
	blob := slotAuditBlob{Name: "EBOOT.BIN", Data: []byte{0x00, 0x81, 0x41, 0x00}}

	h0 := koreanslots.Plan{
		Candidates: []cp932.GlyphKey{originalKey, relocatedKey},
		Mapping:    koreanslots.Mapping{'가': originalKey},
	}
	h0Audit, err := auditMobileExactByteReuse(h0, blob)
	if err != nil {
		t.Fatal(err)
	}
	if len(h0Audit.MappedHits) != 0 {
		t.Fatalf("H0 mapped hits = %#v, want none", h0Audit.MappedHits)
	}

	finalPlan := h0
	finalPlan.Mapping = koreanslots.Mapping{'가': relocatedKey}
	finalAudit, err := auditMobileExactByteReuse(finalPlan, blob)
	if err != nil {
		t.Fatal(err)
	}
	if len(finalAudit.MappedHits) != 1 {
		t.Fatalf("final mapped hits = %#v, want one relocation-only collision", finalAudit.MappedHits)
	}
	if got := finalAudit.MappedHits[0]; got.Rune != '가' || got.Key != relocatedKey || got.Blob != "EBOOT.BIN" {
		t.Fatalf("unexpected final relocation hit: %+v", got)
	}
}

func TestExactByteOffsetsIncludesOverlapsAndHonorsLimit(t *testing.T) {
	got := exactByteOffsets([]byte{1, 1, 1, 1}, []byte{1, 1}, 2)
	if len(got) != 2 || got[0] != 0 || got[1] != 1 {
		t.Fatalf("offsets = %v, want [0 1]", got)
	}
}
