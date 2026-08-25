package com.fsmkh1.zillfontdump;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;

/**
 * Guards and rewrites the opening three-line invocation used for the first
 * end-to-end Korean sentence PoC.
 *
 * The source record is msgsec001 record 7 (ID 10007):
 *   汝、無限のソウルを持つ者よ\n我に応ぜよ\n我が問いに答え、その魂を我に示せ\0
 *
 * The replacement intentionally reuses only the five audited renderer keys
 * selected for this PoC and preserves the two native line-break controls.
 */
final class StartupMessage {
    static final String MEMBER_NAME = "message/msgsec001.dat";
    static final int RECORD_INDEX = 7;

    // CP932 source bytes, including native 0x0A line breaks and trailing NUL.
    private static final byte[] EXPECTED = hex(
            "93f0814196b38cc082cc835c8345838b82f08e9d82c28ed282e60a" +
            "89e482c9899e82ba82e60a" +
            "89e482aa96e282a282c9939a82a6814182bb82cc8db082f089e482c98ea682b900");

    // "테스트 성공\n테스트 성공\n테스트 성공\0"
    // 테 -> E1 A1 (PAF key A1E1)
    // 스 -> E9 A1 (PAF key A1E9)
    // 트 -> E2 B8 (PAF key B8E2)
    // 성 -> E6 BB (PAF key BBE6)
    // 공 -> E6 BF (PAF key BFE6)
    private static final byte[] REPLACEMENT = hex(
            "e1a1e9a1e2b820e6bbe6bf0a" +
            "e1a1e9a1e2b820e6bbe6bf0a" +
            "e1a1e9a1e2b820e6bbe6bf00");

    static final class Record {
        final int offset;
        final int span;

        Record(int offset, int span) {
            this.offset = offset;
            this.span = span;
        }
    }

    private StartupMessage() {}

    static Record inspect(ByteBuffer input) throws IOException {
        ByteBuffer data = input.duplicate().order(ByteOrder.LITTLE_ENDIAN);
        if (data.remaining() < 2) throw new IOException("msgsec001.dat is too small");
        int count = u16(data, 0);
        if (count <= RECORD_INDEX) {
            throw new IOException("msgsec001.dat has only " + count + " records");
        }
        int tableEnd = 2 + count * 2;
        if (tableEnd > data.limit()) {
            throw new IOException("msgsec001.dat offset table extends past member");
        }

        int[] offsets = new int[count];
        int previous = tableEnd;
        for (int i = 0; i < count; i++) {
            int offset = u16(data, 2 + i * 2);
            if (offset < tableEnd || offset < previous || offset > data.limit()) {
                throw new IOException("msgsec001.dat invalid record offset at index " + i);
            }
            offsets[i] = offset;
            previous = offset;
        }

        int start = offsets[RECORD_INDEX];
        int end = RECORD_INDEX + 1 < count ? offsets[RECORD_INDEX + 1] : data.limit();
        int span = end - start;
        if (span < EXPECTED.length) {
            throw new IOException("startup message record is shorter than expected source guard");
        }
        for (int i = 0; i < EXPECTED.length; i++) {
            if (data.get(start + i) != EXPECTED[i]) {
                throw new IOException("startup message source guard mismatch at byte " + i);
            }
        }
        if (REPLACEMENT.length > EXPECTED.length) {
            throw new IOException("internal Korean startup replacement exceeds guarded source length");
        }
        return new Record(start, span);
    }

    static byte[] guardedPatchBytes() {
        // Only the exact guarded source prefix is replaced. Bytes after the
        // original NUL within the record span remain untouched.
        byte[] out = new byte[EXPECTED.length];
        System.arraycopy(REPLACEMENT, 0, out, 0, REPLACEMENT.length);
        return out;
    }

    static int guardedLength() {
        return EXPECTED.length;
    }

    static int replacementLength() {
        return REPLACEMENT.length;
    }

    private static int u16(ByteBuffer data, int offset) {
        return (data.get(offset) & 0xff) | ((data.get(offset + 1) & 0xff) << 8);
    }

    private static byte[] hex(String value) {
        if ((value.length() & 1) != 0) throw new IllegalArgumentException("odd hex length");
        byte[] out = new byte[value.length() / 2];
        for (int i = 0; i < out.length; i++) {
            int hi = Character.digit(value.charAt(i * 2), 16);
            int lo = Character.digit(value.charAt(i * 2 + 1), 16);
            if (hi < 0 || lo < 0) throw new IllegalArgumentException("invalid hex");
            out[i] = (byte) ((hi << 4) | lo);
        }
        return out;
    }
}
