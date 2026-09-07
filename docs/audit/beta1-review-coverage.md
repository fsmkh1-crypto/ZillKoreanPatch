# Beta1 Language Review Coverage v2

> Generated. Do not hand-edit `translations/korean/review-ledger.jsonl` or this summary.

## Authoritative current coverage

- Accepted Korean IDs: **42,016**
- Valid contextual review: **467 (1.111%)**
  - direct `full_read`: **176**
  - direct `manifest_edit`: **291**
  - propagated: **0** (not yet credited)
- `CONTEXT_STALE`: **0**
- `UNREVIEWED` for contextual purposes: **41,549**
- Approved manifest records (historical, non-deduplicated): **291**

A row is valid only when its historical review basis still matches current
`SHA256(JP + NUL + pinned EN + NUL + KO + NUL + layout + NUL + consumer_signature)`.
Mismatch automatically reports `CONTEXT_STALE` and removes the row from valid coverage.

## Orthogonal flags currently derived

- `PERSISTED_LAYOUT`: **27** among ledger rows
- `ALIAS_GROUP`: **0** among ledger rows
- `SOURCE_ANOMALY`: **0** among ledger rows
- `FIXED_BUFFER` / `RUNTIME_PENDING`: **not guessed**; pending integration with a repository-derived consumer/runtime map.

## Strict propagation candidates (not coverage)

- Unique strict signatures: **36,899**
- Duplicate strict-signature groups: **1,986**
- IDs inside such groups: **7,103**
- Potential extra IDs after one representative review per group: **5,117**

Strict signature requires exact Japanese + exact Korean + exact persisted layout + repository-visible consumer signature.
These are candidates only; no automatic KEEP propagation is credited.

## Historical edit manifests

| Batch | Records | Unique IDs | Manifest |
| --- | ---: | ---: | --- |
| 002 | 44 | 44 | `docs/audit/beta1-contextual-copyedit-002-reviewed.json` |
| 003 | 32 | 32 | `docs/audit/beta1-contextual-copyedit-003-reviewed.json` |
| 004 | 10 | 10 | `docs/audit/beta1-contextual-copyedit-004-reviewed.json` |
| 005 | 14 | 14 | `docs/audit/beta1-contextual-copyedit-005-reviewed.json` |
| 006 | 47 | 47 | `docs/audit/beta1-contextual-copyedit-006-reviewed.json` |
| 007 | 13 | 13 | `docs/audit/beta1-contextual-copyedit-007-reviewed.json` |
| 008 | 12 | 12 | `docs/audit/beta1-contextual-copyedit-008-reviewed.json` |
| 009 | 5 | 5 | `docs/audit/beta1-contextual-copyedit-009-reviewed.json` |
| 010 | 36 | 36 | `docs/audit/beta1-contextual-copyedit-010-reviewed.json` |
| 011 | 9 | 9 | `docs/audit/beta1-contextual-copyedit-011-reviewed.json` |
| 012 | 8 | 8 | `docs/audit/beta1-contextual-copyedit-012-reviewed.json` |
| 013 | 24 | 24 | `docs/audit/beta1-contextual-copyedit-013-reviewed.json` |
| 014 | 11 | 11 | `docs/audit/beta1-contextual-copyedit-014-reviewed.json` |
| 015 | 11 | 11 | `docs/audit/beta1-contextual-copyedit-015-reviewed.json` |
| 016 | 7 | 7 | `docs/audit/beta1-contextual-copyedit-016-reviewed.json` |
| 017 | 8 | 8 | `docs/audit/beta1-contextual-copyedit-017-reviewed.json` |

## Completion/quality rule

Coverage and accuracy are separate. Final Beta1 must additionally run a reproducible random second-pass audit
of the valid KEEP population (target sample: 200); a correction rate above 5% requires expanded re-review.
Final scanners must include scanners introduced after the reviewed batches; scanner-zero alone is not evidence
of whole-corpus correctness.
