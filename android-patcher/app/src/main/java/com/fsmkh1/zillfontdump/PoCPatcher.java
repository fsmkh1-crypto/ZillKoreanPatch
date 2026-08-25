package com.fsmkh1.zillfontdump;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.TreeMap;

final class PoCPatcher {
    private static final int IMAGE_DATA_OFFSET = 0x80;
    private static final int ATLAS_ROW_BYTES = 512 / 2; // 512px, 4bpp
    private static final int[] GIM_STARTS = {0xC0, 0x201B0, 0x402A0, 0x60390};

    // Startup-screen smoke-test mapping. These are existing CP932 glyph slots;
    // their PAF keys, metrics, BST links, record count and message bytes stay intact.
    //
    // の -> 가  page 1 x=421 y=379 w=10 h=10  (already proven in PPSSPP)
    // 無 -> 나  page 1 x=15  y=243 w=12 h=12
    // 我 -> 다  page 2 x=435 y=3   w=12 h=11
    // 応 -> 라  page 1 x=195 y=108 w=12 h=11
    // 答 -> 마  page 1 x=150 y=93  w=12 h=11
    // 魂 -> 바  page 1 x=375 y=213 w=12 h=12
    // 示 -> 사  page 1 x=1   y=169 w=11 h=10
    // 者 -> 아  page 1 x=136 y=423 w=10 h=11
    private static final Glyph[] GLYPHS = {
            new Glyph("가", 1, 421, 379, new String[]{
                    "0000000810",
                    "0000000910",
                    "0FFFF40910",
                    "0000830910",
                    "0000A009FB",
                    "0004900910",
                    "003C100910",
                    "07B1000910",
                    "1600000910",
                    "11111FF811"
            }),
            new Glyph("나", 1, 15, 243, new String[]{
                    "000000000000",
                    "0F00000000F0",
                    "0F00000000F0",
                    "0F00000000F0",
                    "0F00000000F0",
                    "0F00000000F0",
                    "0F000000FFF0",
                    "0F00000000F0",
                    "0F00000000F0",
                    "0F00000000F0",
                    "0FFFFFFF00F0",
                    "000000000000"
            }),
            new Glyph("다", 2, 435, 3, new String[]{
                    "000000000000",
                    "0FFFFFFF00F0",
                    "0F00000000F0",
                    "0F00000000F0",
                    "0F00000000F0",
                    "0F000000FFF0",
                    "0F00000000F0",
                    "0F00000000F0",
                    "0F00000000F0",
                    "0FFFFFFF00F0",
                    "000000000000"
            }),
            new Glyph("라", 1, 195, 108, new String[]{
                    "000000000000",
                    "0FFFFFFF00F0",
                    "0000000F00F0",
                    "0000000F00F0",
                    "0000000F00F0",
                    "0FFFFFFFFFF0",
                    "0F00000000F0",
                    "0F00000000F0",
                    "0F00000000F0",
                    "0FFFFFFF00F0",
                    "000000000000"
            }),
            new Glyph("마", 1, 150, 93, new String[]{
                    "000000000000",
                    "0FFFFFFF00F0",
                    "0F00000F00F0",
                    "0F00000F00F0",
                    "0F00000F00F0",
                    "0F00000FFFF0",
                    "0F00000F00F0",
                    "0F00000F00F0",
                    "0F00000F00F0",
                    "0FFFFFFF00F0",
                    "000000000000"
            }),
            new Glyph("바", 1, 375, 213, new String[]{
                    "000000000000",
                    "0FFFFFFF00F0",
                    "0F00000F00F0",
                    "0F00000F00F0",
                    "0F00000F00F0",
                    "0FFFFFFF00F0",
                    "0F00000FFFF0",
                    "0F00000F00F0",
                    "0F00000F00F0",
                    "0F00000F00F0",
                    "0FFFFFFF00F0",
                    "000000000000"
            }),
            new Glyph("사", 1, 1, 169, new String[]{
                    "00000000000",
                    "000F00000F0",
                    "000F00000F0",
                    "00F0F0000F0",
                    "00F0F0000F0",
                    "00F0F00FFF0",
                    "00F0F0000F0",
                    "0F000F000F0",
                    "0F000F000F0",
                    "00000000000"
            }),
            new Glyph("아", 1, 136, 423, new String[]{
                    "0000000000",
                    "00FFF000F0",
                    "0FF0FF00F0",
                    "0F000F00F0",
                    "0F000F00F0",
                    "0F000FFFF0",
                    "0F000F00F0",
                    "0F000F00F0",
                    "0FF0FF00F0",
                    "00FFF000F0",
                    "0000000000"
            })
    };

    private static final Edit[] EDITS = buildEdits();

    private PoCPatcher() {}

    private static final class Glyph {
        final String korean;
        final int page;
        final int x;
        final int y;
        final String[] rows;

        Glyph(String korean, int page, int x, int y, String[] rows) {
            this.korean = korean;
            this.page = page;
            this.x = x;
            this.y = y;
            this.rows = rows;
        }
    }

    private static final class Edit {
        final int relativeOffset;
        final int mask;
        final int value;

        Edit(int relativeOffset, int mask, int value) {
            this.relativeOffset = relativeOffset;
            this.mask = mask;
            this.value = value;
        }
    }

    private static Edit[] buildEdits() {
        // byte offset -> {mask,value}; TreeMap also guarantees streaming order.
        TreeMap<Integer, int[]> merged = new TreeMap<>();
        for (Glyph glyph : GLYPHS) {
            if (glyph.page < 0 || glyph.page >= GIM_STARTS.length) {
                throw new IllegalStateException("invalid GIM page for " + glyph.korean);
            }
            for (int row = 0; row < glyph.rows.length; row++) {
                String pixels = glyph.rows[row];
                for (int column = 0; column < pixels.length(); column++) {
                    int alpha = Character.digit(pixels.charAt(column), 16);
                    if (alpha < 0) throw new IllegalStateException("invalid glyph pixel");
                    int x = glyph.x + column;
                    int y = glyph.y + row;
                    int relative = GIM_STARTS[glyph.page] + IMAGE_DATA_OFFSET + swizzledByteOffset(x, y);
                    boolean high = (x & 1) != 0;
                    int mask = high ? 0xF0 : 0x0F;
                    int value = high ? alpha << 4 : alpha;
                    int[] current = merged.get(relative);
                    if (current == null) {
                        merged.put(relative, new int[]{mask, value});
                    } else {
                        current[1] = (current[1] & ~mask) | value;
                        current[0] |= mask;
                    }
                }
            }
        }
        List<Edit> result = new ArrayList<>(merged.size());
        for (Map.Entry<Integer, int[]> entry : merged.entrySet()) {
            result.add(new Edit(entry.getKey(), entry.getValue()[0], entry.getValue()[1]));
        }
        return result.toArray(new Edit[0]);
    }

    private static int swizzledByteOffset(int pixelX, int pixelY) {
        int byteX = pixelX >>> 1;
        int blocksPerRow = ATLAS_ROW_BYTES / 16;
        return (pixelY / 8) * blocksPerRow * 128
                + (byteX / 16) * 128
                + (pixelY & 7) * 16
                + (byteX & 15);
    }

    static long[] absoluteOffsets(FontExtractor.Inspection inspection) {
        long memberStart = inspection.paArc.extent * (long) Iso9660.SECTOR_SIZE + inspection.zillfont.offset;
        long[] out = new long[EDITS.length];
        for (int i = 0; i < out.length; i++) out[i] = memberStart + EDITS[i].relativeOffset;
        return out;
    }

    static void copyAndPatch(InputStream input, OutputStream output,
                             FontExtractor.Inspection inspection) throws IOException {
        long[] targets = absoluteOffsets(inspection);
        byte[] buffer = new byte[1024 * 1024];
        long position = 0;
        int patchIndex = 0;
        int read;
        while ((read = input.read(buffer)) >= 0) {
            if (read == 0) continue;
            long end = position + read;
            while (patchIndex < targets.length && targets[patchIndex] < end) {
                long target = targets[patchIndex];
                if (target >= position) {
                    int index = (int) (target - position);
                    int oldValue = buffer[index] & 0xFF;
                    Edit edit = EDITS[patchIndex];
                    buffer[index] = (byte) ((oldValue & ~edit.mask) | (edit.value & edit.mask));
                }
                patchIndex++;
            }
            output.write(buffer, 0, read);
            position = end;
        }
        if (position != inspection.isoSize) {
            throw new IOException("ISO copy size mismatch: wrote " + position + " bytes, expected " + inspection.isoSize);
        }
        if (patchIndex != targets.length) {
            throw new IOException("Not all Korean startup-screen glyph edits were applied");
        }
    }

    static int patchByteCount() {
        return EDITS.length;
    }

    static int glyphCount() {
        return GLYPHS.length;
    }
}
