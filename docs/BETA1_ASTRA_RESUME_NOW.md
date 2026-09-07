# Beta1 Astra resume-now handoff

## Current checkpoint: Batch 019 COMPLETE / Batch 020 next

Date: 2026-09-07 KST  
Status: **BETA1 INCOMPLETE — CONTINUE BETA1 ONLY. DO NOT START BETA2.**

Repository: `fsmkh1-crypto/ZillKoreanPatch`  
Working/target branch: `milestone/Beta1`  
Pinned English reference: `HK47196/zill@a98d9ce29f361d666ec23da0dcfd351f24537ffd`

This file is the current resume-now handoff for the exact stop point after Batch 019. Older handoffs are historical context only when they conflict with this file or the v3 checkpoint.

The legitimate remote HEAD immediately before this handoff write was:

`ed9af3ccbb8cc85cd1c98b1cd0da6d7222dbc1be`

This handoff commit is necessarily newer than that SHA. **Before any mutation, fetch the actual remote `milestone/Beta1` HEAD again and preserve any legitimate newer commits. Never reset backward and never force-push.**

## Read first

Before changing anything, read in this order:

1. `docs/BETA1_ASTRA_RESUME_NOW.md` — this file, exact resume point.
2. `docs/BETA1_REVIEW_PIPELINE_V3_CHECKPOINT.md` — current v3 workflow/checkpoint.
3. `docs/BETA1_REVIEW_LEDGER_POLICY.md` — coverage/evidence policy.
4. `docs/audit/beta1-review-coverage.md` — generated authoritative coverage.
5. `docs/audit/beta1-review-distribution.md` — workload distribution.
6. As needed: `AGENTS.md`, `CONTRIBUTING.md`, Korean style/QA documents, queue/apply scripts and workflows.

Do not use the old Batch-006 section previously present in this file as the active continuation point.

## User-locked Beta1 objective

Beta1 is not merely an APK build. All four goals remain required:

1. Whole-dialogue Korean reflow / line-break stability.
2. Proofreading of excessive commas, spacing and punctuation.
3. B-font readability.
4. Final translation, naturalness, names and terminology quality.

Do **not** start Beta2.

Locked B-font profile:

`10x10 / BearingX 1 / advance 12 / gamma 0.60 / 10px / 72dpi / HintingNone`

Locked terminology currently includes:

`로스톨, 페름, 레무온, 아트레이아, 콘스, 기어, 석화수, 녹사, 파르셴, 플린트, 소도, 주작장군, 현무장군`

Do not substitute the D font or silently change locked terms.

## English-patch-first rule

An English patch exists. For every meaningful correction:

1. Read the actual Japanese source meaning and local context first.
2. Inspect the same ID in pinned English and determine **how it handled the expression and why**.
3. Use that evidence to produce natural Korean.

Japanese remains the semantic authority. Do not blindly translate the English text.

For layout, storage, controls, renderer/consumer behavior and reflow, prefer the general mechanism demonstrated by the English patch instead of inventing Korean-only exceptions or hacks. Batch 018's layout-audit fix is the model: generic compiler-parity behavior, not a Korean special case.

## v3 review/coverage contract

Accepted Korean IDs: **42,016**.

Per-ID language basis:

`SHA256(JP + NUL + pinned EN + NUL + KO)`

A changed language basis invalidates only that ID (`CONTEXT_STALE`). Structural/layout evidence is orthogonal and may set `LAYOUT_RECHECK` without erasing language coverage.

Language propagation equivalence requires exact:

- Japanese
- pinned English
- Korean

Propagation is never automatic. For KEEP propagation, inspect every sibling's surrounding context explicitly. No propagated coverage has been credited yet.

Scanners/AI may discover candidates but cannot create KEEP coverage. KEEP must come from an actual direct contextual read or valid explicitly checked propagation.

EDITs must use exact current before-values and the reviewed-copyedit queue. Adding a manifest alone does not trigger the heavy workflow; `docs/audit/beta1-copyedit-queue.json` must advance.

## Batch 019 — COMPLETE

### Direct review scope

Source files:

- `translations/korean/messages/msgsec007.toml`
- `translations/korean/messages/msgsec007-part99.toml`
- `translations/korean/messages/msgsec008-part99.toml`
- `translations/korean/messages/msgsec009-part99.toml`

Physical IDs reviewed directly in source order:

- `70000..70029`
- `80000..80016`
- `90000..90011`

Total physical IDs: **59**.

Every selected ID was directly read in this order:

**Japanese meaning/context → pinned English treatment/reason → Korean accuracy/naturalness**.

No propagation was credited.

Decisions:

- KEEP: **39**
- EDIT: **17**
- exclusions/internal markers: **3** (`70029`, `80016`, `90011`)
- contextual IDs credited: **56**

EDIT IDs:

`70002, 70003, 70008, 70009, 70015, 70018, 70026, 80002, 80004, 80008, 80009, 80010, 80011, 80013, 80015, 90001, 90004`

Representative fixes included:

- `70002`: fullwidth `６개국` → `6개국`.
- `70003`: JP `俺たちからしぼり取った税金` and pinned EN's “taxes wrung out of us” confirm `우리에게 쥐어짠` was wrong; fixed to `우리에게서 쥐어짠`.
- `70008`, `70009`: restored Korean question punctuation where JP/EN were interrogative.
- `80011`: JP `他人事ながら` means roughly “though it is another person's affair / as an outsider”; existing `남 일 같지 않게` reversed the meaning and was corrected.
- `80010`: Japanese `看板息子` was not left as awkward literal `간판 아들`; naturalized using JP context and English treatment.
- `90004`: Japanese `旅立ちの舞台` was not left as unnatural literal `출발 무대`; naturalized using the English handling/context.
- Other accepted edits were limited to defensible grammar, punctuation, meaning and naturalness corrections. Do not turn one-off findings into new global terminology rules without corpus evidence.

### Visible workload versus official coverage

Batch 019 contained many state-heavy `<select>` / `<if>` records. The review covered approximately **224 visible dialogue-state segments** across the 59 physical IDs.

This does **not** mean 224 coverage IDs. The official ledger is ID-based: a physical ID with eight visible state branches still contributes at most one contextual ID. Therefore Batch 019 produced **56 official contextual IDs** (39 KEEP + 17 EDIT) after excluding the three internal markers.

This distinction must be retained going forward:

- **Official completion/coverage:** ID-based against 42,016 accepted IDs.
- **Workload/throughput sizing:** track both physical IDs and actual/estimated visible segments.

For Batch 020 and later throughput evidence, include at least `total_ids` and `visible_segments_reviewed` (or an explicitly named estimate if exact counting was not instrumented). Do not compare a state-heavy 59-ID batch to 59 one-line IDs as if they were equal work.

Evidence: `docs/audit/review/throughput-019.json` now records `visible_segments_reviewed_estimate: 224` and explicitly states that it is workload evidence, not ledger coverage.

## Batch 019 evidence and CI

EDIT manifest:

`docs/audit/beta1-contextual-copyedit-019-reviewed.json`

Queue after completion:

- `docs/audit/beta1-copyedit-queue.json`
- revision: **20**
- manifest: 019

Do not reapply Batch 018 or 019. Before Batch 020, re-fetch the queue in case concurrent work legitimately advanced it; if unchanged, the next revision is 21.

Batch 019 heavy EDIT workflow:

- run: `34079371354`
- conclusion: **SUCCESS**
- semantic commit: `ff7e45c0bcc1bdabe4087acf5339a9d6e0c90b0d`
- review-basis commit: `2c012b9f6180782d9084ebcdc83d814a75c08443`

The heavy workflow passed:

- exact before-value verification
- reviewed-value application
- stale persisted-layout invalidation/re-derivation
- zero layout drift postcondition
- allowed-file restriction
- per-batch Korean gates
- glyph/font/integrity/terminology/consistency/text-sanity checks
- pinned-English consumer/storage/effective-layout contract
- semantic commit, rebase/push and post-rebase review-basis registration

Dense scope:

`docs/audit/review/scope-019.json`

Packet SHA256:

`42e123c4c4cf456c5d46276134ed3594ad35c4e5338c4172877f56bd2d347889`

Final dense KEEP scope workflow:

- run: `34079651223`
- conclusion: **SUCCESS**
- coverage refresh commit: `a1d8525de742fe11ea6ffc16141d280eb991bbb4`

There was one intentionally failing scope run used only to obtain the deterministic packet hash because the local session could not reproduce the required dual Git checkout:

- run: `34079610648`
- placeholder hash intentionally failed packet verification
- actual SHA was captured and immediately sealed
- final scope run passed

Treat this as process/hash-oracle overhead, **not a semantic, translation, layout or runtime failure**.

## Official coverage after Batch 019

Generated authoritative source: `docs/audit/beta1-review-coverage.md`.

- accepted Korean IDs: **42,016**
- valid contextual review: **543 (1.292%)**
  - legacy direct `full_read`: **176**
  - dense `scope_full_read`: **50**
  - direct `manifest_edit`: **317**
  - propagated: **0**
- `CONTEXT_STALE`: **0**
- `LAYOUT_RECHECK`: **0**
- contextual `UNREVIEWED`: **41,473**
- `RUNTIME_PENDING` full population: **47**
- pending batches lacking review basis: **none**

Dense scope table:

- S-018: 21 IDs / KEEP 11 / EDIT 9 / excluded 1
- S-019: 59 IDs / KEEP 39 / EDIT 17 / excluded 3

## Throughput evidence and what not to infer

`docs/audit/review/throughput-019.json`

Current recorded facts:

- total IDs: 59
- direct-read IDs: 59
- visible dialogue-state segments reviewed: approximately 224
- contextual IDs: 56
- edits: 17
- keeps: 39
- exclusions: 3
- successful CI wait observed: 69 seconds total (45 heavy + 24 scope)
- separate packet-hash incident: 11 seconds

`reading_seconds`, `evidence_seconds`, and end-to-end `wall_clock_seconds` were **not separately stopwatch-instrumented**, so they remain `null`. Do not invent them or derive a fake production rate.

Batch 018's 297 seconds similarly represented only its measured direct-review interval, not its total wall-clock cost. Neither 018 nor 019 justifies a corpus ETA from raw ID/hour alone.

## Exact continuation: Batch 020

The next operation is **Batch 020 source-order dense review**. Do not spend another cycle merely discussing the scope; inspect the tree and begin the actual review.

1. Fetch current remote `milestone/Beta1` HEAD and recent commits. Preserve legitimate newer work.
2. Read the current v3 checkpoint, ledger policy, generated coverage, queue and this handoff.
3. Enumerate the **actual Korean message tree after the completed `msgsec009-part99.toml` scope**. Do not invent `msgsec010*` or assume numbering; verify the branch tree/content first.
4. Choose the next coherent source-order scope by **visible review workload**, aiming roughly at 200–300 ordinary visible-dialogue equivalents. Raw physical ID count is secondary and may be much lower for state-heavy files.
5. Review every selected row directly: JP meaning/context first → pinned EN treatment and reason → KO accuracy/naturalness.
6. Count and record both physical IDs and visible dialogue-state segments. If segment counting is approximate, label it explicitly as an estimate.
7. Keep edits conservative and evidence-based: actual semantic error, grammar/particle issue, punctuation error, awkward literal Japanese, terminology inconsistency, or clearly inferior naturalness. Do not rewrite merely for taste.
8. Inspect the pinned English patch for every meaningful edit and understand why it chose its structure. Japanese remains semantic authority.
9. Do not credit duplicate propagation unless every sibling context was explicitly inspected and the exact JP+EN+KO signature rule holds.
10. After the full review decisions are accumulated, create exact-before manifest `docs/audit/beta1-contextual-copyedit-020-reviewed.json`.
11. Re-fetch `docs/audit/beta1-copyedit-queue.json`; if it is still revision 20, advance to revision 21 and point it to 020. Do not assume if concurrent work changed it.
12. Run the heavy reviewed EDIT queue once for the coherent batch, not once per internal subsegment.
13. Create `docs/audit/review/scope-020.json` and deterministic packet evidence. KEEP is credited only after scope verification succeeds.
14. Create `docs/audit/review/throughput-020.json` containing at minimum:
    - `total_ids`
    - `direct_read_ids`
    - `visible_segments_reviewed` or explicitly named estimate
    - `contextual_ids`
    - `edits`, `keeps`, `exclusions`, `propagated_ids`
    - `reading_seconds`
    - `evidence_seconds`
    - `ci_wait_seconds`
    - `incident_seconds`
    - `wall_clock_seconds`
   Use `null` for any timing not genuinely measured.
15. Verify heavy workflow, dense scope workflow, official generated coverage, final remote HEAD, `CONTEXT_STALE`, and `LAYOUT_RECHECK`.
16. Update the checkpoint and this resume-now handoff after 020 completes.

Internal review subsegments are allowed for attention/state management, but they are not separate official coverage unless explicitly registered. Avoid repeated CI fixed cost per small subsegment.

## Safety / concurrency rules

- Work only on `milestone/Beta1` unless the current authoritative documents explicitly require another branch.
- Do not touch Beta2.
- Never force-push.
- Never reset legitimate newer remote commits to an older SHA.
- Re-fetch branch HEAD before every write-sensitive phase.
- Do not rewrite Japanese source or pinned English evidence.
- Do not hand-edit generated ledger/coverage as a substitute for the workflows.
- Do not credit scanner results as contextual review.
- Do not claim runtime/hardware validation that did not occur.
- Do not create Korean-only layout/storage exceptions when the English/reference consumer model provides a general solution.

## Final Beta1 reminder

Even after dense language coverage is complete, Beta1 still requires the final whole-corpus/reflow/consumer/glyph/data/font QA, second-pass KEEP accuracy audit, outstanding runtime-pending evidence handling, and a fresh Android RC/APK as appropriate. A successful copyedit batch alone is not Beta1 completion.
