# A-045 — Mobile forensic log terminal-error preservation

## Observation

The second real Android retail preflight after A-044 authenticated both retail executables and reached the actual BOOT.BIN scanners:

- retail EBOOT SHA-256 binding succeeded;
- retail BOOT.BIN SHA-256 binding succeeded and `executable_format=elf`;
- C5 scanner completed with 691 heuristic candidates;
- `$15` value scanner completed with zero heuristic candidates (non-dispositive by contract);
- font scanner completed with one linked stride-32 candidate at file offset `0xEDF74` / vaddr `0xEDEF4`;
- PR14 policy replay completed;
- H0 and final exact-byte audits ran;
- `FORENSIC MAPPING_FINGERPRINT` was emitted with `changed=12`.

`FORENSIC MOBILE_PREFLIGHT_COMPLETE` was absent, so the preflight still exited before successful completion. The Android clipboard log contained only lines beginning with `FORENSIC`; the terminal builder error was therefore lost even though the UI displayed a failure.

This run is not evidence that A-044 recurred. The BOOT.BIN scanner path demonstrably executed.

## Diagnostic defect

The mobile handoff contract captured every `FORENSIC` line but the builder emitted its terminal preflight/build error only on stderr with a non-FORENSIC prefix. In addition, the exact-byte ownership audit emitted every mapped hit, producing hundreds of pages of low-value clipboard output.

## Correction

1. `build-korean-iso --preflight-only` now emits `FORENSIC MOBILE_PREFLIGHT_ERROR error=...` before returning non-zero.
2. Full mobile ISO build failures emit `FORENSIC MOBILE_BUILD_ERROR error=...` before returning non-zero.
3. H0/final exact-byte audits retain their aggregate counts but emit at most eight detailed mapped-hit records per phase, followed by an omitted-count summary.

## Evidence policy

The observed runtime-freeze investigation remains unchanged:

- a successful run is only non-reproduction for that run;
- a freeze/crash is strong negative evidence;
- zero heuristic scanner candidates do not disprove a runtime path;
- the single font stride-32 candidate is a lead, not a root-cause finding.

The next real retail preflight should produce a compact clipboard log containing the terminal failure reason if it still exits non-zero.
