# A-030 — Preserve call and pointer-memory context around substitution candidates

## Trigger

A-029 tightened the substitution scanner so candidates require register-linked `0x02` prefix evidence rather than loose proximity. The resulting authenticated-retail candidate windows still decoded only a small instruction subset, leaving calls, returns and ordinary pointer loads/stores as opaque `word 0x...` entries.

## Hypothesis

If a future retail EBOOT candidate is the shared substitution formatter or is near it, preserving control-transfer and pointer-memory instructions in the same logged window should materially reduce the static work needed to trace `$15` source and destination state. This improves evidence capture without changing the candidate score or claiming runtime semantics.

## Verification

- Extended the forensic MIPS decoder with `j`, `jal`, `jr`, and `jalr`.
- Extended memory decoding with `lh`, `lhu`, `lw`, `sh`, and `sw` in addition to the existing byte loads/stores.
- Added deterministic unit coverage for each newly decoded instruction family.
- Deliberately left candidate scoring unchanged: only the existing register-linked prefix/opcode relationships determine whether a window is reported.

## Result

### Confirmed implementation facts

- A reported candidate window can now retain nearby direct calls, indirect calls/returns, and common halfword/word pointer loads and stores as readable instructions.
- The new decoding does not make a candidate more or less likely to be reported.
- CI for the preceding syntax repair (`d7efdc293181571c7879d1684b106aa34dabf799`) passed all three workflows before this enhancement was added.

### Interpretation

This does not identify the retail substitution routine. It strengthens the forensic handoff once an authenticated retail asset is available: a surviving candidate can be inspected for caller/callee shape and likely source/destination pointer traffic without first reconstructing every nearby word manually.

## Evidence grade

- **CONFIRMED**: decoder behavior after unit/CI success.
- **STRONG DESIGN EVIDENCE**: retaining calls and pointer-memory operations is directly useful for static source/destination tracing.
- **OPEN**: authenticated-retail candidate existence, actual formatter identity, `$15` source pointer, destination capacity, and freeze causality.

## What this excludes

- It excludes the previous tooling limitation where nearby `jal`, `jr`, `lw`, `sw`, `lh`, `lhu`, and `sh` instructions were necessarily left opaque.
- It does not exclude table-driven dispatch, indirect pointer provenance outside the retained window, or a formatter that does not contain a literal `$15` comparison.

## New question

When authenticated retail EBOOT output is available, do the strongest register-linked candidate windows contain a call/copy shape that can be followed to a concrete `$15` source and bounded destination?

## Commits

- `dff979ffb892c881ee677e2b80eefb44125b8235` — decode calls and pointer memory operations
- `58b8ef2b83c4c7449eb1d00f37c1a3f74d97ea52` — add deterministic decode coverage
