# U6 residual warning population audit

## Scope

U6 is a **payload-neutral audit-instrumentation milestone** over the final U5 runtime patch. It does not change Korean translations, runtime layouts, glyph mappings, fixed strings, or embedded patch-project data. It does change shared Go source at the logging/classification level: `internal/layout/warning_audit.go` adds observational census helpers and `internal/release/korean_contract_chain.go` prints additional `FORENSIC` summaries after the payload-affecting derivation/validation steps have completed.

Purpose: classify the residual non-blocking warning population by upstream-English consumer evidence before considering any further mutation. A warning is not promoted into a runtime contract merely because it is numerous.

Baseline runtime milestone: `838f1b965e89e916d2c780a0e7c70ec0126d4bc7` (`milestone/U5`).

Payload-equivalence evidence is recorded separately in `docs/audit/U6-payload-equivalence.md`.

## Full-corpus result

The repository-only consumer/storage/visual audit checked the complete accepted Korean corpus:

- canonical: 42,016
- checked: 42,016
- upstream-English fixed consumer layouts: 464
- verified narrow-dialogue reflows: 3,829
- hard visual layouts: 141
- consumer contracts: PASS
- visual contracts: PASS
- residual warnings: 16,985

The verified narrow-dialogue population is separately hard-gated:

- checked: 7,382
- reflowed: 3,829
- residual overflow after U5 reflow: **0**

### Exact meaning of `residual_overflow=0`

This gate proves zero residual overflow under the renderer-width model using the static Korean text plus independently established runtime-width reservations. In particular, `<value:$28>` contributes the proven player-name reservation instead of being measured as zero, and authenticated posting bindings use their separately derived reservations.

It does **not** claim that every unknown runtime substitution has a proven worst-case expansion. Rows containing a substitution without an independently established bound remain explicitly tracked below and require either a separately authenticated bound or runtime evidence before a stronger claim can be made.

Therefore the four remaining `line_exceeds_authoring_ceiling` warnings that also fall in the verified narrow-dialogue population are authoring warnings, not residual overflow under the current proven-width model.

## Residual warning census

### `item_description_single_line_overflow`: 13

Exact IDs:

`200534, 200538, 200539, 200540, 200541, 200542, 200543, 200544, 200545, 200546, 200547, 200579, 200580`

All 13 are verified item-description consumers.

Upstream English authority intentionally excludes item descriptions from automatic `Reflow()` and leaves them unreflowed while emitting a warning. Consequently U6 does **not** add Korean-only wrapping or layout mutation for these records. They remain practical-QA candidates if clipping is observed in the real item-description UI.

### `line_exceeds_authoring_ceiling`: 11,985

Ownership/evidence split:

| Evidence | Consumer | Count |
| --- | --- | ---: |
| unknown | C5 | 194 |
| unknown | C5 portrait | 9,212 |
| verified | C5 portrait | 1 |
| unknown | guild commentary | 37 |
| verified | guild region | 41 |
| unknown | unproven | 841 |
| verified category, unproven consumer | unproven | 1,655 |
| verified | verified narrow dialogue | 4 |

These warnings remain heuristic authoring signals unless a separately authenticated consumer contract proves a narrower runtime limit. U6 does not generalize the U5 dialogue-window contract to these populations.

### `runtime_substitution_unbounded`: 4,987

Ownership/evidence split:

| Evidence | Consumer | Count |
| --- | --- | ---: |
| unknown | C22 | 55 |
| verified | C22 | 7 |
| unknown | C5 | 56 |
| unknown | C5 portrait | 2,450 |
| verified | C5 portrait | 211 |
| verified | C5 single page | 47 |
| unknown | guild commentary | 3 |
| unknown | unproven | 659 |
| verified category, unproven consumer | unproven | 285 |
| verified | verified narrow dialogue | 1,214 |

This warning means that the row contains a runtime substitution or tracked format whose final expansion is not universally derivable from static text alone. It is not itself an overflow finding.

## Verified narrow-dialogue substitution boundary

The 1,214 `runtime_substitution_unbounded` rows inside the verified narrow-dialogue population were reclassified at the **unique message-row** level rather than by token-bucket totals:

- total warning rows: **1,214**
- rows whose runtime substitution is only the independently bounded `<value:$28>` player-name token: **809**
- rows containing at least one substitution with no independently proven expansion maximum: **405**

`<value:$28>` itself occurs in 909 verified-narrow-dialogue rows, but token counts overlap: 100 of those rows also contain at least one unbounded token. Consequently `909` must not be subtracted mechanically from `1,214`; the disjoint row-level split is `809 + 405 = 1,214`.

The 405-row population contains these unbounded tokens:

`$01, $15, $1A, $20, $23, $29, $2A, $2B, $2C, $33, $3C`

The repository census hard-pins the `1,214 / 809 / 405` boundary, the exact token set, and the exact sorted 405-message ID set. The ID-set fingerprint is:

`sha256:13bf9c396dff808920032c60c308cf2d1111bae56cebd2c5eccbea954dd3965f`

Every U6 census run recomputes and asserts that fingerprint and emits the exact IDs as `FORENSIC U6_VERIFIED_DIALOGUE_UNBOUNDED_IDS`, so a population change cannot be silently absorbed into the zero-overflow claim.

## Runtime substitution token census

Counts below are message counts per token/consumer bucket. A message may contain more than one token, so token totals are not a disjoint partition of 4,987.

### `<value:$28>`

`$28` is the only current value substitution with an independently established expansion bound: player name, maximum **16 encoded bytes** in the C5 known-expansion audit.

It appears in 4,028 messages across the recorded buckets, including C22, C5/C5-portrait, C5-single-page, verified narrow dialogue, and currently unproven consumers. The renderer-width model explicitly reserves player-name advance for `$28`; it is therefore incorrect to describe every `$28` row as having substitution width zero.

### `<value:$15>`

`$15` appears in **404 messages**:

- unknown guild commentary: 3
- unknown/unproven consumer: 393
- verified category but unproven consumer: 4
- verified narrow dialogue: 4

No independently established runtime expansion maximum exists for `$15`.

The eight `$15` records with verified category evidence remain practical regression priorities:

`10010, 170025, 170207, 170208, 170209, 270043, 1070030, 1070032`

Only four of those eight are also verified narrow dialogue:

`170025, 170207, 170208, 170209`

This list does **not** assert that `$15` is the freeze root cause or that these records are unsafe. It only pins the strongest currently grounded regression shortlist and its dialogue intersection.

Other observed value tokens remain unbounded unless separately proven: `$01`, `$16`, `$17`, `$1A`, `$1B`, `$20`, `$23`, `$24`, `$25`, `$29`, `$2A`, `$2B`, `$2C`, `$33`, `$3C`, `$3D`, plus seven verified `printf`-format records.

## C5 dynamic-expansion interpretation

The existing C5 dynamic audit deliberately keeps three domains separate:

- exact stored compiled bytes,
- static runtime bytes,
- known maximum bytes using only independently proven substitution maxima.

Unknown substitutions are counted as unknown. They are **not** assigned guessed expansion lengths and are not promoted into proven buffer overflow. In particular, stored headroom is not treated as proof of runtime headroom.

## U6 decision

No new runtime/layout/text/glyph mutation is justified by this census alone.

That is intentional: the audit reduced ambiguity without creating a new patch behavior. Further hard gates or fixes require one of the following:

1. an upstream-English consumer/storage contract that applies to the same records,
2. authenticated asset-backed evidence establishing a runtime bound, or
3. practical runtime evidence demonstrating clipping/freeze/crash.

Practical QA priorities after U6:

1. U4 English name-entry grid/input behavior,
2. U5 dialogue wrapping appearance, with the zero-overflow guarantee interpreted under the proven-width model above,
3. the eight verified `$15` regression IDs, especially the four verified-dialogue intersections,
4. the 13 item-description warning IDs if real UI clipping is observed.

U6 is therefore a **payload-neutral audit risk-reduction milestone**, not a new Korean runtime-patch revision.
