# U6 residual warning population audit

## Scope

U6 is an audit-only milestone over the final U5 runtime patch. It does **not** change Korean translations, runtime layouts, glyph mappings, fixed strings, or Android payload contents.

Purpose: classify the residual non-blocking warning population by upstream-English consumer evidence before considering any further mutation. A warning is not promoted into a runtime contract merely because it is numerous.

Baseline runtime milestone: `838f1b965e89e916d2c780a0e7c70ec0126d4bc7` (`milestone/U5`).

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

Therefore the four remaining `line_exceeds_authoring_ceiling` warnings that also fall in the verified narrow-dialogue population are authoring warnings, not residual screen-overflow evidence.

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

This warning means that static layout analysis cannot know the final runtime expansion. It is not itself an overflow finding.

## Runtime substitution token census

Counts below are message counts per token/consumer bucket. A message may contain more than one token, so token totals are not a disjoint partition of 4,987.

### `<value:$28>`

`$28` is the only current value substitution with an independently established expansion bound: player name, maximum **16 encoded bytes** in the C5 known-expansion audit.

It appears in 4,028 messages across the recorded buckets, including C22, C5/C5-portrait, C5-single-page, verified narrow dialogue, and currently unproven consumers. The known bound is used only where the corresponding consumer/storage reasoning is authenticated; it is not generalized into an invented universal runtime contract.

### `<value:$15>`

`$15` appears in **404 messages**:

- unknown guild commentary: 3
- unknown/unproven consumer: 393
- verified category but unproven consumer: 4
- verified narrow dialogue: 4

No independently established runtime expansion maximum exists for `$15`.

The eight `$15` records with verified category evidence are pinned as practical regression priorities:

`10010, 170025, 170207, 170208, 170209, 270043, 1070030, 1070032`

This list does **not** assert that `$15` is the freeze root cause or that these eight records are unsafe. It only provides the strongest currently grounded regression shortlist.

Other observed value tokens remain unbounded unless separately proven: `$01`, `$16`, `$17`, `$1A`, `$1B`, `$20`, `$23`, `$24`, `$25`, `$29`, `$2A`, `$2B`, `$2C`, `$33`, `$3C`, `$3D`, plus seven verified `printf`-format records.

## C5 dynamic-expansion interpretation

The existing C5 dynamic audit deliberately keeps three domains separate:

- exact stored compiled bytes,
- static runtime bytes,
- known maximum bytes using only independently proven substitution maxima.

Unknown substitutions are counted as unknown. They are **not** assigned guessed expansion lengths and are not promoted into proven buffer overflow. In particular, stored headroom is not treated as proof of runtime headroom.

## U6 decision

No new runtime/layout/text/glyph mutation is justified by this census alone.

That is intentional: the audit reduced ambiguity without creating a new regression surface. Further hard gates or fixes require one of the following:

1. an upstream-English consumer/storage contract that applies to the same records,
2. authenticated asset-backed evidence establishing a runtime bound, or
3. practical runtime evidence demonstrating clipping/freeze/crash.

Practical QA priorities after U6:

1. U4 English name-entry grid/input behavior,
2. U5 dialogue wrapping appearance despite residual overflow = 0,
3. the eight verified `$15` regression IDs above,
4. the 13 item-description warning IDs if real UI clipping is observed.

U6 is therefore an **audit-only risk-reduction milestone**, not a new patch revision requiring a separate Android runtime payload.
