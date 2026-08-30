// SPDX-License-Identifier: GPL-3.0-or-later

package message

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/cp932"
)

// VerifyKoreanProjectionCompatibility proves that the Korean semantic path can
// reconstruct every translatable authenticated retail source record byte-for-
// byte. It deliberately audits retail source material rather than contributor
// input, so source literals that the translator-facing validators forbid (for
// example half-width kana or literal angle brackets) must remain admissible in
// this audit-only path.
//
// Fixed source line breaks are explicit in the upstream canonical form but are
// implicit in Korean semantic text. The Korean split rule is therefore applied
// to reconstructed source semantic text and its output is compared directly to
// the authenticated retail record bytes. This checks fixed-line-break omission,
// fragment assignment, movable substitutions, kana controls, and fixed source
// controls without incorrectly re-validating retail source as new translation.
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
		_, koreanText, err := projection.sourceAuditTexts()
		if err != nil {
			return checked, fmt.Errorf("%s: ID %d projection audit source reconstruction: %w", bank.Name, record.ID, err)
		}

		// nil is intentionally audit-only: splitSemanticWith still verifies source
		// substitution/format topology but does not apply contributor natural-text
		// bans to authenticated retail literals.
		values, err := projection.splitSemanticWith(koreanText, nil)
		if err != nil {
			return checked, fmt.Errorf("%s: ID %d Korean source split: %w", bank.Name, record.ID, err)
		}
		koreanBytes, err := projection.materializeValues(values, true, cp932.Encode)
		if err != nil {
			return checked, fmt.Errorf("%s: ID %d Korean source materialization: %w", bank.Name, record.ID, err)
		}
		if !bytes.Equal(koreanBytes, record.Raw) {
			return checked, fmt.Errorf("%s: ID %d Korean projection diverges from authenticated retail bytes (retail=%d bytes Korean=%d bytes)",
				bank.Name, record.ID, len(record.Raw), len(koreanBytes))
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
