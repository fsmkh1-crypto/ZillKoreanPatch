package com.fsmkh1.zillfontdump;

import java.io.IOException;
import java.io.OutputStream;
import java.nio.ByteBuffer;
import java.nio.channels.FileChannel;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.ArrayList;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.zip.ZipEntry;
import java.util.zip.ZipOutputStream;

final class FontExtractor {
    static final String EXPECTED_DISC_ID = "ULJM05410";
    static final String EXPECTED_VERSION = "1.03";
    private static final int EXPECTED_MEMBER_COUNT = 14231;
    private static final int ZILLFONT_INDEX = 13611;
    private static final int JILLBTN_INDEX = 13612;
    private static final long ZILLFONT_OFFSET = 0x3D8F510L;
    private static final long ZILLFONT_MEMBER_SIZE = 0x80470L;
    private static final long JILLBTN_OFFSET = 0x3E0F980L;
    private static final long JILLBTN_MEMBER_SIZE = 0x18E60L;
    private static final long WRAPPER_SIZE = 0x10L;

    static final class Inspection {
        final String discId;
        final String version;
        final long isoSize;
        final Iso9660.Entry paBin;
        final Iso9660.Entry paArc;
        final PaaIndex.Member zillfont;
        final PaaIndex.Member jillbtn;

        Inspection(String discId, String version, long isoSize,
                   Iso9660.Entry paBin, Iso9660.Entry paArc,
                   PaaIndex.Member zillfont, PaaIndex.Member jillbtn) {
            this.discId = discId;
            this.version = version;
            this.isoSize = isoSize;
            this.paBin = paBin;
            this.paArc = paArc;
            this.zillfont = zillfont;
            this.jillbtn = jillbtn;
        }
    }

    private static final class Exported {
        final String path;
        final long payloadSize;
        final String sha256;
        final PaaIndex.Member member;

        Exported(String path, long payloadSize, String sha256, PaaIndex.Member member) {
            this.path = path;
            this.payloadSize = payloadSize;
            this.sha256 = sha256;
            this.member = member;
        }
    }

    private FontExtractor() {}

    static Inspection inspect(FileChannel channel) throws IOException {
        Iso9660 iso = new Iso9660(channel);
        Iso9660.Entry sfo = require(iso, "PSP_GAME/PARAM.SFO");
        Map<String, String> fields = SfoParser.parseStrings(iso.read(sfo));
        String discId = normalizeDiscId(fields.get("DISC_ID"));
        String version = safe(fields.get("DISC_VERSION"));
        if (!EXPECTED_DISC_ID.equals(discId)) {
            throw new IOException("Unsupported game: DISC_ID=" + safe(fields.get("DISC_ID")) +
                    " (expected ULJM-05410)");
        }
        if (!EXPECTED_VERSION.equals(version)) {
            throw new IOException("Unsupported game version: " + version + " (expected 1.03)");
        }

        Iso9660.Entry paBin = require(iso, "PSP_GAME/USRDIR/pa.bin");
        Iso9660.Entry paArc = require(iso, "PSP_GAME/USRDIR/pa.arc");
        PaaIndex index = PaaIndex.parse(iso.read(paBin));
        if (index.count() != EXPECTED_MEMBER_COUNT) {
            throw new IOException("PAA member count=" + index.count() + " (expected " + EXPECTED_MEMBER_COUNT + ")");
        }

        PaaIndex.Member zillfont = index.member(ZILLFONT_INDEX);
        PaaIndex.Member jillbtn = index.member(JILLBTN_INDEX);
        validateMember(zillfont, "font/zillfont.par", ZILLFONT_OFFSET, ZILLFONT_MEMBER_SIZE);
        validateMember(jillbtn, "2d/font/jillbtn.par", JILLBTN_OFFSET, JILLBTN_MEMBER_SIZE);

        long required = Math.max(zillfont.offset + zillfont.size, jillbtn.offset + jillbtn.size);
        if (paArc.size < required) {
            throw new IOException("pa.arc is smaller than the validated ULJM-05410 layout");
        }
        return new Inspection(discId, version, channel.size(), paBin, paArc, zillfont, jillbtn);
    }

    static void export(FileChannel channel, Inspection inspection, OutputStream output,
                       String sourceName) throws IOException {
        Iso9660 iso = new Iso9660(channel);
        try (ZipOutputStream zip = new ZipOutputStream(output)) {
            List<Exported> files = new ArrayList<>();
            files.add(exportMember(iso, inspection.paArc, zip,
                    "font/zillfont.par", inspection.zillfont));
            files.add(exportMember(iso, inspection.paArc, zip,
                    "2d/font/jillbtn.par", inspection.jillbtn));

            String manifest = manifestJson(inspection, sourceName, files);
            zip.putNextEntry(new ZipEntry("manifest.json"));
            zip.write(manifest.getBytes(java.nio.charset.StandardCharsets.UTF_8));
            zip.closeEntry();
        }
    }

    private static Exported exportMember(Iso9660 iso, Iso9660.Entry paArc, ZipOutputStream zip,
                                         String path, PaaIndex.Member member) throws IOException {
        long payloadSize = member.size - WRAPPER_SIZE;
        long absolute = paArc.extent * Iso9660.SECTOR_SIZE + member.offset + WRAPPER_SIZE;
        MessageDigest digest = sha256();
        zip.putNextEntry(new ZipEntry(path));
        iso.copyRange(absolute, payloadSize, zip, digest);
        zip.closeEntry();
        return new Exported(path, payloadSize, hex(digest.digest()), member);
    }

    private static void validateMember(PaaIndex.Member member, String expectedName,
                                       long expectedOffset, long expectedSize) throws IOException {
        if (!expectedName.equals(member.name)) {
            throw new IOException("PAA member " + member.index + " is " + member.name +
                    " (expected " + expectedName + ")");
        }
        if (member.offset != expectedOffset) {
            throw new IOException(member.name + " offset=" + hex0x(member.offset) +
                    " (expected " + hex0x(expectedOffset) + ")");
        }
        if (member.size != expectedSize) {
            throw new IOException(member.name + " size=" + hex0x(member.size) +
                    " (expected " + hex0x(expectedSize) + ")");
        }
    }

    private static Iso9660.Entry require(Iso9660 iso, String path) throws IOException {
        Iso9660.Entry e = iso.find(path);
        if (e == null) throw new IOException("Missing ISO file: " + path);
        return e;
    }

    private static MessageDigest sha256() throws IOException {
        try {
            return MessageDigest.getInstance("SHA-256");
        } catch (NoSuchAlgorithmException e) {
            throw new IOException("SHA-256 unavailable", e);
        }
    }

    private static String manifestJson(Inspection i, String sourceName, List<Exported> files) {
        StringBuilder s = new StringBuilder();
        s.append("{\n");
        s.append("  \"format\": \"zill-font-extract-v1\",\n");
        s.append("  \"extractorVersion\": \"0.1.0\",\n");
        s.append("  \"target\": \"ULJM-05410 v1.03\",\n");
        s.append("  \"discId\": \"").append(json(i.discId)).append("\",\n");
        s.append("  \"discVersion\": \"").append(json(i.version)).append("\",\n");
        s.append("  \"sourceIso\": {\"name\": \"").append(json(sourceName))
                .append("\", \"size\": ").append(i.isoSize).append("},\n");
        s.append("  \"archive\": {\"paBinSize\": ").append(i.paBin.size)
                .append(", \"paArcSize\": ").append(i.paArc.size)
                .append(", \"memberCount\": ").append(EXPECTED_MEMBER_COUNT).append("},\n");
        s.append("  \"files\": [\n");
        for (int n = 0; n < files.size(); n++) {
            Exported f = files.get(n);
            s.append("    {\"path\": \"").append(json(f.path)).append("\", ")
                    .append("\"size\": ").append(f.payloadSize).append(", ")
                    .append("\"sha256\": \"").append(f.sha256).append("\", ")
                    .append("\"paaIndex\": ").append(f.member.index).append(", ")
                    .append("\"arcOffset\": \"").append(hex0x(f.member.offset)).append("\", ")
                    .append("\"memberSize\": ").append(f.member.size).append(", ")
                    .append("\"wrapperBytesStripped\": 16}");
            if (n + 1 < files.size()) s.append(',');
            s.append('\n');
        }
        s.append("  ]\n}");
        return s.toString();
    }

    private static String normalizeDiscId(String s) {
        return safe(s).replace("-", "").replace("_", "").toUpperCase(Locale.ROOT);
    }

    private static String safe(String s) {
        return s == null ? "" : s.trim();
    }

    private static String json(String s) {
        return safe(s).replace("\\", "\\\\").replace("\"", "\\\"")
                .replace("\n", "\\n").replace("\r", "\\r");
    }

    private static String hex0x(long value) {
        return "0x" + Long.toHexString(value).toUpperCase(Locale.ROOT);
    }

    private static String hex(byte[] data) {
        StringBuilder s = new StringBuilder(data.length * 2);
        for (byte b : data) s.append(String.format(Locale.ROOT, "%02x", b & 0xff));
        return s.toString();
    }
}
