// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
)

type slotAuditBlob struct {
	Name string
	Data []byte
}

type slotAuditHit struct {
	Rune    rune
	Key     cp932.GlyphKey
	Blob    string
	Offsets []int
}

type mobileSlotAudit struct {
	CandidateHits int
	MappedHits    []slotAuditHit
}

// auditMobileExactByteReuse compares the literal-only mobile allocation against
// the desktop-style exact-byte ownership rule without changing the allocation.
// It is forensic evidence only: a raw two-byte hit can be accidental, so each
// mapped hit is reported with its owning blob and concrete offsets instead of
// being treated as proof that the slot is unsafe.
func auditMobileExactByteReuse(plan koreanslots.Plan, blobs ...slotAuditBlob) (mobileSlotAudit, error) {
	candidateSet := make(map[cp932.GlyphKey]struct{}, len(plan.Candidates))
	for _, key := range plan.Candidates {
		candidateSet[key] = struct{}{}
	}

	candidateHitSet := make(map[cp932.GlyphKey]struct{})
	mappedByKey := make(map[cp932.GlyphKey]rune, len(plan.Mapping))
	for r, key := range plan.Mapping {
		mappedByKey[key] = r
	}

	var mappedHits []slotAuditHit
	for key := range candidateSet {
		encoded, err := key.Bytes()
		if err != nil {
			return mobileSlotAudit{}, fmt.Errorf("mobile slot audit key 0x%04X: %w", uint16(key), err)
		}
		for _, blob := range blobs {
			offsets := exactByteOffsets(blob.Data, encoded, 8)
			if len(offsets) == 0 {
				continue
			}
			candidateHitSet[key] = struct{}{}
			if r, mapped := mappedByKey[key]; mapped {
				mappedHits = append(mappedHits, slotAuditHit{Rune: r, Key: key, Blob: blob.Name, Offsets: offsets})
			}
		}
	}

	sort.Slice(mappedHits, func(i, j int) bool {
		if mappedHits[i].Key != mappedHits[j].Key {
			return mappedHits[i].Key < mappedHits[j].Key
		}
		return mappedHits[i].Blob < mappedHits[j].Blob
	})
	return mobileSlotAudit{CandidateHits: len(candidateHitSet), MappedHits: mappedHits}, nil
}

func exactByteOffsets(data, needle []byte, limit int) []int {
	if len(needle) == 0 || limit <= 0 {
		return nil
	}
	var out []int
	for base := 0; base+len(needle) <= len(data) && len(out) < limit; {
		rel := bytes.Index(data[base:], needle)
		if rel < 0 {
			break
		}
		offset := base + rel
		out = append(out, offset)
		base = offset + 1
	}
	return out
}
