# Claude review log

Last updated: 2026-08-26

This file records adversarial review outcomes so later reviews can distinguish already-closed findings from open technical debt. Read it together with `docs/CLAUDE_HANDOFF.md` and `docs/CLAUDE_REVIEW_PROTOCOL.md`.

## Review 1 — Gemini v2 exchange

Reviewed baseline/head: through `a80988f364f1028b01340e2338404e3bb8f1052b`.

Accepted MUST-FIX findings:

1. Generic printf matching treated ordinary prose such as `100% Match` as a protected `% M` token.
2. Gemini v2 stale-source verification reused the externally supplied glossary when rebuilding expected source, allowing glossary modifications to compare against themselves.

Resolution:

- runtime-format protection narrowed to the actual `%s` / `%u` forms with optional positional number;
- v2 external glossary is currently required to match repository-owned canonical state, which is an empty object until a real canonical glossary is wired;
- regression tests were added for both findings.

Status: CLOSED.

## Review 2 — Korean message path before first real-sentence PoC

Review target code head: `cb0586012655fdfe6d01d94d252ba84650a1fa5c`.

Claude verified Review 1 findings were genuinely closed.

### MUST-FIX: unmatched Korean replacement IDs were silently ignored

Affected function:

- `internal/message/compile_korean.go` — `CompileBankKorean`

Failure mode:

- the compiler iterated bank records and looked up replacements by source ID;
- a replacement map entry for an ID not present in that bank was never consumed and never rejected;
- a typo or wrong-section replacement could therefore produce a successful build while leaving the intended Japanese record untouched.

Resolution:

- `CompileBankKorean` now tracks every consumed replacement ID;
- after bank traversal, any replacement ID not present in the bank causes a deterministic error;
- unmatched IDs are sorted before error reporting;
- regression test `TestCompileBankKoreanRejectsUnmatchedReplacementID` covers the failure.

Fix commits:

- `c9f4187afb4f3bc28de98047151af6e7d06e2832` — implementation
- `00f270894ce210f63e3fb97828bfb077f521acf3` — regression test

CI for the regression-test head passed:

- `go test ./...`
- `go vet ./...`
- `./zill check`

Status: CLOSED.

### SHOULD-FIX: duplicated stock/Korean projection logic

Affected areas:

- `internal/message/projection.go`
- `internal/message/korean_materialize.go`

The stock and Korean paths duplicated substitution counting, printf-signature validation, reserved-markup checks, and materialization control traversal. This was later refactored so both paths share the projection/control traversal and differ only where Korean renderer-slot encoding is required.

Status: CLOSED before continued bulk translation.

## Review 3 — renderer slot audit

The first audit review returned `BLOCKED` with one valid MUST-FIX:

- `cmd/zill/korean_font.go::loadRetailBindata` opened both `pa` and `pami` even after finding `data/bindata.dat` in `pa`.

Resolution:

- `pa` is checked first and `pami` is now only a fallback;
- `TestLoadRetailBindataDoesNotRequirePamiWhenFoundInPa` proves that a pa-resident bindata succeeds with no pami files present;
- retail bindata is now directly SHA-256 pinned before scan results can affect candidate selection.

During integration review GPT also found that the first audit version scanned raw `EBOOT.BIN` even though the canonical executable manifest and release pipeline operate on decrypted/plaintext `SYSDIR/BOOT.BIN`. The audit was corrected to authenticate both files, use EBOOT for supported-game binding, and scan authenticated BOOT for executable string references.

The follow-up Claude review verified the pa→pami fix, bindata hash pin, BOOT/EBOOT trust boundary and scanner scope. Final gate decision:

`PASS FOR POC CANDIDATE SELECTION`

This is explicitly not a production-safety determination.

### Empirical audit input and candidate narrowing

The authenticated v0.5 audit bundle contained:

- `SYSDIR/BOOT.BIN` SHA-256 `5e294dc84a7f0d50719ecd26cb24ffb3792f2d9445803690845a8f1fa1cb85a3`;
- `SYSDIR/EBOOT.BIN` SHA-256 `2a52012be00c07512dcde932ff6e9eb9b96912c59dd5a25c7c26ef821c124d68`;
- `data/bindata.dat` SHA-256 `3241fc000f3d52fe8522baaa985fd866e29d64d3a0f23ac4e28b66dee957de3e`;
- retail `zillfont.par` and `jillbtn.par` matching the documented fingerprints.

After the message, fixed-string, BOOT and bindata audit, 212 installed two-byte renderer keys remained as audited candidates.

For the first real-sentence PoC, five candidates were selected and then subjected to the additional exact-byte check Claude required. Each selected CP932 byte pair occurs zero times in both authenticated `BOOT.BIN` and authenticated `bindata.dat`:

| Korean | PAF key | Raw bytes | Retail glyph | Page | Cell | Metrics |
| --- | --- | --- | --- | ---: | --- | --- |
| 테 | `A1E1` | `E1 A1` | 癸 | 1 | 405,123 11x12 | bearing 0,-10; advance 12 |
| 스 | `A1E9` | `E9 A1` | 鬘 | 1 | 450,123 12x11 | bearing 0,-10; advance 12 |
| 트 | `B8E2` | `E2 B8` | 篋 | 1 | 90,273 11x11 | bearing 0,-10; advance 12 |
| 성 | `BBE6` | `E6 BB` | 貊 | 1 | 150,288 12x11 | bearing 0,-10; advance 12 |
| 공 | `BFE6` | `E6 BF` | 豼 | 1 | 465,303 12x11 | bearing 0,-10; advance 12 |

Status: PASS FOR EMPIRICAL POC CANDIDATE USE ONLY. Other UI/script/archive classes remain unaudited and these keys must not yet be described as production-safe or generally reusable.

## Review 4 — first real Korean sentence PoC implementation

Reviewed implementation range: `632421435f2b083f6d4512f9ced336dcb31f2d28` through head `828d38f5e29fe08cc2b2d89c5976b71e2c880502`.

Claude checked the Android PoC implementation against the guarded retail startup record, renderer-key mapping, atlas edit isolation, streaming ISO patch path, source/output handling and regression tests.

Verified points include:

- all five custom message byte pairs map exactly to the intended little-endian PAF renderer keys and documented atlas cells;
- all five 10x10 rasters fit their selected cells and the cells do not overlap;
- the first natural-text segment rewrite cannot overwrite either native `0x0A` line break or the `05 05 05` end terminator;
- later text segments remain byte-identical in the reviewed first-line-only variant;
- `PoCPatcher.copyAndPatch` preserves ISO size and applies only sorted declared edits, with duplicate and unapplied-edit guards;
- the app opens the source read-only, re-inspects immediately before writing, uses a separate output URI and deletes partial output on failure;
- tests independently check later-line/control preservation and global unchanged-byte behavior outside declared edits.

Final gate decision:

`PASS FOR FIRST-SENTENCE POC TEST`

The subsequent PPSSPP test rendered `테스트 성공` correctly while leaving the following two Japanese lines intact, empirically proving the renderer/message-remap path. The short padded PoC text was horizontally offset, confirming that production layout must use rebuilt message lengths/offsets rather than trailing-space padding.

### SHOULD-FIX: full startup message member is not SHA-256 pinned

The PoC Android path guards the selected startup record but does not SHA-256 pin the whole `message/msgsec001.dat` member. This remains relevant only if that narrow in-place PoC pattern is reused; the production compiler is moving toward canonical authenticated bank rebuilding instead.

Status: OPEN NON-BLOCKING POC DEBT; reassess at final production source-authentication gate.

## Review 5 — continued bulk Korean translation gate

Claude reviewed the canonical Korean corpus, Python import/refresh pipeline, control-token grammar, Korean semantic/layout ownership, compiler behavior, slot/font planning and representative overlays after roughly 3.3k Korean records had accumulated.

Returned gate decision:

`BLOCKED`

The decision remains recorded exactly as returned. No second Claude review has been performed. Full findings and remediation details are in `docs/CLAUDE_REVIEW_5.md`.

### MUST-FIX 1: destructive refresh could delete valid reflowed Korean wording

`tools/korean/refresh-japanese-refs.py` used a stale local control regex and included `<line-break>` in destructive control comparison. Because it runs automatically, a legitimate Korean line-break change could delete the entire translated row.

Resolution:

- refresh now imports shared `fixed_tokens` from `tools/korean/control_tags.py`;
- `<line-break>` is excluded from destructive fixed-control comparison;
- `next-packet.py` also imports the shared runtime-control grammar;
- regression tests cover legitimate Korean reflow and the full fixed-control contract.

Historical audit found that the old logic had removed eight rows in commit `7620fdd36aa141d32a2851657d1e6618b71b269d`. Exact source comparison showed:

- `30000` and `60011`: false-positive removals with unchanged Japanese; safely restored;
- `30007`, `30017`, `30018`, `30028`, `30029`, `30030`: canonical Japanese had changed; intentionally left for retranslation rather than restoring stale wording.

Recovery commit: `8e246798baf3fcfd4b95af5fcd5ab0476be1f7f2`.

Status: CLOSED.

### MUST-FIX 2: semantic Korean carried legacy Japanese-derived line breaks

Resolution:

- translator-owned `korean` is now required to contain no `<line-break>`;
- wrapping belongs only to optional build-owned `layout`;
- Python import/apply paths and Go corpus loading enforce the invariant;
- legacy semantic line breaks were deterministically migrated to spaces while old visual positions were retained in `layout`;
- `CompileBankKorean` always compiles semantic text with layout disabled and validates optional layout separately;
- translation edits invalidate stale layout.

Status: CLOSED.

### SHOULD-FIX: Hangul-boundary reflow

Claude noted that whitespace-only reflow could not wrap an unspaced Korean word.

Resolution:

- semantic/layout comparison now tokenizes plain text at Unicode-rune granularity while keeping controls atomic;
- generated layout may normalize a whitespace span, replace a whitespace span with a line break, or insert a line break between two adjacent precomposed Hangul syllables;
- tests reject character/control changes, leading/trailing boundaries and repeated zero-width boundaries.

Fix head: `fc98c5d22de5f44e8833c236074b37f171ffdf9a`.

Status: CLOSED.

### SHOULD-FIX: theoretical color/discard annotation edge

A raw `%c` payload equal to literal `<` or `>` could theoretically produce an emitted `<color:c>` / `<discard:c:$XX>` annotation not matched by the current `[^<>]` recognizer. No such canonical emitted form has been found in repository search.

Status: OPEN NON-BLOCKING ROBUSTNESS ITEM FOR FINAL ISO/CONTROL AUDIT.

### Post-review CI/font state

The source-aware recovery increased the required custom raster set from 931 to 932 runes. CI correctly failed closed on missing `U+AEBE`; the deterministic raster workflow regenerated and validated the catalog, committing `167a9fab788b8a36149983ee95df097297f57715`.

Post-remediation CI at `91a09942ccbdf77c3f2db52324653934d01ef2f7` passed:

- `go test ./...`;
- `go vet ./...`;
- Python syntax checks;
- Python Korean tooling regression tests;
- `./zill check`;
- `./zill korean-check`;
- `./zill korean-font-check`.

Current unique accepted Korean corpus after recovery: 3,304 records; 932 custom renderer glyphs required.

## Current next review gate

Do not spend another Claude review merely to re-check the just-remediated bulk-translation blockers. Continued translation must obey the enforced invariant: preserve fixed runtime controls in exact order, emit no semantic `<line-break>`, and leave wrapping to build-owned layout.

The next high-value adversarial review is the **final integrated Korean ISO/build gate**, covering full production message-bank/font application, source authentication, final renderer-slot safety evidence, layout generation and unchanged-byte/resource guarantees.
