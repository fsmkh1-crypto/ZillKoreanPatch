# Claude full QA review — project decision (2026-09-06)

Repository: `fsmkh1-crypto/ZillKoreanPatch`
Branch: `fix/u6-korean-reflow-parity`

This document records the project decision on the independent Claude review of the planned Korean dialogue full-corpus copyedit, terminology normalization, English-patch parity, reflow coverage, control/layout safety, and runtime QA process.

If this document conflicts with an older statement in `docs/KOREAN_DIALOGUE_QA_PROTOCOL.md` or `docs/KOREAN_TRANSLATION_STYLE.md`, the narrower correction in this document applies until the older document is consolidated.

## Verdict

- External review verdict: `PASS WITH CHANGES`
- Project decision: `ACCEPTED — GO AFTER FIXES`
- Mass dialogue copyedit/normalization is blocked until the code/validation blockers below are addressed.

The documentation direction is accepted. The blockers are implementation/validation mismatches, not a rejection of the English-patch-first policy.

## Evidence classes

Use these labels when carrying review findings forward:

- `PROVEN` — directly supported by repository code/tests/data or reproducible measurement.
- `LIKELY` — strongly supported but intent/root-cause inference is not independently encoded by the repository.
- `UNKNOWN` — requires retail assets, runtime evidence, or other evidence not present in repository-only checks.

Do not promote `LIKELY` or `UNKNOWN` to a universal contract without new evidence.

## Accepted blockers

### C-1 — C5/C5-portrait movable-substitution exclusion

Status: `ACCEPTED / PROVEN`

`koreanEnglishDialogueVisualConsumer` currently rejects C5/C5-portrait records whenever `koreanDialogueRuntimeSubstitution` detects a movable substitution. This excludes early Ferme tutorial record `1980005`, which contains `<value:$28>`, from static Korean dialogue reflow.

`$28` is not width-unbounded in the same sense as arbitrary caller data: the layout engine already reserves a worst-case player-name advance and the game contract documents an 8-character / 16-byte player-name maximum. Therefore the project must stop treating every movable value tag as one undifferentiated exclusion class.

Required action:

1. classify substitutions by whether their inline rendered width is proven bounded;
2. admit bounded substitutions such as `$28` to the English-parity source-aware path;
3. keep genuinely unbounded inline substitutions fail-closed/manual until a bound is proven;
4. do not infer safety for `$15/$16/$17/$1A/$1B/$24/$25/$2B` merely because they are movable — each requires a proven measurement/binding contract.

Regression anchors: `1980005`, `560650`, and `280181`.

### C-2 — repository source-aware fragment splitting

Status: `ACCEPTED / PROVEN`

`splitRepositoryReflowFragments` currently treats every tag matched by the generic layout `controlTag` regex as a fixed fragment delimiter. That regex includes `<value:$XX>` tags.

Retail `message.Project` instead keeps `pureMovable` and `callerMovable` substitutions inside a translatable fragment. Repository fallback must mirror that semantic distinction as closely as possible; movable substitutions are not fixed fragment boundaries.

Required action:

- repository fallback must keep the same movable substitution set inside fragments;
- fixed controls remain delimiters;
- source `<line-break>` remains a source-layout hint;
- add focused tests proving value-adjacent text can be measured/reflowed as one fragment.

### C-3 — regression test currently freezes the defect

Status: `ACCEPTED / PROVEN`

The existing C5 test asserts that `560650` remains outside static reflow solely because it contains a movable runtime value. `560650` itself uses `$28`, so the assertion is not a valid future contract once bounded substitutions are distinguished from unbounded substitutions.

Required action:

- replace the blanket dynamic-exclusion assertion;
- require `$28` examples to remain eligible;
- add a separate synthetic/real fixture that proves an actually unbounded inline substitution remains excluded until its bound is established.

### C-4 — `./zill check` is not a Korean data/reflow gate

Status: `ACCEPTED / PROVEN`

`release.Check` performs contributor/English ancillary checks and does not load/run the Korean corpus/reflow chain. Therefore `./zill check` must not be cited as evidence that a Korean dialogue editing batch passed.

Repository Korean gates include, as applicable:

- `go test ./...`
- `go vet ./...`
- `./zill korean-check`
- `./zill korean-font-check`
- `python3 tools/korean/audit-korean-glyph-repertoire.py`
- `python3 tools/korean/qa-layout-drift.py`
- `python3 tools/korean/qa-integrity.py`
- `python3 tools/korean/qa-terminology.py`
- `python3 tools/korean/qa-consistency.py`
- `python3 tools/korean/qa-consistency-triage.py`
- `python3 tools/korean/qa-voice.py`
- `python3 tools/korean/qa-text-sanity.py`

`./zill check` may still be run as a general repository check, but it must be labelled correctly and must never substitute for Korean-specific validation.

### C-5 — Rostorl documentation/data conflict

Status: `ACCEPTED / PROVEN`

The style policy currently identifies `ロストール / Rostorl -> 로스톨` as the intended normalized Korean target, while `translations/terminology/korean-canonical.toml` still contains legacy `로스토올`, and the existing corpus predominantly uses that legacy spelling.

Project decision:

- target normalization remains `로스톨` based on the already approved terminology decision;
- migration is **not complete** until the canonical table and all proven same-entity user-visible occurrences are changed in a dedicated terminology batch and residual variants are audited;
- until that batch lands, do not report repository terminology consistency as complete for Rostorl.

Do not silently change the canonical table alone and then treat thousands of resulting QA mismatches as success. Table + corpus migration + residual audit belong to one reviewed normalization task (possibly split into mechanically linked commits).

## Accepted high-risk gaps

### H-1 — exact English eligible set requires asset-backed projection

Status: `ACCEPTED / PROVEN`

Repository-only `corpus.LoadProject` does not provide authenticated retail tokens required by `message.Project`/English `Reflow`. Therefore an exact English eligible population `E` cannot be proven in repo-only CI.

Rule:

- repository mode may report Korean expected/eligible/excluded sets and exclusion reasons;
- exact `E` and `E-K`/`K-E` parity are `PENDING/UNKNOWN` until an asset-backed run;
- asset-backed parity results must record artifact/bank provenance.

### H-2 / H-3 — excluded population and subset audit

Status: `ACCEPTED`

A residual audit that deliberately mirrors derivation eligibility proves only that subset. `derivation subset PASS != whole-corpus visual PASS`.

Required gate:

- enumerate excluded records by reason and ID;
- separately perform a final whole-corpus visual/layout warning audit;
- do not call subset residual count `0` a whole-corpus reflow pass.

### H-4 — custom glyph capacity

Status: `ACCEPTED`

Large linguistic edits can introduce new Hangul syllables and consume finite reusable two-byte slots. Record before/after glyph repertoire/custom-slot counts for large batches; retain asset-backed slot-allocation failure as a release gate.

## Terminology decision model

`translations/terminology/korean-canonical.toml` is currently a canonical-only two-field table (`japanese`, `korean`). It has no review-state schema. Therefore the external suggestion to add unresolved Ferme directly to that file as `REVIEW` is **not adopted literally**.

Instead:

- unresolved terms live in a review-candidate document/table until decided;
- only approved `KEEP/NORMALIZE` canonical mappings are promoted to `korean-canonical.toml`;
- tooling may later gain an explicit stateful terminology schema, but do not overload the current canonical table with pseudo-state fields without updating the parser and contract.

### Ferme (`フェルム` / `Ferme`)

Status: `REVIEW`

Observed Korean variants include `페름` and `펠름`. English canonical identity is `Ferme`; the correct Korean surface form is not yet proven solely by the Roman spelling and Japanese form.

Rules:

- no global replacement yet;
- remove unsupported `Pelm` wording from project documentation unless clearly labelled as a back-romanization of the Korean variant;
- enumerate all user-visible surfaces and confirm same-entity usage;
- seek stronger provenance/pronunciation evidence before choosing canonical Korean spelling.

## Reflow substitution classes

For QA/reflow purposes, do not use one binary `dynamic` bucket. Record at least:

1. `BOUNDED_INLINE` — inline substitution whose maximum rendered advance is proven by engine/game contract (currently `$28` is the confirmed example).
2. `UNBOUNDED_INLINE` — inline substitution whose maximum rendered advance is not proven in the current path; cannot be declared visually safe by static width alone.
3. `CONTROL_FLOW_OPERAND` — value/control occurrence used by expression/select/predicate grammar rather than rendered inline; visual advance is not treated as ordinary text width.
4. `FIXED_CONTROL` — source-owned fixed controls such as select/count/control topology that must remain fixed and must not be confused with movable anchors.

Classification must come from the actual source projection/token grammar where available; regex presence alone is insufficient for control-flow-vs-inline decisions.

## Validation changes required before mass copyedit

Before full-corpus language editing begins:

- [ ] fix repository movable-fragment splitting;
- [ ] fix C5 bounded-substitution eligibility (`$28` first, no speculative widening);
- [ ] replace defect-preserving regression assertions and add `1980005` anchor;
- [ ] correct Korean validation documentation so `./zill check` is not misrepresented;
- [ ] add/define a final whole-corpus visual/layout audit independent of derivation eligibility;
- [ ] add excluded-population reason/ID reporting;
- [ ] record large-batch glyph repertoire deltas;
- [ ] keep exact English set parity asset-backed/PENDING in repository mode;
- [ ] resolve Rostorl canonical table + corpus migration before terminology normalization is declared complete;
- [ ] keep Ferme as REVIEW until stronger evidence exists.

## Workflow decision

The large pre-flight list should be completed **per batch/baseline**, not mechanically re-read line by line every few minutes of one uninterrupted session. Repeat it when:

- a new batch begins;
- branch/HEAD/baseline changes;
- work resumes after handoff;
- scope/defect class changes materially;
- concurrent changes are detected.

This preserves the safety intent without creating checklist fatigue.

## Go/no-go

Current state: `NO-GO FOR MASS COPYEDIT / GO FOR BLOCKER FIXES AND AUDIT INFRASTRUCTURE`.

Mass copyedit may move to `GO` only after the blocker checklist above is satisfied or each remaining item is explicitly scoped as asset-backed `PENDING/UNKNOWN` with no repository-side unsafe gap hidden by that classification.
