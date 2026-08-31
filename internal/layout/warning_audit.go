// SPDX-License-Identifier: GPL-3.0-or-later

package layout

import (
	"sort"

	"github.com/HK47196/zill/internal/corpus"
)

// WarningAuditBucket groups non-blocking visual warnings by the strongest
// runtime ownership evidence currently available. It is deliberately
// observational: classification must not silently turn an authoring heuristic
// into a runtime contract.
type WarningAuditBucket struct {
	Code     string
	Category string
	Basis    string
	Consumer string
	Count    int
}

// RuntimeSubstitutionAuditBucket groups rows that carry a runtime substitution
// by the exact source token and strongest currently authenticated consumer.
// Count is a row count, not an occurrence count: repeated uses of one token in a
// single message remain one risk-bearing message for census purposes.
type RuntimeSubstitutionAuditBucket struct {
	Token    string
	Basis    string
	Consumer string
	Count    int
}

// WarningOwnership exposes only audit metadata. It does not grant a new runtime
// contract: callers must still require Basis == "verified" before treating the
// category evidence as authenticated.
func (e *Engine) WarningOwnership(id int) (basis, consumer string) {
	_, basis = e.warningCategory(id)
	return basis, e.warningConsumer(id)
}

// AuditWarningPopulation classifies warnings without changing layouts or
// validation semantics. This is the release-facing census used to decide which
// residual populations deserve a separately authenticated hard gate.
func (e *Engine) AuditWarningPopulation(warnings []Warning) []WarningAuditBucket {
	type key struct {
		code, category, basis, consumer string
	}
	counts := make(map[key]int)
	for _, warning := range warnings {
		category, basis := e.warningCategory(warning.MessageID)
		consumer := e.warningConsumer(warning.MessageID)
		counts[key{warning.Code, category, basis, consumer}]++
	}
	out := make([]WarningAuditBucket, 0, len(counts))
	for k, count := range counts {
		out = append(out, WarningAuditBucket{
			Code:     k.code,
			Category: k.category,
			Basis:    k.basis,
			Consumer: k.consumer,
			Count:    count,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Code != out[j].Code {
			return out[i].Code < out[j].Code
		}
		if out[i].Consumer != out[j].Consumer {
			return out[i].Consumer < out[j].Consumer
		}
		if out[i].Basis != out[j].Basis {
			return out[i].Basis < out[j].Basis
		}
		return out[i].Category < out[j].Category
	})
	return out
}

// AuditRuntimeSubstitutionPopulation gives the broad warning population useful
// operational shape without assigning an invented expansion bound. $28 remains
// the only value substitution with an independently proven maximum in the C5
// byte audit; every other token stays explicitly unbounded here.
func (e *Engine) AuditRuntimeSubstitutionPopulation(korean *corpus.KoreanProject) []RuntimeSubstitutionAuditBucket {
	if korean == nil {
		return nil
	}
	type key struct { token, basis, consumer string }
	counts := make(map[key]int)
	for _, row := range korean.Entries {
		seen := make(map[string]struct{})
		for _, token := range valueTag.FindAllString(row.Korean, -1) {
			seen[token] = struct{}{}
		}
		if formatSignatureID(row.ID) && printfConversion.MatchString(visible(row.Korean)) {
			seen["printf"] = struct{}{}
		}
		if len(seen) == 0 {
			continue
		}
		_, basis := e.warningCategory(row.ID)
		consumer := e.warningConsumer(row.ID)
		for token := range seen {
			counts[key{token, basis, consumer}]++
		}
	}
	out := make([]RuntimeSubstitutionAuditBucket, 0, len(counts))
	for k, count := range counts {
		out = append(out, RuntimeSubstitutionAuditBucket{Token: k.token, Basis: k.basis, Consumer: k.consumer, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Token != out[j].Token { return out[i].Token < out[j].Token }
		if out[i].Consumer != out[j].Consumer { return out[i].Consumer < out[j].Consumer }
		return out[i].Basis < out[j].Basis
	})
	return out
}

func (e *Engine) warningCategory(id int) (string, string) {
	for _, r := range e.categories {
		if id < r.First {
			break
		}
		if id <= r.Last {
			return r.Category, r.Basis
		}
	}
	return "uncategorized", "unknown"
}

func (e *Engine) warningConsumer(id int) string {
	// Prefer specific storage/UI consumers over broad category evidence.
	switch {
	case e.itemDescription(id):
		return "item-description"
	case e.has(e.consumers.C22IDs, id):
		return "c22"
	case e.has(e.consumers.C5PortraitIDs, id):
		return "c5-portrait"
	case e.has(e.consumers.C5IDs, id):
		return "c5"
	case e.has(e.consumers.SinglePageC5IDs, id):
		return "c5-single-page"
	case e.has(e.consumers.BoundedLabelIDs, id):
		return "bounded-label"
	case e.has(e.consumers.GuildClientIDs, id):
		return "guild-client"
	case e.has(e.consumers.GuildRegionIDs, id):
		return "guild-region"
	case e.has(e.consumers.GuildCommentaryIDs, id):
		return "guild-commentary"
	case e.narrowText(id):
		return "verified-narrow-dialogue"
	default:
		return "unproven"
	}
}
