# A-058 — A-054 rejected for current HEAD after runtime freeze

Date: 2026-08-30

Status: **A-054 is rejected as the sole root cause for the current forensic HEAD.**

## Exact code provenance

The forensic branch `audit/projection-control-roundtrip` is at:

`da316771262adaf88ccc8f5f665926ae2e665be3`

This HEAD contains the A-055 compiler gate for ID 210065 and the full-corpus scanner-span audit added immediately afterward. The Android release-candidate branch used for the subsequent hands-on ISO test changed only packaging/workflow/launcher files relative to this forensic HEAD; it did not change the Go message compiler, Korean materialization, corpus, layout, font, executable patches, or ISO-authoring logic.

Therefore the game payload used by the corrected RC patcher is the same current forensic payload for purposes of the A-054 test.

## Runtime result

A patched ISO produced from that payload still froze at the 210065 scene once.

Under the project's asymmetric evidence rule, one freeze is strong failure evidence. Under the explicit A-054 acceptance criterion proposed for this experiment, a single freeze is sufficient to reject A-054 as the sole explanation for the current HEAD; there is no evidentiary value in running two additional repetitions merely to complete a 3-run success criterion that has already failed.

## What remains true from A-054

A-054 remains a strong explanation for the historical captured invocation that showed:

- `s4=0`;
- `s3=0x113` after the first scanner call;
- an invalid word at `s0+0x3C0`;
- runaway scanning in `z_un_089661DC`.

That historical capture is not invalidated. What is rejected is the stronger claim that preventing the current compiler from materializing a >=0x100 scanner span is sufficient to eliminate the present 210065 freeze.

## Static corpus gate

The current branch also contains a Go regression test that materializes accepted Korean records through the production record materializer and models the observed retail scanner's maximum span semantics. CI passes with no current materialized Korean record reaching the `0x100` scanner-span boundary.

This means there is no remaining current-corpus offender for this specific >=0x100 inline-span mechanism to mass-fix.

## Next direction

Stop tracer work for now and move to the next independent candidate. Do not reuse same-screen location as evidence that the historical A-054 machine state recurred. The next candidate must explain a freeze that survives the current A-055-safe materialization and full-corpus <0x100 span gate.
