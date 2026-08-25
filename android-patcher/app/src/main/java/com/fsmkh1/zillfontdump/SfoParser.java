package com.fsmkh1.zillfontdump;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.nio.charset.StandardCharsets;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.Map;

final class SfoParser {
    private SfoParser() {}

    static Map<String, String> parseStrings(ByteBuffer input) throws IOException {
        ByteBuffer b = input.duplicate().order(ByteOrder.LITTLE_ENDIAN);
        if (b.remaining() < 20 ||
                (b.get(0) & 0xff) != 0x00 || b.get(1) != 'P' || b.get(2) != 'S' || b.get(3) != 'F') {
            throw new IOException("Invalid PARAM.SFO magic");
        }
        long keyTable = u32(b, 8);
        long dataTable = u32(b, 12);
        long count = u32(b, 16);
        if (count > 4096) throw new IOException("Unreasonable PARAM.SFO entry count");

        Map<String, String> out = new LinkedHashMap<>();
        for (int i = 0; i < (int) count; i++) {
            int off = 20 + i * 16;
            if (off + 16 > b.limit()) throw new IOException("Truncated PARAM.SFO index");
            int keyOff = u16(b, off);
            int fmt = u16(b, off + 2);
            long dataLen = u32(b, off + 4);
            long dataOff = u32(b, off + 12);
            int keyPos = checkedIndex(keyTable + keyOff, b.limit());
            String key = cString(b, keyPos, b.limit() - keyPos);
            if (fmt == 0x0204 || fmt == 0x0004 || fmt == 0x0404) {
                int valuePos = checkedIndex(dataTable + dataOff, b.limit());
                int len = (int) Math.min(dataLen, b.limit() - valuePos);
                out.put(key, cString(b, valuePos, len));
            }
        }
        return Collections.unmodifiableMap(out);
    }

    private static int checkedIndex(long value, int limit) throws IOException {
        if (value < 0 || value >= limit) throw new IOException("PARAM.SFO offset out of range");
        return (int) value;
    }

    private static String cString(ByteBuffer b, int off, int maxLen) {
        int len = 0;
        while (len < maxLen && b.get(off + len) != 0) len++;
        byte[] data = new byte[len];
        for (int i = 0; i < len; i++) data[i] = b.get(off + i);
        return new String(data, StandardCharsets.UTF_8).trim();
    }

    private static int u16(ByteBuffer b, int off) {
        return (b.get(off) & 0xff) | ((b.get(off + 1) & 0xff) << 8);
    }

    private static long u32(ByteBuffer b, int off) {
        return ((long) b.get(off) & 0xff) |
                (((long) b.get(off + 1) & 0xff) << 8) |
                (((long) b.get(off + 2) & 0xff) << 16) |
                (((long) b.get(off + 3) & 0xff) << 24);
    }
}
