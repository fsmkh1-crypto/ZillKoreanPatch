# A-034 — Standalone `$15` substitution runtime scanner

## Trigger

The substitution scanner was already wired into the authenticated retail ISO build, but that unnecessarily coupled forensic EBOOT inspection to the rest of the patch build. The next question only needs the retail executable.

## Hypothesis

A standalone CLI accepting one retail EBOOT will let the same heuristic scanner be run and archived independently, reducing the asset-backed gate from “full ISO build environment” to “authenticated EBOOT available”.

## Verification

Added `tools/forensics/value-runtime-scan/main.go`.

The CLI:
- reads exactly one EBOOT path;
- calls the same `internal/forensics/valuescan.Scan` used by the real ISO-build audit;
- prints candidate file offsets, ELF virtual addresses, scores and reasons;
- prints a configurable decoded byte-distance window around each candidate;
- classifies visible instructions as call/jump/branch/load/store/address-or-immediate/other;
- supports `--limit 0` for all candidates and `--window -1` for the full retained scanner window;
- explicitly states that candidates are heuristic and that zero candidates do not disprove a shared/table-driven dispatcher.

## Result

The authenticated-retail substitution investigation no longer requires constructing an ISO merely to obtain the static candidate disassembly. Once the retail EBOOT is available, the direct next command is:

`go run ./tools/forensics/value-runtime-scan --limit 10 --window -1 RETAIL_EBOOT`

The resulting candidate log is suitable for following nearby calls and pointer/data movement toward the `$15` source and destination capacity.

## Evidence grade

**CONFIRMED** for tool wiring once CI passes.

**OPEN** for authenticated-retail candidates, `$15` source/destination semantics and freeze causality.

## What this excludes

Nothing about root cause. This only lowers the asset requirement for the next static forensic gate.

## New question

Can the strongest authenticated-retail candidate be followed through its direct call/load/store sequence to a concrete `$15` source pointer and bounded destination/copy routine before another device test is considered?