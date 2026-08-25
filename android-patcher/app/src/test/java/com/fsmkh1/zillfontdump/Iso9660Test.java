package com.fsmkh1.zillfontdump;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertNotNull;

import java.io.File;
import java.io.RandomAccessFile;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.nio.charset.StandardCharsets;
import org.junit.Test;

public class Iso9660Test {
    @Test
    public void findsCaseInsensitiveVersionedFile() throws Exception {
        File f = File.createTempFile("iso9660", ".iso");
        f.deleteOnExit();
        try (RandomAccessFile raf = new RandomAccessFile(f, "rw")) {
            raf.setLength(24L * Iso9660.SECTOR_SIZE);
            byte[] pvd = new byte[Iso9660.SECTOR_SIZE];
            pvd[0] = 1;
            System.arraycopy("CD001".getBytes(StandardCharsets.US_ASCII), 0, pvd, 1, 5);
            pvd[6] = 1;
            writeDirRecord(pvd, 156, 20, Iso9660.SECTOR_SIZE, true, new byte[]{0});
            raf.seek(16L * Iso9660.SECTOR_SIZE);
            raf.write(pvd);

            byte[] dir = new byte[Iso9660.SECTOR_SIZE];
            int pos = 0;
            pos += writeDirRecord(dir, pos, 20, Iso9660.SECTOR_SIZE, true, new byte[]{0});
            pos += writeDirRecord(dir, pos, 20, Iso9660.SECTOR_SIZE, true, new byte[]{1});
            byte[] name = "HELLO.TXT;1".getBytes(StandardCharsets.US_ASCII);
            writeDirRecord(dir, pos, 21, 5, false, name);
            raf.seek(20L * Iso9660.SECTOR_SIZE);
            raf.write(dir);
            raf.seek(21L * Iso9660.SECTOR_SIZE);
            raf.write("hello".getBytes(StandardCharsets.US_ASCII));

            Iso9660 iso = new Iso9660(raf.getChannel());
            Iso9660.Entry e = iso.find("hello.txt");
            assertNotNull(e);
            assertEquals(5, e.size);
            ByteBuffer data = iso.read(e);
            byte[] actual = new byte[5];
            data.get(actual);
            assertEquals("hello", new String(actual, StandardCharsets.US_ASCII));
        }
    }

    private static int writeDirRecord(byte[] dst, int pos, int extent, int size,
                                      boolean directory, byte[] name) {
        int len = 33 + name.length + ((name.length & 1) == 0 ? 1 : 0);
        dst[pos] = (byte) len;
        dst[pos + 1] = 0;
        putU32LE(dst, pos + 2, extent);
        putU32BE(dst, pos + 6, extent);
        putU32LE(dst, pos + 10, size);
        putU32BE(dst, pos + 14, size);
        dst[pos + 25] = (byte) (directory ? 2 : 0);
        dst[pos + 28] = 1;
        dst[pos + 29] = 0;
        dst[pos + 30] = 0;
        dst[pos + 31] = 1;
        dst[pos + 32] = (byte) name.length;
        System.arraycopy(name, 0, dst, pos + 33, name.length);
        return len;
    }

    private static void putU32LE(byte[] b, int o, int v) {
        b[o] = (byte) v;
        b[o + 1] = (byte) (v >>> 8);
        b[o + 2] = (byte) (v >>> 16);
        b[o + 3] = (byte) (v >>> 24);
    }

    private static void putU32BE(byte[] b, int o, int v) {
        b[o] = (byte) (v >>> 24);
        b[o + 1] = (byte) (v >>> 16);
        b[o + 2] = (byte) (v >>> 8);
        b[o + 3] = (byte) v;
    }
}
