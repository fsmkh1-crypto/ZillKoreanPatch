package com.fsmkh1.zillfontdump;

import static org.junit.Assert.assertEquals;

import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.nio.charset.StandardCharsets;
import org.junit.Test;

public class PaaIndexTest {
    @Test
    public void parsesMemberNameSizeAndArchiveOffset() throws Exception {
        byte[] name = "2d/font/jillbtn.par\0".getBytes(StandardCharsets.US_ASCII);
        int count = 1;
        int offsetTable = PaaIndex.HEADER_SIZE + PaaIndex.RECORD_SIZE;
        int nameOffset = offsetTable + 4;
        ByteBuffer b = ByteBuffer.allocate(nameOffset + name.length).order(ByteOrder.LITTLE_ENDIAN);
        b.put(0, (byte) 'P');
        b.put(1, (byte) 'A');
        b.put(2, (byte) 'A');
        b.put(3, (byte) 0);
        b.putInt(8, count);
        b.putInt(16, offsetTable);
        b.putInt(PaaIndex.HEADER_SIZE, nameOffset);
        b.putInt(PaaIndex.HEADER_SIZE + 4, 0x18E60);
        b.putInt(offsetTable, 0x3E0F980);
        b.position(nameOffset);
        b.put(name);
        b.position(0);

        PaaIndex index = PaaIndex.parse(b);
        assertEquals(1, index.count());
        PaaIndex.Member member = index.member(0);
        assertEquals("2d/font/jillbtn.par", member.name);
        assertEquals(0x18E60L, member.size);
        assertEquals(0x3E0F980L, member.offset);
    }
}
