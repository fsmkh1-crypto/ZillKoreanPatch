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

    // Startup-screen smoke-test mapping. All eight source glyphs occur on the
    // mandatory new-game intro screen. We only replace atlas pixels; PAF keys,
    // metrics, BST links, record count and message bytes remain unchanged.
    //
    // The 10x10 Korean bitmaps use the exact raster rule that produced the
    // already PPSSPP-proven の -> 가 PoC: UnDotum 10px, text origin (0,-2),
    // grayscale quantized to 4bpp (0..15). Larger Japanese source cells are
    // cleared and the 10x10 Korean bitmap is positioned so its effective
    // bearing matches the proven 가 cell (bearing X=1, Y=-9).
    private static final Glyph[] GLYPHS = {
            new Glyph("の", "가", 1, 421, 379, 10, 10, 0, 0, "000000081000000009100ffff4091000008309100000a009fb0004900910003c10091007b100091016000009100000000400"),
            new Glyph("無", "나", 1, 15, 243, 12, 12, 1, 1, "000000081005000009100b000009100b000009100b000009fb0b000359100dacdb5910054200091000000009100000000400"),
            new Glyph("我", "다", 2, 435, 3, 12, 11, 1, 1, "000000081000000009101ffff60910190000091019000009fb19001559101dbcea3910054200091000000009100000000400"),
            new Glyph("応", "라", 1, 195, 108, 12, 11, 1, 1, "00000008100ffff30910000073091000007309100ffff309fb0a000009100a012439100fedc9391000000009100000000400"),
            new Glyph("答", "마", 1, 150, 93, 12, 11, 1, 1, "000000081000000009102ffff40910290074091029007409fb29007409102ffff40910000000091000000009100000000400"),
            new Glyph("魂", "바", 1, 375, 213, 12, 12, 1, 1, "0100110810290065091029006509102ffff5091029006509fb29006509102ffff50910000000091000000009100000000400"),
            new Glyph("示", "사", 1, 1, 169, 11, 10, 0, 0, "000000081000260009100048000910006a00091000ac0009fb03967009102c10b50910330016091000000009100000000400"),
            new Glyph("者", "아", 1, 136, 423, 10, 11, 0, 1, "000000081001ce6009100952c109100b006509100a004709fb0b005509100952c1091002ce50091000000009100000000400")
    };

    private static final Edit[] EDITS = buildEdits();

    private PoCPatcher() {}

    private static final class Glyph {
        final String japanese;
        final String korean;
        final int page;
        final int x;
        final int y;
        final int width;
        final int height;
        final int pasteX;
        final int pasteY;
        final String raster;

        Glyph(String japanese, String korean, int page, int x, int y,
              int width, int height, int pasteX, int pasteY, String raster) {
            this.japanese = japanese;
            this.korean = korean;
            this.page = page;
            this.x = x;
            this.y = y;
            this.width = width;
            this.height = height;
            this.pasteX = pasteX;
            this.pasteY = pasteY;
            this.raster = raster;
            if (raster.length() != 100) {
                throw new IllegalStateException("10x10 raster length mismatch for " + korean);
            }
            if (pasteX < 0 || pasteY < 0 || pasteX + 10 > width || pasteY + 10 > height) {
                throw new IllegalStateException("Korean raster does not fit source cell for " + japanese);
            }
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
        // byte offset -> {mask,value}. TreeMap also guarantees streaming order.
        TreeMap<Integer, int[]> merged = new TreeMap<>();
        for (Glyph glyph : GLYPHS) {
            if (glyph.page < 0 || glyph.page >= GIM_STARTS.length) {
                throw new IllegalStateException("invalid GIM page for " + glyph.korean);
            }
            for (int row = 0; row < glyph.height; row++) {
                for (int column = 0; column < glyph.width; column++) {
                    int alpha = 0;
                    if (column >= glyph.pasteX && column < glyph.pasteX + 10
                            && row >= glyph.pasteY && row < glyph.pasteY + 10) {
                        int rasterIndex = (row - glyph.pasteY) * 10 + (column - glyph.pasteX);
                        alpha = Character.digit(glyph.raster.charAt(rasterIndex), 16);
                        if (alpha < 0) throw new IllegalStateException("invalid glyph pixel");
                    }
                    int pixelX = glyph.x + column;
                    int pixelY = glyph.y + row;
                    int relative = GIM_STARTS[glyph.page] + IMAGE_DATA_OFFSET
                            + swizzledByteOffset(pixelX, pixelY);
                    boolean high = (pixelX & 1) != 0;
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
        long memberStart = inspection.paArc.extent * (long) Iso9660.SECTOR_SIZE
                + inspection.zillfont.offset;
        long[] out = new long[EDITS.length];
        for (int i = 0; i < out.length; i++) {
            out[i] = memberStart + EDITS[i].relativeOffset;
        }
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
            throw new IOException("ISO copy size mismatch: wrote " + position
                    + " bytes, expected " + inspection.isoSize);
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
