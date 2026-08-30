# A-069 — English-first fixed-data contract parity

## Scope

This audit compares the upstream English fixed-data path with the Korean path for the two non-message fixed-string surfaces currently present in the project: patched `BOOT.BIN`/`EBOOT.BIN` strings and `data/bindata.dat` equipment names.

The purpose is contract parity, not translation-count parity. An engine/storage invariant must be preserved. English content policy is not copied merely because English translates a surface.

## EBOOT fixed strings — PASS / DIFFERENT-BY-DESIGN

English uses a complete 557-field table. Korean intentionally uses a sparse reviewed overlay. That difference is content policy, not a weaker storage contract.

Both paths share the following engine invariants:

- the executable runtime manifest is applied before localization;
- the manifest-patched ELF is authenticated against `patchedELFSHA256`;
- each source string is encoded and matched at its exact offset;
- the byte immediately after the guarded source field must be NUL;
- replacement encoded bytes may not exceed the original fixed-width capacity;
- the field must lie inside the ELF;
- translated fields may not overlap;
- writes clear the full original capacity before copying replacement bytes.

Korean additionally encodes replacement text through the exact renderer-slot mapping and re-runs `elfpatch.VerifyApplied` after the sparse overlay, so localization cannot silently overwrite the runtime-patch spans.

Classification: **PASS** for engine invariants; **DIFFERENT-BY-DESIGN** for complete-English versus sparse-Korean translation coverage.

## BINDATA equipment table — DIFFERENT-BY-DESIGN, authenticated ownership PASS

English translates all 132 equipment-name records in `data/bindata.dat` and therefore replaces that archive member.

Korean currently has no `release/korean/strings/equipment.toml`, so the release does not claim to translate this fixed-data table. This is a translation-coverage difference, not permission to ignore the blob.

The Korean desktop and mobile slot planners still:

- load the authenticated retail `bindata.dat`;
- run the shared `fixeddata.ApplyEquipment` against the upstream English equipment table as a structural/source-layout guard;
- scan the authenticated blob for CP932 ownership evidence;
- feed the exact `bindata.dat` bytes into `koreanslots.BuildPlan`, which excludes exact two-byte references before renderer-slot allocation.

Therefore the untranslated BINDATA member remains byte-identical retail data while still constraining Korean renderer-key ownership.

Classification: **DIFFERENT-BY-DESIGN** for translation coverage; **PASS** for authenticated fixed-data/slot-ownership contract.

## Freeze-evidence interpretation

This audit does not establish a universal freeze cause. It removes another class of static divergence: a Korean fixed-string write exceeding the English fixed-width contract, a localization write clobbering executable patches, or a renderer key being allocated over authenticated BINDATA ownership.

Dynamic substitutions and runtime-only consumers remain separate evidence domains.
