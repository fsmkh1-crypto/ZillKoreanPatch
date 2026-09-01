# U6 C5 single-page stabilization

## Scope

This record closes the historical U6 C5 single-page build failure without changing canonical Korean translation text, weakening the C5 validator, or creating a new U-stage. The repair belongs to U6 stabilization.

The original asset-backed failure checked 15,679 C5 records and reported **9 message IDs / 15 branches** where the effective Korean projection occupied `2 pages` while the authenticated consumer contract allows `maximum 1` page.

| Message ID | Branch(es) | Historical observation |
| ---: | --- | --- |
| 1280007 | b7 | 2 pages > maximum 1 |
| 1280008 | b5, b7 | 2 pages > maximum 1 |
| 1280012 | b1, b2, b5, b6 | 2 pages > maximum 1 |
| 1280017 | b5 | 2 pages > maximum 1 |
| 1280020 | b2, b7 | 2 pages > maximum 1 |
| 1280021 | b3, b5 | 2 pages > maximum 1 |
| 1280043 | b4 | 2 pages > maximum 1 |
| 1280050 | b4 | 2 pages > maximum 1 |
| 1280051 | b4 | 2 pages > maximum 1 |

Total: **9 messages / 15 branches**.

## Authenticated consumer classification

Pinned upstream-English authority: `HK47196/zill@a98d9ce29f361d666ec23da0dcfd351f24537ffd`.

All nine message IDs are members of upstream `release/layout/consumer-map.toml` `single_page_c5_ids`. The local consumer map preserves that classification. The `maximum 1` validation result is therefore a real upstream-English storage contract, not a Korean-only warning.

The nine IDs are not C22 scanner consumers.

### Independent upstream parity verification

The pinned upstream file declares the complete single-page population as the contiguous range `1280000` through `1280053` (54 IDs). Therefore every historical failure ID is an exact upstream member:

`1280007, 1280008, 1280012, 1280017, 1280020, 1280021, 1280043, 1280050, 1280051`.

The pinned upstream `release/layout/consumer-map.toml` Git blob SHA is:

`d902650500743b04ec9d5a38687f68c63580027a`

The same file at the Korean audit branch state used for this verification has the identical Git blob SHA:

`d902650500743b04ec9d5a38687f68c63580027a`

This proves full-file byte parity for `consumer-map.toml`, not merely matching membership for the nine IDs.

The validator semantics are also directly retained from upstream source:

1. `internal/layout/config.go` maps TOML `single_page_c5_ids` into `SinglePageC5IDs`.
2. `internal/layout/validate.go` includes `SinglePageC5IDs` in the C5 validation population.
3. Generic C5 starts with `c5MaxPages`, whose upstream value is `9`.
4. If the record ID is in `SinglePageC5IDs`, validation overrides `maxPages` to `1`.

Accordingly, `2 pages > maximum 1` is the intended upstream contract for these nine records, not a Korean-only limit introduced by this repository.

## Cause and repair

The historical failure was not caused by the Korean semantic translation text. The affected Korean records contain semantic branches (`<end>` and selection controls) rather than authored page breaks that deliberately request two C5 pages.

The failure class came from applying the captured retail string-scanner hardening too broadly. A scanner-derived conservative wrap could introduce enough line breaks to advance the C5 three-line page cursor into a second page. The later, correct C5 validator then rejected that derived state because these consumers are authenticated single-page C5.

Current U6 production code removes that cross-consumer mutation:

1. C5 storage derivation remains owned by the C5 contract.
2. Runtime scanner hardening now uses `DeriveKoreanC22RetailScannerLayouts` rather than the historical all-record scanner derivation.
3. That derivation iterates only authenticated `C22IDs` and additionally requires authenticated retail source compatibility with the captured NUL-terminated scanner model.
4. The C5 validator remains fail-closed and still enforces `maxPages = 1` for `single_page_c5_ids`.
5. Canonical Korean strings are not shortened or rewritten to make the gate pass.

This is the correct repair direction: remove the invalid consumer cross-over rather than weaken the consumer contract.

## Regression gate

`internal/layout/retail_scanner_scope_test.go` now hard-asserts the exact historical nine-ID population:

- every ID must remain authenticated as `single_page_c5_ids`, and
- no ID may appear in `C22IDs` / C22 scanner derivation scope.

This prevents a future parity/refactor change from silently reintroducing the same scanner-to-C5 collision.

## Evidence boundary

This closes the identified build-path defect in source and configuration. It does **not** convert a successful build or playthrough into proof that every historical PSP freeze is solved. Runtime substitution risks and unrelated warning populations retain their own evidence boundaries.

A retail ISO rerun remains useful practical QA, but it is no longer a prerequisite for applying the known structural fix: the invalid scanner ownership that produced this failure class has been removed and is now regression-guarded under U6.
