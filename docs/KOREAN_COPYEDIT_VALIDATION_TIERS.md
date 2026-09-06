# Korean copyedit validation tiers

Date: 2026-09-06

Purpose: keep mass Korean copyediting safe without repeatedly paying for checks whose inputs have not changed.

The rule is **dependency-based reuse**, not “passed once, therefore never run again.” A previous PASS may be reused only while every input relevant to that gate is unchanged.

## Baseline evidence

Before the first semantic copyedit batch, Beta1 checkpoint `43e5a09c81b498724357b223bb6f7805d2a011b3` passed:

- CI run `34033702959`: English-first audits, `go test ./...`, `go vet ./...`, Python tests, layout drift, `./zill korean-check`, `./zill korean-font-check`, and the separate general `./zill check`.
- Korean data CI run `34033702971`: corpus integrity, glyph, layout, terminology, consistency, voice and text/data gates.

The later documentation-only commits up to this policy do not invalidate code/font/parser invariants.

## Tier A — ALWAYS for every accepted semantic-text batch

These gates depend directly on changed translation text and therefore cannot be inherited from a previous batch.

1. Exact proposal/manifest application check: every changed record must match its reviewed `before` value before edit and its approved `after` value after edit.
2. Japanese/source field, message ID and runtime-control/substitution topology unchanged unless the batch explicitly exists to change them.
3. TOML parse / strict Korean corpus load.
4. Persisted-layout lexical drift check. If semantic text changed under an authored/persisted layout, synchronize only through the repository-approved drift mechanism or rederive layout through the appropriate English-parity consumer path.
5. New-Hangul / glyph-coverage delta scan. No unhandled renderer glyph may enter the corpus.
6. Changed-section integrity / terminology / consistency / voice / text-sanity checks applicable to the edited records.
7. Consumer/storage/layout check for the changed records. Text length and punctuation can change encoded byte counts even when controls do not change.
8. Record the exact changed IDs, category counts, layout count, glyph delta and remaining review count in the batch audit.

## Tier B — CHANGE-TRIGGERED

Run these only when their dependency changed.

### Go tests and vet

Run `go test ./...` and `go vet ./...` when any of the following changed:

- `*.go`
- Go module/dependency files
- parser/projection/reflow/consumer/storage/renderer logic
- generated Go code or tests

For a pure TOML copyedit batch with no Go/config/engine change, reuse the most recent green Go baseline and do not duplicate a full Go suite locally merely because another text batch was accepted.

### Full Korean font check / raster regeneration

Run `./zill korean-font-check` and regenerate the raster catalog when:

- new Hangul/renderer glyphs are introduced;
- Korean font parameters, renderer aliases, slot mapping, rasterizer or font source change;
- the lightweight glyph-delta scan cannot prove the existing catalog covers the changed text.

If a pure copyedit introduces **zero new Hangul/glyphs**, retain the prior font/raster proof and do not regenerate the 1,308-glyph catalog.

### English-first implementation parity audit

Re-run structural English-patch parity work when:

- consumer/storage/reflow/layout/parser/compiler behavior changes;
- a new consumer class or previously unknown failure class is discovered;
- a Korean-specific implementation divergence is proposed.

Do not repeat the full parity investigation for punctuation/naturalness-only edits that use already-approved semantics/layout machinery.

### Full repository-wide terminology sweep

Run a full sweep when canonical terminology tables/rules change or a batch intentionally changes a canonical term. Otherwise run the changed-section/changed-ID terminology gate and keep the last full-corpus baseline.

## Tier C — CHECKPOINT / RELEASE ONLY

These are expensive integration proofs and should not be repeated after every similar text batch.

Run at a deliberate checkpoint (for example after a group of sections), after a structural implementation change, or before publishing a new Beta/APK:

- complete `go test ./...` + `go vet ./...` if not already triggered by code changes;
- full 42,016-record accepted-corpus census;
- full Korean English-consumer storage/visual census;
- full static dialogue-reflow census and residual-overflow assertion;
- full warning/runtime-pending census;
- Android RC / patcher native build / debug APK / embedded payload / SHA verification;
- release artifact upload and final manifest.

Android RC is **not** a per-copyedit-batch gate.

## Invalidation matrix

| Change type | Always | Go suite | Full font | English parity | Full corpus checkpoint | Android RC |
|---|---|---|---|---|---|---|
| punctuation/spacing/naturalness only, no new glyph | yes | reuse | reuse | reuse | periodic | no |
| translation wording with no new glyph/control | yes | reuse | reuse | reuse unless consumer failure appears | periodic | no |
| new Hangul/glyph | yes | reuse if no Go | run | reuse | checkpoint | no |
| persisted-layout structural change | yes | targeted layout + relevant tests | as glyph delta requires | run if behavior/rule changes | run | usually no |
| control/substitution change | yes | run relevant/full | as glyph delta requires | run | run | before release |
| Go parser/reflow/consumer/compiler change | yes | run | as relevant | run | run | before release |
| font/raster/slot mapping change | yes | relevant tests | run | compare English font/release rationale | run | before release |
| release/APK preparation | accumulated batch audit | current | current | current | run | run |

## First-batch calibration rule

Section 001 is the first actual Beta1 semantic copyedit batch. After it is applied, run enough remote CI to prove that the Tier-A path and layout synchronization procedure are valid end-to-end once. If it passes without exposing a new failure class, later batches of the **same risk class** may omit duplicate full Go/font/parity work according to this document.

If a later batch exposes a new failure class, invalidate the relevant reused proof, fix it English-patch-first, and promote the new targeted regression check into Tier A or Tier B as appropriate.
