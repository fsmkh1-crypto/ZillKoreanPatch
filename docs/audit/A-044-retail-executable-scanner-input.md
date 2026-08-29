# A-044 — Retail executable scanner input correction

## Trigger

The first Android `--preflight-only` run against the supported ULJM-05410 v1.03 retail ISO authenticated `EBOOT.BIN` successfully, then exited before emitting the first `C5_RUNTIME_SCAN` line.

Observed asset-backed log boundary:

- `RETAIL_EBOOT_BINDING` emitted with the pinned supported SHA-256
- no `C5_RUNTIME_SCAN` line followed
- preflight exited non-zero

## Hypothesis

`auditC5RuntimeCandidates` was feeding authenticated retail `EBOOT.BIN` bytes to scanners that require an executable MIPS ELF/PT_LOAD image. The patch path itself applies executable ELF patches to retail `BOOT.BIN`, while `EBOOT.BIN` is separately fingerprinted/used for ownership and fixed-field evidence.

## Verification

Code trace showed:

- `auditC5RuntimeCandidates` loaded `EBOOT.BIN` and passed those bytes directly to `c5scan.Scan`, `valuescan.Scan`, and `fontscan.Scan`.
- `buildKoreanAlphaExecutable` reads retail `BOOT.BIN` as the source passed to `elfpatch.Apply`.
- PR14/mobile planner audits already load `BOOT.BIN` separately for executable/static literal analysis.
- The asset-backed Android run reached the `EBOOT.BIN` SHA-256 binding log and failed before the first scanner summary, matching failure in the first ELF scanner call.

## Result

Corrected `auditC5RuntimeCandidates` to:

1. retain and log pinned `EBOOT.BIN` authentication as a non-scanner forensic input;
2. load authenticated retail `BOOT.BIN` separately;
3. use authenticated `BOOT.BIN` for C5, `$15`, and font/renderer MIPS ELF scanning;
4. tag scanner output explicitly with `input=BOOT.BIN`;
5. distinguish `scanner_input=false` for EBOOT and `scanner_input=true executable_format=elf` for BOOT.

## Evidence grade

**CONFIRMED diagnostic-input bug.**

This finding does **not** identify the game freeze root cause. It identifies why the first asset-backed preflight could not reach the intended executable scanners.

## What this excludes

The first Android preflight failure is not evidence against the C5, substitution, or renderer hypotheses. Those scanners did not successfully inspect the executable image in that run.

## New question

After rerunning preflight with authenticated `BOOT.BIN` as the scanner input:

- which C5 candidates are found?
- which `$15` candidates are found?
- which linked 0x20-stride font/PAF candidates are found?
- does preflight reach the later PR14, slot-ownership, bank, font-repack, and executable-generation gates?
