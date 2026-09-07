# Beta1 Dense Review Pipeline v3 Checkpoint

## Current: Batch 022 COMPLETE / measured throughput baseline available

Date: 2026-09-07. Repository: `fsmkh1-crypto/ZillKoreanPatch`.
Working branch: `milestone/Beta1`. **Beta1 INCOMPLETE. Beta2 NOT STARTED.**

Always fetch the actual remote `milestone/Beta1` HEAD before mutation. Never reset a legitimate newer commit backward and never force-push.

Pinned English reference: `HK47196/zill@a98d9ce29f361d666ec23da0dcfd351f24537ffd`.
Japanese remains semantic authority. For every correction, inspect the pinned English patch's handling/reason before choosing natural Korean; do not invent Korean-only storage/runtime exceptions where English demonstrates the general contract.

## Authoritative generated coverage after Batch 022

Generated coverage commit: `61185ae08eb1008f4d7385d436a818d24f9ad990`.

- Accepted Korean IDs: **42,016**
- Valid contextual review: **880 (2.094%)**
  - legacy direct `full_read`: **176**
  - dense `scope_full_read`: **271**
  - direct `manifest_edit`: **433**
  - propagated: **0**
- `CONTEXT_STALE`: **0**
- `LAYOUT_RECHECK`: **0**
- contextual `UNREVIEWED`: **41,136**
- pending batches lacking registered review basis: **none**
- full accepted `RUNTIME_PENDING` population: **47**

Batch 022 directly read 134 contextual IDs but increased valid coverage by **131** because `160127`, `160129`, and `160131` already had valid Batch 003 `CONTEXT_EDIT` evidence. This is expected deduplication, not lost coverage.

Coverage and structural state remain orthogonal. Language basis is `SHA256(JP + NUL + pinned EN + NUL + KO)`. Structural basis is tracked separately; one changed ID never invalidates an entire scope.

## Batch 022 — COMPLETE

Source-order baseline review:

- `translations/korean/messages/msgsec016-part99.toml`: `160085`–`160131`
- `translations/korean/messages/msgsec017-part99.toml`: accepted source-order rows `170000`–`170093`
- next accepted source-order row: **`170095`**

Population:

- directly read physical IDs: **135**
- contextual/player-facing IDs: **134**
- KEEP: **106**
- EDIT: **28**
- exclusion/internal marker: **1** (`170039`)
- new valid coverage: **131**
- propagated: **0**

Every selected row was directly read Japanese -> pinned English -> Korean. No scanner-only KEEP and no unread duplicate propagation was credited.

Representative fixes included the previously established `悪いことは言わん` advice idiom at `160096`, `모험가` -> established `모험자`, runtime `<value:$15>` particle avoidance at `170025`, and correcting equipment prompts `170030/170031` from person-directed `누굴/누구` wording to equipment-directed wording.

Manifest: `docs/audit/beta1-contextual-copyedit-022-reviewed.json`.
Queue revision: **24**.
Semantic commit: `7bc94d24bcb97b11b7373e89286c33b06a4833cb`.
Review-basis commit: `39c6bc34d3c58083c6c30c5750e25ebd7068a436`.

### Heavy EDIT validation

Run `34087495742` — **SUCCESS**, 43 sec.

Verified:

- exact before-values and all 28 reviewed edits
- control/layout contract
- persisted layout population **141**
- layout semantic drift **0**
- layouts invalidated **0**
- accepted Korean IDs **42,016**
- custom renderer glyphs **1,308**
- bad glyph characters/records **0**
- integrity/terminology/consistency/text-sanity gates
- pinned-English consumer/storage/effective-layout contract

### KEEP scope validation

Scope: `docs/audit/review/scope-022.json`.
Deterministic packet SHA256:
`fafa7fba14c0d35e13e319c0faccd32068178fef7091c9fa9483f42feefc50ff`.

The existing placeholder-hash discovery procedure caused run `34087696829` to fail only on the expected zero-hash mismatch; no ledger/coverage mutation occurred. It consumed 19 sec of the measured wall-clock window and is recorded as process overhead.

Final scope run `34087777067` — **SUCCESS**, 26 sec:

- all dense packets regenerated and verified
- language ledger rebuilt
- generated coverage committed

## Batch 022 measured throughput baseline

Authoritative timing record: `docs/audit/review/throughput-022.json`.

Measurement window: `2026-09-07T05:29:48.028Z` -> `2026-09-07T05:42:05Z`.

- wall clock to final successful scope: **737.0 sec**
- setup: **90.0 sec**
- direct JP -> EN -> KO reading: **138.8 sec**
- finalizing 28 EDIT replacement strings: **120.0 sec**
- explicitly phase-timed evidence/manifest/queue/scope construction: **207.2 sec**
- successful CI: **69 sec** (heavy 43 + final scope 26)
- deliberate hash-discovery incident: **19 sec**
- unattributed tooling/API/wait remainder: **93.0 sec**

Derived only for this batch:

- reading: **1.036 sec/contextual ID**
- EDIT drafting: **4.286 sec/EDIT**
- evidence: **1.546 sec/contextual ID**
- successful CI: **0.515 sec/contextual ID**
- end-to-end wall: **5.500 sec/contextual ID**
- EDIT rate: **20.9%**

**Do not turn this one baseline into an authoritative whole-corpus ETA.** Batch 022 was materially easier than Batch 021: EDIT rate was 20.9% versus 50.8%, and 022 contained many short location/item/equipment rows. The measurement does, however, disprove the earlier unsupported assumption that successful CI itself is the dominant bottleneck: only 69 of 737 seconds were the two successful CI jobs.

The next throughput decision should be based on this measured baseline plus at least the known workload-mix difference. Do not resume denominator-marker auditing as a substitute for throughput work; the uploaded technical/non-display audit remains separate accounting evidence.

## Locked review rules

- scanners/heuristics identify risk but create zero KEEP coverage by themselves.
- exact propagation signature is Japanese + pinned English + Korean; EN mismatch forbids propagation.
- KEEP propagation requires every sibling context to be displayed and checked.
- EDIT propagation requires exact before-values and the normal heavy gates.
- final KEEP accuracy audit uses a different reviewer/model and over-samples propagated KEEP; correction rate above 5% expands review.
- B font remains locked: 10x10, BearingX 1, advance 12, gamma 0.60, 10px/72dpi, HintingNone.

## Exact next source point

If source-order review continues, begin at accepted ID **`170095`**, after first re-fetching the actual remote branch and current source tree.

Do not silently enlarge the next batch or introduce parallel lanes until the Batch 022 throughput result has been evaluated for the user's speed question. Beta1 is not complete until contextual review is complete under the accepted completion policy, stale/layout flags are closed, final whole-corpus reflow/storage/glyph/data/font QA passes, runtime/source anomalies have dispositions, the second-pass KEEP audit passes, and a new post-copyedit Android Beta1 RC is built and verified. **Do not start Beta2.**
