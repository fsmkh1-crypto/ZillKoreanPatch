# Technical Analysis — Korean Name Input

## 1. Current conclusion

The final Korean patch package confirms that **Oreshika has a custom Korean on-screen keyboard (OSK) and manual Korean surname/given-name entry path implemented in the patched game**.

This is no longer a hypothesis based only on public screenshots. The recovered final package's own regression QA explicitly tracks the Korean OSK, PGF/font, name UI, and name serialization as protected ranges/surfaces.

### Verified static anchors

- Korean OSK initialization: EBOOT file range `0x001288E0..0x0012894F`
- Name UI slot: ISO offset `0x25261800`, size `18,432` bytes
- Final EBOOT size: `3,421,440` bytes
- Final EBOOT SHA-256: `4827D94FF05C75D6754096D657D39EC14FDC21F6B77019CC9A98D7EAFA4BAB89`

### Verified runtime evidence recorded by package QA

- checkpoint: `NAME-KIM-TAEHO-COLDLOAD-PRE`
- direct D2 cold-load evidence: family roster/status displays `태호` correctly
- inherited M12B evidence: full `김 가문 피바람외전` and `태호` cold-load path had already passed

The final D2/DLC13 stage changes only 61 approved EBOOT bytes relative to locked M14 and records **zero overlap** with the Korean OSK, PGF/font, name UI, and name serialization areas. Therefore the name-input implementation itself predates D2 and is inherited from the M12B/M14 lineage.

## 2. Why the implementation is technically interesting

Japanese PSP games normally assume a kana/kanji-oriented name-entry path and game-specific text encoding/font behavior. Korean manual input adds several nontrivial problems:

1. an input UI suitable for Korean characters,
2. mapping input selections to Korean character state,
3. constructing or selecting valid Hangul syllables,
4. rendering them with the patched font system,
5. serializing the resulting Korean name into save data,
6. restoring and rendering it correctly after a cold load.

The package's QA treats these as connected but separately protected surfaces. That strongly suggests this is an engineered input/serialization path rather than a cosmetic replacement of a Japanese keyboard texture.

## 3. What the package proves vs. what it does not

### Proven

- A Korean manual surname/given-name entry function exists.
- The implementation includes EBOOT-side OSK initialization changes.
- A dedicated name UI asset region exists in the ISO.
- Korean names survive save/load and cold-load testing.
- `김`/`태호`-related runtime evidence exists in the validated lineage.
- The final DLC13 changes do not alter the Korean input subsystem.

### Not yet proven

We still cannot claim which exact Hangul-composition algorithm is used.

Possible implementations include:

1. **real jamo composition state machine** — choseong/jungseong/jongseong are tracked and a syllable is assembled at runtime;
2. **precomposed lookup** — selected consonant/vowel states index a table of completed Hangul syllables;
3. **custom game-encoding lookup** — composition occurs logically, but the stored value is a patch-specific glyph code rather than Unicode;
4. **hybrid design** — small jamo/state tables drive a larger pre-generated syllable/glyph table.

The public development filename `korean_osk_kim_composed` is suggestive of real composition, but it is not sufficient proof of implementation details.

## 4. Unicode arithmetic hypothesis

If the patch uses standard modern-Hangul Unicode arithmetic, useful signatures would include:

- syllable base `0xAC00`
- 19 initial consonants
- 21 medial vowels
- 28 final-consonant states including none
- `588 = 21 * 28`
- arithmetic equivalent to:

`SBase + (LIndex * 21 + VIndex) * 28 + TIndex`

These remain **search signatures only**. Their absence would not disprove algorithmic composition because the game may use a custom encoding or lookup table.

## 5. Xdelta structural observations

The recovered patch is a valid VCDIFF/xdelta stream whose decoded target size is exactly the manifest's final ISO size (`847,937,536` bytes).

The name UI slot at ISO `0x25261800` falls inside target window 74, which covers `0x25000000..0x257FFFFF`. That window has substantial compressed delta content, consistent with modified assets/data in that region, but xdelta compression prevents extraction of the resulting bytes without the source ISO.

The EBOOT-side OSK changes necessarily fall in the early part of the ISO image; the first VCDIFF window is also by far one of the most heavily patched windows. This is consistent with executable changes, but is not by itself a mapping from ISO offset to EBOOT file offset.

See `XDELTA_FORENSICS.md` for details.

## 6. Next reverse-engineering step

The decisive next step is to obtain either:

- the clean source ISO matching SHA-256 `03791A833656DDEBADFF7848A6296FC1698B52DFD7052C91572DD8B800FD569A`, then apply the recovered xdelta; or
- the final patched `EBOOT.BIN` matching SHA-256 `4827D94FF05C75D6754096D657D39EC14FDC21F6B77019CC9A98D7EAFA4BAB89`.

Then:

1. extract/decrypt the executable as needed;
2. disassemble around file offset `0x001288E0..0x0012894F`;
3. identify branches/calls from the original Japanese name-entry path;
4. locate tables referenced by the new code;
5. search for `0xAC00`, 19/21/28/588 or equivalent tables;
6. inspect writes to the name buffer during `ㄱ → ㅣ → ㅁ` / `김` entry;
7. test final-consonant migration (`각` + vowel, etc.) to distinguish a real IME state machine from a flat lookup;
8. inspect saved bytes for `김` and `태호` to identify the serialization encoding.

## 7. Assessment

- **Custom Korean OSK existence:** verified, high confidence.
- **Korean-name save/load integration:** verified by package QA, high confidence.
- **OSK implementation lineage:** M12B/M14, high confidence.
- **Exact composition algorithm:** unresolved.
- **Potential portability to other PSP fan translations:** plausible, but cannot be assessed responsibly until the EBOOT code and tables are recovered.
