package com.fsmkh1.zillfontdump;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.nio.charset.StandardCharsets;

final class PaaIndex {
    static final int HEADER_SIZE = 0x20;
    static final int RECORD_SIZE = 0x10;

    static final class Member {
        final int index;
        final String name;
        final long offset;
        final long size;

        Member(int index, String name, long offset, long size) {
            this.index = index;
            this.name = name;
            this.offset = offset;
            this.size = size;
        }
    }

    private final ByteBuffer data;
    private final int count;
    private final long offsetTable;

    private PaaIndex(ByteBuffer data, int count, long offsetTable) {
        this.data = data;
        this.count = count;
        this.offsetTable = offsetTable;
    }

    static PaaIndex parse(ByteBuffer input) throws IOException {
        ByteBuffer b = input.duplicate().order(ByteOrder.LITTLE_ENDIAN);
        if (b.remaining() < HEADER_SIZE || b.get(0) != 'P' || b.get(1) != 'A' ||
                b.get(2) != 'A' || b.get(3) != 0) {
            throw new IOException("pa.bin has unexpected magic");
        }
        long countLong = u32(b, 8);
        long offsetTable = u32(b, 16);
        if (countLong > Integer.MAX_VALUE) throw new IOException("PAA member count too large");
        int count = (int) countLong;
        if ((long) HEADER_SIZE + (long) count * RECORD_SIZE > b.limit()) {
            throw new IOException("PAA member table extends past pa.bin");
        }
        if (offsetTable + (long) count * 4 > b.limit()) {
            throw new IOException("PAA offset table extends past pa.bin");
        }
        return new PaaIndex(b, count, offsetTable);
    }

    int count() {
        return count;
    }

    Member member(int index) throws IOException {
        if (index < 0 || index >= count) throw new IOException("PAA member index out of range: " + index);
        int record = HEADER_SIZE + index * RECORD_SIZE;
        long nameOffset = u32(data, record);
        long size = u32(data, record + 4);
        long tablePosition = offsetTable + (long) index * 4;
        if (nameOffset >= data.limit()) throw new IOException("PAA filename offset out of range at " + index);
        if (tablePosition < 0 || tablePosition + 4 > data.limit()) {
            throw new IOException("PAA archive offset out of range at " + index);
        }
        long archiveOffset = u32(data, (int) tablePosition);
        String name = asciiCString(data, (int) nameOffset);
        return new Member(index, name, archiveOffset, size);
    }

    private static String asciiCString(ByteBuffer b, int offset) throws IOException {
        int end = offset;
        while (end < b.limit() && b.get(end) != 0) {
            if ((b.get(end) & 0x80) != 0) throw new IOException("Non-ASCII PAA filename");
            end++;
        }
        if (end == b.limit()) throw new IOException("Unterminated PAA filename");
        byte[] bytes = new byte[end - offset];
        for (int i = 0; i < bytes.length; i++) bytes[i] = b.get(offset + i);
        return new String(bytes, StandardCharsets.US_ASCII);
    }

    private static long u32(ByteBuffer b, int off) {
        return ((long) b.get(off) & 0xff) |
                (((long) b.get(off + 1) & 0xff) << 8) |
                (((long) b.get(off + 2) & 0xff) << 16) |
                (((long) b.get(off + 3) & 0xff) << 24);
    }
}
