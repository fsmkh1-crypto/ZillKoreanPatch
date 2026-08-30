# A-068 — English-first layout/font metrics parity

Status: PASS (static contract gate added)

## Question
Does the Korean renderer width model remain compatible with the upstream English patcher's layout metrics when the mobile Korean path can allocate any installed double-byte renderer key and then rewrites mapped glyph metrics during full repack?

## English baseline
`release/font/metrics.toml` is the layout engine's authoritative width table. Engine advance ceilings in `internal/layout/rules.go` are evaluated against those metrics.

## Korean difference by design
The mobile Korean planner uses `font.DoubleByteKeys()` because the full atlas/PAF repack is not constrained by existing retail cell geometry. `FullRepackAuthenticatedRetailFont` rewrites mapped Korean glyph metrics to `KoreanTargetAdvance` while preserving renderer keys/BST identity.

That difference is valid only if the English layout metric assigned to every eligible double-byte key equals the Korean target advance. Otherwise a line can be accepted/rejected using one width model and rendered using another.

## Gate
`tools/korean/audit-korean-layout-font-metrics.py` now fails closed unless:

- `KoreanRasterWidth`, `KoreanRasterHeight`, and `KoreanTargetAdvance` are positive;
- Korean raster width does not exceed Korean target advance;
- every valid double-byte renderer key in the English metrics table has exactly `KoreanTargetAdvance`;
- the mobile planner still allocates from `font.DoubleByteKeys()` through `BuildPlan`;
- the mobile font path still performs full repack and semantic post-verification.

The dedicated workflow `.github/workflows/english-first-font-metrics.yml` runs this gate whenever the relevant English metrics, Korean font implementation, mobile planner, or gate itself changes.

## Classification
- English layout metric table: ENGINE-INVARIANT baseline.
- Korean raster/metric rewrite: DIFFERENT-BY-DESIGN.
- Width-model equality across every eligible mapped slot: required ENGINE-INVARIANT parity.

This closes the static model-drift class. It does not by itself prove that every authored Korean line is visually ideal; visual QA remains separate from the engine advance contract.
