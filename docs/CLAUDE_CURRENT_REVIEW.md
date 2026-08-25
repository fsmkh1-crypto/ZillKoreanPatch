# Current Claude review status — first real Korean sentence PoC

Last updated: 2026-08-25

The Android first-sentence implementation review is complete.

Gate decision:

`PASS FOR FIRST-SENTENCE POC TEST`

Review details and the remaining non-blocking SHA-256 finding are recorded in:

- `docs/CLAUDE_REVIEW_LOG.md` — Review 4;
- `research/first-korean-sentence-poc.md`.

## Reviewed empirical behavior

Current `main` intentionally rewrites **only the first natural-text line** of `message/msgsec001.dat` record index 7 / canonical ID 10007.

Expected PPSSPP opening display:

```text
테스트 성공
我に応ぜよ
我が問いに答え、その魂を我に示せ
```

The second and third Japanese lines, both native `0x0A` line breaks, and the native `05 05 05` end terminator must remain unchanged.

The five custom renderer assignments remain PoC candidates only:

| Korean | PAF key | Raw bytes | Existing PAF cell |
| --- | --- | --- | --- |
| 테 | `A1E1` | `E1 A1` | page1 x405 y123 11x12 |
| 스 | `A1E9` | `E9 A1` | page1 x450 y123 12x11 |
| 트 | `B8E2` | `E2 B8` | page1 x90 y273 11x11 |
| 성 | `BBE6` | `E6 BB` | page1 x150 y288 12x11 |
| 공 | `BFE6` | `E6 BF` | page1 x465 y303 12x11 |

## Open non-blocking finding

`FontExtractor.inspect` does not yet SHA-256 pin the full retail `message/msgsec001.dat` member. The target record is still strongly guarded by `StartupMessage.inspect`, so Claude did not block this empirical PoC test.

Before the same pattern is promoted to the production full-corpus path:

1. obtain the authenticated retail SHA-256 of the full `message/msgsec001.dat` member;
2. add a startup-message hash constant and `validateSourceHash` call;
3. add a regression test proving unrelated changes elsewhere in the member are rejected.

## Immediate next milestone

Generate the PoC ISO with the reviewed Android build and observe it in PPSSPP. If the first line renders `테스트 성공` and the following two Japanese lines remain unchanged, record the renderer/message-remap path as empirically successful.

Do not interpret a successful observation as production-wide slot safety, full-corpus capacity or final localization readiness.
