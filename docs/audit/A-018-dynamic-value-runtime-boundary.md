# A-018 — Runtime `<value:$XX>` roles and C5 dynamic-expansion boundary

## Trigger

The audit's runtime message-consumer / dynamic-expansion branch needed a quantified separation between values that are rendered into text and values used only by control flow. Independent review then challenged three important assumptions: the exact C5 substitution undercount, the provenance of the `$28` 16-byte bound, and the lightweight Python role classifier.

## Questions

1. How many accepted Korean records contain `<value:$XX>` tokens?
2. Which uses are inline rendered values, `<if>` predicate operands, or `<select>` selectors?
3. How much of the C5 corpus depends on inline runtime values?
4. Does the C5 static validator include substitution placeholder bytes or eventual runtime bytes in its 256-byte page-capacity proof?
5. Is `$28 = 16 encoded bytes` a chronicle-local assumption or a global player-name input/storage contract?
6. Is ID 10010 unusual because `<value:$15>` is directly adjacent to Korean text?
7. Can the lightweight role classifier misclassify a value used as the right-hand side of a predicate comparison?

## Checks performed

- Added `tools/korean/audit-dynamic-substitutions.py` and wired it into CI/reporting.
- Classified values immediately following `<if>` as predicate operands, immediately following `<select>` as selectors, and all other uses as inline for corpus triage.
- Added per-opcode context samples from the real Japanese/Korean corpus.
- Added a counterexample scan for comparison RHS forms such as `<equal><value:$XX>` that the lightweight classifier would otherwise misclassify as inline.
- Read upstream/local C5 validation and added a direct regression test against the original `walkC5` behavior.
- Added branch/page forensic accounting that retains substitution positions and applies only independently established dynamic maxima.
- Corrected the C5 page-boundary off-by-one so the third line-break byte is counted into the page it terminates before subsequent bytes move to the next page.
- Rechecked upstream `release/layout/README.md` for the provenance of the player-name bound.

## Current-corpus result

For the current audit corpus:

- accepted Korean records: **42,028**
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
- concrete comparison-RHS `<value>` counterexamples detected by the new scan: **0**

The nine currently observed inline opcodes are:

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

The corrected relevant inline-adjacency count is **5,742**, not 6,026.

Independent review correctly noted that the Python classifier is not a full grammar parser. A value used on the right-hand side of a predicate comparison could evade the immediate-prefix rule. The audit now explicitly scans the concrete counterexample form; the current corpus contains **zero** such occurrences. This means the reviewer-proposed blind spot does not presently change the headline counts, but the classifier is still described as a corpus triage tool rather than a grammar proof.

## C5 validator finding

The original `walkC5` behavior is now directly regression-tested:

- static text bytes are appended to `leaf.data`;
- a `substitution` token sets `dynamic = true`;
- **neither the eventual runtime bytes nor the compiled two-byte `02 XX` substitution marker are appended to `leaf.data`.**

Therefore the ordinary C5 page-capacity check counts a substitution as **zero page bytes**. The gap is not merely “runtime text may be longer than the two-byte placeholder”; the placeholder itself is absent from page accounting.

Korean validation explicitly describes dynamic substitutions as a runtime-QA boundary and reports dynamic IDs rather than declaring them safe.

This is an inherited proof boundary, not by itself a newly discovered Korean runtime defect.

## C5 boundary-byte correction

Independent review also correctly identified an off-by-one in both the Korean page validator and the first forensic helper: on every third line break, the code created the next page before incrementing the current page's byte count, so that boundary `0x0A` belonged to neither page.

The Korean validator and forensic accounting now use a shared count-before-transition page cursor:

1. count the line-break byte into the current page;
2. if it is the third break, start a new page for subsequent bytes.

Regression tests cover this exact boundary.

The inherited upstream/English validator copy still carries its historical implementation and is not evidence that the old omission was correct. The Korean path is the runtime target being corrected in this audit branch.

## `$28` evidence — provenance strengthened after review

The initial A-018 wording was too narrow when it said the 16-byte `$28` maximum was established only by the chronicle consumer.

Upstream `release/layout/README.md` explicitly documents:

> Player-name substitutions can use at most 16 encoded bytes under the eight-character input and 17-byte C-string storage contract.

`internal/layout/rules.go` encodes the same model as:

- `playerNameMaxCharacters = 8`
- `playerNameMaxEncodedBytes = playerNameMaxCharacters * 2`

Therefore the best current repository evidence treats **16 encoded bytes as a player-name input/storage bound, not merely a chronicle-local display allowance**.

This supports using 16 bytes as the known maximum for `$28` when `$28` is the player-name substitution in C5. It still does **not** prove that the C5 expansion destination itself is exactly the same 256-byte page buffer being modeled, nor that there is no earlier smaller scratch/staging buffer.

## `$15` / ID 10010 evidence

`$15` is inline in all 438 observed uses and corpus contexts show heterogeneous runtime labels/strings rather than a single fixed semantic type.

ID 10010 is genuinely:

`<value:$15>여 ...`

so the runtime value is directly followed by Korean renderer bytes. But direct inline-value/text adjacency is common: **5,742** occurrences exist in the current corpus.

Generic adjacency is therefore demoted. The remaining high-value question is the **runtime source, maximum encoded length, and destination/staging-buffer contract of `$15` in the character-creation-prompt path**.

## `$33` select grammar caveat

Both inherited production logic and the original forensic walker assumed the special `$33` select form has eight arms plus a sink path. Independent review correctly flagged the duplicate unexplained literal in the new audit code.

The forensic implementation now centralizes this as `c5Select33Arms = 8` and explicitly labels it an **inherited upstream grammar assumption whose independent retail-runtime derivation remains OPEN**. The refactor removes duplicated audit magic; it does not claim to newly prove the value 8.

## Evidence assessment

### CONFIRMED

- Current corpus counts above are deterministic for the audited corpus and the known comparison-RHS classifier counterexample count is zero.
- Original `walkC5` page data counts substitution tokens as zero bytes while marking the leaf dynamic.
- Korean C5 validation knows dynamic substitutions remain outside its complete static proof.
- The Korean page-boundary byte omission was real and has been corrected in the audit branch.
- Upstream documentation ties the 16-byte player-name maximum to an eight-character input and 17-byte C-string storage contract.
- ID 10010's `$15` is inline and directly followed by Korean text.
- Direct inline-value/text adjacency is widespread rather than unique to ID 10010.

### STRONG

- C5 dynamic headroom deserves substantially more attention than generic substitution adjacency because 2,548 C5 records contain inline runtime values while ordinary static page accounting excludes their eventual bytes entirely.
- `$28 = 16 encoded bytes` is presently justified as a player-name-wide bound by upstream's documented input/storage contract.

### OPEN

- Whether any real Korean C5 branch/page reaches or exceeds the relevant runtime destination capacity after all actual substitutions.
- Whether the C5 256-byte page buffer is the direct expansion destination or only a later consumer of an earlier scratch/staging buffer.
- Independent retail-runtime derivation of `$33`'s fixed eight-arm select shape.
- Runtime maximum/source/storage contracts for `$15`, `$16`, `$17`, `$1A`, `$1B`, `$2B`, `$24`, and `$25` outside already documented consumers.
- Whether the observed Bad Execution Address originates from C5 expansion, another message consumer, font/resource state, or an interaction among subsystems.

## Meaning for root-cause ranking

The generic theory “a `<value>` directly adjacent to custom Hangul breaks the renderer” is reduced further.

The more precise theory “runtime-expanded text passes static validation but exceeds an intermediate or consumer buffer” is strengthened as a **falsifiable investigation branch**, not declared more probable solely because it is now measurable.

The renderer-slot/font branch remains independently active, and A-019 records a third cross-cutting EBOOT + mapping/font interaction that should not be forced into either branch prematurely.

## Next gates

1. Run the corrected C5 known-expansion forensic report on authenticated retail assets and record pages exceeding 255 bytes under proven maxima.
2. Establish the runtime maximum/source/storage contract for the remaining eight inline opcodes before adding any guessed bounds.
3. Reverse-engineer whether C5 expansion writes directly into the 256-byte page destination or through an earlier scratch/staging buffer.
4. Independently derive or asset-validate the `$33` eight-arm select grammar assumption.
5. Investigate `$15` specifically in the character-creation-prompt consumer path.
6. Keep the role classifier cross-checked against concrete grammar counterexample scans; do not treat its string-prefix heuristic as a parser proof.
