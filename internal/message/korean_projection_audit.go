// SPDX-License-Identifier: GPL-3.0-or-later

package message

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanslots"
)

// VerifyKoreanProjectionCompatibility proves that the Korean semantic path is
// byte-compatible with the upstream materializer for every translatable source
// record in one authenticated retail bank. It deliberately uses stock Japanese
// source text and an empty Korean mapping, so the only variable under test is
// the Korean-specific semantic splitting rule (not translation wording, glyph
// allocation, layout heuristics, or runtime storage limits).
//
// Fixed source line breaks are explicit in the upstream canonical form but are
// implicit in Korean semantic text. Both forms are materialized with source
// layout enabled and must produce identical record bytes. This directly checks
// the fixed-line-break omission rule, fragment assignment, movable
// substitutions, kana controls, and all fixed source controls against the
// upstream implementation.
func VerifyKoreanProjectionCompatibility(bank corpus.Bank) (int, error) {
	checked := 0
	for _, record := range bank.Records {
		if !hasTranslatableProjectionMaterial(record) {
			continue
		}
		projection, err := Project(record)
		if err != nil {
			return checked, fmt.Errorf("%s: ID %d projection audit: %w", bank.Name, record.ID, err)
		}
		upstreamText, koreanText, err := projection.sourceAuditTexts()
		if err != nil {
			return checked, fmt.Errorf("%s: ID %d projection audit source reconstruction: %w", bank.Name, record.ID, err)
		}

		upstreamBytes, err := projection.Materialize(upstreamText, true)
		if err != nil {
			return checked, fmt.Errorf("%s: ID %d upstream source materialization: %w", bank.Name, record.ID, err)
		}
		koreanBytes, err := projection.MaterializeKorean(koreanText, true, koreanslots.Mapping{})
		if err != nil {
			return checked, fmt.Errorf("%s: ID %d Korean source materialization: %w", bank.Name, record.ID, err)
		}
		if !bytes.Equal(koreanBytes, upstreamBytes) {
			return checked, fmt.Errorf("%s: ID %d Korean projection diverges from upstream materialization (upstream=%d bytes Korean=%d bytes)",
				bank.Name, record.ID, len(upstreamBytes), len(koreanBytes))
		}
		checked++
	}
	return checked, nil
}

func hasTranslatableProjectionMaterial(record corpus.Record) bool {
	for _, token := range record.Tokens {
		if token.Kind == "text" || movableSubstitution(token) {
			return true
		}
	}
	return false
}

func (p *Projection) sourceAuditTexts() (upstream, korean string, err error) {
	var upstreamBuilder, koreanBuilder strings.Builder
	for _, node := range p.nodes {
		if node.fixed {
			upstreamBuilder.WriteString(node.display)
			if node.kind != "line_break" {
				koreanBuilder.WriteString(node.display)
			}
			continue
		}
		if node.fragment < 0 || node.fragment >= len(p.Fragments) {
			return "", "", fmt.Errorf("invalid fragment index %d", node.fragment)
		}
		text, err := sourceFragmentAuditText(p.Fragments[node.fragment])
		if err != nil {
			return "", "", err
		}
		upstreamBuilder.WriteString(text)
		koreanBuilder.WriteString(text)
	}
	return upstreamBuilder.String(), koreanBuilder.String(), nil
}

func sourceFragmentAuditText(fragment Fragment) (string, error) {
	text := strings.ReplaceAll(fragment.SourceLayout, "\n", lineBreak)
	for _, anchor := range fragment.Anchors {
		marker := "{{" + anchor.Name + "}}"
		tag := fmt.Sprintf("<value:$%02X>", anchor.Opcode)
		if !strings.Contains(text, marker) {
			return "", fmt.Errorf("fragment %s is missing source anchor %s", fragment.Key, marker)
		}
		text = strings.Replace(text, marker, tag, 1)
	}
	if reservedAnchor.MatchString(text) {
		return "", fmt.Errorf("fragment %s retains an unresolved source anchor", fragment.Key)
	}
	return text, nil
}
