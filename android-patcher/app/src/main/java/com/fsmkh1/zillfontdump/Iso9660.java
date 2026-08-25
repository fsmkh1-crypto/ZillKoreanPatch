package com.fsmkh1.zillfontdump;

import java.io.EOFException;
import java.io.IOException;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.nio.channels.FileChannel;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.List;
import java.util.Locale;

final class Iso9660 {
    static final int SECTOR_SIZE = 2048;

    static final class Entry {
        final String name;
        final long extent;
        final long size;
        final boolean directory;

        Entry(String name, long extent, long size, boolean directory) {
            this.name = name;
            this.extent = extent;
            this.size = size;
            this.directory = directory;
        }
    }

    private final FileChannel channel;
    private final Entry root;

    Iso9660(FileChannel channel) throws IOException {
        this.channel = channel;
        ByteBuffer pvd = readAt(16L * SECTOR_SIZE, SECTOR_SIZE);
        if ((pvd.get(0) & 0xff) != 1 ||
                !"CD001".equals(ascii(pvd, 1, 5)) ||
                (pvd.get(6) & 0xff) != 1) {
            throw new IOException("Not an ISO9660 primary volume descriptor");
        }
        this.root = parseDirectoryRecord(pvd, 156);
        if (!root.directory) {
            throw new IOException("ISO9660 root entry is not a directory");
        }
    }

    Entry find(String path) throws IOException {
        String[] parts = path.replace('\\', '/').split("/");
        Entry current = root;
        for (String raw : parts) {
            if (raw.isEmpty()) continue;
            if (!current.directory) return null;
            String wanted = normalizeName(raw);
            Entry next = null;
            for (Entry child : list(current)) {
                if (normalizeName(child.name).equals(wanted)) {
                    next = child;
                    break;
                }
            }
            if (next == null) return null;
            current = next;
        }
        return current;
    }

    ByteBuffer read(Entry entry) throws IOException {
        if (entry.size > Integer.MAX_VALUE) {
            throw new IOException("Entry too large for in-memory read: " + entry.name);
        }
        return readAt(entry.extent * SECTOR_SIZE, (int) entry.size);
    }

    void copyRange(long absoluteOffset, long length, java.io.OutputStream out,
                   java.security.MessageDigest digest) throws IOException {
        byte[] buf = new byte[64 * 1024];
        long remaining = length;
        long pos = absoluteOffset;
        while (remaining > 0) {
            int n = (int) Math.min(buf.length, remaining);
            ByteBuffer bb = ByteBuffer.wrap(buf, 0, n);
            readFully(pos, bb);
            out.write(buf, 0, n);
            if (digest != null) digest.update(buf, 0, n);
            pos += n;
            remaining -= n;
        }
    }

    private List<Entry> list(Entry dir) throws IOException {
        if (!dir.directory) throw new IOException("Not a directory: " + dir.name);
        if (dir.size > Integer.MAX_VALUE) throw new IOException("Directory too large");
        ByteBuffer data = readAt(dir.extent * SECTOR_SIZE, (int) dir.size);
        List<Entry> out = new ArrayList<>();
        int pos = 0;
        while (pos < data.limit()) {
            int len = data.get(pos) & 0xff;
            if (len == 0) {
                pos = ((pos / SECTOR_SIZE) + 1) * SECTOR_SIZE;
                continue;
            }
            if (pos + len > data.limit() || len < 34) {
                throw new IOException("Malformed ISO9660 directory record");
            }
            int nameLen = data.get(pos + 32) & 0xff;
            if (33 + nameLen > len) throw new IOException("Malformed ISO9660 filename");
            int first = data.get(pos + 33) & 0xff;
            if (!(nameLen == 1 && (first == 0 || first == 1))) {
                out.add(parseDirectoryRecord(data, pos));
            }
            pos += len;
        }
        return out;
    }

    private Entry parseDirectoryRecord(ByteBuffer data, int pos) throws IOException {
        int len = data.get(pos) & 0xff;
        if (len < 34 || pos + len > data.limit()) throw new IOException("Bad directory record");
        long extent = u32le(data, pos + 2);
        long size = u32le(data, pos + 10);
        boolean directory = (data.get(pos + 25) & 0x02) != 0;
        int nameLen = data.get(pos + 32) & 0xff;
        if (pos + 33 + nameLen > data.limit()) throw new IOException("Bad directory filename");
        String name;
        if (nameLen == 1 && (data.get(pos + 33) & 0xff) == 0) name = ".";
        else if (nameLen == 1 && (data.get(pos + 33) & 0xff) == 1) name = "..";
        else name = new String(slice(data, pos + 33, nameLen), StandardCharsets.US_ASCII);
        return new Entry(name, extent, size, directory);
    }

    private ByteBuffer readAt(long offset, int length) throws IOException {
        ByteBuffer bb = ByteBuffer.allocate(length).order(ByteOrder.LITTLE_ENDIAN);
        readFully(offset, bb);
        bb.flip();
        return bb;
    }

    private void readFully(long offset, ByteBuffer dst) throws IOException {
        long pos = offset;
        while (dst.hasRemaining()) {
            int n = channel.read(dst, pos);
            if (n < 0) throw new EOFException("Unexpected EOF at " + pos);
            if (n == 0) continue;
            pos += n;
        }
    }

    private static String normalizeName(String name) {
        int semi = name.indexOf(';');
        if (semi >= 0) name = name.substring(0, semi);
        while (name.endsWith(".")) name = name.substring(0, name.length() - 1);
        return name.toUpperCase(Locale.ROOT);
    }

    private static String ascii(ByteBuffer b, int off, int len) {
        return new String(slice(b, off, len), StandardCharsets.US_ASCII);
    }

    private static byte[] slice(ByteBuffer b, int off, int len) {
        byte[] out = new byte[len];
        for (int i = 0; i < len; i++) out[i] = b.get(off + i);
        return out;
    }

    private static long u32le(ByteBuffer b, int off) {
        return ((long) b.get(off) & 0xff) |
                (((long) b.get(off + 1) & 0xff) << 8) |
                (((long) b.get(off + 2) & 0xff) << 16) |
                (((long) b.get(off + 3) & 0xff) << 24);
    }
}
