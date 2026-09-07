# Beta1 Review Ledger Policy v2

## Authority

`translations/korean/review-ledger.jsonl` is the generated authoritative progress ledger for Beta1 language/context review. `docs/audit/beta1-review-coverage.md` is its generated human-readable summary. Neither file is hand-edited.

The retired v1 sparse ledger under `docs/audit/beta1-review-ledger.json` must not be used for progress reporting.

## Separate axes

Context review is one axis. Runtime/storage/special-case evidence is orthogonal and must not replace it.

### `context_state`

- `CONTEXT_KEEP`: full contextual review found no semantic/copyedit change necessary at the recorded basis.
- `CONTEXT_EDIT`: contextual review produced an approved semantic/copyedit change at the recorded basis.
- `CONTEXT_STALE`: the recorded review basis no longer matches the current row and is excluded from valid coverage.
- IDs absent from the sparse ledger are contextually `UNREVIEWED` unless another authoritative event is registered.

There is deliberately no `AUTO_ONLY` context state. Automated QA, regex discovery, mechanical migrations and scanner false-positive review are not contextual coverage.

### `evidence`

- `full_read`: the entire recorded scope was read contextually.
- `manifest_edit`: an approved contextual copyedit manifest proves that the edited ID was contextually reviewed.
- `propagated`: reserved for future strict alias/repeated-text propagation after all propagation preconditions are proven. It is reported separately from direct review.

### `flags[]`

Flags are orthogonal attributes, not progress states. Current generated flags include:

- `PERSISTED_LAYOUT`
- `ALIAS_GROUP`
- `SOURCE_ANOMALY`

`FIXED_BUFFER` and `RUNTIME_PENDING` must be derived from the release-owned English-consumer/runtime contract, never hand-entered or guessed. Until that export is integrated, they are explicitly not credited in the ledger.

## Stale-proof review basis

Every valid contextual event records the review-time basis and is checked against the current repository on every ledger build.

The basis is:

`SHA256(Japanese + NUL + pinned English + NUL + Korean + NUL + persisted layout + NUL + consumer_signature)`

The English source is pinned to:

`HK47196/zill@a98d9ce29f361d666ec23da0dcfd351f24537ffd`

The current conservative consumer signature includes physical overlay paths, alias multiplicity, raw consumer metadata when present, and persisted-layout presence. Consumer-sensitive propagation is not credited until the engine-derived consumer map is available.

Historical 001-017 bases are reconstructed at their actual semantic commits from `docs/audit/beta1-context-review-commits.json`. If a historical overlay path moved, the builder recovers the row by ID from the historical tree rather than substituting a current path.

If the current basis differs from the recorded basis, the row becomes `CONTEXT_STALE` automatically and no longer counts toward valid contextual coverage.

## Conservative backfill rule

Historical coverage is intentionally understated.

| Historical evidence | Ledger treatment |
| --- | --- |
| Explicit full-read range such as 001 | `CONTEXT_KEEP` / later `CONTEXT_EDIT` as applicable |
| Applied contextual manifest | `CONTEXT_EDIT` |
| Scanner candidate checked only for one pattern | no contextual credit |
| Scanner false positive | no contextual credit |
| Mechanical terminology migration | no contextual credit |
| Automated/static QA only | no contextual credit |

Discovery is not review.

## Automatic basis registration for new edit batches

The reviewed-copyedit queue is responsible for registering the final semantic review basis for contextual batches.

After all gates pass, it:

1. creates the semantic commit;
2. fetches and rebases against the legitimate current `milestone/Beta1` remote head;
3. captures the final post-rebase semantic SHA;
4. records that SHA in `docs/audit/beta1-context-review-commits.json` for the batch;
5. pushes without force.

This ordering prevents a pre-rebase SHA from becoming the ledger basis. A manifest/scope whose basis has not yet been registered is reported as pending and contributes zero coverage rather than causing fabricated progress.

A future full-read batch with zero semantic edits requires an explicit no-edit review-basis registration mechanism before it may count as KEEP; simply adding a scope is not sufficient.

## Propagation rule

Repeated Japanese text is an optimization opportunity, not automatic coverage.

A propagated review may count only when all group members match at the review basis on:

1. exact Japanese;
2. exact Korean;
3. exact persisted layout;
4. engine-derived consumer signature;
5. relevant orthogonal flags.

The representative must itself be directly context-reviewed. Propagated rows must record `propagated_from`, and reports must always split direct and propagated counts.

Until the engine-derived consumer/flag map is complete, strict duplicate groups are reported only as candidates and contribute zero propagated coverage.

## Coverage versus accuracy

Coverage measures how much of the corpus has a currently valid contextual review basis. It does not prove that every KEEP decision was correct.

Accuracy is checked separately by a reproducible second-pass sample of valid `CONTEXT_KEEP` rows. The default final Beta1 audit target is 200 rows with a recorded deterministic seed. If more than 5% of the sample requires correction, the KEEP population is not considered reliable and expanded re-review is required.

Final scanner-zero is also insufficient by itself. The final scan must include scanners added after the reviewed batches, and must be paired with the random KEEP audit.

## Beta1 language/reflow completion conditions

Before claiming Beta1 language review complete:

- accepted corpus contextual coverage is 100% valid (`KEEP`, `EDIT`, or explicitly proven propagation), with `CONTEXT_STALE = 0`;
- persisted-layout population is revalidated after the final semantic edit;
- line-start prohibited punctuation, newly empty display rows and edge-whitespace postconditions pass;
- known source anomalies and the eight Latin-only title records have reachability/localization dispositions;
- alias consumer consistency passes;
- runtime-unbounded population has a separate runtime disposition;
- final whole-corpus overflow is measured after the final edit and is zero for the proven static population;
- glyph/data/terminology/integrity and pinned-English consumer contracts pass;
- the reproducible KEEP accuracy sample satisfies the accepted threshold;
- the final scanner set includes newly introduced scanners.

Beta2 does not begin until Beta1 is complete.
