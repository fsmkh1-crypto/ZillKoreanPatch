# A-032 — Structured substitution candidate logging

## Trigger

A-030/A-031 made the retail substitution candidate window substantially more readable, but the real build log still required manual inspection of every decoded line to separate calls, branches, memory reads/writes and address construction.

## Hypothesis

Adding presentation-only instruction classes to the authenticated-retail candidate log will reduce analysis error and make `$15` source/destination tracing faster without changing candidate scoring or evidence strength.

## Verification

- Kept `internal/forensics/valuescan` candidate scoring and selection unchanged.
- Added `forensicInstructionKind()` only in the real ISO-build reporting path.
- Each retained candidate-window instruction now logs one of:
  - `call`
  - `jump`
  - `branch`
  - `load`
  - `store`
  - `address-or-immediate`
  - `other`
- Added table-driven unit coverage for representative `jal/jalr`, `j/jr`, conditional branch, load/store, `lui/ori/addiu`, and unrelated instructions.

## Result

Retail asset-backed build logs can now be filtered mechanically for likely control-flow edges and pointer/data movement around the strongest `$15` substitution candidates.

This does **not** identify the dispatcher, `$15` source, destination buffer, or capacity. It does not alter the scanner's heuristic score or promote a candidate to runtime proof.

## Evidence grade

**CONFIRMED** for log classification behavior once CI passes.

**OPEN** for authenticated-retail candidate results and all `$15` runtime semantics.

## What this excludes

Nothing about root cause. This is forensic observability only.

## New question

On an authenticated retail EBOOT candidate, which nearby `call/load/store/address-or-immediate` sequence actually carries the `$15` source pointer and destination buffer, and is a bounded copy/format capacity visible?
