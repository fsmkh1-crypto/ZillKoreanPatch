# Korean Dialogue QA Protocol

This document is the mandatory operating checklist for Korean dialogue review, copy-editing, terminology normalization, layout/reflow fixes, and related runtime QA.

It supplements `AGENTS.md`, `CONTRIBUTING.md`, and `docs/KOREAN_TRANSLATION_STYLE.md`. If rules appear to conflict, the project-wide English-patch-first engine contract in `AGENTS.md` wins.

The purpose of this protocol is not merely to list good practices. It must make it possible to prove what was checked, what was changed, what remains uncertain, and whether the Korean patch still preserves the English patch/engine contract.

---

## 0. Non-negotiable premise

**English patch first.**

Before changing Korean dialogue behavior, layout, storage, controls, or build logic:

1. Find the corresponding English-patch path.
2. Identify why the English patch handles that consumer that way.
3. Preserve the same engine-facing contract in Korean unless Korean/custom-renderer requirements genuinely require a different representation.
4. Document any Korean divergence with evidence.

Do not invent a Korean-only workaround merely because it appears to fix one screen.

For language/terminology work, the English patch is also the primary project reference for proper-name identity/spelling, while the Japanese source is used to recover intended pronunciation. Final Korean spelling follows `docs/KOREAN_TRANSLATION_STYLE.md`.

---

## 1. Mandatory pre-flight checklist

Run this checklist at the start of every Korean dialogue QA/editing work session before modifying files.

- [ ] Confirm the intended working branch and current HEAD.
- [ ] Record the exact baseline SHA from which this session starts.
- [ ] Confirm that protected/base branches are not being modified directly.
- [ ] Confirm there is no unexpected concurrent change already present on the working branch.
- [ ] Read or re-check `AGENTS.md`.
- [ ] Read or re-check `CONTRIBUTING.md` for message/reflow rules.
- [ ] Read or re-check `docs/KOREAN_TRANSLATION_STYLE.md`.
- [ ] Read or re-check this protocol.
- [ ] Identify the exact scope of this session: records, sections, defect class, terminology set, or audit batch.
- [ ] Identify the corresponding English-patch implementation/consumer contract for any engine-facing issue.
- [ ] Record whether English parity is `PASS`, `DIFFERENT-BY-DESIGN`, `MISSING`, or `UNKNOWN`.
- [ ] Confirm that no blind global replacement or mass rewrite will be performed.
- [ ] Establish a before-state for every record that may be edited: file, ID, Japanese, Korean, controls/substitutions, and existing layout.
- [ ] Define the validation gates that must pass before the batch may be accepted.

For the current U6 reflow work, use only `fix/u6-korean-reflow-parity` unless the project owner explicitly changes the branch policy. Do not modify `milestone/U6`, `translation/section001-batch2`, U7, debug, forensic, experiment, or improvement branches as a side effect.

### Pre-flight failure rule

If the working branch, baseline SHA, English contract, or intended scope cannot be established confidently, do not mutate dialogue data yet. Resolve the uncertainty or classify the affected work as `HOLD/UNKNOWN`.

---

## 2. Baseline snapshot and reproducibility

Every editing/audit session must have a reproducible before-state.

Record at minimum:

- repository and branch
- baseline HEAD SHA
- intended comparison base, when applicable
- affected files/sections or candidate-set definition
- known runtime defect IDs/scenes
- test/build command set expected for the batch
- asset/build provenance when an asset-backed test is performed

Do not report a post-edit result without being able to compare it to the recorded baseline.

When the branch moves during the session, re-check whether the movement came from this work or an external/concurrent change before continuing.

---

## 3. Corpus invariants: things dialogue editing must never silently break

For every touched record:

- `id` is immutable.
- Source file/section identity is immutable unless a separate structural migration is explicitly approved.
- `japanese` is immutable and must remain byte-for-byte identical.
- Authenticated Japanese/source-address expectations must remain valid.
- Duplicate-ID behavior must not be changed accidentally; where duplicates exist, the project loader's identity/byte-equivalence contract still applies.
- `korean` contains semantic Korean text, not build-owned wrapping.
- Do not insert `<line-break>` into semantic Korean text.
- Runtime control tokens must remain exactly equivalent to the source contract and in the required order.
- Runtime substitutions such as `<value:$XX>` and printf-style substitutions such as `%s`/`%u`, where present, must be preserved exactly.
- Fixed control literal occupancy must remain valid; do not accidentally remove required literal slots around fixed controls.
- Do not inject raw control bytes or raw Unicode control characters.
- Preserve valid UTF-8 text and do not introduce accidental decomposed/combining-character variants as a side effect of editing.
- TOML structure, keys, and required headers/metadata must remain valid.
- Existing `layout` is not semantic translation. If semantic Korean text changes, the previous generated layout is stale and must be invalidated/removed and regenerated by the current build-owned reflow path.
- Never manually keep an old layout merely because it visually resembles the edited sentence.
- Never manually edit ordinary generated layout to conceal a reflow defect.

A copy-edit that violates any invariant is a failed edit even if the Korean sentence reads better.

### Fail-closed rule

Any unexplained ID/Japanese/control/substitution/layout-contract difference discovered after an edit blocks acceptance of that batch until explained, fixed, or reverted.

---

## 4. Review target: full copy-edit, minimal rewriting

The default scope is **full-corpus proofreading plus minimal polishing**, not wholesale retranslation.

Review every candidate for the following classes.

### A. Spacing and orthography

Check for:

- incorrect Korean spacing
- particles/endings attached or separated unnaturally
- inconsistent spacing around numbers, symbols, names, and UI terms
- obvious spelling/orthographic errors

### B. Commas and punctuation

Check for:

- unnecessary commas copied from English/Japanese pause structure
- repeated commas that make Korean dialogue stilted
- commas inserted between tightly connected Korean clauses where natural speech does not need them
- inconsistent sentence-final punctuation
- malformed ellipses, question/exclamation combinations, or punctuation spacing

Do **not** remove commas globally. Keep punctuation when it prevents ambiguity, marks lists/vocatives/interjections, or is part of deliberate character rhythm.

### C. Natural Korean dialogue

Check for:

- literal Japanese/English word order
- unnatural particles or connective endings
- excessive nominalization
- redundant subjects/objects already obvious from context
- repeated wording that is not intentional characterization
- awkward game-dialogue phrasing
- sentences that are grammatical but clearly not how a Korean speaker would normally say the line

### D. Register and characterization

Preserve:

- Japanese plain/casual vs polite register unless context establishes an exception
- honorifics and titles
- archaic or military speech
- deliberate stutter, hesitation, emotional repetition, catchphrases, and strong character-specific verbal tics

Do not smooth characterization merely to make every speaker sound uniform.

### E. Terminology and proper names

Check for:

- inconsistent spellings of character/place/item/skill names
- alternate transliterations for the same proper noun
- terminology drift between dialogue and UI/system text
- mechanical Japanese long-vowel rendering
- Japanese epenthetic vowels carried into Korean unnecessarily
- mechanical sokuon or kana-to-Hangul conversion

Follow the source hierarchy and transliteration rules in `docs/KOREAN_TRANSLATION_STYLE.md`.

Confirmed normalization rule:

- `ロストール / Rostorl -> 로스톨`
- `로스토올` is not the canonical Korean target.

Known audit candidate from current runtime QA: `フェルム / Ferme` has been observed with inconsistent Korean spellings and must be reviewed as a terminology-consistency case rather than globally replaced on intuition.

---

## 5. Required context packet before changing a non-trivial line

For anything beyond an obvious typo/spacing fix, inspect as much of the following as is available:

1. record ID and section
2. Japanese source
3. current Korean text
4. English patch translation/reference
5. adjacent dialogue lines and scene flow
6. speaker/character register when identifiable
7. control tokens and runtime substitutions
8. existing/generated layout state
9. consumer type and width/storage contract when layout is involved
10. repeated occurrences of the same name/term when terminology is involved

Classify the candidate before editing:

- `COPYEDIT` — spacing, punctuation, obvious grammar/wording correction
- `CONTEXTUAL-REWRITE` — natural Korean requires limited contextual rephrasing
- `TERMINOLOGY` — name/term consistency change
- `LAYOUT-DEFECT` — semantic Korean may be fine; build/reflow/display path is wrong
- `ENGINE-CONTRACT` — consumer/storage/control/materialization behavior requires a code/path fix
- `HOLD` — meaning, name, consumer contract, or context is not sufficiently proven

When uncertain, hold the row rather than guessing.

---

## 6. Proper-name/terminology normalization protocol

Do not perform corpus-wide name replacement until a canonical mapping has been reviewed.

For every candidate name/term:

- [ ] Enumerate Japanese form(s).
- [ ] Enumerate English patch canonical spelling(s).
- [ ] Enumerate all current Korean variant(s).
- [ ] Check whether variants refer to the same entity/term.
- [ ] Apply the long-vowel, epenthetic-vowel, sokuon, r-row/final-consonant, and consonant-identity rules in `KOREAN_TRANSLATION_STYLE.md`.
- [ ] Classify as `KEEP`, `NORMALIZE`, or `REVIEW`.
- [ ] Record evidence for the proposed canonical Korean spelling.
- [ ] Enumerate all affected files/records before replacement.
- [ ] Confirm the replacement cannot alter unrelated substrings or control syntax.
- [ ] After replacement, re-scan the corpus for residual variants.

A terminology normalization is incomplete if old variants remain unintentionally on another user-visible surface.

### Global replacement exception

A global/scripted replacement may be used only after the mapping is approved and all occurrences have been enumerated/proven to represent the same term. The resulting diff must still be reviewed record by record or by an equivalent deterministic audit. Natural-language wording/punctuation changes are never covered by this exception.

---

## 7. Line-break and reflow protocol

### Principle

Do not fix ordinary dialogue wrapping by hand inside semantic Korean text.

The English patch keeps ordinary semantic dialogue unbroken and lets the build/reflow path derive display layout. Korean must preserve that contract unless a consumer is proven to require fixed visual layout.

If a consumer requires deliberately authored/fixed visual layout, represent the Korean exception through the project-approved layout path, not by contaminating semantic Korean with ordinary wrapping tokens.

### For every observed no-wrap/overflow defect

- [ ] Identify the exact record ID(s).
- [ ] Confirm the current Korean semantic text and `layout` state.
- [ ] Check how the corresponding English record reaches its display consumer.
- [ ] Determine the English consumer classification and width/storage rule.
- [ ] Check whether Korean is in the same reflow population.
- [ ] Check runtime substitutions such as `<value:$XX>` that may affect width or eligibility.
- [ ] Determine whether the defect is semantic text, stale/missing layout, consumer classification, projection/materialization, or runtime-only behavior.
- [ ] Fix the missing English-parity contract before adding Korean-specific heuristics.
- [ ] Regenerate derived Korean layout rather than manually authoring ordinary dialogue breaks.
- [ ] Add a regression test for every confirmed systemic escape hatch.

### Mandatory regression anchor: early Ferme/Pelm tutorial event

The early `msgsec198` Ferme/Pelm encounter is a required regression area because runtime QA reported dialogue appearing without expected wrapping.

At minimum, verify record `1980005`, including its runtime `<value:$28>` substitution, against the English consumer/reflow path. Do not consider the reflow audit complete while this case is unexplained.

This record is an anchor, not a one-off exception: use it to find all records that share the same defect class.

---

## 8. Reflow coverage must be proven as a set comparison

Do not claim that reflow was checked merely because representative records pass.

For each relevant consumer/class, define and compare:

- `E` = records eligible/handled by the English patch's established reflow/consumer contract
- `K` = Korean records expected to preserve that same contract
- `MISSING = E - K`
- `EXTRA = K - E`

Every member of `MISSING` or `EXTRA` must be explained as one of:

- `PASS` after correcting classification/data
- `DIFFERENT-BY-DESIGN` with written evidence
- `MISSING` defect requiring correction
- `UNKNOWN/HOLD`

A reflow coverage audit cannot be called complete while unexplained `MISSING`, `EXTRA`, or runtime-substitution exclusions remain.

Where retail token projection/materialization is available, prefer exact consumer/materialized evidence over source-text width guesses. Where only source-aware fallback is possible, label that limitation explicitly.

---

## 9. Candidate discovery rules

Automation may **find candidates** but must not blindly rewrite the corpus.

Permitted automated discovery includes:

- suspicious Korean spacing patterns
- punctuation/comma density
- duplicate or inconsistent proper-name spellings
- repeated phrases
- records with unusually long unbroken semantic text
- English-reflow-eligible records lacking expected Korean derived layout
- runtime-substitution records that are excluded from or fail reflow
- stale layout associated with changed semantic text
- control/substitution mismatches
- residual terminology variants after an approved normalization

Automated edits are prohibited unless the transformation is mechanically proven safe and separately reviewed. Natural-language punctuation and wording changes require contextual review.

Candidate scanners must not be treated as proof that non-candidates are linguistically correct. They are recall aids, not substitutes for contextual full-corpus review.

---

## 10. Severity and stop conditions

Classify defects so cosmetic work cannot hide engine-breaking risk.

- `BLOCKER` — crash/freeze, corrupted control flow, broken runtime substitution, invalid materialization/storage contract, unexplained engine-contract divergence
- `HIGH` — missing/incorrect reflow causing unreadable overflow/clipping, systematic layout exclusion, major semantic mistranslation
- `MEDIUM` — inconsistent terminology/proper name, clear unnatural translation, register error
- `LOW` — spacing, comma, punctuation, minor phrasing polish

### Stop the current batch and investigate/revert when

- an invariant changes unexpectedly,
- a repository/test gate fails because of the batch,
- an engine-facing difference from English cannot be explained,
- unrelated files/records appear in the diff,
- a semantic edit retains stale layout,
- a runtime test shows a new crash/freeze/control/substitution/display regression,
- concurrent branch movement makes the baseline ambiguous.

Do not continue accumulating cosmetic edits on top of an unresolved `BLOCKER` introduced or exposed by the current batch.

---

## 11. Batch and commit discipline

Do not perform one giant opaque rewrite.

Work in reviewable section-sized or defect-class-sized batches. For each batch, maintain a review log containing at least:

| Field | Required content |
| --- | --- |
| File/ID | Message location and record ID |
| Before | Existing Korean text |
| After | Proposed/accepted Korean text |
| Type | COPYEDIT / CONTEXTUAL-REWRITE / TERMINOLOGY / LAYOUT-DEFECT / ENGINE-CONTRACT |
| Severity | BLOCKER / HIGH / MEDIUM / LOW |
| Reason | Why the change is needed |
| Source checked | Japanese / English / adjacent context as applicable |
| English parity | PASS / DIFFERENT-BY-DESIGN / MISSING / UNKNOWN |
| Control safe | PASS/FAIL |
| Layout action | unchanged / invalidated / regenerated |
| Tests | focused/static/build gates run |
| Runtime QA | pending / pass / fail / not applicable |

Keep commits small enough that unintended semantic or structural changes can be audited from the diff.

Where practical:

- keep pure dialogue copy-edit/terminology data changes separate from engine/reflow code changes,
- include the regression test with the engine/reflow fix it proves,
- do not mix unrelated renderer/build-system work into a dialogue QA commit,
- make commit messages describe the defect class or reviewed batch rather than claiming broader coverage than was actually performed.

---

## 12. Validation gate after every editing batch

A batch is not complete until the applicable checks pass.

### Static corpus checks

- [ ] ID set unchanged.
- [ ] Japanese source unchanged byte-for-byte.
- [ ] Source/authentication expectations still pass.
- [ ] TOML parses successfully.
- [ ] No semantic Korean `<line-break>` insertion.
- [ ] Runtime control sequence preserved.
- [ ] Runtime substitutions preserved.
- [ ] Fixed-control literal occupancy preserved.
- [ ] No stale layout remains for edited semantic text.
- [ ] No unintended terminology variants were introduced.

### Reflow/layout checks

- [ ] Re-derive build-owned Korean layouts using the current U6 parity path.
- [ ] Check English/Korean reflow-population set parity for the affected consumer.
- [ ] Check residual overflow/no-wrap candidates.
- [ ] Check runtime-substitution cases separately where width is not statically knowable.
- [ ] Run focused regression tests for any defect fixed in the batch.
- [ ] Confirm any fixed-layout exception is consumer-scoped and evidenced.

### Repository checks

Run the applicable project gates, including:

- `go test ./...`
- `go vet ./...`
- `./zill check`
- focused Korean reflow/layout tests

Remember: `./zill check` is not a substitute for asset-backed build validation or runtime QA.

### Asset-backed/runtime checks

Before release-quality acceptance:

- [ ] Build/materialize with the relevant retail assets where required.
- [ ] Record the exact code HEAD/build source used for the runtime artifact.
- [ ] Validate final encoded/materialized bytes and consumer limits where applicable.
- [ ] Test representative runtime screens in PPSSPP/real hardware as appropriate.
- [ ] Treat one successful runtime attempt as limited evidence only.
- [ ] Record any freeze, overflow, clipping, missing wrap, malformed substitution, or control-flow anomaly as a failure requiring investigation.

---

## 13. Runtime QA evidence log

A runtime QA statement must be tied to a reproducible artifact and scene.

Record at minimum:

| Field | Required content |
| --- | --- |
| Code HEAD | SHA used to build |
| Build/artifact | identifiable build or artifact reference |
| Emulator/hardware | PPSSPP version/device or real hardware, when relevant |
| Record/scene | IDs and how the scene is reached |
| Expected | intended wrapping/rendering/behavior |
| Observed | actual result |
| Result | PASS / FAIL / NOT-REACHED |
| Evidence | screenshot/video/log/reference when available |

A screenshot of one correct screen proves only that screen/build/run. It does not prove all records or all execution paths.

A runtime failure is stronger evidence than a single runtime success. Investigate failures rather than averaging them away with later successful runs.

---

## 14. End-of-batch diff audit

Before committing or handing off:

- [ ] Compare final state against the recorded baseline SHA.
- [ ] Diff contains only intended files and records.
- [ ] No unrelated code, branch, translation batch, or renderer changes slipped in.
- [ ] Japanese/IDs/control contracts remain intact.
- [ ] Every semantic edit has a stated reason.
- [ ] Every terminology change has an approved mapping/evidence.
- [ ] Every engine-facing change includes English parity evidence.
- [ ] Touched semantic rows do not retain stale generated layout.
- [ ] Known regression anchors affected by the change pass their focused tests.
- [ ] Remaining `HOLD`/UNKNOWN items are explicitly listed rather than silently guessed.
- [ ] Test/build/runtime claims match what was actually executed.

Every engine-facing change summary must include:

- `English patch parity checked: YES / N/A`
- `English reference/consumer contract:`
- `Korean divergence, if any:`
- `Evidence for divergence:`

---

## 15. Definition of “full review”

Do not use the phrase **“full runtime review”** or **“all dialogue verified in game”** unless every applicable record/screen has actually been exercised at runtime.

Use precise labels:

- **static full-corpus audit** — all corpus records were checked by static/automated rules
- **contextual full-corpus copy-edit** — all Korean records were contextually reviewed for language quality
- **terminology full-corpus audit** — all defined proper-name/term candidates and user-visible occurrences were checked against the approved mapping
- **reflow coverage audit** — all eligible records were compared as explicit English/Korean consumer/reflow sets and all differences were classified
- **runtime sampled QA** — representative scenes were tested in game
- **runtime full-path QA** — only when the defined runtime path/coverage was actually completed

For every “full” claim, preserve the denominator/coverage definition. A countless statement such as “all relevant cases checked” is insufficient when the relevant set can be enumerated.

This distinction exists to prevent a static audit from being mistaken for exhaustive in-game verification.

---

## 16. Session handoff / interruption protocol

Before stopping a long-running QA/editing session, especially when moving to a new chat/agent or when context is becoming difficult to maintain, write a handoff that contains:

- branch and current HEAD
- session baseline SHA
- completed scope and exact definition of what “completed” means
- commits created during the session
- tests/builds/runtime QA actually executed and results
- unresolved `BLOCKER/HIGH/HOLD/UNKNOWN` items
- next exact record/section/defect class to resume from
- known regression anchors
- whether the working tree/branch is in a safe committed state
- any temporary hypothesis that must **not** be mistaken for a proven root cause

Do not leave the next worker to infer completion from commit names alone.

---

## 17. Session start record template

At the beginning of each work session, record a compact status block in the work log/handoff:

```text
Korean Dialogue QA Pre-flight
Branch:
Baseline HEAD:
Current HEAD:
Scope:
AGENTS.md checked: YES
CONTRIBUTING.md checked: YES
KOREAN_TRANSLATION_STYLE.md checked: YES
KOREAN_DIALOGUE_QA_PROTOCOL.md checked: YES
English patch parity target:
Consumer/contract:
Candidate/reflow coverage set definition:
Known regression anchors:
No blind global edits: CONFIRMED
Before-state captured: YES/PENDING
Required validation gates:
```

Do not begin dialogue mutation until this pre-flight is complete.

---

## 18. Acceptance rule

A batch is accepted only when all of the following are true:

1. intended language/terminology/layout objective is satisfied,
2. corpus invariants pass,
3. English-patch parity is proven or documented as a justified Korean divergence,
4. required static/repository/reflow checks pass,
5. asset-backed/runtime checks required by the defect class have been performed or are explicitly marked pending,
6. the final diff is limited to intended scope,
7. unresolved items are visible as `HOLD/UNKNOWN/PENDING`, not silently treated as PASS.

**No unexplained engine-contract divergence, stale layout, control mismatch, or runtime regression may be hidden by a successful copy-edit or by a single successful runtime screen.**