package com.fsmkh1.zillfontdump;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertNotEquals;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import org.junit.Test;

public class PoCPatcherTest {
    @Test
    public void patchesOnlyDeclaredFontAndMessageBytesAndPreservesSize() throws Exception {
        int isoSize = 900000;
        byte[] source = new byte[isoSize];
        for (int i = 0; i < source.length; i++) source[i] = (byte) (i * 31 + 7);

        Iso9660.Entry paBin = new Iso9660.Entry("pa.bin", 0, 0, false);
        Iso9660.Entry paArc = new Iso9660.Entry("pa.arc", 0, isoSize, false);
        PaaIndex.Member zillfont = new PaaIndex.Member(13611, "font/zillfont.par", 1000, 0x80470L);
        PaaIndex.Member jillbtn = new PaaIndex.Member(13612, "2d/font/jillbtn.par", 0, 0x18E60L);
        PaaIndex.Member startup = new PaaIndex.Member(1234, StartupMessage.MEMBER_NAME, 800000, 1000);
        StartupMessage.Record startupRecord = new StartupMessage.Record(
                128,
                80,
                new StartupMessage.Segment[]{
                        new StartupMessage.Segment(128, 20),
                        new StartupMessage.Segment(149, 20),
                        new StartupMessage.Segment(170, 20),
                });
        FontExtractor.Inspection inspection = new FontExtractor.Inspection(
                "ULJM05410", "1.03", isoSize,
                null, null, paBin, paArc, zillfont, jillbtn,
                null, null, "", paArc, startup, startupRecord);

        ByteArrayOutputStream out = new ByteArrayOutputStream();
        PoCPatcher.copyAndPatch(new ByteArrayInputStream(source), out, inspection);
        byte[] patched = out.toByteArray();

        assertEquals(5, PoCPatcher.glyphCount());
        assertEquals(source.length, patched.length);
        long[] offsets = PoCPatcher.absoluteOffsets(inspection);
        assertEquals(PoCPatcher.patchByteCount(inspection), offsets.length);

        boolean[] declared = new boolean[source.length];
        int actualChanges = 0;
        long previous = -1;
        for (long offset : offsets) {
            assertNotEquals("duplicate or unsorted patch offset", previous, offset);
            int p = (int) offset;
            declared[p] = true;
            if (source[p] != patched[p]) actualChanges++;
            previous = offset;
        }
        assertNotEquals(0, actualChanges);
        for (int i = 0; i < source.length; i++) {
            if (!declared[i]) assertEquals("unexpected byte edit at " + i, source[i], patched[i]);
        }

        // The guarded first natural-text segment begins with the custom
        // renderer bytes assigned to 테.
        int messageStart = 800000 + 128;
        assertEquals(0xE1, patched[messageStart] & 0xff);
        assertEquals(0xA1, patched[messageStart + 1] & 0xff);

        // The synthetic native line-break gap is deliberately untouched.
        assertEquals(source[800000 + 148], patched[800000 + 148]);
        assertEquals(source[800000 + 169], patched[800000 + 169]);
    }
}
