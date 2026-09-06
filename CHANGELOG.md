# ZillKoreanPatch Change Log

This file is the durable release-level index for the Korean patch.
Detailed evidence and implementation rationale belong in the linked release note.

## Recording rule from Beta 1 onward

Every material patch change must record, at minimum:

1. **What changed** — affected subsystem, data, or user-visible behavior.
2. **How it was implemented** — concrete implementation or data-path change.
3. **Why it was changed** — defect, runtime evidence, visual comparison, or other reason.
4. **English-patch reference** — the English consumer/layout/storage behavior used as the authority, and why it applies. If a Korean-only divergence is required, record the evidence for that divergence.
5. **Validation** — exact checks actually completed. Static tests, CI, asset-backed validation, emulator/hardware samples, and full runtime QA must not be conflated.
6. **Known limits / pending work** — anything not proven by the completed validation.

Do not describe a candidate, partial test, or CI-only result as full runtime validation.

---

## Beta 1 — 2026-09-06

**Preserved predecessor:** `milestone/U7`

Beta 1 consolidates the Korean reflow-parity work and adopts the user-selected **B font profile** as the release font baseline.

### Font

- Fixed raster geometry remains **10x10**.
- Runtime placement remains **BearingX 1**.
- Runtime glyph advance remains **12**.
- Alpha is now corrected with **gamma 0.60** immediately before 4bpp quantization.
- The deterministic render contract is versioned as `opentype-10px-72dpi-hinting-none-origin-0,-2-alpha-gamma-0.60-round-4bpp-v2`.
- The generated Korean raster catalog is regenerated automatically from that rule.

**Reason:** side-by-side font comparison selected B as the cleaner visual result. Geometry and runtime advances were deliberately left unchanged so the visual-weight adjustment does not silently change layout/reflow contracts.

### Korean dialogue reflow parity

Beta 1 carries the reflow fixes developed after U7, including the English-consumer-aligned treatment of fixed controls versus movable runtime substitutions, bounded player-name substitution handling, source-aware fragment handling, hard post-width checks, regression anchors, and broader final-layout auditing added on the working line.

**English-patch authority:** semantic projection follows the English consumer model: fixed control-flow nodes delimit cases/fragments while movable runtime substitutions remain part of their semantic fragment. Korean reflow is allowed to diverge only where the Korean renderer has a separately proven metric/consumer requirement.

### Release workflow

- The Android Korean A-054 release-candidate workflow is retargeted from the temporary `fix/u6-korean-reflow-parity` working branch to `milestone/Beta1`.
- Generated font output remains workflow-owned and reproducible rather than manually edited.

### Validation at release preparation time

- Korean raster catalog workflow completed successfully after the gamma 0.60 renderer change and auto-committed regenerated output.
- The B profile has a unit-test lock for the gamma value, render-rule identifier, and representative 8-bit-alpha to 4bpp quantization results.
- Existing Korean reflow/consumer regression tests remain part of the branch history.

These checks establish repository/CI consistency. They do **not**, by themselves, claim complete PSP hardware runtime QA.

See [`docs/BETA1_RELEASE_NOTES.md`](docs/BETA1_RELEASE_NOTES.md) for detailed implementation rationale, English-patch parity rules, validation boundaries, and the preserved branch policy.
