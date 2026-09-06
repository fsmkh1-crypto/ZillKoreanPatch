# Oreshika PSP Korean Patch – Hangul Input Research

This directory isolates research on the Korean fan patch for **俺の屍を越えてゆけ (Ore no Shikabane wo Koeteyuke)** from the Zill O'll patch project.

## Target observed by the user

- Platform: PSP / PPSSPP
- Game ID: `UCJS-10117`
- Game-reported version: `1.01`
- Patched ISO CRC reported by PPSSPP: `7179FB43`
- Final patch package date: `2026-08-01`

## Exact package recovered

The user's Google Drive `작업중` folder contains the exact final distribution package:

`Oreshika_R29_Korean_M14_DLC13_FINAL_Xdelta_20260801.zip`

Locally verified ZIP SHA-256:

`00A50B4B292489476109030B6E7FA1871234A199CCB5A4247BE82877DC5F689C`

The package contains an xdelta patch plus verification documents and patching scripts; it does **not** contain a full game ISO.

The package manifest identifies:

- required clean ISO SHA-256: `03791A833656DDEBADFF7848A6296FC1698B52DFD7052C91572DD8B800FD569A`
- expected final ISO SHA-256: `5B76F8AC94CF4F6AF1ABB48939C33B1B59551801452F247B00FA43E5A7FAAA75`
- final EBOOT SHA-256: `4827D94FF05C75D6754096D657D39EC14FDC21F6B77019CC9A98D7EAFA4BAB89`
- xdelta SHA-256: `08469300D12E37D39A0966C3BBE7CC751F0FB3F0B1989043CE5E2ECDD18C4BDD`

## Korean name-input status

The custom Korean name-entry system is now **verified to exist as an explicit subsystem**, not merely inferred from screenshots.

The package's final regression QA explicitly identifies:

- Korean OSK initialization in EBOOT file range `0x001288E0..0x0012894F`
- name UI slot at ISO offset `0x25261800`, length `18,432` bytes
- cold-load runtime validation of the Korean given name `태호`
- inherited M12B runtime validation of the full `김 가문 피바람외전` name/title path

The final DLC build changes only 61 approved EBOOT bytes from locked M14, and the package QA records zero intersection between those DLC changes and OSK, PGF/font, name UI, and name serialization ranges. Therefore the Korean input implementation belongs to the earlier M12B/M14 lineage and is preserved unchanged in the final distribution.

## What remains unresolved

The package proves the presence and preservation of the custom Korean OSK, but the exact Hangul composition algorithm is still not proven because the distribution contains only a source-dependent xdelta, not the patched EBOOT itself.

To determine whether it uses:

- true choseong/jungseong/jongseong runtime composition,
- a precomposed-syllable lookup table,
- a patch-specific glyph/encoding map, or
- a hybrid state machine,

we need either the exact clean source ISO matching the manifest hash or an extracted final `EBOOT.BIN` from the patched output.

## Reusable PSP Koreanization method

The Oreshika work is also preserved as a **general reconnaissance method for future PSP Koreanization projects**.

The standing rule is to inspect, in order, the game's existing text representation, input/OSK path, renderer/font consumer, and save/load serialization before inventing a custom Hangul renderer or CP932 slot map.

If a game-wide 16-bit/Unicode-capable path is actually proven, reuse or extend that path first. If the game is fundamentally CP932/custom-font keyed like Zill O'll, retain the legacy architecture and use slot remapping only where that is the safer fit.

See `PORTABLE_KOREANIZATION_METHOD.md` for the full decision tree, proof requirements, and validation checklist.

See:

- `ANALYSIS.md` — current technical assessment
- `PACKAGE_EVIDENCE.md` — recovered-package hashes and QA evidence
- `XDELTA_FORENSICS.md` — VCDIFF/xdelta structural findings
- `SOURCES.md` — public evidence and development posts
- `PORTABLE_KOREANIZATION_METHOD.md` — reusable PSP Koreanization reconnaissance and architecture-selection method

## Branch isolation

Research branch: `research/oreshika-korean-input`

Nothing here should be merged into Zill O'll localization work without an explicit decision.
