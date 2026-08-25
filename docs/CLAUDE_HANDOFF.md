# Claude handoff: Zill O'll Infinite Plus Korean patch

Last updated: 2026-08-25

Read this first, then `docs/CLAUDE_REVIEW_PROTOCOL.md`. The project now assumes Claude review at high-risk milestone boundaries.

## 1. Repository and scope

Canonical repository:

- `fsmkh1-crypto/ZillKoreanPatch`

Do not mix this with:

- `fsmkh1-crypto/ZillOcrOverlay` — separate real-time OCR translator project.

Target:

- Zill O'll Infinite Plus (PSP)
- `ULJM-05410`
- retail version `1.03`

Upstream base:

- `HK47196/zill`
- pinned baseline `a98d9ce29f361d666ec23da0dcfd351f24537ffd`

Preserve upstream licensing/provenance. Code is GPL-3.0-or-later; translation content is CC BY-SA 4.0.

## 2. Current collaboration roles

- GPT: lead developer/integrator, canonical GitHub state, font/encoding/build/Android work, primary Korean translator, final merge decisions.
- Claude: adversarial reviewer / QA / design reviewer. Review is expected before major runtime/ISO milestones.
- Gemini: optional overflow or secondary translation worker only; no longer on the critical path.

Retail bytes, deterministic tests, hashes, and actual PPSSPP behavior outrank model consensus.

## 3. Proven renderer/font facts

Authenticated retail members:

- `font/zillfont.par`
  - size `0x80470` / 525,424 bytes
  - SHA-256 `0d3d6d2648870e87a01636cdfc7cc7af8100ea40b71e5ed05f82ac197606584a`
  - PAA index 13611
  - `pa.arc` offset `0x3D8F520`
- `2d/font/jillbtn.par`
  - size `0x18E60` / 101,984 bytes
  - SHA-256 `95b48379092db4db72f890d5a221ba8c4094dd438cb4c4eba98eb5520c7b17aa`
  - PAA index 13612
  - `pa.arc` offset `0x3E0F990`

Correct PAF geometry:

- PAF begins at `jillbtn.par + 0x4490`
- PAF size `0x149D0`
- 2637 complete records
- record size `0x20`
- records begin at PAF `+0x30`
- page distribution: page0=1156, page1=1156, page2=325
- keys strictly ascending, no duplicates
- all 2637 records reachable from the validated BST

PAF `+0x00 u16` is a renderer lookup key derived from CP932/Shift-JIS bytes, not Unicode.

Example:

- `の` -> CP932 bytes `82 CC` -> stored little-endian renderer key `0xCC82`.

Authenticated `の` record used by first PoC:

- record index 2034
- key `0xCC82`
- size 10x10
- x=421, y=379
- bearing x=1, y=-9
- advance 12
- page 1

See `research/korean-font-poc-reproduction.md` for the first PoC byte delta.

## 4. Actual-game PoC status

Observed in PPSSPP:

- single glyph `の -> 가` works;
- later 8-glyph smoke also works with all eight Hangul visible and repeated keys rendering consistently.

Smoke mappings:

- `の -> 가`
- `無 -> 나`
- `我 -> 다`
- `応 -> 라`
- `答 -> 마`
- `魂 -> 바`
- `示 -> 사`
- `者 -> 아`

This proves Hangul raster display, multi-glyph atlas addressing/swizzle, repeated-key rendering, and the renderer-key concept in the actual game.

It does **not** prove that common Japanese glyphs can be sacrificed in production. The smoke build globally replaced common Japanese bitmaps and is test-only.

## 5. Intended production renderer architecture

Current target architecture:

1. determine the actual Korean rune set needed by the localized corpus;
2. identify existing two-byte CP932 renderer keys that are genuinely safe to repurpose;
3. deterministically map Korean runes to those keys;
4. emit the corresponding byte pairs when compiling Korean text;
5. install matching Hangul bitmap/metrics behind those PAF keys;
6. preserve stock Japanese and the existing PAF/BST record structure wherever possible.

Core components already present:

- `internal/cp932/glyphkey.go`
- `internal/koreanslots/slots.go`
- `internal/zillfont/paf.go`
- `internal/message/projection.go`
- `internal/message/korean_materialize.go`
- `internal/message/compile_korean.go`
- `cmd/zill/korean_font.go`

## 6. Previous Claude review and fixes

Previous review baseline:

- `a80988f364f1028b01340e2338404e3bb8f1052b`

Claude identified two concrete Gemini v2 bugs:

1. generic printf regex falsely matched ordinary prose such as `100% Match` as `% M`;
2. glossary stale-source validation reused the external glossary when building the expected source, so glossary modifications could compare equal to themselves.

Both findings were accepted.

Fixes now present:

- `internal/translationexchange/v2_review_fixes.go`
- `internal/translationexchange/v2_review_fixes_test.go`

Current behavior:

- protected printf syntax is aligned to the runtime message compiler's `%s` / `%u` form (with optional positional number);
- current v2 external source requires the repository-owned canonical glossary state (currently an empty object), so arbitrary external glossary injection is rejected;
- regression tests cover both failures.

## 7. New Korean message path added after that review

### Mapping-aware validation and materialization

`internal/message/korean_materialize.go` adds a Korean-specific path where validation and byte emission use the exact same `koreanslots.Mapping`.

Important properties:

- unmapped Hangul fails closed;
- `<value:$XX>` and layout breaks are still handled through existing projection/control semantics;
- stock CP932 runes cannot be silently overridden by a custom Korean mapping;
- raw control characters / half-width kana / reserved markup rules remain enforced;
- final natural-language bytes are emitted via `koreanslots.Encode`.

Tests:

- `internal/message/korean_materialize_test.go`

### Selective Korean bank compiler

`internal/message/compile_korean.go` adds `CompileBankKorean`.

Purpose:

- only explicitly selected Korean records are rebuilt with Korean renderer-slot bytes;
- every non-selected record is copied byte-identically from retail source;
- runtime bank capacity and wide-offset layout are still enforced;
- replacement layout uses the existing semantic-preservation rules.

Tests:

- `internal/message/compile_korean_test.go`

This is intentionally separate from the existing English/stock `CompileBank` path so the first real-sentence PoC remains narrow and reversible.

## 8. Slot-safety status

`cmd/zill/korean_font.go` was made more conservative.

The old diagnostic effectively treated "unused by current English translation text" as reusable. That claim was too strong.

Current diagnostic excludes keys referenced by either:

- current English message text; or
- retail Japanese message text.

It now reports the remainder only as **candidate** two-byte slots and prints an explicit warning that UI/ELF/fixed-data references outside message banks have not yet been excluded.

Therefore:

- message-corpus-unreferenced != proven safe;
- no candidate may be called "safe/reusable" for production until a wider reference audit or another justified conservative policy exists.

## 9. Gemini v2 status

Gemini v2 remains available as an optional external translation exchange, but GPT is now the primary translator.

The v2 design still matters as a safe backup path:

- controls are removed before external translation;
- only indexed natural-language segments are returned;
- repository-owned locked controls are reconstructed locally;
- strict JSON/schema/segment identity checks remain in place.

Do not spend review time redesigning Gemini unless a concrete correctness issue affects current work. Renderer/message/slot safety is now the higher-priority path.

## 10. Current CI state

Latest code milestone before the review-protocol documentation commit:

- `cb0586012655fdfe6d01d94d252ba84650a1fa5c`

At that state, CI passed:

- `go test ./...`
- `go vet ./...`
- `./zill check`

`./zill check` reported 43,116 records, with existing upstream translation states intact.

## 11. CURRENT CLAUDE REVIEW REQUEST

Review code changes from the previous reviewed baseline:

- base: `a80988f364f1028b01340e2338404e3bb8f1052b`
- code head: `cb0586012655fdfe6d01d94d252ba84650a1fa5c`

Primary review targets:

- `internal/translationexchange/v2_review_fixes.go`
- `internal/translationexchange/v2_review_fixes_test.go`
- `internal/message/korean_materialize.go`
- `internal/message/korean_materialize_test.go`
- `internal/message/compile_korean.go`
- `internal/message/compile_korean_test.go`
- `cmd/zill/korean_font.go`

Please review adversarially for concrete failures, especially:

1. Do the two previous review fixes actually close the reported bugs without introducing a bypass/regression?
2. Can `SplitSemanticKorean` validation and `MaterializeKorean` encoding disagree for any rune/control/layout case?
3. Can a malicious/incorrect mapping overwrite stock CP932 behavior or emit invalid renderer-key bytes?
4. Does `CompileBankKorean` truly keep every non-selected retail record byte-identical and preserve ID/order/offset/capacity invariants?
5. Does `CompileBankKorean` need additional rejection for replacement IDs that are not present in the bank, duplicate/ambiguous selection, or cross-bank misuse?
6. Are runtime substitutions, printf signatures, fixed controls, kana controls, authored line breaks, and generated layout handled consistently with the existing production projection path?
7. Is the current Japanese+English message scan a valid conservative improvement, and are there any cases where it can incorrectly label an actually referenced renderer key as a candidate?
8. What is the minimum safe next step to audit UI/ELF/fixed-data references before repurposing a candidate slot?

For every finding, use severity + file/function + concrete failure + minimal reproducer/test + fix direction. Prefer a focused failing test over a broad redesign.

## 12. Next milestone blocked on this review

Do not yet claim any renderer slot is production-safe.

After this review and any fixes, the next milestone is:

**first real Korean sentence in the actual game using candidate-unused renderer keys plus Korean message-byte remapping, without globally replacing common Japanese glyphs.**

That milestone should proceed only after the slot-safety audit is strengthened enough to justify the chosen PoC keys.

## 13. References

- `docs/CLAUDE_REVIEW_PROTOCOL.md`
- `research/korean-font-poc-reproduction.md`
- `research/font-archive-and-korean-encoding.md`
- `internal/zillfont/paf.go`
- `internal/koreanslots/slots.go`
- `internal/cp932/glyphkey.go`
- `internal/message/projection.go`
- `internal/message/korean_materialize.go`
- `internal/message/compile_korean.go`
- `cmd/zill/korean_font.go`

When prose conflicts with current code/tests or authenticated retail evidence, prefer current code/tests and authenticated evidence, then update the prose.
