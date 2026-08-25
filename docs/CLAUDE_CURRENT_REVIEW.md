# Current Claude review status — production-path gate pending

Last updated: 2026-08-25

The Android first-sentence implementation review is complete and the empirical PPSSPP test has now passed.

Previous gate decision:

`PASS FOR FIRST-SENTENCE POC TEST`

Observed PPSSPP behavior:

```text
테스트 성공
我に応ぜよ
我が問いに答え、その魂を我に示せ
```

The first Korean line rendered correctly and the following two Japanese lines remained intact. The renderer/message-remap PoC is therefore empirically proven for the selected five candidate keys.

See:

- `docs/CLAUDE_REVIEW_LOG.md` — Review 4 and subsequent empirical result;
- `research/first-korean-sentence-poc.md`.

## No immediate Claude review requested

Do **not** spend another adversarial review on the already-passed PoC unless a regression directly changes that path.

The next Claude review should be reserved for the higher-value **integrated production Korean translation path** after GPT has implemented and tested it.

## Work to complete before that review

The production-path review should not be requested until the following are substantially implemented:

1. repository-owned canonical Korean corpus storage/import, with Japanese authoritative source and protected structural controls;
2. production renderer-slot allocation driven by the actual Korean corpus glyph set rather than the five PoC glyphs;
3. Korean message compilation connected to the normal build pipeline rather than only the selective PoC API;
4. production font/atlas materialization for the allocated Korean glyph set;
5. the duplicated stock/Korean validation/control traversal refactored or otherwise protected against drift;
6. message-bank source authentication/fail-closed checks suitable for the production patcher;
7. regression tests covering unchanged controls, replacement-ID consumption, deterministic slot mapping, overflow/capacity failure, and byte-isolation guarantees.

## Remaining known non-blocking debt

- `internal/message/projection.go` and `internal/message/korean_materialize.go` duplicate common validation/control logic. This should be closed before the large translation/build phase.
- the PoC Android path does not SHA-256 pin the entire retail `message/msgsec001.dat` member. A production-wide patcher must authenticate the source message data it patches or compile from canonical authenticated sources.
- the 212 renderer slots are audited candidates, not globally production-safe slots; additional resource coverage and/or conservative allocation rules are still needed.

## Next required Claude output

When the integrated production path is ready, replace this document with a focused adversarial review request covering the implementation range and production failure modes. That review will be the gate immediately before starting the large translation batch.
