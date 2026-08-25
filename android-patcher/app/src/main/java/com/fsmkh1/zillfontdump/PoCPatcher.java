package com.fsmkh1.zillfontdump;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import java.util.Map;
import java.util.TreeMap;

final class PoCPatcher {
    private static final int IMAGE_DATA_OFFSET = 0x80;
    private static final int ATLAS_ROW_BYTES = 512 / 2; // 512px, 4bpp
    private static final int[] GIM_STARTS = {0xC0, 0x201B0, 0x402A0, 0x60390};

    // First end-to-end Korean sentence PoC.
    //
    // These five renderer keys survived the message/fixed/BOOT/bindata audit,
    // have zero exact raw-byte occurrences in authenticated BOOT.BIN and
    // bindata.dat, and use suitable existing cells with advance=12 and
    // bearing=(0,-10). They remain PoC candidates, not production-safe slots.
    //
    // The 10x10 Korean bitmaps use the exact raster rule already proven in
    // PPSSPP by the earlier の -> 가 and eight-glyph smoke tests: UnDotum 10px,
    // text origin (0,-2), grayscale rounded to 4bpp (0..15). Each bitmap is
    // pasted at (1,1), yielding the proven effective Korean bearing (1,-9).
    private static final Glyph[] GLYPHS = {
            new Glyph("癸", "테", 0xA1E1, 1, 405, 123, 11, 12, 1, 1,
                    "00000021407fff519380730001938073000193807fff7f9380730001938073236293807fda62938000000193800000002130"),
            new Glyph("鬘", "스", 0xA1E9, 1, 450, 123, 12, 11, 1, 1,
                    "000000000000005500000000aa0000000776800004b7007b401830000281000000000000000000004ffffffff40000000000"),
            new Glyph("篋", "트", 0xB8E2, 1, 90, 273, 11, 11, 1, 1,
                    "000000000002ffffff40028000000002ffffff10028000000002ffffff40000000000000000000004ffffffff40000000000"),
            new Glyph("貊", "성", 0xBBE6, 1, 150, 288, 12, 11, 1, 1,
                    "0001000010001b000460003c00046000a947ff601a50a604602300040460004cffee7000c4001870005cefea100000000000"),
            new Glyph("豼", "공", 0xBFE6, 1, 465, 303, 12, 11, 1, 1,
                    "000000000006ffffff70000000065000007009300000a007004ffffffff4008dffd9100592002a5001aeffd8000000000000")
    };

    private static final FontEdit[] FONT_EDITS = buildFontEdits();

    private PoCPatcher() {}

    private static final class Glyph {
        final String japanese;
        final String korean;
        final int rendererKey;
        final int page;
        final int x;
        final int y;
        final int width;
        final int height;
        final int pasteX;
        final int pasteY;
        final String raster;

        Glyph(String japanese, String korean, int rendererKey, int page, int x, int y,
              int width, int height, int pasteX, int pasteY, String raster) {
            this.japanese = japanese;
            this.korean = korean;
            this.rendererKey = rendererKey;
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
                throw new IllegalStateException("Korean raster does not fit source cell for key "
                        + Integer.toHexString(rendererKey));
            }
        }
    }

    private static final class FontEdit {
        final int relativeOffset;
        final int mask;
        final int value;

        FontEdit(int relativeOffset, int mask, int value) {
            this.relativeOffset = relativeOffset;
            this.mask = mask;
            this.value = value;
        }
    }

    private static final class AbsoluteEdit {
        final long offset;
        final int mask;
        final int value;

        AbsoluteEdit(long offset, int mask, int value) {
            this.offset = offset;
            this.mask = mask;
            this.value = value;
        }
    }

    private static FontEdit[] buildFontEdits() {
        // byte offset -> {mask,value}. TreeMap guarantees deterministic ordering
        // and merges two 4bpp pixels that share one underlying byte.
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
        List<FontEdit> result = new ArrayList<>(merged.size());
        for (Map.Entry<Integer, int[]> entry : merged.entrySet()) {
            result.add(new FontEdit(entry.getKey(), entry.getValue()[0], entry.getValue()[1]));
        }
        return result.toArray(new FontEdit[0]);
    }

    private static int swizzledByteOffset(int pixelX, int pixelY) {
        int byteX = pixelX >>> 1;
        int blocksPerRow = ATLAS_ROW_BYTES / 16;
        return (pixelY / 8) * blocksPerRow * 128
                + (byteX / 16) * 128
                + (pixelY & 7) * 16
                + (byteX & 15);
    }

    private static AbsoluteEdit[] absoluteEdits(FontExtractor.Inspection inspection) {
        if (inspection.startupMessageArc == null || inspection.startupMessage == null
                || inspection.startupRecord == null) {
            throw new IllegalStateException("inspection is missing guarded startup message metadata");
        }
        StartupMessage.ByteEdit[] messageEdits = StartupMessage.patchEdits(inspection.startupRecord);
        List<AbsoluteEdit> edits = new ArrayList<>(FONT_EDITS.length + messageEdits.length);

        long fontMemberStart = inspection.paArc.extent * (long) Iso9660.SECTOR_SIZE
                + inspection.zillfont.offset;
        for (FontEdit edit : FONT_EDITS) {
            edits.add(new AbsoluteEdit(fontMemberStart + edit.relativeOffset, edit.mask, edit.value));
        }

        long recordStart = inspection.startupMessageArc.extent * (long) Iso9660.SECTOR_SIZE
                + inspection.startupMessage.offset + inspection.startupRecord.offset;
        for (StartupMessage.ByteEdit edit : messageEdits) {
            edits.add(new AbsoluteEdit(recordStart + edit.relativeOffset, 0xFF, edit.value));
        }

        edits.sort(Comparator.comparingLong(edit -> edit.offset));
        for (int i = 1; i < edits.size(); i++) {
            if (edits.get(i - 1).offset == edits.get(i).offset) {
                throw new IllegalStateException("overlapping PoC edits at ISO offset " + edits.get(i).offset);
            }
        }
        return edits.toArray(new AbsoluteEdit[0]);
    }

    static long[] absoluteOffsets(FontExtractor.Inspection inspection) {
        AbsoluteEdit[] edits = absoluteEdits(inspection);
        long[] out = new long[edits.length];
        for (int i = 0; i < edits.length; i++) out[i] = edits[i].offset;
        return out;
    }

    static void copyAndPatch(InputStream input, OutputStream output,
                             FontExtractor.Inspection inspection) throws IOException {
        AbsoluteEdit[] edits = absoluteEdits(inspection);
        byte[] buffer = new byte[1024 * 1024];
        long position = 0;
        int patchIndex = 0;
        int read;
        while ((read = input.read(buffer)) >= 0) {
            if (read == 0) continue;
            long end = position + read;
            while (patchIndex < edits.length && edits[patchIndex].offset < end) {
                AbsoluteEdit edit = edits[patchIndex];
                if (edit.offset >= position) {
                    int index = (int) (edit.offset - position);
                    int oldValue = buffer[index] & 0xFF;
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
        if (patchIndex != edits.length) {
            throw new IOException("Not all Korean sentence PoC edits were applied");
        }
    }

    static int patchByteCount(FontExtractor.Inspection inspection) {
        return absoluteEdits(inspection).length;
    }

    static int glyphCount() {
        return GLYPHS.length;
    }
}
