package com.fsmkh1.zillfontdump;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertNotEquals;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import org.junit.Test;

public class PoCPatcherTest {
    @Test
    public void patchesOnlyDeclaredAtlasBytesAndPreservesSize() throws Exception {
        int isoSize = 300000;
        byte[] source = new byte[isoSize];
        for (int i = 0; i < source.length; i++) source[i] = (byte) (i * 31 + 7);

        Iso9660.Entry paBin = new Iso9660.Entry("pa.bin", 0, 0, false);
        Iso9660.Entry paArc = new Iso9660.Entry("pa.arc", 0, isoSize, false);
        PaaIndex.Member zillfont = new PaaIndex.Member(13611, "font/zillfont.par", 1000, 0x80470L);
        PaaIndex.Member jillbtn = new PaaIndex.Member(13612, "2d/font/jillbtn.par", 0, 0x18E60L);
        FontExtractor.Inspection inspection = new FontExtractor.Inspection(
                "ULJM05410", "1.03", isoSize, paBin, paArc, zillfont, jillbtn);

        ByteArrayOutputStream out = new ByteArrayOutputStream();
        PoCPatcher.copyAndPatch(new ByteArrayInputStream(source), out, inspection);
        byte[] patched = out.toByteArray();

        assertEquals(source.length, patched.length);
        long[] offsets = PoCPatcher.absoluteOffsets(inspection);
        assertEquals(PoCPatcher.patchByteCount(), offsets.length);

        boolean[] changed = new boolean[source.length];
        int actualChanges = 0;
        for (long offset : offsets) {
            int p = (int) offset;
            changed[p] = true;
            if (source[p] != patched[p]) actualChanges++;
        }
        assertNotEquals(0, actualChanges);
        for (int i = 0; i < source.length; i++) {
            if (!changed[i]) assertEquals("unexpected byte edit at " + i, source[i], patched[i]);
        }
    }
}
