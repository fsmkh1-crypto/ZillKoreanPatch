// SPDX-License-Identifier: GPL-3.0-or-later

package release

import (
	"fmt"

	"github.com/HK47196/zill/internal/corpus"
	"github.com/HK47196/zill/internal/koreanslots"
)

// KoreanAlphaPlanBuilder receives the exact retail-bound source and Korean
// project that the ISO compiler will use. The callback must not mutate either
// project. Keeping plan generation inside this bound-source lifetime prevents
// planner/compiler drift.
type KoreanAlphaPlanBuilder func(source *corpus.Project, korean *corpus.KoreanProject) (koreanslots.Plan, int, int, error)

type koreanMobilePreparation struct {
	archives   []*archive
	executable []byte
	parameter  []byte
}

func (p *koreanMobilePreparation) close() {
	if p == nil {
		return
	}
	for _, archive := range p.archives {
		if archive != nil && archive.pair != nil {
			_ = archive.pair.Close()
		}
	}
}

// prepareKoreanMobilePayload is the single deterministic preparation path for
// both real mobile ISO builds and preflight. The two entry points differ only in
// retail-ISO lifetime/output authoring; corpus binding, projection, planning,
// English-first contracts, bank compilation, font transform, EBOOT, and SFO
// must not drift into independently maintained implementations.
func prepareKoreanMobilePayload(root, gameDir, version, mode string, planBuilder KoreanAlphaPlanBuilder) (_ *koreanMobilePreparation, err error) {
	source, _, err := corpus.LoadProject(root)
	if err != nil {
		return nil, err
	}
	canonicalKorean, _, err := corpus.LoadKoreanProject(root, source)
	if err != nil {
		return nil, err
	}

	archives, err := openArchives(gameDir)
	if err != nil {
		return nil, err
	}
	prepared := &koreanMobilePreparation{archives: archives}
	committed := false
	defer func() {
		if !committed {
			prepared.close()
		}
	}()

	banks, owners, err := loadRetailBanks(archives)
	if err != nil {
		return nil, err
	}
	if err := corpus.BindBanks(source, banks); err != nil {
		return nil, err
	}

	korean, skippedStructural, err := BuildKoreanBetaProject(source, canonicalKorean)
	if err != nil {
		return nil, err
	}
	if mode == "preflight" {
		fmt.Printf("Korean beta preflight projection: canonical=%d materializable=%d structural_retail=%d\n",
			len(canonicalKorean.Entries), len(korean.Entries), skippedStructural)
	} else {
		fmt.Printf("Korean beta projection: canonical=%d materializable=%d structural_retail=%d\n",
			len(canonicalKorean.Entries), len(korean.Entries), skippedStructural)
	}

	plan, coverage, total, err := planBuilder(source, korean)
	if err != nil {
		return nil, err
	}
	if mode == "preflight" {
		fmt.Printf("Korean preflight coverage: %d/%d records; custom glyphs: %d; reusable slots: %d\n",
			coverage, total, len(plan.CustomRunes), len(plan.Candidates))
	} else {
		fmt.Printf("Korean coverage: %d/%d records; custom glyphs: %d; reusable slots: %d\n",
			coverage, total, len(plan.CustomRunes), len(plan.Candidates))
	}

	contract, err := runKoreanEnglishContractChain(root, mode, source, korean, plan.Mapping)
	if err != nil {
		return nil, err
	}
	compiled, err := compileKoreanBanksWithPlan(source, korean, banks, plan, contract.Layouts)
	if err != nil {
		return nil, err
	}
	if err := addBanks(owners, compiled); err != nil {
		return nil, err
	}
	if mode == "preflight" {
		fmt.Printf("FORENSIC MOBILE_PREFLIGHT_BANKS compiled=%d authenticated_source_banks=%d\n", len(compiled), len(banks))
	}

	fontReplacements, err := prepareKoreanMobileFontReplacements(root, archives, plan)
	if err != nil {
		return nil, err
	}
	if len(fontReplacements) > 0 {
		var pa *archive
		for _, candidate := range archives {
			if candidate != nil && candidate.name == "pa" {
				pa = candidate
				break
			}
		}
		if pa == nil {
			return nil, fmt.Errorf("Korean beta %s: pa archive unavailable", mode)
		}
		pa.replacements = append(pa.replacements, fontReplacements...)
	}
	if mode == "preflight" {
		fmt.Printf("FORENSIC MOBILE_PREFLIGHT_FONT replacements=%d\n", len(fontReplacements))
	}

	prepared.executable, err = buildKoreanAlphaExecutable(root, gameDir, plan.Mapping)
	if err != nil {
		return nil, err
	}
	prepared.parameter, err = buildKoreanAlphaSFO(root, gameDir, version)
	if err != nil {
		return nil, err
	}
	if mode == "preflight" {
		fmt.Printf("FORENSIC MOBILE_PREFLIGHT_EXECUTABLE bytes=%d sfo_bytes=%d output_iso_written=false\n", len(prepared.executable), len(prepared.parameter))
	}

	committed = true
	return prepared, nil
}
