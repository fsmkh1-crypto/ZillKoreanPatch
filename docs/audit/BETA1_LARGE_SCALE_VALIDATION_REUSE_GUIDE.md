# Beta1 large-scale validation reuse guide

Date: 2026-09-09

Purpose: preserve the lessons, successful workflow, limits, and reusable evidence from the Beta1 three-party/independent validation and bounded full-corpus residual-quality pass so a future large-scale audit can start from a proven process instead of redesigning the pipeline.

## 1. Final Beta1 reference state

- Repository: `fsmkh1-crypto/ZillKoreanPatch`
- Branch: `milestone/Beta1`
- Frozen source used for the residual-quality pass: `2c29f53c8c949e3b10f27ec2ea3044abb2278bec`
- Residual-quality semantic commit: `f34341dc7d4675066b4d495ee6b84af41e06350d`
- Residual-quality evidence commit: `a57cf323ab7bddc7f0f11ddc4273ae8ccdf5a9db`
- Post-pass checkpoint before this guide: `e9d1e2269017b21cf1c411db0c6cd2adec873e73`
- Pinned English reference: `HK47196/zill@a98d9ce29f361d666ec23da0dcfd351f24537ffd`
- Accepted Korean IDs: 42,016
- Valid contextual review: 42,016 / 42,016 = 100.000%
- `UNREVIEWED=0`
- `CONTEXT_STALE=0`
- `LAYOUT_RECHECK=0`

Fresh post-pass validation:

- General CI run `34315400302`: SUCCESS
- Android full-corpus RC run `34315400270`: SUCCESS
- RC artifact ID: `10089928239`
- RC artifact name: `zill-korean-a054-full-corpus-patcher`
- RC artifact ZIP digest: `sha256:98049c94446200effe01cc87f23671c8a3df63ce9138cfe450d6352b2888c8ff`
- Raw APK SHA-256: `a757b51ae19b245776bf80526c8cd38cb5da6be2640327c6ef2b1a9d73b96983`

Static/build success remains separate from actual ISO/PPSSPP/PSP runtime proof.

## 2. What the independent audit actually established

A reproducible 200-row KEEP sample was reviewed by a reviewer/model different from the original primary reviewer.

- Sample size: 200
- Raw result: 15 `SHOULD_EDIT`, 185 `KEEP`
- Raw observed correction rate: 7.5%

The 7.5% figure must not be presented as a precise estimate of the true corpus defect rate or as statistical proof that the corpus is above/below a universal quality threshold. Its useful meaning was diagnostic: genuine residual issues remained and their patterns could be identified.

The audit should therefore be used as a trigger for bounded systematic cleanup, not as the first step of an infinite loop of `sample -> fail -> resample -> fail -> resample`.

## 3. Recommended role separation for a future three-party validation

For a future high-assurance audit, separate the roles rather than having the same reviewer re-read their own KEEP decisions.

1. Primary reviewer
   - Performs the main translation/copyedit review.
2. Independent reviewer
   - Receives frozen JP + pinned EN + current KO + context.
   - Does not receive the primary reviewer's verdict/reason.
   - Preferably uses a different model family/person.
3. Adjudicator
   - Resolves `SHOULD_EDIT`, disputes, semantic ambiguity, and overcorrection.
   - Applies the project rules and pinned-English baseline.

The value of the 200-row audit came largely from reviewer independence. A full re-read by the same original reviewer can be more expensive while still repeating the same blind spots.

## 4. Blinding rules that worked

When conducting a blind independent sample audit:

- Freeze the Korean source HEAD before sampling.
- Freeze the pinned English reference SHA.
- Withhold the real corpus IDs from the independent reviewer when practical.
- Withhold the review ledger.
- Withhold the pass/fail threshold or cumulative result during the review.
- Do not apply corrections between packets; all packets must refer to the same source snapshot.
- Do not feed adjudication feedback from earlier packets back to the reviewer before the predetermined sample is complete.

Existing sampler:

`tools/korean/sample-beta1-review-audit.py`

The 2026-09-09 audit used:

- sample size 200
- seed `beta1-final-keep-audit-v2`
- second reviewer label `Claude`
- propagated KEEP population 0
- direct KEEP sample 200

The sampler sorts the final chosen rows before packetization. Therefore adjacent 50-row packets can have very different content composition. Do not compare packet-by-packet error rates as though each packet were an independent random sample. Judge the predetermined 200-row sample as one sample.

## 5. Do not infer internal/player-facing status from blank English alone

A blank pinned-English value is not sufficient evidence that a record is internal or player-invisible.

Use consumer/source metadata and actual context. If internal status cannot be proven, keep the row on the player-facing side for conservative review decisions.

This matters because an audit packet can otherwise contain a misleading concentration of event flags, scene labels, or non-display strings.

## 6. Corpus cleanup method that worked best

After the independent audit, the project deliberately did not perform another full 42,016-row human re-read and did not perform a third-model bulk audit.

Instead:

1. Re-adjudicate the independent audit's concrete `SHOULD_EDIT` rows.
2. Build a Pattern Registry from proven residual issue types.
3. Run a full 42,016-ID census for those patterns.
4. Deduplicate by ID so one row receives one primary adjudication pass even if multiple patterns hit it.
5. Inspect JP -> pinned EN -> KO -> context for hits; do not auto-accept regex hits as errors.
6. Perform exactly one Late Addendum census for new high-value S1/S2 patterns found during the first registry pass.
7. Do not recursively open Late Addendum v2/v3. New non-critical polishing patterns after cutoff go to the next milestone/Beta2.
8. Apply all confirmed corrections together only after the review snapshot work is complete.

This bounded procedure produced Batch 060:

- independent-audit corrections: 15
- systematic companion corrections: 245
- total applied: 260

## 7. Terminology normalization rule

For fixed game terms, named concepts, canonical labels, and character-specific fixed expressions, the same Japanese source term should normally map to the same Korean term.

Decision order:

1. Japanese source: semantic authority.
2. Pinned English patch: inspect how the established patch solved the same term/structure and why.
3. Korean: choose the natural canonical Korean form consistent with project terminology.

Do not mechanically merge semantically distinct Japanese forms merely because their Korean strings look similar.

Confirmed example:

- `勇者` -> `용사`
- `竜字将軍` is a different title and remains `용자장군`; pinned English likewise distinguishes it as `Dragon-Sigil General` rather than `hero`.

Also preserve deliberate source behavior. Example: ID 550369 contains hesitant phonetic `せもんいん？`; Korean `세몬인?` was intentionally preserved because pinned English also treats it as a hesitant reference rather than replacing it with the canonical organization label.

Generic grammar and relationship-sensitive honorifics such as `さん` must not be blindly globally normalized.

## 8. Pattern classes that were useful

Useful full-corpus searches included:

- canonical term variants
- untranslated or literal Japanese residue
- known mistransliteration variants
- spacing around substitution tags such as `<value:...>`
- inconsistent fixed labels
- Japanese-style literal collocations
- suspicious emotion/interjection renderings
- same exact JP with divergent KO, followed by context adjudication
- known typo/spacing structures found by independent audit

High-false-positive searches such as line-initial `야,` or generic `짓` + honorific context should remain supporting signals only. A regex hit is not itself an error.

## 9. Important failure modes learned

### 9.1 Infinite statistical gating

Do not keep sampling until a desired percentage appears. The project owner explicitly chose practical Beta1 quality rather than perfect statistical certification.

Use an independent sample to discover whether residual categories exist, then fix demonstrated categories in a bounded way.

### 9.2 Same-reviewer full re-read

A second full re-read by the same original reviewer is not automatically stronger than a smaller independent audit. It can repeat the same blind spots and attention decreases at very large scale.

### 9.3 Blind mechanical terminology replacement

Never substitute globally without morphology/context checks.

A literal `아타쿠시 -> 이 몸` replacement produced broken candidates such as invalid particles during candidate generation. Candidate generation was correctly failed before apply; replacements were then corrected by grammatical form.

Always validate:

- control/substitution tag sequence unchanged
- exact-before source matches
- no malformed particles introduced
- no doubled phrase fragments
- no accidental semantic merging of distinct terms

### 9.4 Repeated CI during language review

Do not run heavy CI every time a handful of strings changes. Keep candidate corrections outside the live translation until the bounded review pass is finished, then apply once and run heavy QA/finalization once.

### 9.5 New framework/tool development becoming the work

For a closing audit, restrict automation to high-leverage support:

- frozen snapshot/census
- Pattern Registry search + ID dedup
- blind packet generation
- result merge/real-ID restoration

Do not create a new review ledger, scoring model, generalized QA framework, or multi-model orchestration framework unless an existing mechanism truly cannot do the job.

### 9.6 GitHub Actions chaining assumption

A push created with `GITHUB_TOKEN` does not necessarily trigger another workflow automatically. Do not assume publisher -> apply chaining occurred; verify the target workflow run actually exists.

### 9.7 Artifact identification

Never reuse an assumed artifact ID from a previous run. Query the actual workflow run's artifacts and record:

- workflow run ID
- artifact ID
- artifact name
- artifact ZIP digest
- raw APK SHA separately

Artifact ZIP SHA and APK SHA are different values.

## 10. Apply/finalize protocol

Before any translation write:

1. Re-fetch the actual remote `milestone/Beta1` HEAD.
2. Preserve legitimate newer commits from parallel work.
3. Never reset to an older expected SHA merely because a handoff listed it.
4. Never force-push.
5. Verify exact-before Korean values against the current source.
6. Use the existing reviewed-copyedit/materializer/apply/finalize pipeline rather than inventing another one.
7. Run Korean QA and engine/storage parity gates.
8. Regenerate review evidence/coverage after semantic changes.
9. Run fresh general CI and fresh Android RC after the final semantic/evidence commits.

Relevant existing components include:

- `tools/korean/materialize-beta1-reviewed-manifest.py`
- existing reviewed-copyedit apply workflow
- existing evidence-finalization workflow
- `.github/workflows/ci.yml`
- `.github/workflows/android-korean-a054-rc.yml`

## 11. Suggested future decision tree

For a future large-scale validation:

### A. Need independent confidence only

Run one frozen, blinded, predetermined sample with a genuinely independent reviewer. Do not alter source during the sample.

### B. Sample finds repeatable residual categories

Run bounded Pattern Registry full-corpus searches for demonstrated categories, plus at most one Late Addendum.

### C. Sample finds many severe semantic defects or a concentrated broken scene/character

Expand only that proven semantic risk scope first. A whole-corpus second human re-read is a last resort, not the default response.

### D. Project owner wants maximum assurance rather than practical release quality

Optionally add a third independent model/person on a predefined risk population. This was explicitly not required for the Beta1 2026-09-09 closeout, so a future project must decide this before work starts rather than adding reviewers indefinitely after seeing results.

## 12. Beta1 closure principle to reuse

The practical closure rule used here was:

- find real residual defect types independently
- normalize fixed terminology consistently
- census the whole corpus for those demonstrated types
- permit exactly one bounded addendum
- adjudicate with JP semantic authority and pinned-English implementation precedent
- apply in one controlled batch
- rerun full QA and a fresh RC
- stop language polishing at the declared cutoff unless a newly proven major semantic/runtime defect requires reopening a local scope

This protocol is intended to prevent both under-review and endless perfection-driven re-review.