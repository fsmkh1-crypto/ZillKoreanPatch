package com.fsmkh1.zillfontdump;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.nio.CharBuffer;
import java.nio.charset.CharacterCodingException;
import java.nio.charset.CodingErrorAction;
import java.text.Normalizer;
import java.util.ArrayList;
import java.util.List;

/**
 * Guards the opening three-line invocation and rewrites only its first natural-
 * text line for the first end-to-end Korean sentence PoC.
 *
 * The Japanese contributor text is a decoded display projection; retail records
 * may encode kana through ESC K/H/k mode controls and half-width bytes. Therefore
 * this guard validates the actual displayed text and native control structure,
 * instead of assuming one particular byte-for-byte CP932 spelling of the line.
 */
final class StartupMessage {
    static final String MEMBER_NAME = "message/msgsec001.dat";
    static final int RECORD_INDEX = 7;

    private static final String EXPECTED_DISPLAY =
            "汝、無限のソウルを持つ者よ<line-break>" +
            "我に応ぜよ<line-break>" +
            "我が問いに答え、その魂を我に示せ<end>";

    // "테스트 성공" using the five selected PoC renderer keys.
    // 테 -> E1 A1 (PAF key A1E1)
    // 스 -> E9 A1 (PAF key A1E9)
    // 트 -> E2 B8 (PAF key B8E2)
    // 성 -> E6 BB (PAF key BBE6)
    // 공 -> E6 BF (PAF key BFE6)
    private static final byte[] KOREAN_LINE = hex("e1a1e9a1e2b820e6bbe6bf");

    static final class Segment {
        final int offset;
        final int length;

        Segment(int offset, int length) {
            this.offset = offset;
            this.length = length;
        }
    }

    static final class Record {
        final int offset;
        final int span;
        final Segment[] segments;

        Record(int offset, int span, Segment[] segments) {
            this.offset = offset;
            this.span = span;
            this.segments = segments;
        }
    }

    static final class ByteEdit {
        final int relativeOffset;
        final int value;

        ByteEdit(int relativeOffset, int value) {
            this.relativeOffset = relativeOffset;
            this.value = value;
        }
    }

    private static final class DisplayScan {
        final String display;
        final int[] lineBreaks;
        final int endOffset;

        DisplayScan(String display, int[] lineBreaks, int endOffset) {
            this.display = display;
            this.lineBreaks = lineBreaks;
            this.endOffset = endOffset;
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
        if (end <= start) throw new IOException("startup message record is empty");

        DisplayScan scan = scanDisplay(data, start, end);
        if (!EXPECTED_DISPLAY.equals(scan.display)) {
            throw new IOException("startup message display guard mismatch: " + scan.display);
        }
        if (scan.lineBreaks.length != 2) {
            throw new IOException("startup message must contain exactly two native line breaks");
        }

        Segment[] segments = new Segment[]{
                new Segment(start, scan.lineBreaks[0] - start),
                new Segment(scan.lineBreaks[0] + 1, scan.lineBreaks[1] - scan.lineBreaks[0] - 1),
                new Segment(scan.lineBreaks[1] + 1, scan.endOffset - scan.lineBreaks[1] - 1)
        };
        if (segments[0].length < KOREAN_LINE.length) {
            throw new IOException("startup message first line has only " + segments[0].length
                    + " raw bytes; need " + KOREAN_LINE.length);
        }
        return new Record(start, end - start, segments);
    }

    static ByteEdit[] patchEdits(Record record) {
        if (record.segments == null || record.segments.length != 3) {
            throw new IllegalArgumentException("startup record must have exactly three guarded text segments");
        }
        Segment first = record.segments[0];
        if (first.length < KOREAN_LINE.length) {
            throw new IllegalArgumentException("startup first-line segment is too short for Korean PoC text");
        }
        List<ByteEdit> edits = new ArrayList<>(first.length);
        for (int i = 0; i < first.length; i++) {
            int value = i < KOREAN_LINE.length ? KOREAN_LINE[i] & 0xff : 0x20;
            edits.add(new ByteEdit(first.offset - record.offset + i, value));
        }
        return edits.toArray(new ByteEdit[0]);
    }

    static int replacementLineLength() {
        return KOREAN_LINE.length;
    }

    private static DisplayScan scanDisplay(ByteBuffer data, int start, int end) throws IOException {
        StringBuilder out = new StringBuilder();
        List<Integer> lineBreaks = new ArrayList<>();
        String kanaMode = "hiragana";
        int endOffset = -1;

        for (int index = start; index < end; ) {
            int value = data.get(index) & 0xff;
            if (value == 0) break;
            if (index + 2 < end && value == 5
                    && (data.get(index + 1) & 0xff) == 5
                    && (data.get(index + 2) & 0xff) == 5) {
                out.append("<end>");
                endOffset = index;
                index += 3;
                break;
            }
            if (value == 10) {
                out.append("<line-break>");
                lineBreaks.add(index);
                index++;
                continue;
            }
            if (value == 0x1b && index + 1 < end) {
                int command = data.get(index + 1) & 0xff;
                if (command == 'K' || command == 'H' || command == 'k') {
                    kanaMode = command == 'K' ? "katakana" : command == 'H' ? "hiragana" : "halfwidth";
                    index += 2;
                    continue;
                }
                throw new IOException("unexpected renderer escape in startup message at " + (index - start));
            }
            if (isLead(value)) {
                if (index + 1 >= end || !isTrail(data.get(index + 1) & 0xff)) {
                    throw new IOException("invalid CP932 pair in startup message at " + (index - start));
                }
                out.append(decode(new byte[]{data.get(index), data.get(index + 1)}));
                index += 2;
                continue;
            }
            if (value >= 0xA6 && value <= 0xDF) {
                int length = 1;
                if (!"halfwidth".equals(kanaMode) && index + 1 < end) {
                    int next = data.get(index + 1) & 0xff;
                    if (next == 0xDE || next == 0xDF) length = 2;
                }
                byte[] bytes = new byte[length];
                for (int i = 0; i < length; i++) bytes[i] = data.get(index + i);
                String kana = decode(bytes);
                if (!"halfwidth".equals(kanaMode)) {
                    kana = Normalizer.normalize(kana, Normalizer.Form.NFKC);
                    if ("hiragana".equals(kanaMode)) kana = katakanaToHiragana(kana);
                }
                out.append(kana);
                index += length;
                continue;
            }
            if (value >= 0x20 && value <= 0x7e) {
                out.append((char) value);
                index++;
                continue;
            }
            throw new IOException("unexpected control byte 0x"
                    + Integer.toHexString(value) + " in startup message at " + (index - start));
        }

        if (endOffset < 0) throw new IOException("startup message is missing native <end> terminator");
        int[] breaks = new int[lineBreaks.size()];
        for (int i = 0; i < breaks.length; i++) breaks[i] = lineBreaks.get(i);
        return new DisplayScan(out.toString(), breaks, endOffset);
    }

    private static String decode(byte[] bytes) throws IOException {
        try {
            CharBuffer decoded = java.nio.charset.Charset.forName("Shift_JIS")
                    .newDecoder()
                    .onMalformedInput(CodingErrorAction.REPORT)
                    .onUnmappableCharacter(CodingErrorAction.REPORT)
                    .decode(ByteBuffer.wrap(bytes));
            return decoded.toString();
        } catch (CharacterCodingException e) {
            throw new IOException("invalid Shift_JIS in startup message", e);
        }
    }

    private static String katakanaToHiragana(String value) {
        StringBuilder out = new StringBuilder(value.length());
        for (int i = 0; i < value.length(); i++) {
            char c = value.charAt(i);
            if (c >= 'ァ' && c <= 'ヶ') c = (char) (c - 0x60);
            out.append(c);
        }
        return out.toString();
    }

    private static boolean isLead(int value) {
        return value >= 0x81 && value <= 0x9f || value >= 0xe0 && value <= 0xfc;
    }

    private static boolean isTrail(int value) {
        return value >= 0x40 && value <= 0x7e || value >= 0x80 && value <= 0xfc;
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
