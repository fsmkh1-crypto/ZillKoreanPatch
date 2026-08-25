# Zill Infinite Plus font/archive findings and Korean encoding plan

Target: **ULJM-05410 v1.03**

Upstream builder baseline: **HK47196/zill@a98d9ce29f361d666ec23da0dcfd351f24537ffd**.
The upstream code is GPL-3.0-or-later and upstream translation content is CC BY-SA 4.0; retain the original license and attribution files when importing or redistributing derived work.

## Corrected PAF key interpretation

A previous PoC incorrectly interpreted the PAF record field at `+0x00` as a Unicode code point. That conclusion is withdrawn.

`PAF +0x00 u16` is the renderer lookup key formed from the original CP932/Shift-JIS byte sequence, stored as a little-endian integer. For example, PAF value `0xAC82` represents bytes `82 AC`; it is not Unicode U+AC82.

The proved startup PoC is an especially useful concrete example: Japanese `の` is CP932 bytes `82 CC`, therefore its PAF renderer key is `0xCC82`. The authenticated retail PAF record for this key is index 2034 and points to page 1, atlas `(421,379)`, size `10x10`. Replacing only that atlas cell with a Hangul `가` bitmap rendered `가` in PPSSPP while the key, PAF record and message bytes stayed unchanged.

## Korean text strategy

1. Build the final Korean character set from the actual Korean translation.
2. Deterministically map each required Korean Unicode character to a reusable two-byte CP932 renderer key.
3. Teach the message compiler to serialize Korean text to those custom two-byte keys while leaving control constructs intact.
4. Replace the bitmap/metrics associated with exactly the same PAF keys with generated Korean glyphs.
5. For the first output PoC, do not change PAF record count or BST shape/root.
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

Validated font members are identified by **PAA record index, name, size, and upstream retail SHA-256**, while their ARC offsets are read from `pa.bin` rather than hardcoded:

- record #13611: `font/zillfont.par`
  - member size `0x80470`
  - observed retail PAA offset `0x3D8F520`
  - retail SHA-256 `0d3d6d2648870e87a01636cdfc7cc7af8100ea40b71e5ed05f82ac197606584a`
- record #13612: `2d/font/jillbtn.par`
  - member size `0x18E60`
  - observed retail PAA offset `0x3E0F990`
  - retail SHA-256 `95b48379092db4db72f890d5a221ba8c4094dd438cb4c4eba98eb5520c7b17aa`

## PAR

`2d/font/jillbtn.par` contains the `zillfont.paf` sub-block. Direct inspection of the authenticated retail member corrects an earlier 0x10 boundary error:

- PAF magic begins at member-relative **`0x4490`**
- PAF size is **`0x149D0`**
- `0x4490 + 0x149D0 = 0x18E60`, exactly the complete `jillbtn.par` size

`font/zillfont.par` contains four 512x512 4bpp GIM atlas pages: `zillfont_0.gim` through `zillfont_3.gim`.

## PAF

- magic: `paf\0`
- version: `0x000C0201`
- glyph count in header: 2,637
- BST root in header: 1,318
- record offset: `0x30`
- record stride: `0x20`
- all 2,637 records are complete; there is **no truncated final-tail anomaly**

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

Validated invariants from the authenticated retail member:

- keys strictly ascending; duplicates: 0
- child range errors: 0
- BST ordering errors: 0
- cycles: 0
- all 2,637 records reachable from root
- page distribution: page 0 = 1,156; page 1 = 1,156; page 2 = 325
- no record references page 3
- the record table ends exactly at `jillbtn.par` EOF

The Korean PoC preserves these structural invariants and uses slot replacement rather than PAF expansion.

## Extraction / patching boundary

The Android extractor validates `PARAM.SFO` for ULJM-05410 v1.03, parses `pa.bin`, requires the expected member count and font member index/name/size, reads each member from the offset supplied by the PAA index, and verifies the complete member bytes against the upstream retail SHA-256 before export. The ISO is opened through Android SAF in read-only mode.

The Android PoC patcher never overwrites the source ISO. It streams the full source into a newly created ISO and changes only declared bytes inside `font/zillfont.par`. The first proved PoC changed exactly 60 atlas bytes for the `の -> 가` cell; PAF bytes and message bytes were untouched.
