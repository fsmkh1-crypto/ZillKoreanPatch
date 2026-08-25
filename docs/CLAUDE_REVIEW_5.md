# Claude Review 5 — continued bulk Korean translation gate

Date: 2026-08-26
Branch: `translation/section001-batch2`
Reviewer decision: `BLOCKED`

This review was requested after the first empirical Korean renderer/message PoC had passed and the repository had accumulated roughly 3.3k Korean records. The gate was intentionally about **durability of further bulk translation**, not final ISO release wiring.

## MUST-FIX 1 — destructive refresh used the wrong control contract

Claude found that `tools/korean/refresh-japanese-refs.py` maintained a narrower local runtime-control regex and included `<line-break>` in its destructive comparison. Because the script runs automatically after bulk result application, a legitimate Korean reflow could be classified as a control mismatch and the entire translated row removed.

### Resolution

- `refresh-japanese-refs.py` now imports the repository-wide `fixed_tokens` implementation from `tools/korean/control_tags.py`.
- `<line-break>` is therefore excluded from destructive fixed-control comparison.
- `tools/korean/next-packet.py` now imports the shared `RUNTIME_CONTROL_RE` as well, removing the other Python grammar copy noted by the review.
- Python regression tests verify that a Korean row with different line-break placement survives refresh and that the full fixed-control grammar is recognized.

Primary fix commits:

- `dc9a49ac895e90d34d47568e81e2d6974120102d`
- `2221f706aae51425d7ee7abcd8c829dea0e7c23c`
- `f7a18e12d9574ee32cbf6985a4dea7c8e2426cfb`

### Historical destructive-run audit

The old destructive logic had in fact run. Commit `7620fdd36aa141d32a2851657d1e6618b71b269d` removed eight Korean overlay rows:

- `30000`
- `30007`
- `30017`
- `30018`
- `30028`
- `30029`
- `30030`
- `60011`

A source-aware recovery audit compared each removed row with its historical parent and the current canonical Japanese source. Two rows (`30000`, `60011`) were true false-positive removals with unchanged Japanese and were restored. The other six had genuinely changed canonical Japanese and therefore remain queued for retranslation rather than being restored against stale source text.

Recovery commit:

- `8e246798baf3fcfd4b95af5fcd5ab0476be1f7f2` — restored the two safe rows as semantic Korean plus separate layout metadata.

The recovery workflow reported:

- 2 false-positive rows restored;
- 6 stale-source rows left for retranslation;
- 3,304 unique accepted Korean records after recovery;
- Python regression suite passed.

Status: **CLOSED**.

## MUST-FIX 2 — semantic Korean and layout line breaks were conflated

Claude found that `CompileBankKorean` treated any `<line-break>` in `KoreanRecord.Text` as translator-authored and therefore prevented later build-owned reflow. Existing rows had inherited Japanese break positions, and future imports could continue adding the same shape.

### Resolution

- Canonical translator-owned `korean` text is now required to contain **no `<line-break>`**.
- `layout` remains the only field allowed to contain build-owned wrapping.
- Both Python import paths reject semantic Korean containing `<line-break>`.
- The Go Korean corpus loader and `WithKorean` enforce the same invariant.
- `CompileBankKorean` always materializes semantic `Text` with layout disabled, then independently validates and materializes optional `Layout`.
- A deterministic one-time migration removed inherited line-break tokens from semantic Korean, converting each legacy break boundary to a semantic space and preserving the previous visual break positions in `layout`.
- Regression tests cover semantic-break rejection, separate layout acceptance, stale-layout invalidation, and compiler behavior.

Primary fix commits:

- `bc5596183a4a9f8b8bebe87f7cfa91dbdc93b316`
- `b3a46ad945f06aedfbf9a3aefe4a4ff2d2dd0702`
- `e5f1d339819bacdb45c0f16997847bb932b0849c`
- `1def4b9ffc70ebe61bec2db05a43b9aa4cd617b9`
- `7d7f03456f0193b0d2ffd1ddc5deb9e511d296ef`
- `355b8c90e23babedada179fb6a36605b78f69370`
- `209dcc30e616c1a88a7e86cf2b57972883a5c4ba`
- `116f4bdb4a3a71264f7b7be52efeb37e437ae2ad`

Status: **CLOSED**.

## SHOULD-FIX 3 — reflow inside unspaced Hangul runs

Claude noted that the old `preservesSemantics` implementation could only replace complete whitespace spans with `<line-break>` and could not wrap a long unspaced Hangul word.

### Resolution

`internal/message/compile.go` now tokenizes ordinary text at Unicode-rune granularity while keeping annotated controls atomic. Generated layout may:

1. normalize one complete semantic whitespace span to another whitespace span;
2. replace a complete semantic whitespace span with one line break; or
3. insert a zero-width line break between two adjacent precomposed Hangul syllables.

It still rejects character insertion/deletion/reordering, runtime-control reordering, leading/trailing boundaries and repeated boundaries at the same semantic position.

Fix/test commits:

- `640ce773a790d5d9586e05311fad733e17aeac99`
- `e934c4b0494e50a87818523ce9d53d489ac88085`
- `fc98c5d22de5f44e8833c236074b37f171ffdf9a`

Status: **CLOSED before final layout integration**.

## SHOULD-FIX 4 — duplicate Python runtime-control regexes

Closed as part of MUST-FIX 1: `refresh-japanese-refs.py` and `next-packet.py` now use `control_tags.py` as the shared grammar source.

Status: **CLOSED**.

## SHOULD-FIX 5 — theoretical `%c` angle-bracket byte edge

The review noted that a raw color/discard byte equal to literal `<` or `>` could create an emitted annotation that the current `[^<>]` recognizer cannot match. Repository search found no canonical emitted `<color:<...`, `<color:>...`, or analogous discard form, so this is not presently demonstrated in the retail corpus.

Status: **OPEN NON-BLOCKING ROBUSTNESS ITEM FOR FINAL ISO/CONTROL AUDIT**.

## Post-review font synchronization

Recovering the two valid rows increased the required custom renderer set from 931 to 932 runes. CI correctly failed closed on the one missing raster (`U+AEBE`). The Korean raster workflow was explicitly retriggered from a Korean-overlay audit change and regenerated/validated the catalog. It committed the refreshed catalog as:

- `167a9fab788b8a36149983ee95df097297f57715` — `Update generated Korean raster catalog`.

## Current interpretation

Claude's recorded decision remains `BLOCKED` because that was the result returned before these fixes. There has been no second Claude review and this file does not rewrite that decision retroactively.

The two MUST-FIX conditions identified by the review are now implemented and regression-tested. The only known review item still open is the narrow, unconfirmed color/discard annotation robustness edge, which Claude itself classified as SHOULD-FIX before final ISO build rather than a bulk-translation blocker.
