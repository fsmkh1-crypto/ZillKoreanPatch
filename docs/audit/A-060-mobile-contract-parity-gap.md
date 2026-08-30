# A-060 — Mobile Korean ISO path bypassed newly enforced English contracts

Date: 2026-08-30

Status: **confirmed build-path parity defect; fixed in `9956cf7c396f0c198f74b8d533c81ea1d0e16ddf`**.

## Finding

The normal Korean release path and the Android/mobile ISO-only path had drifted.

`BuildKoreanAlpha` had been updated to derive and enforce the upstream English consumer/storage contracts before compiling banks, but `BuildKoreanAlphaISOOnly` still performed only the older C5 storage derivation/validation before compiling the ISO used by the APK.

Therefore a protection could pass review and CI in the normal release path while the APK-generated ISO never executed that protection. This is the same class of failure as a layout or validation rule being present in source but not connected to the materialization path that reaches the game.

## Fix

The mobile path now runs the same pre-compile sequence as the normal Korean release path:

1. derive C5 storage layouts;
2. derive upstream-English C22 consumer layouts using actual Korean renderer bytes;
3. apply C22-only retail scanner hardening where evidenced;
4. validate the full upstream-English fixed consumer/storage contract set;
5. validate Korean C5 branch/page storage;
6. compile the banks and add them to the archive replacement set.

Canonical Korean remains unchanged; all generated wrapping remains build-owned layout.

## Evidence policy

This fixes a real build-path contract omission. It does **not** prove that every freeze shares this cause and does not turn a future non-freeze run into proof of universal safety.

## Next parity-audit targets

- BOOT/EBOOT reserved runtime spans and post-overlay verification;
- BINDATA / fixed-data ownership and renderer-key reservation;
- archive replacement ownership and final rebuilt-archive provenance;
- font PAF/atlas mapping parity and final asset verification;
- ensure every Android/mobile entry point shares the same gates as the normal release path.
