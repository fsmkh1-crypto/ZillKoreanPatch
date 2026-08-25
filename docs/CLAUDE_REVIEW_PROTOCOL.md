# Claude review protocol

Last updated: 2026-08-25

This project assumes Claude is an adversarial reviewer at milestone boundaries. Review is not optional ceremony: before a change crosses from isolated/unit-tested code into a real ISO/font/runtime patch, the current implementation should be reviewed for concrete correctness failures.

## 1. Roles

- GPT: lead developer/integrator, canonical GitHub state, implementation, Korean translation, final merge decisions.
- Claude: adversarial code/design/translation QA reviewer.
- Gemini: optional overflow/secondary translation worker only; not part of the critical path.

The final arbiter is retail data, deterministic tests, hashes, and actual PPSSPP behavior. Model agreement is not evidence by itself.

## 2. Review gates

Request a Claude review at least at these boundaries:

1. after a new parsing/encoding/font/message architecture is implemented;
2. after fixing a prior adversarial review, before relying on the fix downstream;
3. before changing authenticated retail assets or emitting a new ISO PoC;
4. before promoting a PoC path into the production build path;
5. before a large translation batch is considered canonical;
6. whenever a safety claim changes from "candidate" to "safe/reusable/proven".

Small documentation-only or test-only commits do not require a separate review unless they change an asserted fact.

## 3. What Claude should do

Claude should review current `main`, not old chat assumptions. Start with `docs/CLAUDE_HANDOFF.md`, then inspect the exact files/commits named in the current review request.

For every finding, prefer this format:

- severity: MUST-FIX / SHOULD-FIX / NOTE;
- file + function (and line if available);
- concrete failure mode;
- minimal reproducible test or byte-level counterexample;
- proposed fix direction;
- whether the issue blocks the next milestone.

Do not report vague possibilities without a plausible path to failure. Conversely, do not waive a concrete correctness bug because CI is green.

## 4. Required distinctions

Every review should distinguish:

- **proven/observed**: authenticated retail bytes, deterministic test result, actual PPSSPP observation;
- **inference**: strongly suggested but not directly proven;
- **proposal**: intended future architecture.

A reviewer challenging a proven fact should provide a counterexample, failing test, or conflicting retail evidence.

## 5. Review scope discipline

Prefer focused review of the changed subsystem. Do not redesign unrelated working code.

For renderer/message work, prioritize:

- validation and encoding using the exact same Korean mapping;
- preservation of runtime controls and fixed source-owned bytecode;
- record ordering/IDs and untouched-record byte identity;
- runtime bank capacity and offset correctness;
- renderer-key byte order and CP932 semantics;
- candidate slot safety claims;
- failure-closed behavior for unmapped Hangul or unsafe slot reuse;
- layout/line-break semantics versus Japanese source wrapping.

For translation-exchange work, prioritize:

- exact JSON schema;
- stale/modified source rejection;
- protected-token isolation;
- glossary/context trust boundaries;
- segment count/order/identity;
- Markdown/parser corruption paths.

## 6. Review result handling

GPT should independently verify each concrete finding against current code before changing anything.

- Valid MUST-FIX: add a regression test, fix it, rerun CI.
- Valid SHOULD-FIX: fix now if it can affect the next milestone; otherwise document it explicitly.
- Invalid/non-reproducible: record why it does not apply and continue.

After fixes, the next high-risk milestone may proceed only when `go test ./...`, `go vet ./...`, and `./zill check` pass, plus any subsystem-specific checks.

## 7. Current review checkpoint

Current reviewed baseline before the latest implementation work:

- `a80988f364f1028b01340e2338404e3bb8f1052b`

Claude's previous review found:

1. false-positive generic printf matching (`100% Match` -> `% M`);
2. glossary stale-source validation comparing external glossary data against itself.

Both were accepted as valid and fixed with regression tests.

Current review target head:

- `cb0586012655fdfe6d01d94d252ba84650a1fa5c`

Review the range from the prior baseline to the current head, with special attention to the fixes and new Korean renderer/message path.

Key new files/changes:

- `internal/translationexchange/v2_review_fixes.go`
- `internal/translationexchange/v2_review_fixes_test.go`
- `internal/message/korean_materialize.go`
- `internal/message/korean_materialize_test.go`
- `internal/message/compile_korean.go`
- `internal/message/compile_korean_test.go`
- `cmd/zill/korean_font.go`

The next blocked milestone is the first real Korean sentence using reusable renderer slots without globally replacing common Japanese glyphs. Do not proceed to claim slots are safe or generate that PoC until this checkpoint review is considered.
