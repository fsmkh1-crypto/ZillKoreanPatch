# Zill Infinite Plus font/archive findings and Korean encoding plan

Target: **ULJM-05410 v1.03**

Upstream builder baseline: **HK47196/zill@a98d9ce29f361d666ec23da0dcfd351f24537ffd**.
The upstream code is GPL-3.0-or-later and upstream translation content is CC BY-SA 4.0; retain the original license and attribution files when importing or redistributing derived work.

## Corrected PAF key interpretation

A previous PoC incorrectly interpreted the PAF record field at `+0x00` as a Unicode code point. That conclusion is withdrawn.

`PAF +0x00 u16` is the renderer lookup key formed from the original CP932/Shift-JIS byte sequence, stored as a little-endian integer. For example, PAF value `0xAC82` represents bytes `82 AC`; it is not Unicode U+AC82. Therefore the original font must not be described as containing hundreds of pre-existing Hangul syllables based on those numeric values.

The existing conclusions about PAF BST structure, page selection, coordinates and metrics remain valid.

## Korean text strategy

1. Build the final Korean character set from the actual Korean translation.
2. Deterministically map each required Korean Unicode character to a reusable two-byte CP932 renderer key.
3. Teach the message compiler to serialize Korean text to those custom two-byte keys while leaving control constructs intact.
4. Replace the bitmap/metrics associated with exactly the same PAF keys with generated Korean glyphs.
5. For the first output PoC, do not change PAF record count, BST shape/root, final truncated tail behavior, or page 3 usage.
6. Only investigate PAF growth after measuring the unique Korean glyph set and proving that reusable slots are insufficient.

Known slot counts from the English release analysis:

- total installed glyph slots: 2,637
- two-byte slots: 2,487
- English release referenced glyphs: 1,240 total / 1,116 two-byte
- unreferenced by current English text: 1,397 total / 1,371 two-byte

These counts establish PoC capacity only; they do not prove capacity for the complete Korean script.

## PAA / ARC

- `PSP_GAME/USRDIR/pa.bin`
- `PSP_GAME/USRDIR/pa.arc`
- `pa.bin` magic: `PAA\0`
- header size: `0x20`
- record size: `0x10`
- record count: 14,231
- ARC members are stored sequentially with 16-byte alignment.

### Offset/wrapper correction from retail ISO validation

The first on-device extractor run against the user's retail ULJM-05410 v1.03 ISO showed that the PAA index record for member #13611 points to `0x3D8F520`, not the previously recorded `0x3D8F510`. This exposed an earlier convention error: the 16 bytes immediately before the indexed member were incorrectly treated as a wrapper belonging to every member.

The PAA index member offset is authoritative and points at the actual stored member bytes. The upstream English release manifest independently defines the source member itself as size `0x80470` for `font/zillfont.par` and `0x18E60` for `2d/font/jillbtn.par`, with SHA-256 values over those complete member bytes. Therefore the extractor must not add `0x10` to a PAA index offset or strip `0x10` from a member.

This is also forced by the `jillbtn.par` geometry: PAF starts at member-relative `0x44A0` and has size `0x149C0`; `0x44A0 + 0x149C0 = 0x18E60`, exactly the complete member size. There is no room for an additional member-local 16-byte wrapper.

Validated font members are therefore identified by **PAA record index, name, size, and upstream retail SHA-256**, while their ARC offsets are read from `pa.bin` rather than hardcoded:

- record #13611: `font/zillfont.par`
  - member size `0x80470`
  - observed retail PAA offset `0x3D8F520`
  - retail SHA-256 `0d3d6d2648870e87a01636cdfc7cc7af8100ea40b71e5ed05f82ac197606584a`
- record #13612: `2d/font/jillbtn.par`
  - member size `0x18E60`
  - offset is read from the PAA index
  - retail SHA-256 `95b48379092db4db72f890d5a221ba8c4094dd438cb4c4eba98eb5520c7b17aa`

## PAR

`2d/font/jillbtn.par` contains:

- sub-block 0: `JillBtn.gim`
- sub-block 1: `JillBtn.qtx`
- sub-block 2: `zillfont.paf`
- PAF begins at member-relative `0x44A0`
- PAF size `0x149C0`

`font/zillfont.par` contains four 512x512 4bpp GIM atlas pages: `zillfont_0.gim` through `zillfont_3.gim`.

## PAF

- magic: `paf\0`
- version: `0x000C0201`
- glyph count: 2,637
- BST root: 1,318
- record offset: `0x30`
- record stride: `0x20`

Record layout:

```text
+0x00 u16  CP932/Shift-JIS renderer key
+0x02 u8   bitmap width
+0x03 u8   bitmap height
+0x04 u16  atlas X
+0x06 u16  atlas Y
+0x08 s16  bearing X
+0x0A s16  bearing Y
+0x0C u32  advance
+0x10 s32  BST left child
+0x14 s32  BST right child
+0x18 u32  GIM page
+0x1C u32  0
```

Validated invariants:

- keys strictly ascending; duplicates: 0
- child range errors: 0
- BST ordering errors: 0
- cycles: 0
- all 2,637 records reachable from root
- page distribution: page 0 = 1,156; page 1 = 1,156; page 2 = 324
- no complete record references page 3
- applying the page field reduces atlas overlap to two pixels, strongly confirming that `+0x18` is the GIM page index
- final glyph record has only the 0x10-byte core; the final 0x10-byte tail is omitted at EOF

The first Korean rendering PoC must preserve these structural invariants and use slot replacement rather than PAF expansion.

## Extraction APK boundary

The Android extractor is deliberately read-only. It validates `PARAM.SFO` for ULJM-05410 v1.03, parses `pa.bin`, requires the expected member count and font member index/name/size, reads each member from the offset supplied by the PAA index, and verifies the complete member bytes against the upstream retail SHA-256 before export. It writes the two complete PAR members and `manifest.json` to a ZIP. No member-local wrapper bytes are stripped. The ISO is opened through Android SAF in read-only mode and is never overwritten.
