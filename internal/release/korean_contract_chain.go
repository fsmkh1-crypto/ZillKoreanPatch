// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"fmt"
	"sort"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanslots"
	"github.com/HK47196/zill/internal/layout"
)

type koreanEnglishContractChain struct {
	Engine          *layout.Engine
	Layouts         map[int]string
	DerivedC5       int
	DerivedConsumer int
	DerivedDialogue int
	DerivedVisual   int
	DerivedScanner  int
	Warnings        []layout.Warning
	DynamicC5       []int
}

// runKoreanEnglishContractChain is the single production contract path shared by
// desktop, mobile and preflight releases. Keeping derivation order and validation
// here prevents one entry point from silently drifting away from the upstream
// English consumer/visual contract as happened during the forensic cycle.
func runKoreanEnglishContractChain(root, entrypoint string, source *corpus.Project, korean *corpus.KoreanProject, mapping koreanslots.Mapping) (koreanEnglishContractChain, error) {
	var out koreanEnglishContractChain
	layouts := make(map[int]string)
	for _, row := range korean.Entries {
		if row.Layout != "" {
			layouts[row.ID] = row.Layout
		}
	}
	engine, err := loadLayout(root)
	if err != nil {
		return out, err
	}
	out.Engine = engine

	// C5 first, then upstream-English fixed consumers, verified dialogue reflow,
	// hard visual contracts, and finally the separately observed A-054 scanner
	// hardening. English Reflow inserts dialogue line breaks before emitting its
	// overflow warnings; Korean must mirror that order rather than warning on an
	// otherwise-unbroken device layout.
	layouts, out.DerivedC5, err = engine.DeriveKoreanC5StorageLayouts(source, korean, layouts, mapping)
	if err != nil {
		return out, err
	}
	layouts, out.DerivedConsumer, err = engine.DeriveKoreanEnglishConsumerLayouts(source, korean, layouts, mapping)
	if err != nil {
		return out, err
	}
	layouts, out.DerivedDialogue, err = engine.DeriveKoreanEnglishDialogueLayouts(source, korean, layouts, mapping)
	if err != nil {
		return out, err
	}
	layouts, out.DerivedVisual, err = engine.DeriveKoreanEnglishVisualLayouts(source, korean, layouts, mapping)
	if err != nil {
		return out, err
	}
	layouts, out.DerivedScanner, err = engine.DeriveKoreanC22RetailScannerLayouts(source, korean, layouts, mapping)
	if err != nil {
		return out, err
	}

	if err := engine.ValidateKoreanEnglishConsumerContracts(source, korean, layouts, mapping); err != nil {
		return out, err
	}
	out.Warnings, err = engine.AuditKoreanEnglishVisualWarnings(source, korean, layouts, mapping)
	if err != nil {
		return out, err
	}
	out.DynamicC5, err = validateKoreanRuntimeStorage(root, source, korean, layouts, mapping)
	if err != nil {
		return out, err
	}
	out.Layouts = layouts

	warningCounts := make(map[string]int)
	for _, warning := range out.Warnings {
		warningCounts[warning.Code]++
	}
	codes := make([]string, 0, len(warningCounts))
	for code := range warningCounts {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	fmt.Printf("FORENSIC KOREAN_ENGLISH_CONTRACT_CHAIN entrypoint=%s materializable=%d c5_layouts=%d consumer_layouts=%d dialogue_layouts=%d visual_layouts=%d scanner_layouts=%d storage=PASS visual=PASS warnings=%d dynamic_c5=%d\n",
		entrypoint, len(korean.Entries), out.DerivedC5, out.DerivedConsumer, out.DerivedDialogue, out.DerivedVisual, out.DerivedScanner, len(out.Warnings), len(out.DynamicC5))
	for _, code := range codes {
		fmt.Printf("FORENSIC KOREAN_ENGLISH_WARNING entrypoint=%s code=%s count=%d severity=warning\n", entrypoint, code, warningCounts[code])
	}

	// Keep release logs compact: collapse detailed category buckets into the
	// dimensions that determine follow-up risk work. Category-level detail remains
	// available through AuditWarningPopulation for offline/targeted inspection.
	type summaryKey struct {
		code, basis, consumer string
	}
	summary := make(map[summaryKey]int)
	for _, bucket := range engine.AuditWarningPopulation(out.Warnings) {
		summary[summaryKey{bucket.Code, bucket.Basis, bucket.Consumer}] += bucket.Count
	}
	keys := make([]summaryKey, 0, len(summary))
	for key := range summary {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].code != keys[j].code {
			return keys[i].code < keys[j].code
		}
		if keys[i].consumer != keys[j].consumer {
			return keys[i].consumer < keys[j].consumer
		}
		return keys[i].basis < keys[j].basis
	})
	for _, key := range keys {
		fmt.Printf("FORENSIC KOREAN_ENGLISH_WARNING_SUMMARY entrypoint=%s code=%s basis=%s consumer=%s count=%d severity=warning\n",
			entrypoint, key.code, key.basis, key.consumer, summary[key])
	}
	return out, nil
}
