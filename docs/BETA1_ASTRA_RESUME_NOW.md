# Beta1 Astra resume-now handoff

## Exact stop point: Batch 021 COMPLETE / Batch 022 next

Date: 2026-09-07. Repository: `fsmkh1-crypto/ZillKoreanPatch`.
Working branch: `milestone/Beta1`. **Beta1 INCOMPLETE. No Beta2.**

Always re-fetch actual remote `milestone/Beta1` HEAD before any mutation. This handoff is written above the generated Batch 021 coverage commit `939b3623df0cfb9a61f0ba9cbc862e7ad7f62dc3`; later documentation/CI commits are legitimate and must not be reset away. Never force-push.

Read first:

1. this file
2. `docs/BETA1_REVIEW_PIPELINE_V3_CHECKPOINT.md`
3. `docs/BETA1_REVIEW_LEDGER_POLICY.md`
4. `docs/audit/beta1-review-coverage.md`
5. `docs/audit/beta1-review-distribution.md`
6. `docs/audit/review/throughput-021.json` and `scope-021.json` when exact 021 evidence is needed

Pinned English reference: `HK47196/zill@a98d9ce29f361d666ec23da0dcfd351f24537ffd`.
Japanese is semantic authority. For fixes, inspect how the pinned English patch handled the same text/consumer and why, then write natural Korean. Do not mechanically translate English and do not invent Korean-only runtime/storage exceptions where English demonstrates the general solution.

## Batch 021 work actually completed

Direct source-order review:

- `msgsec014-part99.toml`: `140000`–`140021`
- `msgsec015-part99.toml`: `150000`–`150021`
- `msgsec016-part99.toml`: `160000`–`160084`

Result:

- directly read IDs: **129**
- approximate visible dialogue/state segments: **236**
- KEEP: **62**
- EDIT: **64**
- internal marker/comment exclusions: **3** (`140021`, `150021`, `160023`)
- contextual IDs added: **126**
- propagation: **0**

Every selected ID and displayed state was directly read JP -> pinned EN -> KO. No scanner-only KEEP or unread duplicate propagation was credited.

Important repair classes:

- `딩갈` -> canonical `딘갈` in directly reviewed rows
- incorrect Korean particles/causatives and awkward literal Japanese constructions
- speaker-register and punctuation/spacing cleanup
- runtime `<value>` grammar: hard-coded Korean particles after unknown substituted values were avoided by sentence restructuring, following the English patch's value-neutral handling principle
- `邪竜` common-noun meaning was distinguished from the `死竜` place-name wording

Manifest: `docs/audit/beta1-contextual-copyedit-021-reviewed.json`.
Queue: `docs/audit/beta1-copyedit-queue.json`, revision **23**, already applied.
Semantic commit: `d6a497378e1515ab7c55e1333dc2390877a1f037`.
Review-basis commit: `77529032a81329d9797bd23f149e62ce1452dfbd`.

## Validation and one storage incident

Initial heavy run `34083101354` failed only because ID `160063` occupied **152 bytes** in a consumer with a **151-byte** maximum. All earlier apply/layout/Korean QA steps had passed.

The wording was shortened without weakening the contract:

`아주 오래전부터 존재한 폐허...` -> `오래된 폐허...`

Final heavy EDIT run: **34083780080 SUCCESS**.

Successful-run evidence:

- 64 exact reviewed changes applied
- persisted layout population 141
- layout drift 0
- layouts invalidated 0
- accepted Korean records 42,016
- custom renderer glyphs 1,308; bad glyph characters/records 0
- Korean integrity/terminology/consistency/text-sanity passed
- pinned-English consumer/storage/effective-layout contract passed

Scope: `docs/audit/review/scope-021.json`.
Packet SHA256: `af440d9dfc720af08ff1e653fe3d402544ec818e8f7dddcb4173731bf8f63377`.

A placeholder packet hash was used once solely to have the CI verifier expose its deterministic computed hash. Run `34083975845` therefore failed only on the deliberate hash mismatch and made no coverage mutation. The sealed scope then ran successfully:

- final dense scope run: **34084089156 SUCCESS**
- generated coverage commit: `939b3623df0cfb9a61f0ba9cbc862e7ad7f62dc3`

Do not treat either the storage retry or the hash-discovery run as an unresolved regression.

## Current authoritative coverage

From generated `docs/audit/beta1-review-coverage.md` after Batch 021:

- accepted: **42,016**
- valid contextual review: **749 (1.783%)**
  - legacy direct full_read: 176
  - dense scope_full_read: 165
  - direct manifest_edit: 408
  - propagated: 0
- `CONTEXT_STALE`: **0**
- `LAYOUT_RECHECK`: **0**
- contextual `UNREVIEWED`: **41,267**
- pending unregistered review-basis batches: **none**
- full accepted `RUNTIME_PENDING`: **47**

## Timing evidence

`docs/audit/review/throughput-021.json` records only what was actually measured.

- final heavy job: 47 sec
- final scope job: 24 sec
- successful CI total: 71 sec
- reading/evidence/wall-clock timing: null; no separate stopwatch was maintained

Do not convert these CI windows into human review throughput or a corpus ETA.

## Exact next operation: Batch 022

The next unreviewed source-order ID is **`160085`**. Re-fetch the actual branch and enumerate the current file tree before selecting the rest of the scope; do not infer the next filename mechanically.

Continue with a coherent workload sized by visible/state complexity rather than a fixed raw-ID count. For every selected row, read all states JP -> pinned EN -> KO. Inspect every sibling context before propagation; otherwise keep propagation at zero.

For edits:

1. create exact-before `beta1-contextual-copyedit-022-reviewed.json`;
2. recheck current queue, then increment beyond revision 23 to trigger the heavy workflow;
3. do not hand-edit generated ledger/coverage;
4. create and seal `scope-022.json` with the deterministic packet SHA;
5. run one heavy EDIT workflow and one dense scope workflow for the coherent batch;
6. fetch official generated coverage and final remote HEAD before declaring 022 complete.

## Locked Beta1 requirements

Four user goals remain mandatory:

1. whole-dialogue Korean reflow stability
2. comma/spacing/punctuation proofreading
3. B-font readability
4. translation/naturalness/name/terminology correction

B font remains locked at 10x10 / BearingX 1 / advance 12 / gamma 0.60 / 10px / 72dpi / HintingNone.

Beta1 is not finished merely because an APK can build. Before final acceptance, contextual coverage must be completed, stale/layout flags closed, final whole-corpus reflow/storage/glyph/data/font gates passed, runtime/source anomalies dispositioned, a different-reviewer second-pass KEEP accuracy audit passed, and a **new post-copyedit Android Beta1 RC** built and verified. Do not start Beta2.
