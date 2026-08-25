# First real Korean sentence PoC

Date: 2026-08-25

Status: implementation review passed; PPSSPP observation pending.

This PoC is the first test that combines both halves of the intended Korean renderer path:

1. a retail message record is rewritten to custom two-byte renderer keys;
2. the corresponding existing PAF-key atlas cells are replaced with Korean glyph bitmaps.

Unlike the earlier eight-glyph startup smoke test, it does not rely on the original Japanese message already containing the glyphs being replaced.

## Authenticated audit input

The Android v0.5 audit extractor authenticated and exported these retail assets from ULJM-05410 v1.03:

| Asset | Size | SHA-256 |
| --- | ---: | --- |
| `font/zillfont.par` | 525424 | `0d3d6d2648870e87a01636cdfc7cc7af8100ea40b71e5ed05f82ac197606584a` |
| `2d/font/jillbtn.par` | 101984 | `95b48379092db4db72f890d5a221ba8c4094dd438cb4c4eba98eb5520c7b17aa` |
| `SYSDIR/BOOT.BIN` | 3484281 | `5e294dc84a7f0d50719ecd26cb24ffb3792f2d9445803690845a8f1fa1cb85a3` |
| `SYSDIR/EBOOT.BIN` | 3484624 | `2a52012be00c07512dcde932ff6e9eb9b96912c59dd5a25c7c26ef821c124d68` |
| `data/bindata.dat` | 151552 | `3241fc000f3d52fe8522baaa985fd866e29d64d3a0f23ac4e28b66dee957de3e` |

The manifest also confirms `data/bindata.dat` is member index 0 of `pa`, at archive offset `0x10`.

## Slot-audit result

The audit combines references from:

- current English message text;
- retail Japanese message text;
- canonical fixed EBOOT/equipment strings;
- plausible CP932 NUL-terminated literals recovered from authenticated decrypted `BOOT.BIN`;
- plausible CP932 NUL-terminated literals recovered from authenticated `data/bindata.dat`.

There are 2487 installed two-byte renderer keys. After the above union, 212 remain as **audited candidates**.

Claude's follow-up audit review returned `PASS FOR POC CANDIDATE SELECTION`, with an explicit requirement that each key actually chosen for the PoC also receive an exact raw-byte occurrence check in BOOT and bindata to compensate for the binary scanner's known one-glyph blind spot.

## Selected five PoC keys

All five selected keys have zero exact raw-byte occurrences in both authenticated `BOOT.BIN` and `data/bindata.dat`, and are absent from the audited message/fixed reference union.

| Korean | Renderer key | Message bytes | Retail glyph | PAF index | Page | x | y | w | h | bearing | advance |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| 테 | `0xA1E1` | `E1 A1` | 癸 | 1455 | 1 | 405 | 123 | 11 | 12 | 0,-10 | 12 |
| 스 | `0xA1E9` | `E9 A1` | 鬘 | 1458 | 1 | 450 | 123 | 12 | 11 | 0,-10 | 12 |
| 트 | `0xB8E2` | `E2 B8` | 篋 | 1774 | 1 | 90 | 273 | 11 | 11 | 0,-10 | 12 |
| 성 | `0xBBE6` | `E6 BB` | 貊 | 1812 | 1 | 150 | 288 | 12 | 11 | 0,-10 | 12 |
| 공 | `0xBFE6` | `E6 BF` | 豼 | 1867 | 1 | 465 | 303 | 12 | 11 | 0,-10 | 12 |

These are PoC candidates only. No claim is made that they are production-safe across every UI/script/archive resource.

## Target message

Target: `message/msgsec001.dat`, record index 7, canonical ID 10007.

Contributor display projection:

```text
汝、無限のソウルを持つ者よ<line-break>我に応ぜよ<line-break>我が問いに答え、その魂を我に示せ<end>
```

For the reviewed PoC, only the first displayed line is replaced with:

```text
테스트 성공
```

The second and third Japanese lines are intentionally left byte-identical. This makes the empirical pass/fail criterion stricter: the custom renderer-key/message-remap path must work for the first line without disturbing the following native text or controls.

Custom renderer bytes for the replacement line:

```text
E1 A1 E9 A1 E2 B8 20 E6 BB E6 BF
```

The retail record may contain renderer kana-mode escapes (`ESC K/H/k`) that are hidden by the contributor display projection. The Android guard therefore parses the retail record and verifies its displayed Japanese text and native control structure rather than assuming the visible TOML string is a direct byte-for-byte CP932 encoding.

The two native `0x0A` line breaks and the native `05 05 05` `<end>` terminator are not overwritten. Only the first natural-text byte segment is rewritten; unused capacity in that first segment is filled with ASCII spaces so the member size and every downstream record offset remain unchanged. The second and third natural-text segments remain byte-identical.

## Glyph rasterization

The atlas edits use the same raster rule empirically proven by the earlier PPSSPP Hangul tests:

- UnDotum 10 px;
- text origin `(0,-2)`;
- grayscale quantized to 4-bit values 0..15;
- 10x10 raster;
- pasted at `(1,1)` inside each selected source cell.

The underlying PAF key, record index, BST links, page, record count and archive geometry remain unchanged.

## Android PoC patcher safety boundary

The v0.6 patcher:

- opens the original ISO read-only;
- validates DISC_ID/DISC_VERSION;
- authenticates BOOT, EBOOT, font members and bindata;
- locates and validates the exact startup message display before enabling the patch;
- re-runs inspection before output generation;
- streams the original ISO into a separate destination;
- edits only sorted declared atlas/message byte offsets;
- requires output size to equal source ISO size;
- deletes partial output on failure.

Claude's implementation review returned:

`PASS FOR FIRST-SENTENCE POC TEST`

The review also recorded one non-blocking SHOULD-FIX: the full `message/msgsec001.dat` member is not yet SHA-256 pinned. The guarded record parser still fails closed for this PoC target, so the review did not block empirical testing. A full-member retail fingerprint should be added before this pattern is reused as a production-wide translation path.

## Success criterion

Start a new game in PPSSPP. The opening invocation should render its first line as Korean while the following two lines remain Japanese:

```text
테스트 성공
我に応ぜよ
我が問いに答え、その魂を我に示せ
```

A successful observation proves, for the selected PoC path:

- custom Korean message bytes reach the existing renderer lookup;
- the chosen PAF keys resolve normally;
- the matching Korean atlas bitmaps are read from the expected page/cells;
- message remapping and font replacement work together without globally replacing common Japanese glyphs;
- later text/control bytes in the same guarded record survive the first-line rewrite unchanged.

It would still not prove production-wide slot safety, full-corpus capacity, final Korean metrics/layout, or whole-game translation correctness.
