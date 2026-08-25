# Claude review log

Last updated: 2026-08-25

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

The stock and Korean paths duplicate substitution counting, printf-signature validation, reserved-markup checks, and materialization control traversal. Current behavior is consistent and covered by tests, so this does not block the first real-sentence PoC. However, future changes could drift if one path is updated without the other.

Recommended future direction:

- factor shared validation/control traversal into common helpers while injecting only the natural-text encoder/validator difference.

Status: OPEN NON-BLOCKING TECHNICAL DEBT. Address before large-scale production translation/build work, not before the immediate PoC unless another change touches these duplicated rules.

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

## Current next review gate

The slot-audit gate has passed for PoC candidate selection. The next required review target is the implementation of the first end-to-end Korean sentence PoC itself:

- guarded recognition of retail `msgsec001` record 7 / ID 10007;
- preservation of native kana-mode, line-break and end controls;
- custom renderer-byte remap for `테스트 성공`;
- five atlas cell replacements using the audited PoC candidates above;
- streaming ISO copy/patch safety and exact edit boundaries.

The PoC remains unproven until both the implementation review passes and the generated ISO is observed in PPSSPP. The duplicated stock/Korean production projection logic remains open non-blocking technical debt for later production integration.
