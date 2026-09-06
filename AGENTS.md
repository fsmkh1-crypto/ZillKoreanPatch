# ZillKoreanPatch Project Rules

## NON-NEGOTIABLE EXECUTION GATE — EXPLICIT USER START REQUIRED

**STOP RULE: Do not mutate this repository unless the project owner has given an explicit, task-local start signal for the exact work being executed.**

This rule is evaluated before all technical, QA, translation, review, or English-patch-parity rules.

An explicit start signal is a direct instruction to execute, such as `시작`, `ㄱㄱ`, `진행해`, `적용해`, `수정해`, `해놔`, or equivalent wording that clearly authorizes actual changes.

Without such a signal, work is strictly read-only.

Before an explicit start signal, the agent MAY:

- inspect repository files, history, code, data, tests, and documentation;
- compare the Korean path with the English patch;
- investigate defects and root causes;
- analyze third-party or Claude review findings;
- prepare plans, candidate lists, proposed changes, validation plans, and checklists.

Before an explicit start signal, the agent MUST NOT:

- create, modify, rename, move, or delete repository files;
- edit Korean dialogue or terminology data;
- modify code, tests, workflows, CI, or build scripts;
- regenerate or persist layout/build outputs;
- commit or push repository changes;
- apply an otherwise obvious fix merely because its direction is clear.

### Task-local authorization

A start signal authorizes only the bounded task that the project owner actually approved.

Examples:

- `문서 보강해` authorizes the requested documentation changes only.
- `펠름 줄바꿈 수정해` authorizes that defect fix and the directly required regression tests/validation only.
- `클로드 리뷰 검토해` authorizes analysis of the review; it does not authorize applying the review findings.

Do not infer broader authorization from:

- earlier approvals for another task;
- the overall project direction;
- a clear technical next step;
- a failing test;
- an obvious defect;
- a third-party review result;
- `PASS WITH CHANGES`, `CRITICAL`, `MINIMUM CHANGES`, or similar review language.

If execution would expand materially beyond the authorized scope, STOP and wait for a new explicit task-local start signal.

### Mandatory first pre-flight check

Immediately before the first repository write, confirm:

- [ ] `Explicit start signal received: YES`
- [ ] `Authorized scope identified`
- [ ] Correct repository, branch, and HEAD confirmed
- [ ] Baseline SHA recorded
- [ ] Applicable project rules checked
- [ ] English-patch contract identified where applicable
- [ ] No unrelated work is included
- [ ] Required validation gates are known

**If `Explicit start signal received` is not YES, no other condition may authorize repository mutation. STOP.**

Authorization must be re-checked after handoff/new chat/new agent, after completion of the approved task, after an unexpected branch/baseline change, or whenever the scope materially expands.

Technical correctness never overrides missing execution authorization.

## NON-NEGOTIABLE PROJECT PREMISE — ENGLISH PATCH FIRST

Where the existing English patch implements an engine-facing contract, that implementation is the primary reference for the Korean patch.

This premise applies project-wide, including but not limited to:

- message compilation and materialization
- consumer-specific storage limits
- line, page, and whole-record byte limits
- control-token topology and semantics
- dynamic substitutions such as `<value:$XX>`
- layout generation and page boundaries
- fixed-size fields and bounded labels
- bank/archive capacity and replacement semantics
- font/glyph mapping, renderer slots, atlas/PAF transforms
- static/precomputed font transforms
- BOOT.BIN / EBOOT.BIN / bindata handling
- ISO staging, authoring, and final payload provenance

The Korean path may diverge only where Korean text or the custom renderer genuinely requires different representation. Such a divergence must still preserve the underlying engine contract and must be documented with evidence.

A Korean-specific heuristic MUST NOT supersede an established English-patch contract unless there is evidence that the English contract is inapplicable to that consumer.

## Required order of work

When a Korean-path defect, freeze candidate, overflow, layout anomaly, or rendering issue is found:

1. Find the corresponding English-patch path and identify its consumer/engine contract first.
2. Check whether the Korean path preserves the same contract end-to-end.
3. Classify each comparison as one of:
   - `PASS`
   - `DIFFERENT-BY-DESIGN` — only with written evidence
   - `MISSING`
   - `UNKNOWN`
4. Fix `MISSING` contracts before inventing a new Korean-only rule.
5. Keep consumer-specific contracts scoped to the consumers that actually use them; do not generalize one consumer's scanner or buffer rule across unrelated record types.
6. Validate using the final encoded/materialized bytes where possible, not Unicode character count or source-text estimates.
7. Keep canonical Korean translation semantics separate from build-owned layout/projection whenever possible.
8. Fail closed when an established engine contract cannot be proven after materialization.

## Evidence discipline

- One successful runtime test means only that the freeze was not reproduced in that run.
- One freeze/crash is strong failure evidence.
- Do not treat PASS and FAIL symmetrically.
- Do not promote a hypothesis to universal root cause without sufficient evidence.
- Prefer static, compile-time, asset-backed, and exact-materialization gates before repeated runtime probing.
- A same-screen freeze does not by itself prove the same machine-state cause.

## Review requirement

Every change that touches Korean message/layout/rendering/storage/build paths must answer:

- `English patch parity checked: YES / N/A`
- `English reference/consumer contract:`
- `Korean divergence, if any:`
- `Evidence for divergence:`

Unexplained divergence from an established English-patch engine contract is a release blocker.

## Mandatory Korean dialogue QA pre-flight

Before any Korean dialogue copy-editing, translation adjustment, layout/reflow change, or dialogue-runtime QA work, read and follow:

- `CONTRIBUTING.md`
- `docs/KOREAN_TRANSLATION_STYLE.md`
- `docs/KOREAN_DIALOGUE_QA_PROTOCOL.md`
- `docs/CLAUDE_FULL_REVIEW_DECISION_2026-09-06.md` while the accepted review blockers remain open

The pre-flight checklist in `docs/KOREAN_DIALOGUE_QA_PROTOCOL.md` is mandatory per batch/baseline. Repeat it whenever the branch/HEAD/baseline changes, work resumes after handoff, scope changes materially, or concurrent changes are detected. Do not mutate Korean dialogue records until the applicable checklist is complete for the current batch.

In particular:

- do not use blind global punctuation/spacing replacements;
- keep `japanese`, record IDs, runtime controls, and substitutions invariant;
- keep semantic Korean separate from generated `layout`;
- invalidate/regenerate stale layout after semantic edits;
- investigate ordinary line-break/reflow defects through the English consumer/reflow contract before adding Korean-specific behavior;
- distinguish static full-corpus audit, contextual full-corpus copy-edit, reflow coverage audit, sampled runtime QA, and true full-path runtime QA precisely;
- do not call `./zill check` a Korean dialogue/reflow gate; use the Korean-specific gates and QA scanners documented by the accepted review decision;
- do not claim whole-corpus visual safety from a residual audit that only mirrors derivation eligibility;
- classify movable substitutions by proven rendered-width/grammar behavior rather than treating all movable `<value:$XX>` tags as one dynamic class.
