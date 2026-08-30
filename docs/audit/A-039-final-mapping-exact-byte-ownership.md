# A-039 — Audit exact-byte ownership after Minimal87 relocation

## Trigger

The mobile planner already ran `auditMobileExactByteReuse()` against the H0 mapping before applying Minimal87 relocation. That was useful for historical H0 comparison, but the actual ISO is built from the **relocated final mapping**.

A spare renderer key selected during Minimal87 relocation could therefore have an exact two-byte occurrence in retail BOOT/EBOOT/BINDATA even when the H0-mapped key for that rune did not. The pre-relocation audit would not report that final-plan collision.

## Hypothesis

Running the same exact-byte ownership audit again after `plan.Mapping = mapping` will close this observability gap without changing slot allocation or promoting raw byte hits to runtime-ownership proof.

## Verification

- Retained the existing H0-phase audit and explicitly labeled its output `phase=h0`.
- Re-ran `auditMobileExactByteReuse()` after all Minimal87 relocations and after assigning the final mapping.
- Added `phase=final` summary and per-hit logs against the same retail BOOT/EBOOT/BINDATA byte slices.
- Added a synthetic regression where:
  - the H0 key has no exact-byte hit;
  - the relocation target does occur in EBOOT;
  - the H0 audit reports zero mapped hits;
  - the final-plan audit reports exactly one relocation-only mapped hit.

## Result

The asset-backed mobile preflight can now distinguish ownership evidence for the historical H0 mapping from the exact mapping that will actually be used to encode Korean text and construct the full-repack font.

This closes a real static blind spot: relocation-selected keys can no longer escape the exact-byte report merely because they were not part of the H0 mapping.

## Evidence grade

- **CONFIRMED** for the pre-/post-relocation audit distinction and synthetic regression once CI passes.
- **OPEN** for whether the authenticated retail asset set produces any final-plan mapped exact-byte hits.
- **OPEN** for whether any raw exact-byte hit represents a real hidden renderer consumer rather than accidental binary coincidence.

## What this excludes

It excludes treating the old H0-only exact-byte audit as coverage of the final Minimal87 mapping.

It does not prove that a hit is unsafe, nor that absence of hits proves a renderer key is safe. Direct tables, transformed/indexed references, special-key dispatch and other hidden consumers may not contain the literal two-byte renderer key.

## New question

On the actual retail asset-backed preflight, do any mappings appear only in `phase=final` collision output? If so, classify those concrete offsets before selecting the next runtime experiment.