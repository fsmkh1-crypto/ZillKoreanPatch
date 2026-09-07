# Beta1 Astra resume-now handoff

## Exact stop point: Batch 022 COMPLETE / throughput baseline measured

Date: 2026-09-07. Repository: `fsmkh1-crypto/ZillKoreanPatch`.
Working branch: `milestone/Beta1`. **Beta1 INCOMPLETE. No Beta2.**

Always re-fetch actual remote `milestone/Beta1` HEAD before any mutation. Never reset legitimate newer work and never force-push.

Read first:

1. this file
2. `docs/BETA1_REVIEW_PIPELINE_V3_CHECKPOINT.md`
3. `docs/BETA1_REVIEW_LEDGER_POLICY.md`
4. `docs/audit/beta1-review-coverage.md`
5. `docs/audit/beta1-review-distribution.md`
6. `docs/audit/review/throughput-022.json`
7. `docs/audit/review/scope-022.json`

Pinned English reference: `HK47196/zill@a98d9ce29f361d666ec23da0dcfd351f24537ffd`.
Japanese is semantic authority. For fixes, inspect the pinned English patch handling/reason before writing natural Korean. Do not mechanically translate English and do not invent Korean-only runtime/storage exceptions where English demonstrates the general solution.

## Why Batch 022 was run

The user questioned whether contextual coverage was progressing too slowly. Earlier discussion drifted into denominator/technical-marker auditing without first measuring actual review time. Batch 022 was therefore deliberately kept near the existing scale and **used as a pure measured baseline**. Do not silently mix the next throughput decision with more denominator auditing, propagation experiments, or arbitrary batch enlargement.

The uploaded technical/non-display candidate audit remains useful accounting evidence, but it is a separate question from review throughput.

## Batch 022 result

Direct source-order read:

- `msgsec016-part99.toml`: `160085`–`160131`
- `msgsec017-part99.toml`: accepted rows `170000`–`170093`

Population:

- physical/direct IDs: **135**
- contextual/player-facing IDs: **134**
- KEEP: **106**
- EDIT: **28**
- technical exclusion: **1** (`170039`)
- propagation: **0**
- new coverage: **131**
- prior valid-review overlap: `160127`, `160129`, `160131` from Batch 003

Every selected row was directly reviewed JP -> pinned EN -> KO. No scanner-only KEEP was credited.

Important fixes included:

- `160096`: Japanese `悪いことは言わん` advice idiom repaired consistently with the pinned English meaning and earlier directly reviewed instances.
- `160098/160100/160110` and guild wording: established `모험자` terminology restored.
- `160101`: `風桜` normalized to contextual `풍앵나무`.
- `160125`: awkward `거성` replaced with natural castle/residence wording based on JP `居城` and EN handling.
- `170025`: runtime `<value:$15>` particle avoided by sentence restructuring.
- `170030/170031`: equipment selection prompts corrected from person-directed `누굴/누구` wording.

Manifest: `docs/audit/beta1-contextual-copyedit-022-reviewed.json`.
Queue revision: **24**.
Semantic commit: `7bc94d24bcb97b11b7373e89286c33b06a4833cb`.
Review-basis commit: `39c6bc34d3c58083c6c30c5750e25ebd7068a436`.

## Validation

Heavy reviewed-copyedit run `34087495742`: **SUCCESS**, 43 sec.

- 28 exact edits applied
- persisted layout population 141
- layout drift 0
- invalidated layouts 0
- custom renderer glyphs 1308
- glyph/data/integrity/terminology/consistency/text-sanity passed
- pinned-English consumer/storage contract passed

Scope: `docs/audit/review/scope-022.json`.
Packet SHA256: `fafa7fba14c0d35e13e319c0faccd32068178fef7091c9fa9483f42feefc50ff`.

Existing hash-discovery procedure run `34087696829` failed only on the deliberate zero-hash mismatch. It made no ledger/coverage mutation and consumed 19 sec of the measured wall window.

Final dense scope run `34087777067`: **SUCCESS**, 26 sec.
Generated coverage commit: `61185ae08eb1008f4d7385d436a818d24f9ad990`.

## Current authoritative coverage

- accepted: **42,016**
- valid contextual review: **880 (2.094%)**
  - legacy `full_read`: 176
  - dense `scope_full_read`: 271
  - direct `manifest_edit`: 433
  - propagated: 0
- `CONTEXT_STALE`: **0**
- `LAYOUT_RECHECK`: **0**
- contextual `UNREVIEWED`: **41,136**
- pending review-basis batches: **none**

## Throughput measurement — central evidence for the user's speed question

See `docs/audit/review/throughput-022.json`.

Measured from `2026-09-07T05:29:48.028Z` until final successful scope completion at `05:42:05Z`:

- end-to-end wall: **737.0 sec** (12m17s)
- setup: **90.0 sec**
- direct tri-language reading: **138.8 sec**
- edit drafting: **120.0 sec** for 28 EDITs
- explicitly timed evidence construction: **207.2 sec**
- successful CI: **69 sec** (43 + 26)
- deliberate hash-discovery overhead: **19 sec**
- unattributed API/tool/wait remainder: **93.0 sec**

Derived for this batch only:

- reading: **1.036 sec/contextual ID**
- EDIT drafting: **4.286 sec/EDIT**
- evidence: **1.546 sec/contextual ID**
- successful CI: **0.515 sec/contextual ID**
- end-to-end: **5.500 sec/contextual ID**
- EDIT rate: **20.9%**

Do not treat 5.5 sec/ID as an authoritative whole-corpus rate. Batch 022 was easier than Batch 021, whose EDIT rate was 50.8%, and contains many short location/item/equipment rows. The useful conclusion already supported by measurement is narrower: **successful CI is not the dominant bottleneck**; the previous claim that CI/queue overhead was the main problem was unsupported.

The user asked to use this baseline to decide the actual speed strategy. Do not return to technical-marker tier splitting as if it solved throughput. Evaluate the measured 022 result first.

## Exact next source point

If source-order review resumes, the next accepted Korean row after `170093` is **`170095`**. Re-fetch the actual branch and source file before mutation.

Do not automatically start a larger Batch 023 or parallel lanes until the user has evaluated the 022 baseline. The next decision is a throughput-strategy decision, not a denominator-audit task.

## Locked Beta1 requirements

Four user goals remain mandatory:

1. whole-dialogue Korean reflow stability
2. comma/spacing/punctuation proofreading
3. B-font readability
4. translation/naturalness/name/terminology correction

B font remains locked at 10x10 / BearingX 1 / advance 12 / gamma 0.60 / 10px / 72dpi / HintingNone.

Beta1 is not finished merely because an APK can build. Final acceptance still requires the project-defined contextual review/accuracy conditions, stale/layout closure, final whole-corpus reflow/storage/glyph/data/font QA, runtime/source anomaly dispositions, different-reviewer second-pass KEEP audit, and a new post-copyedit Android Beta1 RC. **Do not start Beta2.**
