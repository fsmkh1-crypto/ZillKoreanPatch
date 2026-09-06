# Xdelta / VCDIFF Forensics

## Scope

This note analyzes the recovered final xdelta structurally without reconstructing or redistributing the game ISO.

Patch:

`OreShika_UCJS10117_R29_Korean_M14_DLC13_FINAL_20260801.xdelta`

SHA-256:

`08469300D12E37D39A0966C3BBE7CC751F0FB3F0B1989043CE5E2ECDD18C4BDD`

## Container format

The patch begins with the VCDIFF magic bytes:

`D6 C3 C4 00`

It is therefore a valid VCDIFF/xdelta stream.

The stream uses a secondary compressor. The secondary-compressor ID is `1`, corresponding to xdelta's DJW/static-Huffman secondary compression.

The xdelta application header contains build-environment metadata. Any local absolute developer paths/usernames are intentionally omitted from this repository. One non-sensitive basename visible in that metadata is `cvn-osky.iso`, consistent with an OSK-oriented intermediate build, but this filename alone is not treated as proof of implementation details.

## Decoded target geometry

The VCDIFF stream contains `102` target windows.

The sum of decoded target-window sizes is:

`847,937,536` bytes (`0x328A8000`)

This exactly matches the output ISO size declared by the final package manifest.

Most target windows are 8 MiB.

## Korean name UI region

Package QA identifies the Korean name UI slot as:

- ISO offset: `0x25261800`
- size: `18,432` bytes

This location falls in VCDIFF target window **74**, whose decoded target range is:

`0x25000000..0x257FFFFF`

Relative offset of the name UI slot inside that target window:

`0x261800`

Window 74 characteristics observed from the VCDIFF structure:

- target size: 8 MiB
- compressed delta contribution: approximately 201,747 bytes
- data section length: 84,138
- instruction section length: 41,337
- address section length: 76,241
- all three sections use secondary compression

This confirms that the final patch materially changes data in the same broad 8 MiB target window as the documented name UI slot. It does **not** by itself prove which individual delta instructions correspond to the name UI bytes.

## Other patch-dense target windows

Some of the largest compressed-delta windows are located around:

- target `0x00000000`
- target `0x24800000`
- target `0x25000000`
- target `0x25800000`
- target `0x28000000`
- target `0x29800000`
- target `0x2F000000`

The concentration around `0x24800000..0x26000000` is compatible with substantial UI/resource modifications near the documented name UI region.

The first target window is also one of the most heavily modified, which is compatible with executable/metadata changes. However, without reconstructing the ISO filesystem, no direct mapping from VCDIFF target offset to EBOOT file offset is asserted here.

## Why the xdelta alone cannot reveal the composition code

The final patch is source-dependent and its VCDIFF sections are secondary-compressed. The distribution does not contain the final EBOOT as a standalone file.

Therefore the xdelta alone is insufficient to recover the exact bytes at EBOOT file offset `0x001288E0..0x0012894F` without the matching clean source ISO.

The decisive path is:

1. obtain the user's legally dumped clean ISO matching package SHA-256;
2. apply this exact xdelta;
3. verify output ISO SHA-256;
4. extract the final EBOOT and verify its package SHA-256;
5. disassemble/diff the documented OSK range and its referenced tables.

Until then, claims about Unicode arithmetic, jamo state machines, or lookup-table design remain hypotheses rather than verified reverse engineering.
