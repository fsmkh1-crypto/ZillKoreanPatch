# A-070 — English-first Android payload/export provenance

Status: **MISSING → PASS** for Android transport provenance; shared English/Korean ISO authoring remains **PASS**.

## Scope

This audit follows the English-first project premise through the final Android delivery boundary. The upstream English release and Korean desktop/mobile release already converge on the shared verified PAA rebuild and `authorTranslatedISO` path. Android adds two transport surfaces after those engine-facing contracts: the APK-embedded project tree and the final Storage Access Framework output URI.

These Android-specific checks are **KOREAN-ADDITIONAL** contracts. They do not change an English engine rule; they ensure the exact assets and ISO already validated by the shared engine path are the bytes the Android app actually consumes and exports.

## Finding 1 — cached APK project payload was version-addressed, not content-addressed

Before this audit, `MainActivity.ensureProjectAssets` could reuse `files/zillroot-beta-current` when `payload-version.txt` matched the packaged version. `ProjectAssetIntegrity` checked a required-file list, but the cache did not prove every copied payload byte or reject extra stale files.

That was a provenance gap: the Go builder could receive a same-version but incomplete, stale, or modified project tree.

### Fix

The Android release workflow now creates `assets/zillroot/payload-manifest.sha256` over the complete embedded project tree, verifies it before APK build, extracts the finished APK, and verifies the same manifest again against the packaged assets.

`ProjectAssetIntegrity.verifyPayload` now:

- requires the manifest and critical project files;
- validates each manifest digest/path strictly;
- verifies SHA-256 for every manifest member;
- rejects missing members;
- rejects extra/stale files not represented by the manifest.

Both `PayloadRepairApplication` startup repair and `MainActivity.ensureProjectAssets` verify the cached/copied tree before it can reach `libzill.so`.

Classification: **MISSING → PASS**.

## Finding 2 — exported SAF document was not re-opened after copy

The Go mobile build already authors and reopens the temporary output ISO through the shared `authorTranslatedISO` provenance path, which byte-compares the staged PSP_GAME tree with the ISO. Android then copied that verified ISO to a user-selected `content://` URI but previously reported success immediately after the copy.

A storage provider truncation or write anomaly at this last transport boundary therefore was not independently detected.

### Fix

After `copyFileToUri`, Android now reopens both the verified temporary ISO and the exported URI, computes exact byte length and SHA-256 for each, and reports success only when both match. Any mismatch fails the operation and enters the existing cleanup path for the destination document.

Classification: **MISSING → PASS**.

## Entrypoint result

The production launcher remains `MainActivity`. Its builder invocation is the same `libzill.so build-korean-iso` command whose Go command path reaches `BuildKoreanAlphaISOOnly`; `--preflight-only` reaches `PreflightKoreanAlphaISOOnly`. Those two release functions are already protected by the English-first consumer/storage gates before bank compilation.

`FreezeCaptureActivity` and `FreezeTraceService` are forensic utilities, not alternate ISO patching entrypoints. They do not author a Korean ISO and therefore are not substitutes for the production path.

## Evidence interpretation

This closes deterministic Android asset/export provenance gaps. It does **not** establish that every runtime freeze is fixed. A successful APK build or successful game run remains non-reproduction evidence unless the specific runtime mechanism is independently proven.
