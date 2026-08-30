# A-042 — Android embedded forensic dependency closure

## Trigger

A-041 made the Android wrapper preserve retail `FORENSIC` output, but APK compilation alone does not prove the embedded project root contains every file now required by the expanded preflight. The Android workflow historically embedded only translations, Korean font/EBOOT inputs, two release string tables, and executable/system patch manifests.

## Verification

The current preflight directly requires additional repository-root files at runtime:

- layout loading / C5 consumer analysis requires `release/layout` data, including `consumer-map.toml` and category/rule inputs;
- `auditPR14HistoricalPolicies()` explicitly reads `docs/audit/fixtures/pr14-eboot-full.toml`.

Before this correction, neither `release/layout/**` nor the PR14 fixture was copied into `assets/zillroot`. Therefore an APK could compile and package successfully while the real phone-side `build-korean-iso` later failed during asset-backed forensic preflight.

## Fix

The Android workflow now:

- copies the complete `release/layout` directory into the embedded project root;
- copies `docs/audit/fixtures/pr14-eboot-full.toml` into the matching embedded path;
- adds both paths to workflow push/pull-request triggers;
- adds explicit pre-package `test -f` checks for the key runtime dependencies;
- adds APK-content assertions for:
  - `release/layout/consumer-map.toml`
  - `release/layout/categories.toml`
  - `docs/audit/fixtures/pr14-eboot-full.toml`
  - the A-041 payload-version marker.

## Result

The Android package gate now verifies the files needed by the current retail forensic preflight are physically present in the APK rather than relying on successful Java/Go compilation as a proxy.

This closes a concrete packaging/runtime gap. It does not execute the real retail ISO path in CI because CI still has no game asset.

## Evidence grade

- **CONFIRMED** for dependency identification from current source paths.
- **CONFIRMED** for workflow embedding/assertion behavior once Android Actions passes.
- **OPEN** for the real asset-backed preflight result until the app is run against the supported retail ISO.

## New question

After Android CI passes, the next high-value step is no longer another speculative scanner. Use the current APK with the supported retail ISO, complete one local ISO build, copy `진단 로그`, and analyze the authenticated static evidence before deciding whether gameplay or a live debugger is necessary.
