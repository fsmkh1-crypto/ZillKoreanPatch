# Beta1 Review Ledger v2 checkpoint

Date: 2026-09-07 KST
Status: Beta1 INCOMPLETE; review-ledger v2 infrastructure COMPLETE and authoritative for contextual coverage.
Repository: `fsmkh1-crypto/ZillKoreanPatch`
Branch: `milestone/Beta1`

## Current authoritative remote checkpoint

At checkpoint writing time, the latest generated-ledger semantic state was:

- remote HEAD before this documentation commit: `41f9cd5e5d459608d971d9ecd695b81aa8cbe8ab`
- commit: `audit: refresh stale-proof Beta1 review ledger v2`
- ledger workflow run: `34069205708` — SUCCESS
- pinned English authority: `HK47196/zill@a98d9ce29f361d666ec23da0dcfd351f24537ffd`

Always fetch actual `milestone/Beta1` before mutation. Preserve legitimate newer commits. Never reset/force-push backward.

## Authoritative review coverage

Generated source of truth:

- `translations/korean/review-ledger.jsonl`
- `docs/audit/beta1-review-coverage.md`
- policy: `docs/BETA1_REVIEW_LEDGER_POLICY.md`

Current generated figures:

- accepted Korean IDs: **42,016**
- valid contextual review: **467 (1.111%)**
  - direct `full_read`: **176**
  - direct `manifest_edit`: **291**
  - propagated: **0**
- `CONTEXT_STALE`: **0**
- contextually unreviewed: **41,549**
- pending batches without a registered semantic review basis: **none**

Historical scanner hits, scanner false positives, mechanical terminology migration and automated/static QA are deliberately not counted as contextual review coverage.

## Stale-proof basis

Historical review rows are reconstructed at their actual semantic commits from:

`docs/audit/beta1-context-review-commits.json`

The core review basis is:

`SHA256(Japanese + NUL + pinned English + NUL + Korean + NUL + persisted layout + NUL + conservative consumer signature)`

If the current basis differs from the historical review basis, the row becomes `CONTEXT_STALE` automatically and is removed from valid coverage.

Historical overlay path moves are handled by recovering the ID from the historical Git tree rather than assuming the current file path existed at the review commit.

## English-first consumer/runtime flags

The review-ledger workflow checks out the pinned English patch and requires these release-owned inputs to match the Korean branch byte-for-byte:

- `release/layout/consumer-map.toml`
- `release/layout/categories.toml`

`FIXED_BUFFER` is derived from the same English storage contracts enforced by upstream `internal/layout/validate.go`, not from hand-maintained tags or string heuristics.

`RUNTIME_PENDING` is derived from the current whole-corpus Korean English-consumer contract audit (`TestCurrentKoreanCorpusEnglishConsumerStorageContracts`) and its `FORENSIC KOREAN_DIALOGUE_RUNTIME_PENDING` evidence, not guessed from value-token spelling.

Current whole accepted population:

- `FIXED_BUFFER`: **19,380**
- `RUNTIME_PENDING`: **47**

Among the 467 currently reviewed IDs:

- `FIXED_BUFFER`: **241**
- `RUNTIME_PENDING`: **1**
- `PERSISTED_LAYOUT`: **27**
- `ALIAS_GROUP`: **0**
- `SOURCE_ANOMALY`: **0**

Pinned English consumer/category contract SHA-256:

`eb64f6fe551f1dd39f3d96db07ff30b698571c1269bf63ac3c5881c44f93be6f`

## Strict repeated-text propagation

Propagation is not yet credited as coverage.

Current strict grouping requires exact:

1. Japanese
2. Korean
3. persisted layout
4. pinned-English engine consumer signature / relevant runtime-storage attributes
5. physical alias/storage signature

Current candidate census:

- unique strict signatures: **37,351**
- duplicate strict-signature groups: **1,659**
- IDs inside those groups: **6,324**
- potential extra IDs after one directly reviewed representative per valid group: **4,665**
- currently credited propagated IDs: **0**

Do not promote these 4,665 automatically. A representative must be directly reviewed and the exact group contract must still match at the review basis.

## New-batch semantic-basis automation

`.github/workflows/beta1-reviewed-copyedit-queue.yml` now registers contextual-edit semantic bases automatically.

After gates pass it:

1. commits semantic Korean changes;
2. fetches and rebases onto legitimate current `milestone/Beta1`;
3. captures the final post-rebase semantic SHA;
4. records that SHA for the batch in `docs/audit/beta1-context-review-commits.json`;
5. makes a separate audit commit if needed;
6. pushes without force.

A future manifest/scope with no registered basis contributes zero coverage and is shown as `PENDING_BASIS`; it is never silently counted.

Important: this automatic post-rebase registration has been implemented but has not yet been exercised by a new 018 semantic batch. Treat 018 as the first live proof of that integration and inspect its queue output carefully.

A zero-edit full-read-only batch still needs an explicit no-edit basis-registration path before it may contribute KEEP coverage. Do not count a scope alone.

## Accuracy audit

`tools/korean/sample-beta1-review-audit.py` selects a reproducible second-pass sample from valid direct `CONTEXT_KEEP` rows.

Default final target:

- sample: up to 200 valid direct KEEP rows
- deterministic seed: `beta1-final-keep-audit-v1`
- if more than 5% require correction, the KEEP population requires expanded re-review

Coverage and accuracy are deliberately separate metrics.

## Existing layout safeguards already present

`tools/korean/qa-layout-drift.py` already guards against:

- newly stranded prohibited line-start punctuation
- newly empty display lines
- newly introduced line-edge whitespace
- semantic whitespace loss

Ordinary semantic edits invalidate stale persisted layout for reflow rather than treating old layout as permanent authority.

## Immediate next work

Do not start Beta2.

Before contextual batch 018:

1. fetch current remote HEAD and ensure this checkpoint or legitimate newer commits are present;
2. generate a fresh post-017/current aligned JP/EN/KO corpus, or obtain exact-current rows directly;
3. use the ledger for every reviewed KEEP/EDIT decision;
4. treat 018 as the first live proof of queue auto-registration;
5. continue the two-track strategy:
   - risk-driven candidate review (spacing, punctuation, translationese, terminology, suspicious controls/layout)
   - sequential contextual full-read coverage, with strict propagation used only when proven;
6. do not choose a fixed daily count merely for appearance; use the observed strict-group structure and actual review throughput.

Previously noted strong candidate to re-check against exact current JP/EN/KO:

- ID `510078`: likely collapsed spaces around `때믿을` / `걸발견해서` plus unnecessary comma before `와 달라고`; never apply without current exact-before and source/context check.
- ID `10191`: summary text appears to contain missing sentence punctuation around `...수수께끼 귀족이었다 주인공은...`; inspect JP/pinned EN/current KO before changing.

Final Beta1 still requires whole-corpus contextual coverage, final post-edit reflow/overflow measurement, persisted-layout and consumer/alias/source-anomaly dispositions, runtime pending disposition, B-font/glyph/data/terminology/integrity checks, accuracy sampling, and a new final Android Beta1 RC. No existing pre-copyedit APK is final Beta1.
