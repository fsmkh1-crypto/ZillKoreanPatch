package com.fsmkh1.zillfontdump;

import static org.junit.Assert.assertEquals;

import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.nio.charset.StandardCharsets;
import java.util.Map;
import org.junit.Test;

public class SfoParserTest {
    @Test
    public void parsesUtf8StringEntry() throws Exception {
        byte[] key = "DISC_ID\0".getBytes(StandardCharsets.UTF_8);
        byte[] value = "ULJM-05410\0".getBytes(StandardCharsets.UTF_8);
        int keyTable = 20 + 16;
        int dataTable = keyTable + key.length;
        ByteBuffer b = ByteBuffer.allocate(dataTable + value.length).order(ByteOrder.LITTLE_ENDIAN);
        b.put(0, (byte) 0x00);
        b.put(1, (byte) 'P');
        b.put(2, (byte) 'S');
        b.put(3, (byte) 'F');
        b.putInt(4, 0x00000101);
        b.putInt(8, keyTable);
        b.putInt(12, dataTable);
        b.putInt(16, 1);
        b.putShort(20, (short) 0);
        b.putShort(22, (short) 0x0204);
        b.putInt(24, value.length);
        b.putInt(28, value.length);
        b.putInt(32, 0);
        b.position(keyTable);
        b.put(key);
        b.position(dataTable);
        b.put(value);
        b.position(0);
        b.limit(dataTable + value.length);

        Map<String, String> fields = SfoParser.parseStrings(b);
        assertEquals("ULJM-05410", fields.get("DISC_ID"));
    }
}
