# Current Claude review request — first real Korean sentence PoC

Last updated: 2026-08-25

Read these first:

- `docs/CLAUDE_HANDOFF.md`
- `docs/CLAUDE_REVIEW_PROTOCOL.md`
- `docs/CLAUDE_REVIEW_LOG.md`
- `research/first-korean-sentence-poc.md`

The renderer-slot audit follow-up has already received Claude's gate decision:

`PASS FOR POC CANDIDATE SELECTION`

Do not re-review the already-closed pa→pami blocker unless this new implementation regressed it. This review is the next gate: the Android implementation that turns five audited candidate renderer keys plus one guarded retail startup record into the first end-to-end Korean sentence ISO PoC.

A pass here authorizes empirical PPSSPP testing only. It must not be interpreted as production-wide slot safety.

## Current implementation target

Review current `main`, especially:

- `android-patcher/app/src/main/java/com/fsmkh1/zillfontdump/StartupMessage.java`
- `android-patcher/app/src/main/java/com/fsmkh1/zillfontdump/FontExtractor.java`
- `android-patcher/app/src/main/java/com/fsmkh1/zillfontdump/PoCPatcher.java`
- `android-patcher/app/src/main/java/com/fsmkh1/zillfontdump/MainActivity.java`
- `android-patcher/app/src/test/java/com/fsmkh1/zillfontdump/StartupMessageTest.java`
- `android-patcher/app/src/test/java/com/fsmkh1/zillfontdump/PoCPatcherTest.java`
- `research/first-korean-sentence-poc.md`

Relevant implementation commits begin at `632421435f2b083f6d4512f9ced336dcb31f2d28` and continue through the current head.

The prior eight-glyph smoke code is no longer the active PoC path; compare with history only when useful for verifying the already-proven atlas swizzle/raster method.

## Intended PoC behavior

The retail target is:

- archive member: `message/msgsec001.dat`;
- record index: 7;
- canonical ID: 10007;
- contributor display:
  `汝、無限のソウルを持つ者よ<line-break>我に応ぜよ<line-break>我が問いに答え、その魂を我に示せ<end>`.

The patch should make those three displayed lines read:

```text
테스트 성공
테스트 성공
테스트 성공
```

The five custom renderer assignments are:

| Korean | PAF key | Raw bytes | Existing PAF cell |
| --- | --- | --- | --- |
| 테 | `A1E1` | `E1 A1` | page1 x405 y123 11x12 |
| 스 | `A1E9` | `E9 A1` | page1 x450 y123 12x11 |
| 트 | `B8E2` | `E2 B8` | page1 x90 y273 11x11 |
| 성 | `BBE6` | `E6 BB` | page1 x150 y288 12x11 |
| 공 | `BFE6` | `E6 BF` | page1 x465 y303 12x11 |

Each chosen raw pair was additionally checked to occur zero times in authenticated retail `BOOT.BIN` and `data/bindata.dat`, as required by the previous Claude review. They remain PoC candidates, not production-safe slots.

## Important retail-message detail

Do not assume the TOML display projection is the retail byte spelling.

`internal/corpus/bank.go::displayText` shows that retail records can contain `ESC K/H/k` kana-mode controls and half-width kana bytes that are omitted/normalized in contributor display, while:

- `0x0A` projects to `<line-break>`;
- `05 05 05` projects to `<end>`.

`StartupMessage.inspect` therefore parses the actual retail record, reproduces the relevant display semantics, and requires an exact expected displayed source with exactly two native line breaks and one native end marker. It records the three natural-text byte segments.

`StartupMessage.patchEdits` is supposed to edit only those three natural-text byte spans. It writes the 11-byte custom Korean line at the start of each span and fills the unused part of that span with ASCII spaces. It must leave the original line-break controls and `05 05 05` end terminator untouched. No message offset table entry or archive member size should change.

## Review questions — MUST answer

1. **Retail display parser fidelity**
   - Is the subset implemented in `StartupMessage.scanDisplay` sufficient and fail-closed for this exact guarded record?
   - Check direct Shift_JIS text, double-byte lead/trail handling, half-width kana, dakuten/handakuten, `ESC K/H/k`, NFKC normalization and katakana→hiragana conversion.
   - Could a malformed or differently controlled record still produce the expected display and yield unsafe segment boundaries?

2. **Control preservation**
   - Prove that the two `0x0A` line breaks and `05 05 05` end terminator cannot be overwritten by `patchEdits` / `PoCPatcher.absoluteEdits`.
   - Check off-by-one behavior at every segment boundary.
   - Check that filling unused natural-text bytes with `0x20` cannot leak across a control or alter later records.

3. **Message-bank integrity**
   - Verify record offsets, member size and all downstream record positions stay unchanged.
   - Determine whether leaving original bytes after `<end>`/NUL inside the record span can matter.
   - Check the source guard is strong enough that a wrong section/record/member cannot be silently patched.

4. **Renderer byte/key correspondence**
   - Verify the custom message pairs correspond exactly to the PAF keys as little-endian renderer keys (`E1 A1` → `A1E1`, etc.).
   - Verify all five atlas coordinates/page/cell sizes match the documented retail PAF records.
   - Verify each 10x10 bitmap fits the selected cell and that the `(1,1)` placement/metrics assumption is consistent with the earlier empirically proven Hangul raster path.

5. **Atlas edit isolation**
   - Review `buildFontEdits`, swizzle calculation, nibble masks and merged-byte logic.
   - Prove it edits only pixels in the five declared source cells.
   - Check that cells cannot overlap or produce colliding byte edits unexpectedly.

6. **Streaming ISO edit correctness**
   - Review absolute-offset construction for both font and message edits.
   - Check sorting, duplicate detection, chunk-boundary handling, source-size check and unapplied-edit detection.
   - Look for any way an edit before the current streaming position could be silently skipped.

7. **Source/output safety**
   - Verify the app still opens the source read-only, uses a distinct output URI, re-inspects before writing, preserves ISO size, and deletes partial output on failure.
   - Check whether the inspection comparison is strong enough for TOCTOU given that `fresh` performs all hashes/source guards again.

8. **Tests**
   - Assess whether `StartupMessageTest` and `PoCPatcherTest` actually exercise the intended invariants rather than merely mirroring implementation.
   - Give a concrete missing regression test for every MUST-FIX finding.

9. **Safety language**
   - Ensure docs/UI/code do not call these five keys production-safe, globally unused or generally reusable.

## Known out-of-scope/non-blocking items

Unless the new implementation directly interacts with them, do not block this PoC for:

- full whole-game UI/script/archive slot audit;
- full Korean corpus translation;
- final production atlas packing/metrics;
- duplicated stock/Korean Go projection logic already logged as technical debt;
- Gemini translation-exchange redesign.

## Required output

For each finding provide:

- severity: MUST-FIX / SHOULD-FIX / NOTE;
- file + function;
- concrete failure mode;
- minimal reproducible test or byte example when possible;
- fix direction;
- whether it blocks generating/testing the first-sentence PoC ISO.

End with exactly one gate decision:

- `BLOCKED`; or
- `PASS FOR FIRST-SENTENCE POC TEST`.

If GitHub write access is unavailable, output the complete review in chat for manual transfer to GPT.
