# Beta1 Dense Review Pipeline v3 Checkpoint

## Current: Batch020 COMPLETE / Batch021 next

Date:2026-09-07. Current exact continuation is `docs/BETA1_ASTRA_RESUME_NOW.md`.
Batch020 directly read84 IDs, approximately281 visible segments: KEEP53,EDIT27,
excluded4,contextual80,propagated0. Semantic1867f6229a4d398bbfecb15327fbf0f8642139b0;
heavy34081021753 SUCCESS; scope34081135516 SUCCESS (packet verified first attempt).
Generated84c30b504afbb167551466ab3c4d0e59da2a2ed5 proves623/42016 (1.483%):
full_read176,scope_full_read103,manifest_edit344,UNREVIEWED41393,
CONTEXT_STALE0,LAYOUT_RECHECK0,pending basis none. Runtime PENDING47 unchanged.
Glyphs1308; layout drift0; invalidated0; no engine/font changes or runtime/APK proof.
Queue revision21 points to already applied020; recheck actual queue before021.
Next observed source file014-part99 must be reverified in current tree.
Use notes/scope/throughput020 for exact IDs, reasons, timings and workflow evidence.
The measured64s is successful CI job windows, not end-to-end review time.
All current user restrictions and v3 policy remain in force. Beta1 incomplete.

---

## Historical checkpoint019 (superseded for current state/next action)

# Beta1 Dense Review Pipeline v3 Checkpoint

Date: 2026-09-07
Status: **Batch 019 COMPLETE / Beta1 INCOMPLETE / Beta2 NOT STARTED**

This is the current dense-review checkpoint. For review-policy details, `docs/BETA1_REVIEW_LEDGER_POLICY.md` and `docs/audit/review/README.md` remain authoritative. Generated coverage is `docs/audit/beta1-review-coverage.md` and must not be hand-edited.

## Current repository baseline

Repository: `fsmkh1-crypto/ZillKoreanPatch`
Branch: `milestone/Beta1`

The legitimate remote HEAD observed immediately before this checkpoint update was:

`c67b9188b8149c49dd65d08e239291f67f8fa395`

That commit records Batch 019 throughput evidence. This checkpoint commit is intentionally written above it. Always fetch the actual remote HEAD again before any mutation. Never reset legitimate newer commits backward and never force-push.

Pinned English reference:

`HK47196/zill@a98d9ce29f361d666ec23da0dcfd351f24537ffd`

English-patch-first rule remains mandatory: determine Japanese meaning/context first, inspect what the pinned English patch did and why, then choose natural Korean. Japanese is the semantic source; English is the implementation/context reference, not a replacement source language. Do not invent Korean-only storage/runtime/reflow exceptions where the English implementation demonstrates the general solution.

## Authoritative coverage after Batch 019

Generated source: `docs/audit/beta1-review-coverage.md`

- accepted Korean IDs: **42,016**
- valid contextual review: **543 (1.292%)**
  - legacy direct `full_read`: **176**
  - dense `scope_full_read`: **50**
  - direct `manifest_edit`: **317**
  - propagated: **0**
- `CONTEXT_STALE`: **0**
- `LAYOUT_RECHECK`: **0**
- contextual `UNREVIEWED`: **41,473**
- approved registered manifest records: **317**
- pending batches lacking registered review basis: **none**
- full accepted `RUNTIME_PENDING` population: **47**

Language review and structural state remain separate:

`language_basis = SHA256(JP + NUL + pinned EN + NUL + KO)`

`structural_basis = SHA256(layout + NUL + physical_consumer_signature)`

One changed row can stale only that ID. Structural changes may set `LAYOUT_RECHECK` without erasing valid language coverage.

## Batch 018 result retained

Batch 018 reviewed `msgsec005.toml` + `msgsec006.toml`, 21 IDs directly.

- KEEP 11
- EDIT 9
- exclusion 1 (`50014` internal section marker)
- propagated 0
- packet SHA256 `9cee205f04bb9952a27850329920fb8bc36661b703efbc01758612031845ab47`
- heavy edit run `34071663361`
- semantic commit `f157624fa1abf293453fa885f4c36a750e60ffd9`
- review-basis commit `3066cda3d36b52c740ea6fe36c34c379ded6b244`
- KEEP scope run `34071881826`

Batch 018 exposed a real generic layout-audit defect: repeated semantic whitespace boundaries were collapsed by a set-based Python comparison while the Go compiler preserved boundary cardinality. Commit `9b76712bb1cabc30106a9ea4d48efdee8cdcd770` changed `qa-layout-drift.py` to compiler-parity semantic-unit logic. No Korean-only exception was added.

Its historical `297 sec / 21 IDs` interval is **not** end-to-end production throughput. It represented the recorded direct-review window and excluded fixed/incident costs. Do not extrapolate either 163 h or 824 h remaining-work estimates from Batch 018.

## Batch 019 production-scale pilot — COMPLETE

Reviewed source files, in actual repository/source order:

1. `translations/korean/messages/msgsec007.toml`
2. `translations/korean/messages/msgsec007-part99.toml`
3. `translations/korean/messages/msgsec008-part99.toml`
4. `translations/korean/messages/msgsec009-part99.toml`

Important tree fact: ordinary `msgsec008.toml` does **not** exist. Never infer the next filename mechanically; enumerate the actual tree before selecting each new scope.

Batch 019 population:

- direct-read IDs: **59**
- contextual IDs credited: **56**
- KEEP: **39**
- EDIT: **17**
- exclusions/internal markers: **3** (`70029`, `80016`, `90011`)
- propagated: **0**
- observed EDIT rate over contextual IDs: **17/56 = 30.36%**; pilot statistic only, not a corpus-wide estimate

Every selected ID was directly read in source order as:

1. Japanese meaning/context
2. pinned English treatment and likely reason
3. current Korean accuracy/naturalness

No scanner-only KEEP and no repeated-text propagation was credited.

### Batch 019 EDIT evidence and execution

EDIT IDs:

`70002, 70003, 70008, 70009, 70015, 70018, 70026, 80002, 80004, 80008, 80009, 80010, 80011, 80013, 80015, 90001, 90004`

Key correction classes included:

- `王者` mistranslated as `왕자` -> champion/winner sense corrected.
- `他人事ながら` had been rendered as `남 일 같지 않게`, reversing the meaning; corrected to the speaker's detached-but-pleased sense.
- Korean particle/grammar defects such as `우리에게 쥐어짠` were corrected.
- missing question punctuation was restored where Japanese/English both establish an actual question.
- literal Japanese calques such as `간판 아들` and `출발 무대` were naturalized after checking how the English patch avoided literal phrasing.
- fullwidth numeral `６개국` was normalized to Korean-visible `6개국`.

Evidence:

- manifest: `docs/audit/beta1-contextual-copyedit-019-reviewed.json`
- queue revision: **20**
- heavy edit workflow: `34079371354` — **SUCCESS**
- semantic commit: `ff7e45c0bcc1bdabe4087acf5339a9d6e0c90b0d`
- review-basis registration: `2c012b9f6180782d9084ebcdc83d814a75c08443`

Heavy gate passed all stages:

- exact before-value verification
- control topology preservation
- reviewed value application
- stale persisted-layout invalidation/re-derivation
- zero layout drift
- changed-file restriction / `git diff --check`
- Korean glyph/font/integrity/terminology/consistency/text-sanity gates
- pinned-English consumer/storage/effective-layout contract
- semantic commit + final post-rebase review-basis registration

### Batch 019 KEEP scope evidence

- scope: `docs/audit/review/scope-019.json`
- packet SHA256: `42e123c4c4cf456c5d46276134ed3594ad35c4e5338c4172877f56bd2d347889`
- final dense-scope workflow: `34079651223` — **SUCCESS**
- coverage refresh commit: `a1d8525de742fe11ea6ffc16141d280eb991bbb4`

The local session could not reproduce the Git checkouts required to calculate the deterministic packet hash directly. A placeholder hash was therefore used once to let the scope verifier expose its computed actual hash. That deliberate verification run `34079610648` failed only at packet-hash comparison and reported the actual SHA above. The scope was immediately sealed with that SHA and the final scope workflow passed packet regeneration, ledger rebuild, and coverage commit. This was process overhead, not a translation or runtime failure.

### Batch 019 throughput evidence

File: `docs/audit/review/throughput-019.json`

The measurement policy was corrected after Batch 018. Do not invent missing stopwatch data.

Actually observed from workflow timestamps:

- successful heavy edit CI: **45 sec**
- successful dense scope CI: **24 sec**
- combined observed successful CI window: **69 sec**
- hash-oracle incident run: **11 sec**, recorded separately

Not separately instrumented, therefore intentionally `null`:

- `reading_seconds`
- `evidence_seconds`
- `wall_clock_seconds`
- class-specific rapid/ordinary/high-risk stopwatch values

Physical ID count is a poor workload proxy because state-heavy `<select>`/`<if>` IDs contain many visible dialogue states. Future batching should remain workload-aware rather than mechanically target an exact raw ID count.

## Dense review / propagation rules that remain locked

- scanners and AI heuristics discover risk but create **zero KEEP coverage** by themselves.
- exact language propagation signature is Japanese + pinned English + Korean.
- any pinned-English mismatch forbids propagation.
- KEEP propagation requires every sibling context to be displayed and checked before credit.
- propagated KEEP remains separately counted and later over-sampled in second-pass accuracy QA.
- EDIT propagation still requires exact-before values and all heavy edit gates.
- final KEEP accuracy audit uses a different reviewer/model; correction rate above 5% expands re-review.

## Beta1 completion requirements remain unchanged

Beta1 is not complete until all four user goals are satisfied:

1. whole-dialogue line-break/reflow stability;
2. excessive comma, bad spacing and punctuation proofreading;
3. locked B-font readability;
4. translation/naturalness/names/terminology correction.

Before declaring completion, valid contextual coverage must reach 100%, `CONTEXT_STALE=0`, all `LAYOUT_RECHECK` must be closed, final persisted-layout/static overflow/glyph/data/terminology/integrity/consumer contracts must pass, source anomalies/runtime-unbounded rows must have dispositions, and the reproducible second-pass KEEP audit must pass. **Do not start Beta2.**

## Next operation: Batch 020

1. fetch the current legitimate `milestone/Beta1` HEAD and recent commits;
2. read `docs/BETA1_ASTRA_RESUME_NOW.md`, this checkpoint, ledger policy, generated coverage, and review distribution;
3. enumerate the actual message-file tree after the completed `msgsec009-part99.toml` scope; do not invent filenames;
4. choose the next coherent source-order workload using roughly 250-300 ordinary-ID-equivalent visible dialogue as a planning target, reducing raw ID count for state-heavy rows;
5. directly read every selected row JP -> pinned EN -> KO;
6. inspect all sibling contexts before any propagation;
7. create exact-before manifest `020` only for approved edits;
8. update `docs/audit/beta1-copyedit-queue.json` to the next revision to trigger the heavy queue; adding a manifest alone does not trigger it;
9. create deterministic `scope-020.json` and packet evidence for unchanged rows;
10. record timing only where actually instrumented; keep unmeasured fields null;
11. verify heavy EDIT CI, light KEEP scope CI, regenerated official coverage, and final remote HEAD before stopping.

Track A risk discovery and Track B source-order dense full-read remain separate. Batch 020 should continue Track B while using Track A findings only as supplementary risk signals.
