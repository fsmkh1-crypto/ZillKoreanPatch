# Korean font PoC reproduction record

Target: **Zill O'll Infinite Plus ULJM-05410 v1.03**

This document records the first PPSSPP-observed Korean glyph PoC so it is reproducible rather than a one-off manual experiment.

## Observed result

The Japanese startup-screen glyph `の` rendered as Hangul `가` in PPSSPP.

- observation: real game boot in PPSSPP, not an atlas viewer
- PPSSPP version: **not recorded yet**
- source ISO: authenticated ULJM-05410 v1.03 retail image selected read-only through Android SAF
- output ISO: a separate newly-created ISO; the source ISO was never overwritten
- observed scope: the original one-glyph `の -> 가` PoC only; the later eight-glyph smoke build is separately tracked and still requires on-device observation

## Authenticated retail inputs

Extractor output:

- `font/zillfont.par`
  - size: `0x80470` / 525,424 bytes
  - SHA-256: `0d3d6d2648870e87a01636cdfc7cc7af8100ea40b71e5ed05f82ac197606584a`
  - PAA index: 13611
  - `pa.arc` member offset: `0x3D8F520`
- `2d/font/jillbtn.par`
  - size: `0x18E60` / 101,984 bytes
  - SHA-256: `95b48379092db4db72f890d5a221ba8c4094dd438cb4c4eba98eb5520c7b17aa`
  - PAA index: 13612
  - `pa.arc` member offset: `0x3E0F990`

## Renderer-key proof

The PAF sub-block begins at `jillbtn.par + 0x4490`, has size `0x149D0`, and contains 2,637 complete `0x20`-byte records beginning at PAF `+0x30`.

Japanese `の` is CP932 bytes `82 CC`. The PAF stores the byte sequence as a little-endian renderer key, therefore the record key is `0xCC82`; it is not a Unicode code point.

Authenticated retail record:

- record index: 2034
- key: `0xCC82` (`82 CC` = `の`)
- width/height: `10 x 10`
- atlas coordinate: `x=421, y=379`
- bearing: `x=1, y=-9`
- advance: `12`
- BST children: `2033`, `2035`
- GIM page: `1`

The successful PoC did **not** modify this PAF record, its renderer key, the BST, record count, or the startup message bytes. It only changed texels in the referenced `zillfont.par` atlas cell.

## Historical implementation

The first streaming Android patch implementation is preserved in repository history around commit `2db4b42be061a6aca4402f2fab2f880788c14a9e` (`Test streaming Korean font PoC byte patching`). It copied the complete source ISO to a separate output and patched a declared set of 60 bytes inside member #13611.

The currently maintained Android smoke patcher generalizes this into glyph definitions plus PSP 4bpp swizzle address calculation; it no longer requires hand-maintaining a flat offset/value list for each added glyph.

## Exact first-PoC byte delta

The following offsets are **relative to the start of authenticated `font/zillfont.par`**. Retail byte -> patched byte:

`0x380E2: 11→01`; `0x380E3: A1→00`; `0x380E4: FF→00`; `0x380E5: FF→00`; `0x380E6: 1F→18`; `0x380E7: 11→10`

`0x380F2: 11→01`; `0x380F3: FC→00`; `0x380F4: 45→00`; `0x380F5: 1F→00`; `0x380F6: F8→19`; `0x380F7: 11→10`

`0x38102: 81→01`; `0x38103: 4F→FF`; `0x38104: F1→FF`; `0x38105: 1D→04`; `0x38106: 81→19`; `0x38107: 1F→10`

`0x38112: F1→01`; `0x38113: 18→00`; `0x38114: F1→80`; `0x38115: 18→03`; `0x38116: 11→19`; `0x38117: 1F→10`

`0x38122: F1→01`; `0x38123: 11→00`; `0x38124: F1→A0`; `0x38125: 11→00`; `0x38126: 11→F9`; `0x38127: 1F→1B`

`0x388B2: F1→01`; `0x388B3: 11→00`; `0x388B4: F8→94`; `0x388B5: 11→00`; `0x388B6: 11→19`; `0x388B7: 1F→10`

`0x388C2: F1→01`; `0x388C3: 11→30`; `0x388C4: 8F→1C`; `0x388C5: 11→00`; `0x388C6: 81→19`; `0x388C7: 1F→10`

`0x388D2: F1→01`; `0x388D3: 48→B7`; `0x388D4: 1F→01`; `0x388D5: 11→00`; `0x388D6: F5→19`; `0x388D7: 15→10`

`0x388E2: 81→11`; `0x388E3: FF→06`; `0x388E4: 16→00`; `0x388E5: 81→00`; `0x388E6: 8F→19`; `0x388E7: 11→10`

`0x388F2: 11→01`; `0x388F3: 11→00`; `0x388F4: 11→00`; `0x388F5: FF→00`; `0x388F6: 18→04`; `0x388F7: 11→10`

These are exactly 60 changed bytes; all other bytes in the extracted `zillfont.par` remained identical in the first PoC.

## Current eight-glyph smoke mapping

The current startup smoke build uses existing renderer slots and replaces only their atlas cells:

- `の -> 가`: page 1, `(421,379)`, `10x10`
- `無 -> 나`: page 1, `(15,243)`, `12x12`
- `我 -> 다`: page 2, `(435,3)`, `12x11`
- `応 -> 라`: page 1, `(195,108)`, `12x11`
- `答 -> 마`: page 1, `(150,93)`, `12x11`
- `魂 -> 바`: page 1, `(375,213)`, `12x12`
- `示 -> 사`: page 1, `(1,169)`, `11x10`
- `者 -> 아`: page 1, `(136,423)`, `10x11`

The maintained implementation derives byte addresses from `(page, x, y)` using the PSP 4bpp swizzle mapping and applies masked nibble edits while streaming the ISO copy.

## Reproduction boundary

To reproduce the first result:

1. Use an authenticated ULJM-05410 v1.03 source ISO.
2. Validate the two font members by PAA index/name/size/SHA-256.
3. Select the source ISO read-only in the Android patcher.
4. Create a separate output ISO.
5. Apply only the declared atlas edits inside `font/zillfont.par`.
6. Verify output size equals source ISO size.
7. Boot the output in PPSSPP and start a new game.
8. On the fixed startup text screen, the original `の` position must display Hangul `가`.

The PPSSPP application version should be recorded on the next verification pass so the environment record is complete.
