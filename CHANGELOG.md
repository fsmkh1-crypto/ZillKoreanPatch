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

## Beta1 contextual copyedit 006 — 2026-09-06 (Beta1 incomplete)

- Applied47 reviewed JP/EN/KO corrections across38 overlay files: collapsed spacing47,
  contextual punctuation12, wrong particles2, source-confirmed Dyneskal typo1
  (categories overlap). Exact before/after/source evidence is in manifest006.
- Reused the English-consumer-aligned semantic-only apply queue; JP, controls and B font
  are preserved. Layout synchronization found zero drift; custom glyphs remain1308.
- Semantic commit `26473dfd2665a09ffc49d70d99ad034d53c635c7`; queue workflow
  `34064093896` SUCCESS, including Korean/glyph/data and English storage gates.
- Prior section001 and term migration are now applied; the older checkpoint below is
  historical, not current status. Contextual002–006 are applied. Full corpus review,
  final reflow audits/new APK and runtime evidence remain incomplete.
- See `docs/BETA1_ASTRA_RESUME_NOW.md` for current authoritative continuation.

## Beta1 finalization checkpoint — 2026-09-06 (incomplete)

- Independently verified starting SHA `34fc3e23c9c2350c8d27eb71f002e6fab23e4bdb`, U7 ancestry (43 commits ahead, zero behind), B renderer/catalog identity, and successful pre-copyedit Android run `34031107941`.
- Added exact source/accepted reconciliation: 43,116 source IDs = 42,016 accepted + 1,100 missing overlays, classified individually. Eight Latin-only event titles remain REVIEW. Historical filter totals overlap 57 already accepted rows, explaining why their sum is not the missing-overlay count.
- Contextually reviewed section 001 IDs 10000–10175 (176 records). Preserved 29 proposed corrections with Japanese/English/before/after evidence. **No semantic Korean edits applied and no layouts regenerated.** This is not an accepted copyedit batch.
- Updated the obsolete Ferme REVIEW instruction to the user-approved target 페름; corpus/table migration remains pending. Clarified all four required Beta1 goals in release notes.
- English reference: upstream `internal/message/projection.go` (blob `5a86124c4b5330ee71f9e443cfde921e0b4ef0ae`) separates editable fragments and fixed controls and keeps movable substitutions in their source fragment. No engine behavior changed.
- Validation: 47 existing Python tests passed; baseline Korean integrity, glyph and layout-drift checks passed. Local Go test setup could not load uncached dependencies with network unavailable; this is an environment blocker, not a PASS. Baseline CI static dialogue scope is 22,137 rows with zero overflow and 47 runtime-width PENDING rows. See handoff for commands and checkpoint CI status.
- Remaining: apply/validate the proposed batch, review all other accepted records, normalize terms across surfaces, resolve title REVIEW items, run final audits/new APK, then safe U-series cleanup. No Beta2 work.

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
