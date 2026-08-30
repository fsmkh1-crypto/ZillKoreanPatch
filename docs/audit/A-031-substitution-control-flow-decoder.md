# A-031 — Substitution candidate control-flow decoder fidelity

## Trigger

The register-linked `valuescan` scanner can preserve candidate windows from an authenticated retail EBOOT, but source/destination tracing still depends on the disassembly text being accurate enough to follow calls, branches and pointer construction. The first decoder extension exposed two review issues in synthetic CI: a malformed test assertion and a mistaken expected MIPS pseudo-direct jump target.

## Hypothesis

Without changing candidate scoring, decoding the common call, pointer-memory, upper-immediate and control-flow instructions around a candidate should make a future authenticated-retail log materially more useful for tracing the `$15` producer and destination. The decoder itself must be regression-tested so a readable but wrong address is not mistaken for evidence.

## Verification

- Retained candidate scoring and executable-segment selection unchanged.
- Added readable decoding for `j`, `jal`, `jr`, `jalr`, `lh`, `lhu`, `lw`, `sh`, `sw` and `lui`.
- `j`/`jal` targets are reconstructed with the MIPS pseudo-direct rule using `(PC+4)[31:28]` plus the shifted 26-bit target field.
- Conditional branch output now reports the computed PC-relative destination, including negative displacements.
- Synthetic tests cover call/jump decoding, memory operations, `lui`, a pseudo-direct jump whose upper nibble must come from the PC, a forward branch and a negative branch.
- CI exposed and corrected two test defects rather than accepting a misleading expected address.

## Result

### Confirmed implementation facts

- Candidate windows can now expose common control-flow and pointer-construction instructions in human-readable form rather than as opaque words.
- Jump target reconstruction follows the architectural pseudo-direct address rule in the tested decoder path.
- Branch destinations are derived from `PC+4 + sign_extend(imm16<<2)` in the tested decoder path.
- Scanner ranking/scoring semantics are unchanged by this work.

### Interpretation

This improves the next static step after a retail candidate is found: calls, branches and likely global-pointer construction can be followed directly from the persisted window before widening the disassembly. It does not identify the formatter, `$15` source or copy destination by itself.

## Evidence grade

- **CONFIRMED**: decoder behavior covered by synthetic tests once CI passes.
- **FORENSIC ONLY**: any future candidate address until authenticated-retail control/data flow establishes its runtime role.
- **OPEN**: authenticated-retail candidate existence, actual `$15` producer, source maximum, destination buffer/capacity and causal relation to the freeze.

## What this excludes

- Treating raw 26-bit `j`/`jal` fields as complete runtime addresses.
- Treating branch immediates as useful control-flow destinations without PC-relative reconstruction.
- Treating an opaque `word 0x...` around a candidate as sufficient when it is a common pointer/call instruction now covered by the decoder.

It does not exclude table-driven dispatch, indirect calls outside the small persisted window, GP-relative state, or routines that require a wider function-level disassembly.

## New question

When an authenticated retail EBOOT is available, does the highest-ranked register-linked candidate expose a call or pointer-construction chain that can be followed to the `$15` source and the exact destination capacity without another device run?

## Related commits

- `58b8ef2b83c4c7449eb1d00f37c1a3f74d97ea52` — decode calls and pointer-memory operations
- `252fadf065c8f43df6fd3af67bae7bc3add475fe` — correct pseudo-direct jump regression expectation
- `05b2e36e50f0446a882bb91a694bb0d301f344ad` — decode branch targets and `lui`
- `1a9f849a9570e39b3f339f22c017cf1707465d38` — cover branch-target and `lui` decoding
