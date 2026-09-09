# Beta1 Language Review Coverage v3

> Generated. Do not hand-edit `translations/korean/review-ledger.jsonl` or this summary.

## Authoritative current coverage

- Accepted Korean IDs: **42,016**
- Valid contextual review: **39,431 (93.848%)**
  - legacy direct `full_read`: **176**
  - dense `scope_full_read`: **36,896**
  - direct `manifest_edit`: **2,359**
  - propagated: **0** (not yet credited)
- `CONTEXT_STALE`: **0**
- `LAYOUT_RECHECK`: **0** (orthogonal; does not erase language coverage)
- `UNREVIEWED` for contextual purposes: **2,585**
- Approved registered manifest records (historical, non-deduplicated): **2,365**
- Pending batches lacking a registered review basis: **none**

Language stale is ID-granular and uses `SHA256(JP + NUL + pinned EN + NUL + KO)`.
Structural drift uses `SHA256(layout + NUL + physical_consumer_signature)` and sets `LAYOUT_RECHECK` only.
For dense scopes, the ledger reconstructs each ID at `reviewed_commit`; one changed ID never invalidates the rest of its scope.

## Orthogonal flags and structural QA

- `PERSISTED_LAYOUT`: **42** among ledger rows
- `ALIAS_GROUP`: **0** among ledger rows
- `SOURCE_ANOMALY`: **1** among ledger rows
- `LAYOUT_RECHECK`: **0** among ledger rows; does not invalidate language coverage
- `FIXED_BUFFER`: **17884** among ledger rows; full accepted population **19380**
- `RUNTIME_PENDING`: **43** among ledger rows; full accepted population **47**
- English consumer/category contract SHA-256: `eb64f6fe551f1dd39f3d96db07ff30b698571c1269bf63ac3c5881c44f93be6f`

Consumer/storage/runtime metadata remains in each ledger row for traceability, but it is not part of language propagation equivalence.

## Language propagation candidates (not coverage)

Candidate signature is exact Japanese + exact pinned English + exact Korean. EN mismatch is an unconditional split.
- Unique language signatures: **35,731**
- Duplicate groups: **2,663**
- IDs inside duplicate groups: **8,948**
- Potential extra IDs: **6,285**

## Dense review scopes

| Scope | IDs | KEEP | EDIT | Excluded | File |
| --- | ---: | ---: | ---: | ---: | --- |
| S-018 | 21 | 11 | 9 | 1 | `docs/audit/review/scope-018.json` |
| S-019 | 59 | 39 | 17 | 3 | `docs/audit/review/scope-019.json` |
| S-020 | 84 | 53 | 27 | 4 | `docs/audit/review/scope-020.json` |
| S-021 | 129 | 62 | 64 | 3 | `docs/audit/review/scope-021.json` |
| S-022 | 135 | 106 | 28 | 1 | `docs/audit/review/scope-022.json` |
| S-023 | 430 | 391 | 39 | 0 | `docs/audit/review/scope-023.json` |
| S-024 | 700 | 666 | 34 | 0 | `docs/audit/review/scope-024.json` |
| S-025 | 700 | 623 | 77 | 0 | `docs/audit/review/scope-025.json` |
| S-027 | 700 | 633 | 67 | 0 | `docs/audit/review/scope-027.json` |
| S-028 | 1,200 | 1,139 | 61 | 0 | `docs/audit/review/scope-028.json` |
| S-029 | 1,200 | 1,026 | 174 | 0 | `docs/audit/review/scope-029.json` |
| S-030 | 1,200 | 1,118 | 82 | 0 | `docs/audit/review/scope-030.json` |
| S-031 | 1,200 | 1,178 | 22 | 0 | `docs/audit/review/scope-031.json` |
| S-032 | 1,200 | 1,171 | 29 | 0 | `docs/audit/review/scope-032.json` |
| S-033 | 1,200 | 1,182 | 18 | 0 | `docs/audit/review/scope-033.json` |
| S-034 | 1,200 | 1,170 | 30 | 0 | `docs/audit/review/scope-034.json` |
| S-035 | 1,200 | 1,175 | 25 | 0 | `docs/audit/review/scope-035.json` |
| S-036 | 1,200 | 1,172 | 28 | 0 | `docs/audit/review/scope-036.json` |
| S-037 | 1,200 | 1,172 | 28 | 0 | `docs/audit/review/scope-037.json` |
| S-038 | 1,200 | 1,157 | 43 | 0 | `docs/audit/review/scope-038.json` |
| S-039 | 1,200 | 1,161 | 39 | 0 | `docs/audit/review/scope-039.json` |
| S-040 | 1,200 | 1,043 | 157 | 0 | `docs/audit/review/scope-040.json` |
| S-041 | 1,200 | 1,182 | 18 | 0 | `docs/audit/review/scope-041.json` |
| S-042 | 1,200 | 1,166 | 34 | 0 | `docs/audit/review/scope-042.json` |
| S-043 | 1,200 | 1,177 | 23 | 0 | `docs/audit/review/scope-043.json` |
| S-044 | 1,200 | 1,160 | 40 | 0 | `docs/audit/review/scope-044.json` |
| S-045 | 1,200 | 1,186 | 14 | 0 | `docs/audit/review/scope-045.json` |
| S-046 | 1,200 | 1,167 | 33 | 0 | `docs/audit/review/scope-046.json` |
| S-047 | 1,200 | 1,169 | 31 | 0 | `docs/audit/review/scope-047.json` |
| S-048 | 1,200 | 1,153 | 47 | 0 | `docs/audit/review/scope-048.json` |
| S-049 | 1,200 | 1,138 | 62 | 0 | `docs/audit/review/scope-049.json` |
| S-050 | 1,200 | 1,131 | 69 | 0 | `docs/audit/review/scope-050.json` |
| S-051 | 1,200 | 1,079 | 121 | 0 | `docs/audit/review/scope-051.json` |
| S-052 | 1,200 | 1,102 | 98 | 0 | `docs/audit/review/scope-052.json` |
| S-053 | 1,200 | 1,157 | 43 | 0 | `docs/audit/review/scope-053.json` |
| S-054 | 1,200 | 1,095 | 105 | 0 | `docs/audit/review/scope-054.json` |
| S-055 | 1,200 | 1,140 | 60 | 0 | `docs/audit/review/scope-055.json` |
| S-056 | 1,200 | 1,100 | 100 | 0 | `docs/audit/review/scope-056.json` |
| S-057 | 1,200 | 1,158 | 42 | 0 | `docs/audit/review/scope-057.json` |

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
| 018 | REGISTERED | 9 | 9 | `docs/audit/beta1-contextual-copyedit-018-reviewed.json` |
| 019 | REGISTERED | 17 | 17 | `docs/audit/beta1-contextual-copyedit-019-reviewed.json` |
| 020 | REGISTERED | 27 | 27 | `docs/audit/beta1-contextual-copyedit-020-reviewed.json` |
| 021 | REGISTERED | 64 | 64 | `docs/audit/beta1-contextual-copyedit-021-reviewed.json` |
| 022 | REGISTERED | 28 | 28 | `docs/audit/beta1-contextual-copyedit-022-reviewed.json` |
| 023 | REGISTERED | 47 | 47 | `docs/audit/beta1-contextual-copyedit-023-reviewed.json` |
| 024 | REGISTERED | 34 | 34 | `docs/audit/beta1-contextual-copyedit-024-reviewed.json` |
| 025 | REGISTERED | 77 | 77 | `docs/audit/beta1-contextual-copyedit-025-reviewed.json` |
| 026 | REGISTERED | 7 | 7 | `docs/audit/beta1-contextual-copyedit-026-reviewed.json` |
| 027 | REGISTERED | 67 | 67 | `docs/audit/beta1-contextual-copyedit-027-reviewed.json` |
| 028 | REGISTERED | 61 | 61 | `docs/audit/beta1-contextual-copyedit-028-reviewed.json` |
| 029 | REGISTERED | 174 | 174 | `docs/audit/beta1-contextual-copyedit-029-reviewed.json` |
| 030 | REGISTERED | 83 | 83 | `docs/audit/beta1-contextual-copyedit-030-reviewed.json` |
| 031 | REGISTERED | 22 | 22 | `docs/audit/beta1-contextual-copyedit-031-reviewed.json` |
| 032 | REGISTERED | 29 | 29 | `docs/audit/beta1-contextual-copyedit-032-reviewed.json` |
| 033 | REGISTERED | 37 | 37 | `docs/audit/beta1-contextual-copyedit-033-reviewed.json` |
| 034 | REGISTERED | 31 | 31 | `docs/audit/beta1-contextual-copyedit-034-reviewed.json` |
| 035 | REGISTERED | 25 | 25 | `docs/audit/beta1-contextual-copyedit-035-reviewed.json` |
| 036 | REGISTERED | 28 | 28 | `docs/audit/beta1-contextual-copyedit-036-reviewed.json` |
| 037 | REGISTERED | 28 | 28 | `docs/audit/beta1-contextual-copyedit-037-reviewed.json` |
| 038 | REGISTERED | 43 | 43 | `docs/audit/beta1-contextual-copyedit-038-reviewed.json` |
| 039 | REGISTERED | 39 | 39 | `docs/audit/beta1-contextual-copyedit-039-reviewed.json` |
| 040 | REGISTERED | 157 | 157 | `docs/audit/beta1-contextual-copyedit-040-reviewed.json` |
| 041 | REGISTERED | 18 | 18 | `docs/audit/beta1-contextual-copyedit-041-reviewed.json` |
| 042 | REGISTERED | 34 | 34 | `docs/audit/beta1-contextual-copyedit-042-reviewed.json` |
| 043 | REGISTERED | 23 | 23 | `docs/audit/beta1-contextual-copyedit-043-reviewed.json` |
| 044 | REGISTERED | 40 | 40 | `docs/audit/beta1-contextual-copyedit-044-reviewed.json` |
| 045 | REGISTERED | 14 | 14 | `docs/audit/beta1-contextual-copyedit-045-reviewed.json` |
| 046 | REGISTERED | 33 | 33 | `docs/audit/beta1-contextual-copyedit-046-reviewed.json` |
| 047 | REGISTERED | 31 | 31 | `docs/audit/beta1-contextual-copyedit-047-reviewed.json` |
| 048 | REGISTERED | 47 | 47 | `docs/audit/beta1-contextual-copyedit-048-reviewed.json` |
| 049 | REGISTERED | 62 | 62 | `docs/audit/beta1-contextual-copyedit-049-reviewed.json` |
| 050 | REGISTERED | 69 | 69 | `docs/audit/beta1-contextual-copyedit-050-reviewed.json` |
| 051 | REGISTERED | 121 | 121 | `docs/audit/beta1-contextual-copyedit-051-reviewed.json` |
| 052 | REGISTERED | 98 | 98 | `docs/audit/beta1-contextual-copyedit-052-reviewed.json` |
| 053 | REGISTERED | 43 | 43 | `docs/audit/beta1-contextual-copyedit-053-reviewed.json` |
| 054 | REGISTERED | 105 | 105 | `docs/audit/beta1-contextual-copyedit-054-reviewed.json` |
| 055 | REGISTERED | 60 | 60 | `docs/audit/beta1-contextual-copyedit-055-reviewed.json` |
| 056 | REGISTERED | 100 | 100 | `docs/audit/beta1-contextual-copyedit-056-reviewed.json` |
| 057 | REGISTERED | 42 | 42 | `docs/audit/beta1-contextual-copyedit-057-reviewed.json` |

## Completion/quality rule

Coverage and accuracy are separate. Final Beta1 uses a reproducible second-pass sample of valid KEEP rows.
The second-pass reviewer/model must differ from first-pass review, and propagated KEEP must be sampled above its population share.
A correction rate above 5% requires expanded re-review. Scanner-zero alone is never whole-corpus proof.
