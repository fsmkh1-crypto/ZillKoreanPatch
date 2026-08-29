# A-018 — Runtime `<value:$XX>` roles and C5 dynamic-expansion boundary

## Trigger

The audit's second leading root-cause branch is runtime message-consumer / dynamic-expansion memory safety. Earlier reasoning treated all `<value:$XX>` uses too uniformly and did not quantify how many are rendered substitutions versus control-flow reads.

## Questions

1. How many accepted Korean records contain `<value:$XX>` tokens?
2. Which uses are inline rendered values, `<if>` predicate operands, or `<select>` selectors?
3. How much of the C5 corpus depends on inline runtime values?
4. Does the upstream/Korean C5 static validator include the eventual substitution bytes in its 256-byte page-capacity proof?
5. Is ID 10010 unusual because `<value:$15>` is directly adjacent to Korean text?

## Checks performed

- Added `tools/korean/audit-dynamic-substitutions.py` and wired it into CI.
- Added a dedicated audit-report workflow that publishes deterministic corpus counts to PR #17.
- Classified a `<value>` immediately following `<if>` as a predicate operand, immediately following `<select>` as a selector, and all other uses as inline until stronger runtime evidence says otherwise.
- Added per-opcode context samples from the actual Japanese/Korean corpus.
- Read upstream `internal/layout/validate.go` and Korean `internal/layout/validate_korean.go` to compare the static C5 accounting with dynamic substitutions.

## Current-corpus result

At commit `004d3359357b567761615a9c1650b0fbb54c6318` the audit branch contains 42,028 accepted Korean records. This supersedes the earlier project baseline of 42,016 for the current branch state.

Across that accepted corpus:

- records containing any `<value>`: **4,980**
- all `<value>` occurrences: **10,123**
- distinct opcodes: **18**
- inline rendered-value occurrences: **6,566**
- predicate-value occurrences: **3,272**
- selector-value occurrences: **285**
- records with at least one inline value: **4,458**
- inline values immediately followed by non-space natural text: **5,742**
- C5 records with at least one inline value: **2,548**
- C22 records with at least one inline value: **62**
- inline-value records not mapped to the currently audited fixed-consumer sets: **1,846**

The inline opcodes are only nine of the 18 total opcodes:

- `$28`: 5,052
- `$15`: 438
- `$1A`: 394
- `$1B`: 385
- `$16`: 187
- `$17`: 60
- `$2B`: 40
- `$24`: 5
- `$25`: 5

Control-flow-only or predominantly control-flow opcodes include `$01`, `$20`, `$23`, `$29`, `$2A`, `$2C`, `$33`, `$3C`, and `$3D`.

## Important correction to the first inventory

The first audit output reported 10,123 `value_tag_occurrences` and 6,026 immediate non-space adjacencies without distinguishing rendered values from control-flow values. That was too broad for dynamic text-expansion reasoning. In particular, `<select><value:$20>%...` was being counted as an apparent direct adjacency even though it is selector syntax.

The corrected audit separates roles. The relevant inline-adjacency count is **5,742**, not 6,026.

This correction is part of the evidence trail and the earlier number must not be used as a rendered-text-expansion count.

## C5 validator finding

Upstream `internal/layout/validate.go` and Korean `internal/layout/validate_korean.go` both lower the record and walk C5 control flow. When `walkC5` encounters a token of kind `substitution`, it sets `dynamic = true` but does **not** append the eventual runtime substitution bytes to the leaf data used for page-size accounting.

The page-capacity check therefore proves only the statically materialized bytes are below the 256-byte C5 page buffer. It deliberately does not prove that the final runtime-expanded page remains below 256 bytes.

Korean validation preserves this interpretation explicitly: `ValidateKoreanC5` says dynamic substitutions remain a runtime-QA boundary, and `KoreanC5DynamicIDs` reports such records separately rather than declaring them safe.

This is not a newly introduced Korean bug in the validator; it is an inherited/known proof boundary. It becomes more important for Korean because the final translated static payload and runtime substitution mix differ from English.

## `$28` evidence

Upstream chronicle validation explicitly accounts for each `<value:$28>` as up to `playerNameMaxEncodedBytes` (16 encoded bytes). This establishes a known maximum for `$28` in that consumer contract.

The general C5 page validator does not apply that 16-byte allowance. Therefore C5 records containing `$28` can pass static validation while still lacking a proof that a maximum-length player name fits the page after runtime expansion.

This is a **high-value static gap**, not yet proof that the observed freeze is caused by a `$28` overflow. A runtime consumer may have additional behavior or a different staging buffer that still keeps those records safe; that must be established rather than assumed.

## `$15` / ID 10010 evidence

`$15` is an inline value in all 438 observed uses. Corpus contexts show that it is not a player-name-only opcode: it appears in positions representing varying runtime labels/strings such as locations, objects, and other context-dependent names.

ID 10010 is:

`<value:$15>여 ...`

and therefore really does materialize with `$15` immediately followed by Korean renderer bytes. However, direct inline-value adjacency is common: 5,742 such occurrences exist in the current corpus. The adjacency property by itself is therefore substantially less distinctive than originally suspected.

What remains important about ID 10010 is the still-unproven **runtime expansion/storage contract of `$15` in that character-creation-prompt path**, not merely adjacency.

## Evidence assessment

### CONFIRMED

- The current corpus counts and role split above are deterministic for the audited commit.
- C5 static page accounting omits the eventual bytes produced by substitution tokens.
- Korean C5 validation knows this limitation and only marks such records as dynamic runtime-QA risks.
- `$28` is assigned a 16-encoded-byte maximum in the upstream chronicle contract.
- ID 10010's `$15` is an inline value followed directly by Korean text.
- Direct inline-value/text adjacency is widespread and is not unique to ID 10010.

### STRONG

- C5 dynamic headroom deserves higher priority than generic substitution adjacency because 2,548 C5 records contain inline runtime values while static page proof excludes their eventual bytes.

### OPEN

- Whether any actual Korean C5 branch/page exceeds 255 bytes after worst-case runtime substitutions.
- Runtime maximum and source/storage contract for `$15`, `$16`, `$17`, `$1A`, `$1B`, `$2B`, `$24`, and `$25` outside already documented consumers.
- Whether the observed Bad Execution Address originates from C5 expansion or another consumer/staging buffer.

## Meaning for root-cause ranking

The generic hypothesis "a `<value>` directly adjacent to custom Hangul breaks the renderer" is reduced because thousands of inline adjacencies exist.

The more precise hypothesis "a runtime-expanded message passes static validation but exceeds a consumer or intermediate buffer" is strengthened, especially for C5, because the validator's dynamic proof boundary is now directly demonstrated in code and affects thousands of Korean records.

This does **not** eliminate the independent renderer-slot/font branch. The project already has a confirmed renderer-slot ownership counterexample (0x87), so the two branches remain independently active.

## Next gates

1. Extend C5 forensic accounting to retain substitution positions per branch/page and add known maxima where contracts are proven, starting with `$28` = 16 encoded bytes.
2. Report C5 pages that are statically valid but would exceed 255 bytes under proven substitution maxima.
3. Reverse-engineer or otherwise prove the runtime maxima/storage source for the remaining eight inline opcodes before assigning arbitrary bounds.
4. Determine whether C5 expansion writes directly into the 256-byte page buffer or passes through an earlier scratch/staging buffer with a different capacity.
5. Keep ID 10010 as a `$15`-contract target, not as an adjacency-only target.
