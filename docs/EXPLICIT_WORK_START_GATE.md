# Explicit Work Start Gate

This document is a project-wide execution gate. It controls **when repository mutation is authorized**. It applies before any translation, code, test, workflow, documentation, build, or data change.

## STOP RULE

**Do not mutate the repository unless the project owner has given an explicit, task-local start signal for the exact work being executed.**

Examples of explicit start signals include direct instructions such as:

- `시작`
- `ㄱㄱ`
- `진행해`
- `적용해`
- `수정해`
- `해놔`

The wording does not have to match these examples exactly, but it must clearly authorize execution rather than merely request analysis, planning, review, comparison, or discussion.

If explicit execution authorization is absent or ambiguous, **STOP before repository mutation**.

## What is allowed before an explicit start signal

Read-only work is allowed when it serves planning or review:

- inspect repository files and history;
- inspect English-patch behavior and engine contracts;
- analyze defects and likely root causes;
- compare Korean and English paths;
- review third-party findings;
- prepare plans, checklists, proposed diffs, or candidate lists;
- identify tests or validation gates that would be required after execution starts.

These activities do **not** themselves authorize a write.

## What is prohibited before an explicit start signal

Do not perform any repository mutation, including:

- create, edit, rename, move, or delete files;
- change Korean dialogue/TOML data;
- change code, tests, workflows, CI, or build scripts;
- commit or push changes;
- apply terminology normalization;
- regenerate or persist derived layouts;
- make a test pass by modifying implementation or fixtures;
- make any unrelated cleanup discovered during analysis.

A clear defect, an obvious fix, a failing test, or a high-confidence review finding is **not** permission to execute.

## Task-local scope rule

Execution authorization is scoped to the work the project owner actually approved.

Examples:

- `문서 보강해` authorizes the requested documentation changes, not reflow code changes.
- `펠름 줄바꿈 수정해` authorizes the relevant defect fix and required regression tests/validation, not unrelated corpus copy-editing.
- `클로드 리뷰 검토해` authorizes review/analysis unless the owner separately authorizes applying the findings.

Do not infer broader permission from project direction, previous approvals, or the fact that the next technical step appears obvious.

If the authorized scope materially changes, obtain a new explicit task-local start signal before mutating the newly added scope.

## Third-party review is never a start signal

A Claude/code-review/audit result may establish what should be changed, but it does not authorize making those changes.

Statements such as:

- `PASS WITH CHANGES`
- `MINIMUM CHANGES BEFORE START`
- `CRITICAL`
- `fix required`

are evidence and planning inputs only. They are not project-owner execution authorization.

## Mandatory execution pre-flight

Immediately before the first write in an authorized batch, all applicable items must be true:

- [ ] `Explicit start signal received: YES`
- [ ] `Authorized scope identified:` exact files/records/defect class or bounded task
- [ ] Correct repository and working branch confirmed
- [ ] Current HEAD confirmed and baseline SHA recorded
- [ ] Protected/base/unrelated branches excluded
- [ ] Applicable `AGENTS.md` and project QA/style rules checked
- [ ] English-patch reference/consumer contract identified for engine-facing work
- [ ] No unrelated files or opportunistic cleanup are included
- [ ] Validation gates for the authorized batch are known
- [ ] Rollback/comparison point is known

**If the first checkbox is not true, no subsequent checkbox can authorize execution. STOP.**

## Authorization does not persist indefinitely

A start signal applies to the current bounded task/batch. Re-check authorization when:

- the requested scope is completed;
- work resumes after a handoff/new chat/new agent;
- the branch or baseline changes unexpectedly;
- a new defect class is discovered that requires additional code/data changes;
- review findings expand the work beyond the approved scope;
- the user returns from planning/review to execution after an interruption.

When in doubt, return to read-only analysis rather than expanding the write scope.

## Reporting rule

When execution begins, state the authorized scope concisely. When reporting completion, distinguish:

- what the owner explicitly authorized;
- what files were actually changed;
- what was only analyzed or left pending;
- the baseline and resulting HEAD when repository changes were made.

Never describe planned or reviewed changes as already applied.

## Relationship to other project rules

This gate is evaluated **before** technical rules such as English-patch parity, Korean dialogue QA, terminology normalization, reflow, or runtime validation.

Technical correctness cannot override missing execution authorization.

Once the gate is open for a bounded task, all other project rules remain mandatory; authorization is permission to execute, not permission to bypass QA or English-patch-first constraints.
