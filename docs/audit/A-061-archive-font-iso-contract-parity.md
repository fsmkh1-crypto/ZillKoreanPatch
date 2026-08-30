# A-061 — Archive / font / authored-ISO contract parity audit

Status: static/compile-path audit complete; runtime freeze safety not inferred.

## Scope

Audit the Korean release path against engine/release contracts already enforced by the upstream English path, focusing on the final message/font materialization chain:

`Korean semantic text -> layout -> CompileBankKorean -> PAA replacement -> rebuilt pa/pami -> staged PSP_GAME -> authored ISO`

## Findings

### PAA archive replacement — PASS

`paa.Pair.Rebuild` is already fail-closed. It resolves replacement indexes before writing, rebuilds aligned index/archive files, reopens the temporary pair, checks index bytes/member metadata, and compares every rebuilt member payload byte-for-byte with either the requested replacement or the original retail member before publication.

Therefore compiled message banks and font replacements cannot silently change between `addBanks` / font replacement registration and the rebuilt `pa` / `pami` files.

### Korean mobile font atlas / PAF — PASS

`prepareKoreanMobileFontReplacements` already follows the full-repack model rather than the old atlas-only shortcut. It authenticates the retail PAF member, requires every mapped custom rune to have a raster, performs `FullRepackAuthenticatedRetailFont`, and then runs `VerifyFullRepackSemantics` over the patched atlas/PAF before the replacements are admitted to the archive.

The planner also reserves renderer keys observed in authenticated `BOOT.BIN` and authenticated `bindata.dat`, in addition to fixed renderer keys. This remains a Korean-specific superset of the language-independent engine contract rather than an exception to it.

### Authored ISO boundary — GAP CLOSED

Before this audit, PAA rebuild proved the staged archive payloads, but `authorTranslatedISO` did not reopen the newly authored ISO and compare its `PSP_GAME` payloads back to the staged tree. This left the last build boundary trusted rather than independently proven.

Commit `75d64f3220912f9f1e574c43a70417eff3940dca` closes that gap. `authorTranslatedISO` now reopens the authored ISO and compares every regular file under staged `PSP_GAME` byte-for-byte with the corresponding ISO payload. Any omission, truncation, extension, or byte difference removes the staged ISO and fails the build. The gate is shared by English and Korean release paths.

Expected successful-build evidence:

`FORENSIC ISO_PSP_GAME_PROVENANCE files=<n> bytes=<n> exact=true`

## Interpretation

This audit does not establish that the freeze is fixed. It establishes a stronger provenance chain: once a Korean record/font payload passes its consumer/compiler contract and is registered as a PAA replacement, the archive layer proves the exact replacement bytes and the ISO layer now proves the exact staged `PSP_GAME` bytes reached the authored image.

Remaining parity work should therefore concentrate above the archive boundary: message consumer membership/contracts, dynamic substitution bounds, and any renderer/runtime consumer not represented in the current consumer map.
