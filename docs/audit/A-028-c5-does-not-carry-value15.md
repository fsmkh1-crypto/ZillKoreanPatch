# A-028 — C5 does not explicitly carry `$15`; pivot to shared substitution dispatch

## Trigger

A-025 initially used upstream's `callerMovable` projection name as a clue that ID 10010's `$15` value might be supplied by the C5/event-script call site. That interpretation needed verification against the retained CDC command grammar before it could guide runtime work.

## Hypothesis

If C5 directly supplies `$15`, its command arguments should contain a substitution value/source in addition to display metadata and message IDs. If not, the runtime producer must be resolved after the message record is selected, likely in shared message-formatting state or a substitution dispatcher.

## Verification

- Inspected `internal/cdccontext/context.go` and the retained `retailConsumer` decoder for C5.
- Rechecked `internal/gamefmt/cdc` parsing to confirm C5 arguments are authored CDC scalar arguments.
- Rechecked upstream `callerMovable` history and separated editor/projection behavior from retail runtime behavior.
- Added authenticated-retail focus logging of the raw C5 command and association handle for ID 10010.
- Added a heuristic MIPS `valuescan` scanner for executable regions that may decode `0x02 <opcode>` substitution controls; `$15` immediates are scored strongly when present but are not required.

## Result

### Confirmed

- Retained C5 syntax requires three to seven integer arguments and a semicolon.
- The first C5 argument is display mode, the second is entity-association handle, and all remaining arguments are message IDs.
- There is no explicit string/substitution argument in the C5 command shape.
- Therefore the specific model “C5 directly passes the `$15` replacement string as an authored argument” is false.
- `callerMovable` is a translation projection/editor classification and does not establish retail runtime provenance.

### Forensic tooling added

- ID 10010 retail focus output now includes the exact raw C5 command and association handle.
- `internal/forensics/valuescan` scans executable PT_LOAD segments only and reports heuristic candidate regions containing a `0x02` immediate plus nearby byte-load/control-flow behavior. Nearby `0x15` adds score.
- The scanner is intentionally non-authoritative: a candidate is a disassembly target, zero candidates do not refute a substitution dispatcher, and a `$15` immediate is not required if the implementation uses indexed/jump-table dispatch.

## Evidence grade

- **CONFIRMED**: C5 explicit argument contract and absence of a direct substitution-string argument.
- **SUPERSEDED**: wording that treated `callerMovable` as evidence that the C5 caller itself supplies `$15`.
- **OPEN**: shared formatter/substitution dispatcher address, `$15` source, source capacity, expansion destination, and freeze causality.
- **FORENSIC ONLY**: `valuescan` candidates until authenticated retail EBOOT results are disassembled and data/control flow is verified.

## What this excludes

- C5 authored arguments directly carrying the `$15` replacement string.
- Using the `callerMovable` identifier itself as runtime-producer proof.

It does not exclude a C5-triggered shared formatter using association/global state after message lookup, nor does it exclude a smaller scratch/staging buffer before the retained 256-byte page destination.

## New question

Identify the authenticated-retail routine that handles byte `0x02`, then follow its second-byte opcode dispatch to `$15` and determine the exact source pointer, copy/format routine, and destination capacity before any further device test.
