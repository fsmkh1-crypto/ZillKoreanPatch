# Preliminary Analysis — Korean Name Input

## 1. Verified observations

The Korean fan patch has an explicit Korean name-entry feature rather than merely translated static text.

Public development posts dated 2026-07-29 show attached screenshot filenames containing:

- `01_final_iso_korean_osk_kim_composed(1).jpg`
- `15_final_name_ui_iso_game_given_name_taeho.jpg`
- `30_final_name_ui_iso_surname_field_kim_fullscreen_verified(1).jpg`
- `25_final_name_ui_iso_given_name_taeho_confirmed.jpg`

The post titles also explicitly describe Korean-name input and the resolution of the name-setting portion.

This is strong evidence that the patch author added or replaced the game's name-entry path with a Korean **OSK (on-screen keyboard)** capable of producing at least composed Hangul surname/given-name strings such as `김` and `태호`.

### What is NOT yet verified

Without the patch package or a patched-vs-clean binary diff, it is not yet proven whether the implementation:

1. performs true choseong/jungseong/jongseong composition at runtime,
2. uses a lookup table of precomposed Hangul syllables,
3. maps a smaller custom keyboard to a pre-generated syllable table,
4. hooks Sony/PSP OSK routines, or
5. implements the keyboard entirely inside game code/UI scripts.

The filename token `korean_osk_kim_composed` strongly suggests composition behavior, but a filename alone is not enough to prove the algorithm.

## 2. Why this is technically interesting

A normal Japanese PSP name-entry screen can rely on kana/kanji tables that match the original game's encoding and font assumptions. Korean input is harder because modern Hangul has 11,172 precomposed syllables, while a practical Korean keyboard needs only a small set of jamo plus a composition state machine.

If this patch really composes Hangul interactively, it avoids an enormous flat syllable-selection UI and is substantially more elegant than simply exposing thousands of glyphs.

## 3. Binary signatures to look for

When the exact patch package or patched ISO components become available, compare against a clean `UCJS-10117 v1.01` image and inspect at least:

### EBOOT / executable

- New input-handler branches near the original name-entry code.
- New state variables for initial/medial/final jamo.
- New lookup tables for Korean keyboard cells.
- Constants associated with Hangul composition.

If Unicode-style arithmetic is used, diagnostic constants/formulas may include:

- Hangul syllable base: `0xAC00`
- 19 choseong
- 21 jungseong
- 28 jongseong states including no-final
- `588 = 21 * 28`
- composition pattern equivalent to:

  `SBase + (LIndex * 21 + VIndex) * 28 + TIndex`

Presence of these constants would be strong evidence of algorithmic Hangul composition. Their absence would not disprove composition because the patch may use a custom game encoding.

### Font / glyph resources

- Added Hangul glyph atlas pages.
- Width table expansion.
- Character-code-to-glyph-index tables.
- Whether only used syllables are packed or a broad Hangul subset/full set is present.

### Name UI resources

- Keyboard textures/layout definitions.
- Cursor-grid dimensions changed from Japanese OSK layout.
- New labels for surname/given-name fields.
- Any jamo keyboard pages or mode switching.

## 4. High-value experiment once binaries are available

1. Hash all files in clean v1.01 and patched build.
2. Identify changed ISO members.
3. Diff `EBOOT.BIN`/decrypted executable first.
4. Search changed code/data for UTF-8/UTF-16 Korean, jamo, `0xAC00`, 19/21/28/588 constants.
5. Extract image/font assets and compare dimensions/counts.
6. Trace the name-entry function in PPSSPP debugger:
   - enter `ㄱ`
   - enter `ㅣ`
   - observe transition to `기`
   - add `ㅁ`
   - observe `김`
   - then test final-consonant migration by adding a vowel after a final.
7. Record memory writes to the name buffer after each keystroke.
8. Determine whether the stored name is Unicode, Shift-JIS-compatible custom code, or patch-specific glyph indices.

This dynamic trace will distinguish a real IME-like state machine from a simple precomposed lookup table very quickly.

## 5. Preliminary assessment

**Confidence: medium-high** that the patch contains a custom Korean name-entry/OSK implementation.

**Confidence: low** on the exact composition algorithm until the patch binary is recovered and diffed.

The implementation is worth preserving and documenting because a compact Korean IME/OSK for a PSP-era game can potentially be generalized to other Korean fan-translation projects if its code is sufficiently isolated from Oreshika-specific UI and encoding assumptions.
