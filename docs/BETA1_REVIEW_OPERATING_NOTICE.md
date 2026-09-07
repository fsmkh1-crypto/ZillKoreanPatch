# Beta1 contextual review operating notice

Status: **ACTIVE / project-owner direction**  
Date: 2026-09-08  
Applies to: GPT, Astra, Claude, Codex, and any other agent performing Beta1 contextual review or related throughput work on `milestone/Beta1`.

This notice defines the current default operating direction for Beta1 contextual review. Read it before starting or redesigning a review batch. If an older handoff or throughput note conflicts with this document on review method or throughput strategy, **this notice controls unless the project owner explicitly supersedes it**.

Repository mutation still requires the explicit task-local authorization defined in `AGENTS.md`. Always re-fetch the actual remote `milestone/Beta1` HEAD immediately before mutation; never reset legitimate newer work backward and never force-push.

## 1. Locked linguistic and English-patch roles

Direct contextual review keeps the established order:

**Japanese -> pinned English -> Korean**

The roles are not symmetric:

- **Japanese is the semantic authority.** Meaning, speaker voice, register, intent, character nuance, and source interpretation are decided from Japanese and context.
- **Pinned English is a proven patch/reference implementation.** Inspect how and why it handles engine-facing constraints, substitutions, controls, storage, reflow, consumer behavior, and known runtime hazards. It may also help interpretation, but it is **not** the semantic authority and must not flatten Japanese voice into English wording.
- **Korean is the target under review.** Judge semantic accuracy against Japanese and natural Korean style while preserving established runtime/structural contracts.

Pinned English reference:

`HK47196/zill@a98d9ce29f361d666ec23da0dcfd351f24537ffd`

Do not reframe this project as EN-first translation. Do not mechanically translate English. Conversely, do not invent Korean-only runtime/storage exceptions when the English patch demonstrates an applicable engine contract.

## 2. Known freeze/runtime causes remain guarded by structural QA

Previously identified freeze/runtime classes have been moved into code/tests/contracts where they can be checked mechanically, including fixed-buffer/storage limits, consumer-specific limits, control topology, dynamic substitutions, reflow/page constraints, CP932/materialized-byte checks, and relevant renderer/glyph contracts.

Therefore contextual language review must not repeatedly re-solve known runtime contracts by hand. Human review handles language/context; established structural/runtime gates handle known machine constraints. Unknown runtime paths and `RUNTIME_PENDING` still require their separate dispositions and final runtime QA.

This separation is a reason to preserve the accurate JP -> EN -> KO review method rather than weakening linguistic review for speed.

## 3. Throughput evidence changes the optimization priority

Batch 022 is the measured baseline for process optimization:

- end-to-end wall: **737.0 sec**
- setup: **90.0 sec**
- direct JP -> pinned EN -> KO reading: **138.8 sec**
- EDIT drafting: **120.0 sec**
- evidence/manifest/queue/scope construction: **207.2 sec**
- successful CI: **69 sec**
- deliberate hash-discovery overhead: **19 sec**
- unattributed tool/API/wait remainder: **93.0 sec**

Derived from that measurement:

- reading alone: **18.8%** of wall time
- reading + EDIT drafting: **35.1%** of wall time
- non-language/operational work: **64.9%** of wall time
- end-to-end: **5.500 sec/contextual ID**
- reading: **1.036 sec/contextual ID**

The main optimization target is therefore **operational overhead and phase switching**, not weakening the linguistic review input.

A method that halves only the 138.8-second reading phase can improve the 737-second baseline by at most about 9% before secondary effects. By contrast, evidence/setup/retry/tooling compression attacks a larger measured share while preserving review quality.

## 4. Current default strategy: Fast operating pipeline, not Fast language algorithm

Until new measurements justify a change, keep the linguistic decision method fixed and optimize workflow organization.

For each review scope:

1. **Pre-flight once.** Re-fetch remote HEAD, confirm pinned English, current source tree, review rules, source-order range, and baseline commit. Build/freeze the deterministic review population.
2. **Read the whole scope before editing.** Review every selected row directly in source order as JP -> pinned EN -> KO with surrounding context. During this phase do not modify translation files, regenerate layout, create manifests, or run CI. Record only EDIT candidates/reasons; KEEP requires no per-row prose artifact.
3. **Draft EDITs in one phase.** After the direct read is complete, revisit only EDIT candidates, re-check Japanese/context and English handling/reason, and finalize Korean replacements. Repeated identical fixes may reuse drafting work, but contextual coverage rules still require the applicable direct-review evidence.
4. **Re-fetch remote HEAD immediately before mutation.** Preserve legitimate concurrent work. Verify exact-before values. Do not reset backward or force-push.
5. **Apply the accepted EDIT set in bulk.** Prefer one semantic edit application/commit per coherent batch unless the edit set itself is so large or heterogeneous that splitting is required for safe diagnosis.
6. **Run heavy validation once after the batch edit set is complete.** Known control/storage/reflow/glyph/data/font/terminology/pinned-English consumer gates remain mandatory. Do not run the full heavy gate after every small subset of edits.
7. **Construct evidence once.** After heavy validation succeeds, create/update the reviewed manifest, dense scope evidence, review basis, ledger inputs, and coverage outputs together rather than interleaving them with language reading.
8. **Run final lightweight scope/ledger validation once.** KEEP-only dense-scope evidence uses deterministic packet regeneration and ledger/basis integrity rather than a second unnecessary heavy full-corpus gate.
9. **Record timing by phase.** See section 6.

The prohibited anti-pattern is repeated context switching such as:

`read a few rows -> edit -> build tool -> CI -> evidence -> read more rows -> another CI -> another document pass`

Do not build new scanners, orchestration tools, or review frameworks merely because they are possible. New tooling must address a measured bottleneck and remain within explicit user authorization.

## 5. Batch-size policy: do not change two variables at once

Completion of 700- and 1,200-ID scopes proves feasibility, not optimality. A larger batch can increase fatigue/late-scope judgment drift and can make a failed heavy validation more expensive to diagnose.

Therefore the next throughput-controlled comparison must **not simultaneously change both batch size and operating pipeline**. Keep the already planned/current batch scale for the controlled measurement and change only the overhead-compression behavior described above.

After a measured comparison exists, batch size may be adjusted based on evidence such as:

- EDIT rate and edit complexity;
- dialogue/character-context density;
- late-scope quality signals;
- heavy-validation failure/diagnosis cost;
- end-to-end seconds per reviewed ID.

Do not declare 1,200 IDs the default optimum merely because S-028/S-029 completed at that size.

## 6. Mandatory timing fields for the next controlled batch

Record these separately in seconds so the old 93-second unattributed bucket is not repeated:

- `setup_seconds`
- `reading_seconds` — direct JP -> pinned EN -> KO review only
- `edit_drafting_seconds`
- `apply_seconds`
- `evidence_seconds` — manifest/scope/review-basis/ledger/coverage construction
- `heavy_ci_success_seconds`
- `final_scope_ci_success_seconds`
- `retry_or_failed_ci_seconds`
- `tool_api_wait_seconds`
- `wall_clock_seconds`

Also record reviewed IDs, KEEP, EDIT, exclusions, and EDIT rate.

Primary comparison metrics against Batch 022:

- **end-to-end seconds per reviewed ID**
- **non-language operational seconds per reviewed ID**
- reading seconds per reviewed ID
- EDIT drafting seconds per EDIT
- retry/failure seconds

Reading share of wall time is useful as a secondary signal only. A target such as reading share rising toward 40% can indicate overhead compression, but it is not sufficient by itself because EDIT rate and drafting cost vary by scope.

## 7. Deferred experiments

The following are **deferred, not adopted as the default pipeline**:

- EN+KO-first Fast review with conditional Japanese lookup;
- Gemini or other parallel linguistic reviewer lanes intended primarily to reduce direct reading time;
- further threshold/scanner work whose main purpose is to reduce the 138.8-second reading phase.

The scope-023 Fast experiment showed that an EN-first linguistic shortcut can miss Japanese speaker/register/speech-act information. More importantly, Batch 022 timing shows that direct reading was only 18.8% of the measured wall clock. These experiments may be revisited later if operational overhead is first reduced and measurement then shows language reading has become the dominant bottleneck.

Do not spend project time extending these experiments unless the project owner explicitly reopens them.

## 8. KEEP accuracy audit timing

Do not perform a large independent KEEP audit after every ordinary batch merely as a throughput ritual. That would recreate the overhead this operating direction is intended to remove.

The existing Beta1 completion policy remains authoritative:

- final/reproducible KEEP accuracy audit uses a different reviewer/model;
- propagated KEEP, if any, is over-sampled;
- correction rate above 5% expands re-review;
- scanner-zero is not accuracy proof.

A targeted spot check may still be justified by a concrete anomaly, but routine full sample auditing remains a later accuracy gate unless the project owner directs otherwise.

## 9. What must not be traded away for speed

Throughput optimization must not weaken:

- direct contextual reading requirements;
- Japanese semantic authority;
- pinned-English engine-contract comparison where applicable;
- Korean style/terminology consistency;
- exact IDs/Japanese/control/substitution invariants;
- heavy gates for actual EDIT commits;
- deterministic scope/ledger evidence;
- final runtime/static completion conditions;
- repository concurrency and no-force-push rules.

The optimization target is **wasted orchestration and repeated phase transitions**, not review quality.

## 10. Agent handoff rule

Any agent resuming Beta1 contextual review must state or verify that it has read this notice before proposing a new throughput architecture. Historical handoffs may contain superseded throughput discussion; preserve their factual measurements but follow this notice for the current operating direction.

If a future measured batch changes the conclusion, update this notice (or explicitly supersede it) rather than allowing contradictory agent-local plans to accumulate across chats.
