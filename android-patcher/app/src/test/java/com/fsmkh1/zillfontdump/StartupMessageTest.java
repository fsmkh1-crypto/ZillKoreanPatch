package com.fsmkh1.zillfontdump;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertTrue;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import org.junit.Test;

public class StartupMessageTest {
    private static final byte[] SOURCE = hex(
            "93f0814196b38cc082cc835c8345838b82f08e9d82c28ed282e60a" +
            "89e482c9899e82ba82e60a" +
            "89e482aa96e282a282c9939a82a6814182bb82cc8db082f089e482c98ea682b900");

    @Test
    public void validatesExactRecordAndBuildsShorterGuardedPatch() throws Exception {
        byte[] bank = bankWithRecord(SOURCE);
        StartupMessage.Record record = StartupMessage.inspect(ByteBuffer.wrap(bank));
        assertEquals(0x20, record.offset);
        assertTrue(record.span >= SOURCE.length);
        assertEquals(SOURCE.length, StartupMessage.guardedLength());
        assertTrue(StartupMessage.replacementLength() < StartupMessage.guardedLength());

        byte[] patch = StartupMessage.guardedPatchBytes();
        assertEquals(SOURCE.length, patch.length);
        // Replacement begins with the selected renderer bytes for 테.
        assertEquals(0xE1, patch[0] & 0xff);
        assertEquals(0xA1, patch[1] & 0xff);
        // Everything after the replacement's terminating NUL is zero-filled
        // only within the guarded source prefix.
        for (int i = StartupMessage.replacementLength(); i < patch.length; i++) {
            assertEquals(0, patch[i]);
        }
    }

    @Test(expected = IOException.class)
    public void rejectsModifiedSourceRecord() throws Exception {
        byte[] source = SOURCE.clone();
        source[4] ^= 1;
        StartupMessage.inspect(ByteBuffer.wrap(bankWithRecord(source)));
    }

    @Test(expected = IOException.class)
    public void rejectsMalformedOffsetTable() throws Exception {
        byte[] bank = bankWithRecord(SOURCE);
        ByteBuffer.wrap(bank).order(ByteOrder.LITTLE_ENDIAN).putShort(2 + 7 * 2, (short) 1);
        StartupMessage.inspect(ByteBuffer.wrap(bank));
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

    private static byte[] hex(String value) {
        byte[] out = new byte[value.length() / 2];
        for (int i = 0; i < out.length; i++) {
            out[i] = (byte) Integer.parseInt(value.substring(i * 2, i * 2 + 2), 16);
        }
        return out;
    }
}
