# Beta1 Dense Review Pipeline v3 Checkpoint

## Current: Batch 021 COMPLETE / Batch 022 next

Date: 2026-09-07. Repository: `fsmkh1-crypto/ZillKoreanPatch`.
Working branch: `milestone/Beta1`. **Beta1 INCOMPLETE. Beta2 NOT STARTED.**

Always fetch the actual remote `milestone/Beta1` HEAD before mutation. Never reset a legitimate newer commit backward and never force-push.

Pinned English reference: `HK47196/zill@a98d9ce29f361d666ec23da0dcfd351f24537ffd`.
Japanese remains semantic authority. For every correction, inspect the pinned English patch's handling/reason before choosing natural Korean; do not invent Korean-only storage/runtime exceptions where English demonstrates the general contract.

## Authoritative generated coverage after Batch 021

Generated coverage commit: `939b3623df0cfb9a61f0ba9cbc862e7ad7f62dc3`.

- Accepted Korean IDs: **42,016**
- Valid contextual review: **749 (1.783%)**
  - legacy direct `full_read`: **176**
  - dense `scope_full_read`: **165**
  - direct `manifest_edit`: **408**
  - propagated: **0**
- `CONTEXT_STALE`: **0**
- `LAYOUT_RECHECK`: **0**
- contextual `UNREVIEWED`: **41,267**
- pending batches lacking registered review basis: **none**
- full accepted `RUNTIME_PENDING` population: **47**

Coverage and structural state remain orthogonal. Language basis is `SHA256(JP + NUL + pinned EN + NUL + KO)`. Structural basis is tracked separately; one changed ID never invalidates an entire scope.

## Batch 021 — COMPLETE

Direct source-order review:

- `translations/korean/messages/msgsec014-part99.toml`: entire file, `140000`–`140021`
- `translations/korean/messages/msgsec015-part99.toml`: entire file, `150000`–`150021`
- `translations/korean/messages/msgsec016-part99.toml`: `160000`–`160084`

Population:

- directly read IDs: **129**
- approximate visible dialogue/state segments: **236**
- KEEP: **62**
- EDIT: **64**
- exclusions/internal markers: **3** (`140021`, `150021`, `160023`)
- contextual IDs credited: **126**
- propagated: **0**

Every ID/state was directly read as Japanese meaning/context -> pinned English handling/reason -> current Korean. No scanner-only KEEP and no unread duplicate propagation was credited.

Key correction classes included canonical `딘갈`, grammar/particle repairs, speaker-register cleanup, punctuation/spacing, Japanese literal-calque removal, runtime `<value>` particle avoidance, and distinction of `邪竜` from the `死竜` place-name wording.

Manifest: `docs/audit/beta1-contextual-copyedit-021-reviewed.json`.
Queue revision: **23**.
Semantic commit: `d6a497378e1515ab7c55e1333dc2390877a1f037`.
Review-basis commit: `77529032a81329d9797bd23f149e62ce1452dfbd`.

### Heavy EDIT validation

Initial heavy run `34083101354` failed only at the final English-consumer storage contract because ID `160063` used **152 bytes** against a **151-byte** guild-region maximum. Exact-before/apply/layout/Korean QA/glyph/font/terminology/text-sanity had already passed.

The fix did not bypass the storage rule. The Korean wording was shortened while preserving Japanese meaning and the English consumer model:

`아주 오래전부터 존재한 폐허...` -> `오래된 폐허...`

Final heavy run `34083780080` — **SUCCESS**.

Verified in the successful run:

- exact before-values and 64 reviewed edits
- control/layout contract
- persisted layout population **141**
- layout semantic drift **0**
- layouts invalidated **0**
- accepted Korean IDs **42,016**
- custom renderer glyphs **1,308**
- bad glyph characters/records **0**
- integrity/terminology/text-sanity gates
- pinned-English consumer/storage/effective-layout contract
- semantic commit and post-rebase review-basis registration

### KEEP scope validation

Scope: `docs/audit/review/scope-021.json`.
Deterministic packet SHA256:
`af440d9dfc720af08ff1e653fe3d402544ec818e8f7dddcb4173731bf8f63377`.

A deliberate placeholder-hash discovery run `34083975845` failed only on the expected packet-hash comparison and made no ledger/coverage mutation. It exposed the CI-reconstructed hash above.

Final scope run `34084089156` — **SUCCESS**:

- all dense packets regenerated and verified
- language ledger rebuilt from the reviewed commit and pinned English
- generated coverage committed

This placeholder-hash run is process overhead, not a translation/runtime regression.

## Timing evidence

`docs/audit/review/throughput-021.json` is authoritative for measured timing.

- successful heavy job window: **47 sec**
- successful scope job window: **24 sec**
- combined successful CI window: **71 sec**
- `reading_seconds`, `evidence_seconds`, `wall_clock_seconds`: **null** because no separate stopwatch was maintained

Do not infer review throughput or remaining-corpus ETA from those CI-only timings.

## Locked review rules

- scanners/heuristics identify risk but create zero KEEP coverage by themselves.
- exact propagation signature is Japanese + pinned English + Korean; EN mismatch forbids propagation.
- KEEP propagation requires every sibling context to be displayed and checked.
- EDIT propagation requires exact before-values and the normal heavy gates.
- final KEEP accuracy audit uses a different reviewer/model and over-samples propagated KEEP; correction rate above 5% expands review.
- B font remains the locked profile: 10x10, BearingX 1, advance 12, gamma 0.60, 10px/72dpi, HintingNone.

## Exact next operation: Batch 022

Start from **ID `160085`**, but first fetch actual remote HEAD and re-enumerate the current source tree. Continue source order; do not infer filenames mechanically.

1. Read the current handoff, this checkpoint, ledger policy, generated coverage and distribution.
2. Select a coherent workload based on visible/state complexity rather than raw ID count alone.
3. Read every selected row/state JP -> pinned EN -> KO.
4. Inspect all sibling contexts before any propagation; otherwise propagation remains zero.
5. Create exact-before manifest `022` only for approved edits.
6. Increment the queue from revision 23 only after rechecking its actual current value.
7. Run one heavy EDIT workflow for the coherent batch, then one dense KEEP scope workflow.
8. Record only actually measured timing; do not fabricate class or wall-clock values.
9. Re-fetch official coverage and final remote HEAD before declaring Batch 022 complete.

Beta1 is not complete until contextual review is complete, stale/layout flags are closed, final whole-corpus reflow/storage/glyph/data/font QA passes, runtime/source anomalies have dispositions, the second-pass KEEP audit passes, and a new post-copyedit Android Beta1 RC is built and verified. **Do not start Beta2.**
