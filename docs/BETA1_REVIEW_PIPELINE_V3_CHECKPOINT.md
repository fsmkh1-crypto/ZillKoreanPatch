# Beta1 Dense Review Pipeline v3 Checkpoint

Date: 2026-09-07

This document supersedes the v2 review-ledger checkpoint for language-review workflow decisions. Beta2 remains out of scope.

## Current remote baseline

At checkpoint preparation, the legitimate `milestone/Beta1` head had advanced through the v3 structural refresh and lightweight dense-scope validation. Never reset legitimate newer commits backward; fetch current remote HEAD before any mutation.

Pinned English authority:

`HK47196/zill@a98d9ce29f361d666ec23da0dcfd351f24537ffd`

## Authoritative coverage

Generated source: `docs/audit/beta1-review-coverage.md`

- accepted Korean IDs: 42,016
- valid contextual review: 467 (1.111%)
  - historical full-read: 176
  - dense scope full-read: 0
  - manifest edit: 291
  - propagated: 0
- CONTEXT_STALE: 0
- LAYOUT_RECHECK: 0
- contextual UNREVIEWED: 41,549
- FIXED_BUFFER full population: 19,380
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

## Second-pass accuracy

Tool:

`tools/korean/sample-beta1-review-audit.py`

Requirements:

- deterministic seed
- second reviewer/model must differ from first reviewer when first identity is known
- propagated KEEP is deliberately over-sampled (default target 50% of the sample when population permits)
- correction rate >5% requires expanded re-review

## Next operation: 018 pilot

Do not resume the old 5-30-record-only workflow as the primary review unit.

018 should be the first dense sequential full-read pilot:

1. fetch current legitimate `milestone/Beta1` HEAD;
2. choose one coherent source section/scope;
3. generate the deterministic JP/EN/KO packet at that HEAD;
4. actually read the entire packet in source order;
5. display all sibling contexts before approving any repeated-text propagation;
6. record edits separately in the reviewed edit manifest;
7. create `scope-018.json` with exact IDs, edit IDs, exclusions, reviewer, commit, and packet hash;
8. let the lightweight scope gate credit KEEP rows;
9. let the heavy edit queue apply EDIT rows;
10. record actual elapsed review effort as seconds/ID by class (rapid-scan / ordinary / high-risk) before setting future throughput targets.

Track A remains risk-ranked candidate cleanup. Track B remains source-order dense review. The two tracks serve different purposes and must not be conflated.
