# A-068 — English-first layout/font metrics parity

Status: MISSING → PASS

## Question
Does the Korean renderer width model remain compatible with the upstream English patcher's visual contracts when the mobile Korean path can allocate any installed double-byte renderer key and then rewrites mapped glyph metrics during full repack?

## English baseline
`release/font/metrics.toml` is the authoritative width table for stock text. The English `Reflow` path applies engine advance ceilings from `internal/layout/rules.go`. Most over-width cases are authoring warnings, but character profiles are release-blocking: a projected profile line may not exceed `profileAdvance`, and a profile fragment may not exceed `profileMaxLines`.

## Korean difference by design
The mobile Korean planner uses `font.DoubleByteKeys()` because full atlas/PAF repack is not constrained by existing retail cell geometry. `FullRepackAuthenticatedRetailFont` keeps the selected renderer key/BST identity but rewrites every mapped Korean glyph to `KoreanTargetAdvance`.

The first version of this audit incorrectly assumed every eligible double-byte English metric must already equal the Korean target advance. The dedicated CI immediately disproved that assumption: six legitimate double-byte punctuation keys have nominal English advances smaller than 12 (`0x5c81`, `0x6381`, `0x6681`, `0x6781`, `0x6881`, `0x7e81`). Rejecting those keys would copy an English *glyph-specific policy* instead of preserving the actual engine contract.

The correct model is therefore:

- unmapped stock characters use the upstream English metric table;
- mapped Korean custom runes use the metric actually installed by the Korean PAF transform, `zillfont.KoreanTargetAdvance`;
- upstream English release-blocking visual conditions are evaluated with that mixed, produced-renderer model;
- English warning-only visual ceilings remain warnings/QA concerns and are not silently promoted to Korean build failures.

## Implementation
`internal/layout/validate_korean_visual.go` adds `ValidateKoreanEnglishVisualContracts` and a Korean-aware renderer measurement path. For mapped custom runes it uses `KoreanTargetAdvance`; for stock text it uses the same English glyph metrics. It mirrors the English hard character-profile contracts: `profileAdvance` and `profileMaxLines`.

`ValidateKoreanEnglishConsumerContracts` now invokes this visual validator after storage contracts pass, so desktop, mobile ISO, and mobile preflight all inherit it through their existing English-first release gate.

## Static gate
`tools/korean/audit-korean-layout-font-metrics.py` now fails closed unless:

- Korean raster geometry and target advance remain internally valid;
- the mobile planner still allocates through `font.DoubleByteKeys()` and `BuildPlan`;
- the mobile font path still performs full repack and semantic verification;
- the Korean visual validator explicitly uses `zillfont.KoreanTargetAdvance` for mapped runes;
- the English hard profile limits remain present;
- the release-blocking English consumer gate still calls the Korean visual validator.

The gate reports, but does not reject, nominal double-byte English widths that differ from the mapped Korean advance because those differences are intentional and are resolved by the produced PAF metric override.

## Classification
- English stock metric table and hard profile limits: ENGINE-INVARIANT baseline.
- Korean mapped-rune metric rewrite: DIFFERENT-BY-DESIGN.
- Measuring mapped Korean runes with the produced PAF advance before applying English hard limits: required ENGINE-INVARIANT parity.

This closes the deterministic width-model drift for upstream release-blocking visual contracts. Warning-only visual authoring ceilings and runtime-dependent substitutions remain QA/reporting concerns, not newly invented build blockers.
