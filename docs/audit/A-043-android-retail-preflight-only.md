# A-043 — Android authenticated retail preflight-only path

## Trigger

A-041/A-042 made the Android patcher preserve forensic output, refresh embedded project data after APK updates, and include the runtime files required by the new audits. The remaining workflow still required a complete translated ISO to be authored before the user could retrieve those static results.

That made the expensive archive/ISO write path a prerequisite for evidence that is logically available earlier.

## Hypothesis

The mobile build can expose a lower-cost asset-backed gate that uses the same authenticated retail source, bound banks, Korean projection, final Minimal87 slot plan and deterministic generation checks as the real ISO builder, but stops before rebuilding archives or authoring an output ISO.

## Verification / implementation

- Added `release.PreflightKoreanAlphaISOOnly`.
- The preflight:
  - validates the source ISO through the same `openRetailISO` contract used by the real mobile build,
  - loads the canonical source/Korean projects,
  - opens retail archives and binds authenticated retail banks,
  - builds the same runtime-materializable Korean project,
  - invokes the same mobile plan builder used by the real ISO build,
  - therefore runs C5 runtime candidate, PR14 historical-policy, H0/final exact-byte collision and mapping-fingerprint audits,
  - validates consumer-specific Korean storage,
  - compiles the Korean banks and checks replacement ownership,
  - prepares the mobile font replacements (including their existing semantic checks),
  - builds the Korean executable and PARAM.SFO in memory,
  - then stops before archive rebuilding, PSP_GAME staging and ISO authoring.
- Added `build-korean-iso --preflight-only`; this mode rejects `--out` so it cannot accidentally be described as having produced a patched ISO.
- Added an Android `RETAIL 진단만 실행` action that:
  - copies the already inspected source ISO into app-private temporary storage,
  - invokes `build-korean-iso --preflight-only`,
  - captures every `FORENSIC` line,
  - exposes those results through the existing `진단 로그 복사` action,
  - deletes the temporary ISO/extracted tree afterward,
  - never asks for or writes a destination ISO.
- Added explicit terminal markers:
  - `FORENSIC MOBILE_PREFLIGHT_BEGIN output_iso_written=false`
  - `FORENSIC MOBILE_PREFLIGHT_BANKS ...`
  - `FORENSIC MOBILE_PREFLIGHT_FONT ...`
  - `FORENSIC MOBILE_PREFLIGHT_EXECUTABLE ... output_iso_written=false`
  - `FORENSIC MOBILE_PREFLIGHT_OK ... output_iso_written=false`
  - `FORENSIC MOBILE_PREFLIGHT_COMPLETE output_iso_written=false`

## Result

The next authenticated-retail evidence collection no longer requires a playable or even fully authored Korean ISO. It can stop at the deterministic pre-runtime boundary and return the forensic evidence directly.

This deliberately does **not** prove runtime safety. A successful preflight means only that the authenticated retail source, bank binding, corpus/materialization, slot plan, static storage checks, font generation semantics and executable/SFO generation all completed on that asset without producing an output ISO.

## Evidence grade

**CONFIRMED** for the code-path separation once Go/Android CI passes.

**OPEN** for the actual authenticated-retail result until the preflight is run against the supported retail ISO.

## What this excludes

Nothing about the freeze mechanism by itself. It removes ISO authoring and device gameplay as prerequisites for the next static evidence gate.

## New question

When the supported retail ISO is run through this preflight, which remaining branch survives the combined asset-backed evidence: substitution/message staging, renderer/font lookup/relocation, exact-byte slot ownership, or another mechanism?
