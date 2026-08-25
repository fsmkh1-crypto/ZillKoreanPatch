# Claude handoff: Zill O'll Infinite Plus Korean patch

Last updated: 2026-08-25

This is the working handoff for Claude. Read this before reviewing or changing code.

## 1. Repository and scope

Canonical working repository:

- `fsmkh1-crypto/ZillKoreanPatch`

Do **not** mix this project with `fsmkh1-crypto/ZillOcrOverlay`. `ZillOcrOverlay` is a separate real-time OCR translator project and is not the Korean patch implementation.

Target game:

- Zill O'll Infinite Plus (PSP)
- `ULJM-05410`
- retail version `1.03`

Upstream base:

- `HK47196/zill`
- pinned upstream baseline: `a98d9ce29f361d666ec23da0dcfd351f24537ffd`

Preserve upstream licensing/provenance. Code is GPL-3.0-or-later; translation content is CC BY-SA 4.0.

## 2. Collaboration roles

Current division of labor:

- GPT: lead/integration, GitHub canonical state, code, font/encoding, build/Android patcher, final merge decisions.
- Gemini: external translation worker only. It has no GitHub access. Input/output is file-based JSONL.
- Claude: adversarial reviewer / QA / design reviewer. Prefer review comments or a branch/PR over direct edits to `main` unless explicitly asked.

The final arbiter is not model consensus. Retail bytes, deterministic tests, hashes, and actual-game PPSSPP behavior win.

When reviewing, distinguish clearly between:

- **observed/proven from retail or actual game**
- **inference**
- **proposal**

If challenging a proven claim, provide a concrete counterexample, byte-level evidence, or a reproducible failing test.

## 3. Empirically proven renderer/font facts

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

Corrected PAF geometry (these values supersede earlier mistaken offsets):

- PAF begins at `jillbtn.par + 0x4490`
- PAF size `0x149D0`
- exactly 2637 complete records
- record size `0x20`
- records begin at PAF `+0x30`
- page distribution: page0=1156, page1=1156, page2=325
- keys are strictly ascending, no duplicates
- BST integrity was validated with all 2637 records reachable

Important semantic point:

- PAF `+0x00 u16` is a **renderer lookup key derived from CP932/Shift-JIS bytes**, not Unicode.
- Example: Japanese `の` encodes as CP932 bytes `82 CC`, so the stored little-endian renderer key is `0xCC82`.

Authenticated `の` record used by the first PoC:

- record index 2034
- key `0xCC82`
- size 10x10
- atlas coordinate x=421, y=379
- bearing x=1, y=-9
- advance 12
- page 1

See `research/korean-font-poc-reproduction.md` for the preserved first-PoC byte delta and retail details.

### Actual-game PoC status

The first single-glyph PoC was observed in the real game under PPSSPP: `の -> 가`.

The later eight-glyph smoke test was also observed in the real game under PPSSPP, with all eight Hangul glyphs visible on the fixed new-game screen and repeated mappings rendering consistently:

- `の -> 가`
- `無 -> 나`
- `我 -> 다`
- `応 -> 라`
- `答 -> 마`
- `魂 -> 바`
- `示 -> 사`
- `者 -> 아`

This proves multi-glyph atlas addressing/swizzle and renderer slot reuse in the actual game, not just in an atlas viewer.

**Documentation warning:** `research/korean-font-poc-reproduction.md` was written before the later 8-glyph screenshot verification and still says that the 8-glyph build required on-device observation. That sentence is stale. The screenshot itself has not yet been committed as a repository artifact.

The 8-glyph smoke architecture is test-only because it globally replaces common Japanese glyph bitmaps. Final architecture must not sacrifice stock Japanese glyphs.

## 4. Intended final Korean renderer architecture

Current intended approach:

1. Measure the actual unique Korean Unicode runes required by the translated corpus.
2. Select reusable existing **two-byte CP932 renderer keys** that are safe to repurpose.
3. Deterministically map Korean runes to those existing keys.
4. Emit those custom CP932 byte pairs when compiling Korean message text.
5. Put matching Hangul bitmaps/metrics behind those same PAF keys.
6. Preserve record count, renderer key ordering/BST structure, and stock Japanese glyphs wherever possible.

Relevant existing code:

- `internal/cp932/glyphkey.go`
- `internal/koreanslots/slots.go`
- `cmd/zill/korean_font.go`
- `internal/zillfont/paf.go`
- `internal/message/projection.go`

`internal/koreanslots` already supports deterministic rune -> renderer-key allocation and custom encoding, but the production message materialization path still uses stock `cp932.Encode` and is **not yet wired** to the Korean mapping.

### Capacity caution

Prior analysis found roughly 1371 two-byte renderer slots unreferenced by the upstream English translation corpus. Treat this as capacity evidence only, **not proof that all such slots are safe to overwrite in the retail game**. UI/system/runtime strings outside the translated corpus may still reference some keys.

A safe allocator needs a stronger definition of “reusable” (for example, a whole-game reference scan or a conservative reserved-key policy).

## 5. Gemini translation exchange: why v2 exists

Gemini initially claimed it could preserve inline control tokens such as `<end>`, `<line-break>`, `<value:$28>`, etc. A real 10-record test showed this was unreliable:

- it deleted `<end>` / `<line-break>`
- it turned `<value:$28>` into a Markdown/search-link-like construct
- it broke the requested JSON schema in other ways

Therefore the design changed: **Gemini must never receive or return literal game control tokens.**

## 6. Current Gemini v2 implementation

The segment-safe v2 exchange is now implemented and CI-passing.

Key files:

- `internal/translationexchange/v2.go`
- `internal/translationexchange/v2_json.go`
- `internal/translationexchange/strict_json.go`
- `internal/translationexchange/v2_test.go`
- `cmd/zill/gemini_exchange.go`
- `.github/workflows/gemini-batch.yml`

Current schema name:

- `zill-gemini-v2`

Core design:

- literal controls/placeholders/printf tokens are removed before external export
- exact locked parts remain maintainer-side only
- natural-language fragments are exported as indexed segments
- Gemini returns only indexed Korean segments plus QA metadata
- segment count/order/index must match exactly
- returned segments containing protected tokens are rejected
- accepted results are reconstructed using repository-owned locked parts
- the checker regenerates canonical source from the current repository and rejects stale/modified source JSONL

The v2 test suite covers, among other things:

- no control leakage into exported segments/context/reference
- exact reconstruction from translated segments + locked source controls
- strict JSON contract
- null arrays / missing fields / unknown fields / Markdown / blank lines rejection
- segment length/index drift rejection
- protected-token injection rejection

Recent v2 commit chain:

- `4b68c93c6b736e81a3bb0fc08dc530ab44beb974` — segment-safe v2 core
- `0a4b4063feb12b165af957c186581a3cbb6387b8` — strict v2 JSONL IO
- `37b80e15ac39865604b6d995f5539eaf85550cbe` — exact v2 response schema
- `fd46b92fddb0a5d327b7916aa72ae2e0b841bb0d` — CLI switched to v2
- `9f41c27716d8b5ea738079bdda860a82f9556254` — v2 tests
- `8e538c43b16ed9c9642e97d494975fe696a992e1` — checker wired to canonical project root
- `a80988f364f1028b01340e2338404e3bb8f1052b` — control-free dialogue export workflow / section selection

CI on the final v2 state passed `go test ./...`, `go vet ./...`, and existing `./zill check`.

A 10-record actual-dialogue v2 export from section 187 was also generated successfully by the workflow.

### Documentation warning

`docs/gemini-translation-exchange.md` was written for the earlier v1 design and is now partly stale. Treat v2 code/tests as authoritative until that document is updated.

## 7. Important v2 review questions for Claude

Please review the v2 design adversarially, especially these questions:

1. **Semantic placeholder context:** literal runtime placeholders are intentionally hidden from Gemini. Does the current control-free `full_text`/segment structure remove too much semantic information for natural translation (for example a player-name vocative)? If so, propose safe metadata that conveys semantics without exposing mutable control syntax.
2. **Line-break ownership:** v2 currently treats protected tokens as locked source material. Review how this interacts with the existing message projection/layout system, which already distinguishes semantic text from layout/reflow. We must not accidentally freeze Japanese line wrapping into final Korean layout if the layout layer is supposed to reflow it.
3. **Reconstruction ordering:** inspect `SplitV2` / `LockedPart.AfterSegment` / `ReconstructV2` for edge cases involving leading controls, adjacent controls, whitespace-only runs, repeated controls, and records with no ordinary text.
4. **Stale-source verification:** confirm that rebuilding canonical v2 source and comparing it to the returned source file cannot be bypassed by glossary/context fields or ordering details.
5. **Schema strictness:** verify there is no permissive JSON path that can silently accept malformed or extra data.

If you find a concrete problem, cite file + line/function and preferably add a focused failing test before proposing the fix.

## 8. Work that was intentionally stopped / not yet implemented

The user stopped the next implementation step because the session was taking too long. Do not assume the following exists yet:

- production unused-slot safety scan
- final Korean slot allocation based on a completed corpus
- production message-byte remapping to Korean custom keys
- wiring `koreanslots.Encode` into the message compiler/materializer
- final PAF metrics/atlas packing for a full Korean subset
- canonical Korean TOML/import format
- full translation corpus
- final Korean line-wrapping/layout tuning
- end-to-end production ISO build containing a real Korean sentence via unused slots

The next major renderer milestone is: **render the first real Korean sentence in the actual game using unused/reusable renderer slots and message-byte remapping, without globally replacing common Japanese glyphs.**

## 9. Suggested Claude work order

Default reviewer task order:

1. Review the Gemini v2 changes listed above and report only concrete correctness/design issues.
2. Review the proposed unused-slot + message-remap architecture against current `koreanslots`, `message`, `zillfont`, and PAF code.
3. Identify the minimum safe implementation boundary for the first real Korean-sentence PoC.
4. Do not redesign proven retail geometry unless evidence contradicts it.
5. If coding is requested, use a branch/PR with focused tests; avoid large unrelated refactors.

## 10. Useful references

- `research/korean-font-poc-reproduction.md`
- `research/upstream-provenance.md`
- `internal/zillfont/paf.go`
- `internal/koreanslots/slots.go`
- `internal/cp932/glyphkey.go`
- `internal/message/projection.go`
- `internal/translationexchange/v2.go`
- `internal/translationexchange/v2_json.go`
- `internal/translationexchange/v2_test.go`
- `cmd/zill/gemini_exchange.go`

When in doubt, prefer current `main`, authenticated retail data, current tests, and actual-game PPSSPP observations over older chat assumptions or stale prose docs.
