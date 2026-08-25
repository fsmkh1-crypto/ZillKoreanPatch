package com.fsmkh1.zillfontdump;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;

final class PoCPatcher {
    // 60 atlas-byte edits only. These replace the CP932 glyph for の (key 0xCC82)
    // with a 10x10 Hangul '가' bitmap while preserving the PAF key, metrics, BST,
    // record count and every other byte in the font member.
    private static final int[] RELATIVE_OFFSETS = new int[]{
            229602,229603,229604,229605,229606,229607,
            229618,229619,229620,229621,229622,229623,
            229634,229635,229636,229637,229638,229639,
            229650,229651,229652,229653,229654,229655,
            229666,229667,229668,229669,229670,229671,
            231602,231603,231604,231605,231606,231607,
            231618,231619,231620,231621,231622,231623,
            231634,231635,231636,231637,231638,231639,
            231650,231651,231652,231653,231654,231655,
            231666,231667,231668,231669,231670,231671
    };

    private static final byte[] VALUES = new byte[]{
            1,0,0,0,24,16,
            1,0,0,0,25,16,
            1,(byte)255,(byte)255,4,25,16,
            1,0,(byte)128,3,25,16,
            1,0,(byte)160,0,(byte)249,27,
            1,0,(byte)148,0,25,16,
            1,48,28,0,25,16,
            1,(byte)183,1,0,25,16,
            17,6,0,0,25,16,
            1,0,0,0,4,16
    };

    static final String PATCHED_ZILLFONT_SHA256 =
            "aa9c158889ef5947e815e9e81ebabfa2c7cef7f6506fa1cabad890e8e1dbb4bf";

    private PoCPatcher() {}

    static long[] absoluteOffsets(FontExtractor.Inspection inspection) {
        long memberStart = inspection.paArc.extent * (long) Iso9660.SECTOR_SIZE + inspection.zillfont.offset;
        long[] out = new long[RELATIVE_OFFSETS.length];
        for (int i = 0; i < out.length; i++) out[i] = memberStart + RELATIVE_OFFSETS[i];
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
                    buffer[(int) (target - position)] = VALUES[patchIndex];
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
            throw new IOException("Not all Korean glyph PoC edits were applied");
        }
    }

    static int patchByteCount() {
        return RELATIVE_OFFSETS.length;
    }
}
