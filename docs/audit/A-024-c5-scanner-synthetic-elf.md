# A-024 — C5 runtime scanner synthetic ELF pipeline verification

## Trigger

The C5 runtime scanner initially had unit coverage for heuristic window scoring, but that did not prove the complete path from an ELF container through executable-segment selection into candidate reporting.

## Hypothesis

A malformed ELF reader, wrong program-header filter, or file-offset mapping bug could make later retail scanner results meaningless even if the isolated score function was correct.

## Verification

Commit `5d2ad08e29a838585545aea06690652fd7c2817f` added a synthetic ELF32 little-endian MIPS fixture containing a C5-like instruction cluster.

The regression test verifies:

- the ELF is accepted as MIPS;
- only executable `PT_LOAD` segments are scanned;
- the embedded contract-like sequence produces exactly one candidate;
- the reported candidate file offset is `0x100`;
- the candidate reaches the configured score threshold;
- the same bytes inside a non-executable `PT_LOAD` produce zero candidates.

## Result

The scanner's ELF parse -> architecture check -> executable segment selection -> window scoring -> candidate file-offset path is exercised end to end by deterministic synthetic tests.

This does not establish that the heuristic uniquely identifies the retail C5 handler. Candidate output remains forensic evidence to inspect, not proof. Likewise, zero candidates on a retail executable would not refute the retained C5 contract because compiler instruction selection may differ from the heuristic pattern.

## Evidence grade

- **CONFIRMED**: synthetic scanner pipeline behavior and executable-segment filtering.
- **OPEN**: retail C5 handler identity, actual buffer destination and exact producer/consumer dataflow.

## What this excludes

- A future retail candidate cannot be dismissed as merely an artifact of the previously untested ELF-segment plumbing.
- It does not exclude heuristic false positives or false negatives.

## New questions

1. What candidates, if any, appear against the authenticated ULJM05410 1.03 EBOOT?
2. For each candidate, do nearby instructions establish the 256-byte destination, three-line page split, or nine-page bound?
3. Can caller/callee dataflow connect such a candidate to the C5 CDC display path and dynamic substitution expansion?

## Commit

- `5d2ad08e29a838585545aea06690652fd7c2817f` — `Test C5 scanner through synthetic MIPS ELF`
