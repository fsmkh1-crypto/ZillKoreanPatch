package com.fsmkh1.zillfontdump;

import static org.junit.Assert.assertArrayEquals;
import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertTrue;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.nio.charset.Charset;
import java.util.Arrays;
import org.junit.Test;

public class StartupMessageTest {
    @Test
    public void rewritesOnlyFirstNaturalTextLineAndPreservesControlsAndLaterLines() throws Exception {
        byte[] bank = bankWithRecord(directRecord());
        StartupMessage.Record record = StartupMessage.inspect(ByteBuffer.wrap(bank));
        assertEquals(0x20, record.offset);
        assertEquals(3, record.segments.length);
        assertTrue(record.segments[0].length >= StartupMessage.replacementLineLength());

        byte[] original = bank.clone();
        StartupMessage.ByteEdit[] edits = StartupMessage.patchEdits(record);
        assertEquals(record.segments[0].length, edits.length);
        for (StartupMessage.ByteEdit edit : edits) {
            bank[record.offset + edit.relativeOffset] = (byte) edit.value;
        }

        StartupMessage.Segment first = record.segments[0];
        assertEquals(0xE1, bank[first.offset] & 0xff);
        assertEquals(0xA1, bank[first.offset + 1] & 0xff);
        for (int i = StartupMessage.replacementLineLength(); i < first.length; i++) {
            assertEquals(0x20, bank[first.offset + i] & 0xff);
        }

        int firstBreak = first.offset + first.length;
        assertEquals(10, original[firstBreak] & 0xff);
        assertEquals(original[firstBreak], bank[firstBreak]);

        StartupMessage.Segment second = record.segments[1];
        StartupMessage.Segment third = record.segments[2];
        assertArrayEquals(
                Arrays.copyOfRange(original, second.offset, second.offset + second.length),
                Arrays.copyOfRange(bank, second.offset, second.offset + second.length));
        assertArrayEquals(
                Arrays.copyOfRange(original, third.offset, third.offset + third.length),
                Arrays.copyOfRange(bank, third.offset, third.offset + third.length));

        int secondBreak = second.offset + second.length;
        assertEquals(10, original[secondBreak] & 0xff);
        assertEquals(original[secondBreak], bank[secondBreak]);
        int endControl = third.offset + third.length;
        assertEquals(5, original[endControl] & 0xff);
        assertEquals(5, original[endControl + 1] & 0xff);
        assertEquals(5, original[endControl + 2] & 0xff);
        assertArrayEquals(
                Arrays.copyOfRange(original, endControl, endControl + 3),
                Arrays.copyOfRange(bank, endControl, endControl + 3));
    }

    @Test
    public void acceptsRendererKanaModeSpellingOfSoul() throws Exception {
        StartupMessage.Record record = StartupMessage.inspect(ByteBuffer.wrap(bankWithRecord(kanaModeRecord())));
        assertEquals(3, record.segments.length);
        assertTrue(record.segments[0].length >= StartupMessage.replacementLineLength());
    }

    @Test(expected = IOException.class)
    public void rejectsModifiedDisplayedSourceRecord() throws Exception {
        byte[] source = directRecord();
        source[4] ^= 1;
        StartupMessage.inspect(ByteBuffer.wrap(bankWithRecord(source)));
    }

    @Test(expected = IOException.class)
    public void rejectsMalformedOffsetTable() throws Exception {
        byte[] bank = bankWithRecord(directRecord());
        ByteBuffer.wrap(bank).order(ByteOrder.LITTLE_ENDIAN).putShort(2 + 7 * 2, (short) 1);
        StartupMessage.inspect(ByteBuffer.wrap(bank));
    }

    private static byte[] directRecord() throws Exception {
        ByteArrayOutputStream out = new ByteArrayOutputStream();
        out.write(sjis("汝、無限のソウルを持つ者よ"));
        out.write(10);
        out.write(sjis("我に応ぜよ"));
        out.write(10);
        out.write(sjis("我が問いに答え、その魂を我に示せ"));
        out.write(new byte[]{5, 5, 5});
        out.write(0);
        return out.toByteArray();
    }

    private static byte[] kanaModeRecord() throws Exception {
        ByteArrayOutputStream out = new ByteArrayOutputStream();
        out.write(sjis("汝、無限の"));
        out.write(0x1b);
        out.write('K');
        out.write(sjis("ｿｳﾙ"));
        out.write(0x1b);
        out.write('H');
        out.write(sjis("を持つ者よ"));
        out.write(10);
        out.write(sjis("我に応ぜよ"));
        out.write(10);
        out.write(sjis("我が問いに答え、その魂を我に示せ"));
        out.write(new byte[]{5, 5, 5});
        out.write(0);
        return out.toByteArray();
    }

    private static byte[] sjis(String value) {
        return value.getBytes(Charset.forName("Shift_JIS"));
    }

    private static byte[] bankWithRecord(byte[] record) {
        final int count = 9;
        final int tableEnd = 2 + count * 2;
        final int recordStart = 0x20;
        final int next = recordStart + record.length + 5;
        byte[] bank = new byte[next + 4];
        ByteBuffer b = ByteBuffer.wrap(bank).order(ByteOrder.LITTLE_ENDIAN);
        b.putShort(0, (short) count);
        for (int i = 0; i < count; i++) {
            int offset;
            if (i < StartupMessage.RECORD_INDEX) offset = tableEnd;
            else if (i == StartupMessage.RECORD_INDEX) offset = recordStart;
            else offset = next;
            b.putShort(2 + i * 2, (short) offset);
        }
        System.arraycopy(record, 0, bank, recordStart, record.length);
        return bank;
    }
}
