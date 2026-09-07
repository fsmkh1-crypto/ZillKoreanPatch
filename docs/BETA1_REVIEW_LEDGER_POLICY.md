# Beta1 Review Ledger Policy v3

## Authority

`translations/korean/review-ledger.jsonl` is the generated authoritative progress ledger for Beta1 language/context review. `docs/audit/beta1-review-coverage.md` is its generated summary. Neither is hand-edited.

Historical batch manifests and `docs/audit/review/scope-*.json` are source evidence. Dense scope rules are specified in `docs/audit/review/README.md`.

## Separate axes

### Language context state

- `CONTEXT_KEEP`: a direct contextual review found no language change necessary at the recorded basis.
- `CONTEXT_EDIT`: contextual review produced an approved semantic/copyedit change.
- `CONTEXT_STALE`: the ID's recorded **language** basis no longer matches current JP/EN/KO and therefore contributes zero valid coverage.
- IDs absent from the sparse ledger are contextually `UNREVIEWED`.

There is no `AUTO_ONLY` context state. Scanner hits, automated QA, mechanical migrations, and pattern-only checks do not count as contextual coverage.

### Evidence

- `full_read`: conservative historical whole-scope read such as 001.
- `scope_full_read`: deterministic dense packet was actually read; unchanged IDs become KEEP.
- `manifest_edit`: approved contextual edit manifest.
- `propagated`: reserved for explicitly recorded repeated-text propagation and always reported separately.

### Structural/runtime flags

These do not replace language state:

- `PERSISTED_LAYOUT`
- `LAYOUT_RECHECK`
- `ALIAS_GROUP`
- `SOURCE_ANOMALY`
- `FIXED_BUFFER`
- `RUNTIME_PENDING`

`FIXED_BUFFER` and `RUNTIME_PENDING` come from the release-owned pinned-English consumer/runtime contract, never from manual guesses.

## v3 basis split

Language review and structural QA have different invalidation rules.

`language_basis = SHA256(Japanese + NUL + pinned English + NUL + Korean)`

`structural_basis = SHA256(layout + NUL + physical_consumer_signature)`

Pinned English:

`HK47196/zill@a98d9ce29f361d666ec23da0dcfd351f24537ffd`

Only `language_basis` mismatch produces `CONTEXT_STALE`. A structural mismatch sets `LAYOUT_RECHECK` and leaves unchanged language coverage valid. Layout/reflow, font, or consumer-contract changes therefore cannot erase language review merely because derived storage presentation changed.

Historical 001-017 rows are migrated by reconstructing their real semantic commits from `docs/audit/beta1-context-review-commits.json`; no old v2 hash is trusted as a v3 language hash. Dense scopes use the same ID-level reconstruction from `reviewed_commit`.

**Scope stale is always ID-granular.** A single changed ID inside a 400-row scope can stale only that ID. The other 399 remain valid if their language bases still match.

## Dense scope and packet evidence

A dense scope stores the exact ID population, edits, exclusions, `reviewed_commit`, pinned English SHA, reviewer identity, and `packet_sha256`.

The packet generator is deterministic: fixed field order, numeric ID order, UTF-8, LF newlines. Lightweight CI regenerates the packet from `(reviewed_commit, english_sha, id_set, source_files)` and rejects a hash mismatch. Packet SHA is therefore repository-verifiable evidence of what was presented for review, not an unverifiable reviewer assertion.

For a scope:

`KEEP IDs = ids - edit_ids - exclusions`

`edit_ids` obtain `CONTEXT_EDIT` only through the reviewed semantic manifest path. Aggregate KEEP counts without a reconstructable ID set are forbidden.

## Conservative historical backfill

| Historical evidence | Contextual credit |
| --- | --- |
| Explicit full-read scope | KEEP / later EDIT as applicable |
| Applied contextual manifest | EDIT |
| Scanner candidate checked for one pattern | none |
| Scanner false positive | none |
| Mechanical terminology migration | none |
| Automated/static QA | none |

Discovery is not review.

## Edit queue and CI split

EDIT commits keep the existing heavy gates: exact before-value, control topology/data integrity, layout-drift postconditions, glyph/font/terminology/text checks, and pinned-English consumer/storage contract.

KEEP-only dense scopes do not alter `translations/`. They use lightweight packet regeneration + ledger/basis integrity checks and reuse cached structural consumer/runtime evidence. Heavy Go/storage gates are not rerun merely because unchanged strings were reviewed.

The reviewed copyedit workflow has a repository-wide concurrency group and captures the final post-rebase semantic SHA before registering edit basis. No force-push is permitted.

The consumer/storage test remains responsible for effective-layout re-derivation. Canonical rows whose persisted layout was invalidated are re-derived before residual static overflow and English-consumer validation, so a separate duplicate full-corpus overflow pass is not required for every small edit batch.

## Propagation

Language-equivalence candidate signature:

1. exact Japanese;
2. exact pinned English;
3. exact Korean.

Any pinned-English mismatch **forbids propagation without exception** and sends the row to individual review.

Layout, consumer, fixed-buffer/runtime class, and physical alias are not language-equivalence inputs, but remain preserved in the ledger for structural QA and later diagnosis.

KEEP and EDIT use the same linguistic equivalence signature. The greater hidden-error risk is propagated KEEP because a bad KEEP creates no semantic diff. Therefore:

- the representative itself must be directly reviewed;
- every group member's surrounding context must be displayed when approving group propagation;
- propagated rows record `propagated_from` and are always reported separately;
- second-pass accuracy QA over-samples propagated KEEP above its population share;
- EDIT propagation still passes exact-before manifest and all normal edit gates.

No propagation candidate contributes coverage until an explicit propagation evidence record exists.

## Coverage versus accuracy

Coverage is the count of IDs with currently valid language evidence. Accuracy is measured independently.

Final Beta1 uses a reproducible second-pass KEEP audit with a reviewer/model different from the first pass. Target baseline is 200 rows; propagated KEEP is deliberately over-sampled. More than 5% requiring correction triggers expanded re-review.

Scanner-zero is never accepted as whole-corpus proof by itself.

## Completion conditions

Before claiming Beta1 language/reflow complete:

- accepted corpus valid contextual coverage is 100%, with `CONTEXT_STALE = 0`;
- all `LAYOUT_RECHECK` conditions are closed by current layout/consumer QA;
- persisted-layout population is revalidated after the final semantic edit;
- line-start prohibited punctuation, newly empty display rows, and edge-whitespace postconditions pass;
- known source anomalies and eight Latin-only title records have dispositions;
- alias consumer consistency passes;
- runtime-unbounded population has a separate runtime disposition;
- final whole-corpus static overflow is freshly measured after the final edit and is zero for the proven population;
- glyph/data/terminology/integrity and pinned-English consumer contracts pass;
- reproducible second-pass accuracy audit passes;
- final scanner set includes scanners added after earlier reviewed batches.

Beta2 does not begin until Beta1 is complete.
