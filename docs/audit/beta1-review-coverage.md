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
- Approved registered manifest records (historical, non-deduplicated): **291**
- Pending batches lacking a registered review basis: **none**

A row is valid only when its historical review basis still matches current
`SHA256(JP + NUL + pinned EN + NUL + KO + NUL + layout + NUL + consumer_signature)`.
Mismatch automatically reports `CONTEXT_STALE` and removes the row from valid coverage.

## Orthogonal flags and strict propagation candidates

- `PERSISTED_LAYOUT`: **27** among ledger rows
- `ALIAS_GROUP`: **0** among ledger rows
- `SOURCE_ANOMALY`: **0** among ledger rows
- `FIXED_BUFFER`: **241** among ledger rows; full accepted population **19380**
- `RUNTIME_PENDING`: **1** among ledger rows; full accepted population **47**
- English consumer/category contract SHA-256: `eb64f6fe551f1dd39f3d96db07ff30b698571c1269bf63ac3c5881c44f93be6f`

Strict propagation candidates are derived from exact JP + exact KO + exact persisted layout + pinned-English engine consumer signature + physical alias/storage signature. They are candidates only and contribute zero coverage until explicitly propagated from a directly reviewed representative.

- Unique strict signatures: **37351**
- Duplicate strict-signature groups: **1659**
- IDs inside such groups: **6324**
- Potential extra IDs: **4665**

## Historical edit manifests

| Batch | Status | Records | Unique IDs | Manifest |
| --- | --- | ---: | ---: | --- |
| 002 | REGISTERED | 44 | 44 | `docs/audit/beta1-contextual-copyedit-002-reviewed.json` |
| 003 | REGISTERED | 32 | 32 | `docs/audit/beta1-contextual-copyedit-003-reviewed.json` |
| 004 | REGISTERED | 10 | 10 | `docs/audit/beta1-contextual-copyedit-004-reviewed.json` |
| 005 | REGISTERED | 14 | 14 | `docs/audit/beta1-contextual-copyedit-005-reviewed.json` |
| 006 | REGISTERED | 47 | 47 | `docs/audit/beta1-contextual-copyedit-006-reviewed.json` |
| 007 | REGISTERED | 13 | 13 | `docs/audit/beta1-contextual-copyedit-007-reviewed.json` |
| 008 | REGISTERED | 12 | 12 | `docs/audit/beta1-contextual-copyedit-008-reviewed.json` |
| 009 | REGISTERED | 5 | 5 | `docs/audit/beta1-contextual-copyedit-009-reviewed.json` |
| 010 | REGISTERED | 36 | 36 | `docs/audit/beta1-contextual-copyedit-010-reviewed.json` |
| 011 | REGISTERED | 9 | 9 | `docs/audit/beta1-contextual-copyedit-011-reviewed.json` |
| 012 | REGISTERED | 8 | 8 | `docs/audit/beta1-contextual-copyedit-012-reviewed.json` |
| 013 | REGISTERED | 24 | 24 | `docs/audit/beta1-contextual-copyedit-013-reviewed.json` |
| 014 | REGISTERED | 11 | 11 | `docs/audit/beta1-contextual-copyedit-014-reviewed.json` |
| 015 | REGISTERED | 11 | 11 | `docs/audit/beta1-contextual-copyedit-015-reviewed.json` |
| 016 | REGISTERED | 7 | 7 | `docs/audit/beta1-contextual-copyedit-016-reviewed.json` |
| 017 | REGISTERED | 8 | 8 | `docs/audit/beta1-contextual-copyedit-017-reviewed.json` |

## Completion/quality rule

Coverage and accuracy are separate. Final Beta1 must additionally run a reproducible random second-pass audit
of the valid KEEP population (target sample: 200); a correction rate above 5% requires expanded re-review.
Final scanners must include scanners introduced after the reviewed batches; scanner-zero alone is not evidence
of whole-corpus correctness.
