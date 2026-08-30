# A-041 — Android retail forensic log handoff and embedded-payload freshness

## Trigger

The integrated `build-korean-iso` path already runs the authenticated retail EBOOT C5, `$15`, and font scanners on the user's real ISO. However, the Android wrapper consumed the builder's merged stdout/stderr one line at a time, displayed only the current line, retained only a 12-line failure tail, and discarded all successful forensic output. Separately, `ensureProjectAssets()` reused `zillroot-beta-v6/.ready` indefinitely, so installing a newer APK over an existing installation could continue using an older extracted translation/project payload.

## Hypotheses

1. Asset-backed static evidence can be collected from the user's existing retail ISO without gameplay if the Android wrapper preserves `FORENSIC` lines after a successful local build.
2. An APK-local payload version marker can prevent stale extracted translations/configuration from surviving an app update.

## Verification

### Forensic output preservation

`MainActivity` now:

- captures every builder output line whose prefix is `FORENSIC`;
- preserves the captured text after both successful and failed builds;
- exposes a disabled-by-default `진단 로그 복사` button;
- enables the button whenever at least one forensic line was recovered;
- copies the complete captured forensic text to the Android clipboard;
- leaves the status text selectable as a secondary manual path.

The builder itself is unchanged by this UI capture: C5/value/font candidate scoring and all evidence grades remain exactly as before.

### Embedded project payload freshness

The Android workflow now writes `assets/zillroot/payload-version.txt` containing the checkout `GITHUB_SHA` after assembling the embedded translation/release payload.

`ensureProjectAssets()` now:

1. reads the APK-packaged `payload-version.txt`;
2. compares it with the extracted copy under `zillroot-beta-current`;
3. reuses the extracted project only when the versions match exactly;
4. otherwise deletes the stale extracted tree and copies the current APK assets again;
5. verifies the copied version marker before invoking the Korean builder.

The APK payload inspection step also requires the version marker to exist in the built package.

## Result

A future Android patch run can produce the asset-backed static evidence needed by A-037/A-040 without requiring the user to play the game first or manually extract EBOOT.BIN. After the ISO build, the user can copy the forensic log and return it for analysis.

The same run is also protected against silently compiling with an older extracted translation/release payload after installing a newer APK.

## Evidence grade

- **CONFIRMED** for source-level forensic capture and version-marker logic once Android/CI builds pass.
- **OPEN** for the actual retail scanner results until the app is run against the SHA-256-supported retail ISO.
- **OPEN** for runtime freeze causality; collecting the log is a pre-game static gate, not a gameplay result.

## What this excludes

- Losing successful retail scanner output merely because the Android status view advances to later build lines.
- Assuming that an updated APK necessarily refreshed the previously extracted `zillroot-beta-v6` project data.

It does not prove that every Android filesystem/provider behaves identically at runtime, nor does it turn heuristic candidates into runtime proof.

## New question

Once Android CI passes, run one local ISO build with the current APK, copy the forensic log, and inspect the authenticated EBOOT candidate results plus final exact-byte mapping audit. Only then decide whether the next expensive step should be static disassembly of a surviving candidate or a narrowly instrumented PPSSPP/device experiment.
