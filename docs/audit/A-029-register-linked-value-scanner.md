# A-029 — Register-linked substitution runtime scanner

## Trigger

The first `valuescan` implementation searched executable MIPS windows for a literal `0x02`, nearby byte loads, branches, a literal `0x15`, pointer increments and scaled-index shapes. Those features were useful for triage but could be unrelated instructions merely colocated within the same 48-instruction window.

## Hypothesis

Requiring register-flow linkage between the byte load, literal construction and equality branch should materially reduce false positives without inventing runtime semantics. In particular, a stronger candidate should show a byte-loaded register compared against a register that was actually loaded with literal `0x02`; a focus `$15` comparison should be scored only when the same relationship exists for literal `0x15`.

## Verification

- Reworked `internal/forensics/valuescan` so scan anchors are literal constructions from the MIPS zero register, not arbitrary arithmetic containing immediate `2`.
- The scanner now recognizes a prefix comparison only when `beq`/`bne` directly compares a byte-load destination register with a register containing literal `0x02`.
- Literal `0x15` contributes focus evidence only when an equality branch directly compares it with a byte-loaded register.
- Pointer-increment evidence is counted only when it advances the base register used by one of those linked byte loads.
- Scaled-index evidence is counted only when the byte-loaded register participating in the linked `$15` comparison is used as the scaled index.
- Added synthetic negative tests for colocated-but-unlinked literals/loads/branches and for `addiu reg,nonzero,2`, which is arithmetic rather than literal construction.
- Retained the existing synthetic positive ELF test and executable-segment requirement.

## Result

### Confirmed implementation facts

- Unrelated `0x02`, `0x15`, byte-load and branch instructions in one scan window no longer satisfy the reporting threshold merely by coexistence.
- The positive synthetic dispatcher shape remains reportable because its byte-load and literal registers are joined by actual equality branches.
- A candidate remains a heuristic disassembly target. Register linkage improves specificity but does not establish that the routine is the retail message substitution dispatcher.

### Interpretation

This gate makes authenticated-retail output more useful: a future candidate log will encode a concrete local data-flow relationship rather than a loose bag of nearby instruction features. That is a stronger static lead for tracing the shared `0x02 <opcode>` formatter path after C5 selects message 10010.

## Evidence grade

- **CONFIRMED**: scanner behavior and synthetic regression behavior once CI passes.
- **STRONG DESIGN EVIDENCE**: register-linked scoring is more specific than proximity-only scoring for the intended dispatcher shape.
- **OPEN**: whether authenticated retail EBOOT produces any candidate, whether a candidate is the actual substitution dispatcher, the runtime producer of `$15`, its maximum encoded length, and any causal link to the observed freeze.

## What this excludes

- It excludes treating unrelated nearby literals as equivalent to a register-linked prefix/opcode comparison.
- It excludes using arbitrary arithmetic with immediate `2` as a scan anchor.
- It does not exclude implementations that use a jump table, table-driven decode, delay-slot/data-flow pattern outside the scan radius, or no literal `0x15` at all.

## New questions

1. Does an authenticated retail EBOOT yield fewer and better-ranked candidates after register-link filtering?
2. For any surviving candidate, what calls into and out of the routine, and where does the substitution source pointer originate?
3. Is `$15` dispatched through a table rather than a direct literal comparison?
4. Can the surviving candidate be connected to the C5/message renderer path without relying on runtime gameplay testing?

## Related commits

- `958c6ae89b66b443585b46586a2682b301502ab9` — persist substitution candidate disassembly context
- `32e59b1859df1a370c4bce4f993449e9cc5cd914` — require register-linked substitution scan evidence
- `a6c4d1db06c69385e84c8002f23e3617898b331d` — cover unrelated substitution scanner literals
