# Beta1 Dense Review Pipeline v3 Checkpoint

Date: 2026-09-07

This document supersedes the v2 review-ledger checkpoint for language-review workflow decisions. Beta2 remains out of scope.

## Current remote baseline

At Batch 018 completion, the legitimate `milestone/Beta1` head had advanced through the semantic edit application, review-basis registration, and dense-scope coverage refresh. The post-018 remote head observed before this checkpoint update was:

`7a4ee47b6591c18c73ca89e756008c276f472089`

Never reset legitimate newer commits backward; fetch current remote HEAD before any mutation.

Pinned English authority:

`HK47196/zill@a98d9ce29f361d666ec23da0dcfd351f24537ffd`

## Authoritative coverage after Batch 018

Generated source: `docs/audit/beta1-review-coverage.md`

- accepted Korean IDs: 42,016
- valid contextual review: 487
  - historical full-read: 176
  - dense scope full-read: 11
  - manifest edit: 300
  - propagated: 0
- CONTEXT_STALE: 0
- LAYOUT_RECHECK: 0
- contextual UNREVIEWED: 41,529
- RUNTIME_PENDING full population: 47

The 467 pre-v3 review records were migrated by reconstructing their real historical semantic commits. No v2 hash was reinterpreted as a v3 hash.

## v3 invalidation model

Language:

`SHA256(JP + NUL + pinned EN + NUL + KO)`

A mismatch demotes only that ID to `CONTEXT_STALE`.

Structure:

`SHA256(layout + NUL + physical_consumer_signature)`

A mismatch sets `LAYOUT_RECHECK`; unchanged language coverage stays valid. Engine-derived consumer/runtime metadata remains in each ledger row but is orthogonal to language equivalence.

Dense scopes are also stale-checked ID by ID by reconstructing `reviewed_commit`. One changed row never invalidates the rest of a scope.

## Workload measurement

Generated source: `docs/audit/beta1-review-distribution.md`

- rows with non-`<end>` controls: 5,021
- rows with sentence punctuation: 31,115
- conservative rapid-scan candidates: 10,407

Rapid-scan candidate is only workload triage:

- visible KO <= 20
- visible JP <= 20
- no non-`<end>` controls
- no sentence punctuation

It grants zero contextual coverage by itself. `FIXED_BUFFER` is not used as a shortcut.

Visible KO length bins:

- 1-10: 9,786
- 11-20: 8,873
- 21-40: 11,209
- 41-80: 8,564
- 81-160: 2,985
- 161+: 599

## Language propagation candidates

Candidate equivalence is now exactly:

1. exact Japanese
2. exact pinned English
3. exact Korean

Any English mismatch forbids propagation without exception.

Current candidate census:

- unique language signatures: 35,725
- duplicate groups: 2,667
- IDs inside duplicate groups: 8,958
- potential extra IDs after one representative: 6,291

These remain candidate-only and contribute zero propagated coverage until explicit evidence is recorded.

KEEP and EDIT have the same linguistic equivalence signature. KEEP is the harder-to-detect failure mode, so group propagation requires every member's surrounding context to be displayed and propagated KEEP must be over-sampled in second-pass QA. EDIT still passes exact-before manifest and heavy edit gates.

## Deterministic dense review packets

Tool:

`tools/korean/build-beta1-review-packet.py`

Scope evidence:

`docs/audit/review/scope-*.json`

Policy/schema:

`docs/audit/review/README.md`

The actual packet bytes are deterministic UTF-8/LF with fixed field order and numeric ID order. `packet_sha256` is verified by CI by regenerating the packet from `(reviewed_commit, pinned English SHA, exact ID set, source files)`.

Scope KEEP population is reconstructed as:

`ids - edit_ids - exclusions`

No aggregate-only KEEP declaration is allowed.

## KEEP / EDIT CI split

### KEEP-only dense scope

Workflow: `.github/workflows/beta1-review-scope.yml`

- regenerate and verify packet hash
- reconstruct ID-level language basis
- rebuild ledger
- reuse cached structural consumer/runtime evidence
- no heavy Go/storage rerun merely for unchanged translations

The lightweight path has been run successfully after cached structural evidence was created.

### EDIT

Workflow: `.github/workflows/beta1-reviewed-copyedit-queue.yml`

Still requires:

- exact before-values
- approved manifest
- layout invalidation/drift postconditions
- glyph/font/integrity/terminology/consistency/text-sanity gates
- pinned-English consumer/storage + effective-layout contract
- semantic commit
- rebase against legitimate remote HEAD
- final post-rebase semantic SHA registration
- push without force

The queue now has a workflow-level concurrency group; concurrent reviewed batches cannot share the rebase/push window.

The consumer/storage test derives effective English-consumer/dialogue layouts before residual static overflow checks, so deleted persisted layouts are re-derived in the existing heavy edit gate. Do not add a redundant whole-corpus overflow pass for every small edit batch.

## Batch 018 dense full-read pilot — complete

Scope:

- `translations/korean/messages/msgsec005.toml`
- `translations/korean/messages/msgsec006.toml`
- 21 IDs read directly in source order

Decisions:

- KEEP: 11
- EDIT: 9
- exclusion/special: 1 (`50014`, internal section marker; pinned English intentionally blank)
- propagated: 0
- contextual coverage credited: 20

EDIT IDs:

`50004, 50006, 50008, 50010, 50011, 60001, 60002, 60003, 60004`

Evidence and execution:

- edit manifest: `docs/audit/beta1-contextual-copyedit-018-reviewed.json`
- KEEP scope: `docs/audit/review/scope-018.json`
- throughput evidence: `docs/audit/review/throughput-018.json`
- packet SHA256: `9cee205f04bb9952a27850329920fb8bc36661b703efbc01758612031845ab47`
- heavy edit workflow run: `34071663361`
- semantic edit commit: `f157624fa1abf293453fa885f4c36a750e60ffd9`
- review-basis/registry commit: `3066cda3d36b52c740ea6fe36c34c379ded6b244`
- KEEP scope workflow run: `34071881826`

Every Batch 018 row was directly read as Japanese meaning/context -> pinned English treatment -> Korean. No repeated-text propagation was credited.

### Batch 018 layout-audit incident

Editing ID `50006` reduced an ordinary whitespace run from three spaces to one. The persisted layout contained three corresponding consecutive `<line-break>` boundaries. The Go compiler's `internal/message.preservesSemantics` correctly rejected the stale layout, but `tools/korean/qa-layout-drift.py` had represented gap positions as a set and therefore lost repeated-boundary cardinality.

Commit `9b76712bb1cabc30106a9ea4d48efdee8cdcd770` fixed the Python audit generically by introducing semantic-unit parsing and a `preserves_layout_semantics()` predicate that mirrors the Go compiler contract. No Korean-only exception was added. This is the required pattern: prefer the general consumer/storage/runtime contract, informed by the pinned English implementation, over a Korean-specific workaround.

### Batch 018 timing interpretation

`docs/audit/review/throughput-018.json` records a measured 297-second direct-review interval for 21 IDs (14.142857 seconds/ID, 254.545455 IDs/hour). That interval must **not** be interpreted as end-to-end Batch 018 wall-clock throughput. It excludes some evidence preparation, CI waiting, and the one-time `50006` audit incident and retry work.

Conversely, total observed session time must not be divided by 21 and extrapolated as pure per-ID reading cost because it contains batch-fixed and incident costs. Neither extrapolation is authoritative for the remaining corpus.

The 018 class-specific stopwatch fields remain null because those classes were not timed separately. Do not synthesize missing timings.

018 was deliberately a small pilot and was not representative: 9/21 rows required EDIT, while the observed workload classification was high-risk 9, ordinary 11, rapid 1. No corpus-wide EDIT rate or production schedule may be inferred from this sample.

## Batch 019 measurement and sizing

Batch 019 is the first production-scale dense-review pilot. The abandoned 20-30-ID planning range is not authoritative.

Scope size should target approximately **250-300 ordinary-ID-equivalent visible review workload**, not blindly 250-300 physical IDs. A state-heavy `<select>`/`<if>` row can represent several ordinary dialogue rows and must be weighted accordingly. Internal subsegments may be used to accumulate review decisions, but CI/evidence fixed overhead should be amortized into one coherent Batch 019 edit queue and one dense KEEP scope where practical.

Batch 019 throughput evidence must distinguish, when actually measured:

- `wall_clock_seconds`
- `reading_seconds`
- `evidence_seconds`
- `ci_wait_seconds`
- `incident_seconds`

Rapid / ordinary / high-risk class-specific stopwatch values may be populated only if those intervals are separately measured. Otherwise they remain null.

The Batch 019 EDIT rate is an observed pilot statistic only. It should be used to identify whether the next bottleneck is direct review or the heavy edit queue, not assumed to be the corpus-wide rate.

Batch 019 has **not yet received any KEEP, EDIT, propagated, or exclusion coverage** at this checkpoint. Its actual source scope must be verified from the current branch tree before review; do not infer or invent the file after `msgsec007.toml`.

## Second-pass accuracy

Tool:

`tools/korean/sample-beta1-review-audit.py`

Requirements:

- deterministic seed
- second reviewer/model must differ from first reviewer when first identity is known
- propagated KEEP is deliberately over-sampled (default target 50% of the sample when population permits)
- correction rate >5% requires expanded re-review

## Next operation: 019 production-scale pilot

1. fetch current legitimate `milestone/Beta1` HEAD;
2. enumerate the actual source files after the completed 005/006 scope from the branch tree;
3. choose a coherent source-order scope totaling roughly 250-300 ordinary-ID-equivalent visible workload;
4. generate the deterministic JP/EN/KO review basis at that HEAD;
5. actually read every selected row in source order: Japanese source meaning/context first, pinned English treatment and rationale second, natural Korean judgment third;
6. inspect all sibling contexts before approving any repeated-text propagation;
7. record exact-before EDIT evidence separately and apply it through the heavy reviewed-edit queue;
8. credit unchanged rows only through the deterministic dense KEEP scope;
9. record separated timing evidence without inventing class timing;
10. use the resulting EDIT mix and separated costs to size later production scopes.

Track A remains risk-ranked candidate cleanup. Track B remains source-order dense review. The two tracks serve different purposes and must not be conflated.
