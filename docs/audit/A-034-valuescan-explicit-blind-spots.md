# A-034 — Explicit valuescan blind-spot regressions

## Trigger

Independent review of `internal/forensics/valuescan` identified two structurally plausible MIPS patterns that the scanner does not recognize:

- multi-instruction small-constant construction such as `lui` + `ori` rather than a single `addi/addiu/ori` from `$zero`;
- equality implemented as `subu` followed by `beqz/bnez`, where the compared source registers do not appear directly in the branch instruction.

These limitations were not sufficiently explicit in A-029.

## Hypothesis

Encoding these unsupported shapes in synthetic ELF regressions should prevent future maintainers from treating a zero-candidate result as broad evidence against a shared `$15` dispatcher.

## Verification

Added two synthetic regression cases in `internal/forensics/valuescan/valuescan_test.go`:

1. `TestScanBlindSpotMultiInstruction15Literal`
   - valid linked `0x02` prefix path;
   - `$15` is built via `lui r6,0` + `ori r6,r6,0x15`;
   - a direct branch compares the byte-loaded opcode register to `r6`;
   - expected result is zero candidates because current `loadsLiteral` intentionally recognizes only single-instruction literal construction from `$zero`.

2. `TestScanBlindSpotSubuThenBeqzEquality`
   - valid linked `0x02` prefix path;
   - `$15` is loaded directly;
   - equality is expressed as `subu r7,r5,r6` followed by `beq r7,r0,...`;
   - expected result is zero candidates because current register-link logic only inspects branch operands and does not propagate provenance through arithmetic.

## Result

The scanner's blind spots are now executable regression knowledge rather than documentation-only caveats.

This does **not** mean these patterns occur in the retail EBOOT. It means a future retail result of zero candidates cannot exclude them without separate disassembly/dataflow inspection.

## Evidence grade

- **CONFIRMED** for current scanner behavior once CI passes.
- **OPEN** for whether either pattern is present in authenticated retail code.

## What this excludes

Nothing about the runtime root cause. These regressions deliberately document unsupported implementations.

The scanner also remains blind or incomplete for other patterns already noted elsewhere, including jump tables, table-driven dispatch, indirect constant loads from data, range checks such as `slti/sltiu`, and relevant dataflow outside the retained scan radius.

## New question

Once an authenticated retail EBOOT is available, does direct inspection around candidate or near-candidate code reveal any of these unsupported dispatch/dataflow shapes?