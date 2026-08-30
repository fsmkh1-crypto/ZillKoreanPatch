# A-065 — English-first executable and renderer-slot ownership audit

## Scope

Restarted parity audit of the Korean executable and renderer-slot paths using the upstream English patch contract as the primary baseline.

## BOOT/EBOOT result — PASS

The Korean executable path applies the same `patches/executable/manifest.toml` used by the English release before any Korean fixed-string overlay. The manifest authenticates the supported retail `BOOT.BIN` and the patched result. `ApplyKoreanEBOOT` additionally authenticates the manifest-patched ELF using the same patched-ELF fingerprint used by the English fixed-string compiler.

After the sparse Korean overlay, `applyKoreanFixedEBOOT` calls `elfpatch.VerifyApplied`, so every guarded runtime-patch span is rechecked after localization writes. This prevents Korean fixed strings from silently clobbering message-arena, wide-offset, or other executable patches.

Classification: **PASS**.

## Renderer-slot ownership result — MISSING, corrected

`koreanslots.BuildPlan` is explicitly the production slot allocator and supports fail-closed exclusion of any candidate renderer key whose exact two-byte representation occurs in caller-authenticated binary blobs.

The mobile planner authenticated `BOOT.BIN`, `EBOOT.BIN`, and `bindata.dat`, and scanned literal CP932 ownership, but previously called `BuildPlan` without those blobs. Exact-byte collisions were only reported afterwards by `auditMobileExactByteReuse`, which was documented as non-blocking forensic evidence.

That was weaker than the project-owned production slot contract and left the mobile path able to repurpose a renderer key that still occurred in an authenticated engine/fixed-data blob.

Correction:

- feed authenticated `BOOT.BIN`, `EBOOT.BIN`, and `bindata.dat` directly to `koreanslots.BuildPlan`;
- retain literal CP932 reservations as additional evidence;
- require both the initial and final exact-byte audits to report zero candidate/mapped collisions;
- fail the build if any such collision survives;
- CI now checks that this fail-closed chain remains wired.

Classification after correction: **PASS** for the static ownership contract.

## Evidence interpretation

This removes one static renderer-key ownership risk. It is not proof that all runtime freezes are fixed, and successful execution remains non-reproduction evidence only.
