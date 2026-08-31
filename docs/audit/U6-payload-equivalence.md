# U5 / U6 payload-equivalence closure

## Purpose

U6 was intended to reduce uncertainty by adding warning-classification instrumentation, not to create a new Korean runtime patch. Because U6 did modify shared production-source files for observational logging, this note records the strongest byte-level equivalence evidence available without authenticated retail game assets.

Baseline U5 functional commit: `838f1b965e89e916d2c780a0e7c70ec0126d4bc7`.

U6 audit milestone used for the comparison: `24bb3cb573c5405185ec4da0e5da195c31cf68c2`.

## Production-source delta

The U5 -> U6 repository diff changes only audit/test/docs/workflow surfaces plus two shared Go source files used for classification/logging:

- `internal/layout/warning_audit.go` adds observational warning ownership/token census helpers.
- `internal/release/korean_contract_chain.go` adds `FORENSIC` warning-summary output after layout derivation, consumer validation, warning collection, runtime-storage validation, and `out.Layouts = layouts` have already completed.

The added code does not assign to `layouts`, `out.Layouts`, Korean translation text, glyph mappings, fixed strings, or runtime patch manifests.

Accordingly U6 should be described as **payload-neutral audit instrumentation**, not as source-untouched.

## Android embedded project-data byte comparison

Compared workflow artifacts:

- U5 Android RC artifact: `9745661709`
- U6 Android RC artifact: `9749484704`

For each artifact, `app-debug.apk` was extracted and the complete `assets/zillroot/` tree was compared byte-for-byte after excluding only commit-specific metadata:

- `payload-version.txt` — intentionally contains the build commit SHA.
- `payload-manifest.sha256` — necessarily changes because it includes the hash of `payload-version.txt`.

Result:

- compared embedded files per build: **617**
- differing embedded patch-project files: **0**
- deterministic composite SHA-256 for the 617-file tree: `6126493b93bd47d67ae8ea1fc3460758c127acf2a916cbbbe4f89eb3d9fd20f3` for both U5 and U6.

This proves that the Android-embedded Korean translation/release/layout/patch project data consumed by the runtime patcher is byte-identical between final U5 and U6, apart from commit-identification metadata.

## What this does not prove

The whole APK digest is not expected to match because U6 changes Go source and therefore the embedded `libzill.so`, and because `payload-version.txt` records a different commit SHA.

Likewise this repository/Android-artifact comparison does **not** claim byte-for-byte identity of a patched retail `PSP_GAME/`, ISO, or xdelta output. Producing that stronger comparison requires applying both U5 and U6 builders to the same authenticated retail input, which is not present in repository CI.

The current strongest non-device evidence is therefore the conjunction of:

1. byte-identical 617-file embedded patch-project data,
2. source-diff inspection showing U6's shared-path additions are observational after the payload-affecting derivation/validation steps,
3. full standard CI and Korean consumer/storage/visual census success.

A retail-output byte comparison may be added later if the same authenticated retail input is available to both builds, but its absence is not evidence of a payload delta.
