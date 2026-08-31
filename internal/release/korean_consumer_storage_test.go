// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/cp932"
	"github.com/HK47196/zill/internal/koreanslots"
)

// TestCurrentKoreanCorpusEnglishConsumerStorageContracts is the repository-only
// counterpart of the device builder's fixed-consumer and visual gates. It
// exhausts every upstream-English asset-independent contract and warning class
// over all 42,016 accepted Korean rows using Korean renderer metrics.
//
// The derivation order intentionally mirrors runKoreanEnglishContractChain:
// fixed English consumers -> verified English dialogue reflow -> hard visual
// layouts -> validation/warning census. This keeps the full-corpus audit from
// warning on an unbroken repository-only layout that the actual builder would
// have reflowed before the same visual audit.
//
// C5 projection/storage validation is deliberately not repeated here because it
// requires authenticated retail source tokens. The shared desktop/mobile/
// preflight contract chain runs that exact asset-backed gate before compilation.
//
// Do NOT call BuildKoreanBetaProject here. That projection needs authenticated
// retail records to classify editability; on an asset-free repository load its
// Record.Raw fields are intentionally empty and every accepted row can otherwise
// be misclassified as structural. This census therefore validates the canonical
// overlay directly and hard-asserts its checked population.
func TestCurrentKoreanCorpusEnglishConsumerStorageContracts(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	source, _, err := corpus.LoadProject(root)
	if err != nil { t.Fatal(err) }
	korean, _, err := corpus.LoadKoreanProject(root, source)
	if err != nil { t.Fatal(err) }
	const wantCanonical = 42016
	if len(korean.Entries) != wantCanonical {
		t.Fatalf("canonical Korean corpus drift: got %d want %d", len(korean.Entries), wantCanonical)
	}

	texts, err := korean.RuntimeTexts(source)
	if err != nil { t.Fatal(err) }
	mapping := make(koreanslots.Mapping)
	for _, r := range koreanslots.RequiredCustomRunes(texts) {
		mapping[r] = cp932.GlyphKey(0xAC82)
	}

	layouts := make(map[int]string)
	for _, row := range korean.Entries {
		if row.Layout != "" { layouts[row.ID] = row.Layout }
	}
	engine, err := loadLayout(root)
	if err != nil { t.Fatal(err) }
	layouts, derivedEnglish, err := engine.DeriveKoreanEnglishConsumerLayouts(source, korean, layouts, mapping)
	if err != nil { t.Fatal(err) }
	layouts, derivedDialogue, err := engine.DeriveKoreanEnglishDialogueLayouts(source, korean, layouts, mapping)
	if err != nil { t.Fatal(err) }
	dialogueChecked, dialogueOverflowIDs, err := engine.AuditKoreanEnglishDialogueResiduals(source, korean, layouts, mapping)
	if err != nil { t.Fatal(err) }
	if len(dialogueOverflowIDs) != 0 {
		t.Fatalf("verified dialogue residual overflow after English-authority reflow: count=%d ids=%v", len(dialogueOverflowIDs), dialogueOverflowIDs)
	}
	t.Logf("FORENSIC KOREAN_DIALOGUE_REFLOW_SUMMARY checked=%d derived=%d residual_overflow=%d scope=verified_narrow_text", dialogueChecked, derivedDialogue, len(dialogueOverflowIDs))
	layouts, derivedVisual, err := engine.DeriveKoreanEnglishVisualLayouts(source, korean, layouts, mapping)
	if err != nil { t.Fatal(err) }
	if err := engine.ValidateKoreanEnglishConsumerContracts(source, korean, layouts, mapping); err != nil { t.Fatal(err) }
	warnings, err := engine.AuditKoreanEnglishVisualWarnings(source, korean, layouts, mapping)
	if err != nil { t.Fatal(err) }

	warningCounts := make(map[string]int)
	var itemWarningIDs []int
	runtimeWarningIDs := make(map[int]struct{})
	for _, warning := range warnings {
		warningCounts[warning.Code]++
		if warning.Code == "item_description_single_line_overflow" {
			itemWarningIDs = append(itemWarningIDs, warning.MessageID)
		}
		if warning.Code == "runtime_substitution_unbounded" {
			runtimeWarningIDs[warning.MessageID] = struct{}{}
		}
	}
	codes := make([]string, 0, len(warningCounts))
	for code := range warningCounts { codes = append(codes, code) }
	sort.Strings(codes)
	for _, code := range codes {
		t.Logf("FORENSIC KOREAN_ENGLISH_WARNING_CENSUS code=%s count=%d severity=warning", code, warningCounts[code])
	}
	sort.Ints(itemWarningIDs)
	t.Logf("FORENSIC U6_ITEM_DESCRIPTION_WARNING_IDS count=%d ids=%v upstream_behavior=leave_unreflowed_and_warn", len(itemWarningIDs), itemWarningIDs)

	type summaryKey struct { code, basis, consumer string }
	summary := make(map[summaryKey]int)
	for _, bucket := range engine.AuditWarningPopulation(warnings) {
		summary[summaryKey{bucket.Code, bucket.Basis, bucket.Consumer}] += bucket.Count
	}
	keys := make([]summaryKey, 0, len(summary))
	for key := range summary { keys = append(keys, key) }
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].code != keys[j].code { return keys[i].code < keys[j].code }
		if keys[i].consumer != keys[j].consumer { return keys[i].consumer < keys[j].consumer }
		return keys[i].basis < keys[j].basis
	})
	for _, key := range keys {
		t.Logf("FORENSIC U6_WARNING_SUMMARY code=%s basis=%s consumer=%s count=%d", key.code, key.basis, key.consumer, summary[key])
	}
	for _, bucket := range engine.AuditRuntimeSubstitutionPopulation(korean) {
		bound := "unproven"
		if bucket.Token == "<value:$28>" { bound = "player-name-max-16-bytes" }
		t.Logf("FORENSIC U6_RUNTIME_SUBSTITUTION token=%s basis=%s consumer=%s messages=%d bound=%s", bucket.Token, bucket.Basis, bucket.Consumer, bucket.Count, bound)
	}

	var verified15 []int
	var verifiedDialogue15 []int
	valueToken := regexp.MustCompile(`<value:\$[0-9A-F]{2}>`)
	var verifiedDialogueRuntime []int
	var verifiedDialogueUnknown []int
	unknownTokens := make(map[string]struct{})
	knownBoundOnly := 0
	for _, row := range korean.Entries {
		basis, consumer := engine.WarningOwnership(row.ID)
		if strings.Contains(row.Korean, "<value:$15>") && basis == "verified" {
			verified15 = append(verified15, row.ID)
		}
		if _, warned := runtimeWarningIDs[row.ID]; !warned || consumer != "verified-narrow-dialogue" {
			continue
		}
		verifiedDialogueRuntime = append(verifiedDialogueRuntime, row.ID)
		tokens := valueToken.FindAllString(row.Korean, -1)
		onlyKnownBound := len(tokens) > 0
		for _, token := range tokens {
			if token == "<value:$15>" {
				verifiedDialogue15 = append(verifiedDialogue15, row.ID)
			}
			if token == "<value:$28>" {
				continue
			}
			onlyKnownBound = false
			unknownTokens[token] = struct{}{}
		}
		if onlyKnownBound {
			knownBoundOnly++
			continue
		}
		verifiedDialogueUnknown = append(verifiedDialogueUnknown, row.ID)
	}
	sort.Ints(verified15)
	sort.Ints(verifiedDialogue15)
	sort.Ints(verifiedDialogueRuntime)
	sort.Ints(verifiedDialogueUnknown)
	unknownTokenList := make([]string, 0, len(unknownTokens))
	for token := range unknownTokens { unknownTokenList = append(unknownTokenList, token) }
	sort.Strings(unknownTokenList)

	wantVerifiedDialogue15 := []int{170025, 170207, 170208, 170209}
	wantUnknownTokens := []string{"<value:$01>", "<value:$15>", "<value:$1A>", "<value:$20>", "<value:$23>", "<value:$29>", "<value:$2A>", "<value:$2B>", "<value:$2C>", "<value:$33>", "<value:$3C>"}
	if len(verifiedDialogueRuntime) != 1214 || knownBoundOnly != 809 || len(verifiedDialogueUnknown) != 405 {
		t.Fatalf("verified-dialogue substitution boundary drift: total=%d known_bound_only=%d unbounded_unique=%d", len(verifiedDialogueRuntime), knownBoundOnly, len(verifiedDialogueUnknown))
	}
	if !slices.Equal(verifiedDialogue15, wantVerifiedDialogue15) {
		t.Fatalf("verified-dialogue $15 regression set drift: got %v want %v", verifiedDialogue15, wantVerifiedDialogue15)
	}
	if !slices.Equal(unknownTokenList, wantUnknownTokens) {
		t.Fatalf("verified-dialogue unbounded token set drift: got %v want %v", unknownTokenList, wantUnknownTokens)
	}

	t.Logf("FORENSIC U6_VALUE15_VERIFIED_IDS count=%d ids=%v runtime_bound=unproven", len(verified15), verified15)
	t.Logf("FORENSIC U6_VALUE15_VERIFIED_DIALOGUE_IDS count=%d ids=%v runtime_bound=unproven", len(verifiedDialogue15), verifiedDialogue15)
	t.Logf("FORENSIC U6_VERIFIED_DIALOGUE_SUBSTITUTION_BOUNDARY total=%d known_bound_only=%d unbounded_unique=%d unbounded_tokens=%v", len(verifiedDialogueRuntime), knownBoundOnly, len(verifiedDialogueUnknown), unknownTokenList)
	t.Logf("FORENSIC U6_VERIFIED_DIALOGUE_UNBOUNDED_IDS count=%d ids=%v", len(verifiedDialogueUnknown), verifiedDialogueUnknown)

	checked := len(korean.Entries)
	if checked != wantCanonical { t.Fatalf("consumer census checked %d rows, want %d", checked, wantCanonical) }
	t.Logf("FORENSIC KOREAN_CONSUMER_STORAGE_SUMMARY canonical=%d checked=%d english_layouts=%d dialogue_layouts=%d visual_layouts=%d contracts=PASS visual=PASS warnings=%d exact_asset_gates=C5,CompileBankKorean",
		len(korean.Entries), checked, derivedEnglish, derivedDialogue, derivedVisual, len(warnings))
}
