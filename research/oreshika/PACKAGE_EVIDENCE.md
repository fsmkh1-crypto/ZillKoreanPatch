# Recovered Final Package Evidence

## Package recovered from user's Drive

Drive folder: `작업중`

Recovered filename:

`Oreshika_R29_Korean_M14_DLC13_FINAL_Xdelta_20260801.zip`

Observed Drive copy size: `4,273,123` bytes.

Locally computed ZIP SHA-256:

`00A50B4B292489476109030B6E7FA1871234A199CCB5A4247BE82877DC5F689C`

## Distribution contents

The ZIP contains:

- `APPLY_PATCH.bat`
- `apply_patch.ps1`
- `D2_GENERAL_REGRESSION_QA.json`
- `D2_일반회귀_최종게이트_20260801.md`
- `DLC13_STATIC_QA.json`
- `OreShika_UCJS10117_R29_Korean_M14_DLC13_FINAL_20260801.xdelta`
- `PACKAGE_MANIFEST.json`
- `README_먼저_읽기.md`
- `SHA256SUMS.txt`
- `최종_검증_보고서_20260801.md`
- `tools/xdelta3.exe`

It does **not** contain a full copyrighted PSP ISO.

## Manifest identity

Schema:

`oreshika.r29.m14.dlc13.final.package.v1`

Release date:

`2026-08-01`

Status:

`GO_WITH_VERIFIED_SCOPE`

### Declared scope

- Korean localization: yes
- original unsubtitled movies: yes
- Korean manual name entry: yes
- DLC gods unlocked: 13
- bag expansion: no
- automatic warehouse overflow: no
- dungeon timer multiplier: no

## Cryptographic identities

### Required clean source ISO

- product code: `UCJS-10117`
- size: `847,937,536` bytes
- SHA-256: `03791A833656DDEBADFF7848A6296FC1698B52DFD7052C91572DD8B800FD569A`

### Expected final ISO

- size: `847,937,536` bytes
- SHA-256: `5B76F8AC94CF4F6AF1ABB48939C33B1B59551801452F247B00FA43E5A7FAAA75`

### Locked M14 ISO

- size: `847,937,536` bytes
- SHA-256: `67044D03969BBA344A39F29E21CC490D3A8724365704C542CE1C1696BBC35450`

### Final EBOOT

- size: `3,421,440` bytes
- SHA-256: `4827D94FF05C75D6754096D657D39EC14FDC21F6B77019CC9A98D7EAFA4BAB89`

### Final xdelta

- size: `4,111,746` bytes
- manifest SHA-256: `08469300D12E37D39A0966C3BBE7CC751F0FB3F0B1989043CE5E2ECDD18C4BDD`
- independently recomputed from recovered ZIP: same value

## Korean OSK / name-system evidence

The final regression QA explicitly preserves these surfaces:

- Korean OSK initialization: EBOOT `0x001288E0..0x0012894F`
- name UI slot: ISO `0x25261800`, `18,432` bytes

The QA records zero intersection between final DLC-stage modifications and the OSK, PGF/font, name UI, and name serialization surfaces.

The final D2 build differs from locked M14 by only `61` approved EBOOT bytes, all attributed to DLC functionality.

Therefore the Korean OSK/name-input implementation is inherited unchanged from the earlier validated M12B/M14 lineage.

## Runtime evidence recorded by the package

Named checkpoint:

`NAME-KIM-TAEHO-COLDLOAD-PRE`

The D2 regression directly cold-loaded this save and recorded correct `태호` rendering in family roster and status screens.

Referenced direct evidence files:

- `21_valid_name_checkpoint_family_roster_taeho.jpg`
- `22_valid_name_checkpoint_status_taeho.jpg`

The QA also references inherited M12B evidence:

- `runtime_28_osk_kim_coldload_title_PASS.png`
- full `김 가문 피바람외전`
- `태호` cold-load pass

The package explicitly explains that D2 can inherit that full-name result because its 61-byte DLC delta does not touch OSK, PGF, name UI, or name serialization ranges.

## Verification boundary

The package itself states that verification covers its documented static and PPSSPP runtime checkpoints, not every possible long-play branch.

The distribution README prioritizes PPSSPP; PSP hardware and PS Vita Adrenaline are outside its guaranteed scope.

## Repository policy

This research branch records hashes, metadata, technical observations, and sanitized evidence. The recovered binary ZIP/xdelta remains in the user's Drive rather than being republished here until redistribution permission/licensing is known.
