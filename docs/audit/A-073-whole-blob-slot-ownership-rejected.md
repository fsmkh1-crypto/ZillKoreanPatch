# A-073 — Arbitrary whole-blob renderer ownership gate rejected

## Trigger

A real Android/mobile preflight failed after the English-first restart with:

`mobile beta conservative slot allocation: need 1308 custom glyphs but only 0 reusable two-byte slots are available`

The same run authenticated the expected retail EBOOT/BOOT inputs and reported 2,487 installed double-byte font keys. Therefore the zero-candidate result was produced by the ownership rule itself, not by a missing or unauthenticated font asset.

## Rejected rule

The mobile planner had been changed to pass complete authenticated `BOOT.BIN`, `EBOOT.BIN`, and `bindata.dat` byte blobs to `koreanslots.BuildPlan`, whose `ExcludeExactByteReferences` helper removes a candidate glyph key whenever the same two bytes occur anywhere in any supplied blob.

That rule treated every arbitrary two-byte alias in machine code or opaque binary data as proof that the engine owns the corresponding CP932 glyph. On multi-megabyte executable/data inputs this excluded all reusable slots. The premise was over-broad and is rejected.

This is the same class of mistake as applying a consumer-specific scanner contract universally: authentication of a blob proves provenance, not semantic interpretation of every byte pair inside the blob.

## Corrected contract

Renderer ownership now uses evidence that is scoped to data the engine actually interprets as text:

- exact fixed-string keys already known from structured fixed-data tables;
- CP932 literal keys recovered by `slotaudit.ScanCP932Literals` from authenticated `BOOT.BIN`;
- CP932 literal keys recovered from authenticated `bindata.dat` after the 132×17-byte equipment layout/source guards pass;
- stock glyph keys required by the final materialized Korean text itself.

Complete BOOT/EBOOT/BINDATA byte blobs are no longer release-blocking exact-byte exclusion inputs.

Whole-blob exact-byte collision scans remain forensic telemetry only. A collision can motivate a targeted reverse-engineering check, but cannot by itself reserve a renderer slot or fail a build.

## Code changes

- `cmd/zill/build_korean_mobile_plan.go`
  - uses `BuildPlan(texts, installed, rendererKeySetSlice(reserved))`;
  - retains authenticated BOOT/EBOOT/BINDATA loading;
  - reserves structured BOOT/BINDATA CP932 literal scans;
  - keeps initial/final whole-blob collision reports as `diagnostic-only` without failing.
- `cmd/zill/build_korean.go`
  - uses the same structured renderer-ownership policy.
- English-first parity/entrypoint audits now fail if the disproven whole-blob hard gate is reintroduced.

## Evidence interpretation

This failure is strong evidence that the previous whole-blob ownership gate was invalid as a production invariant. It is not evidence that renderer-slot reuse is universally safe, nor does it prove the runtime freeze is fixed.

A future concrete renderer consumer discovered in BOOT/EBOOT/BINDATA must be added as structured ownership evidence for that consumer rather than reintroducing arbitrary byte-pair exclusion across the entire binary.
