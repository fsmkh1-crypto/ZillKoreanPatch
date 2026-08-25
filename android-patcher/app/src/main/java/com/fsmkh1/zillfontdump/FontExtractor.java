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

        Inspection(String discId, String version, long isoSize,
                   Iso9660.Entry paBin, Iso9660.Entry paArc) {
            this.discId = discId;
            this.version = version;
            this.isoSize = isoSize;
            this.paBin = paBin;
            this.paArc = paArc;
        }
    }

    private static final class Exported {
        final String path;
        final long payloadSize;
        final String sha256;
        final long arcOffset;
        final long memberSize;

        Exported(String path, long payloadSize, String sha256, long arcOffset, long memberSize) {
            this.path = path;
            this.payloadSize = payloadSize;
            this.sha256 = sha256;
            this.arcOffset = arcOffset;
            this.memberSize = memberSize;
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
        ByteBuffer binHead = iso.read(paBin);
        if (binHead.remaining() < 4 || binHead.get(0) != 'P' || binHead.get(1) != 'A' ||
                binHead.get(2) != 'A' || binHead.get(3) != 0) {
            throw new IOException("pa.bin has unexpected magic; refusing fixed-offset extraction");
        }
        long required = Math.max(ZILLFONT_OFFSET + ZILLFONT_MEMBER_SIZE,
                JILLBTN_OFFSET + JILLBTN_MEMBER_SIZE);
        if (paArc.size < required) {
            throw new IOException("pa.arc is smaller than the validated ULJM-05410 layout");
        }
        return new Inspection(discId, version, channel.size(), paBin, paArc);
    }

    static void export(FileChannel channel, Inspection inspection, OutputStream output,
                       String sourceName) throws IOException {
        Iso9660 iso = new Iso9660(channel);
        try (ZipOutputStream zip = new ZipOutputStream(output)) {
            List<Exported> files = new ArrayList<>();
            files.add(exportMember(iso, inspection.paArc, zip,
                    "font/zillfont.par", ZILLFONT_OFFSET, ZILLFONT_MEMBER_SIZE));
            files.add(exportMember(iso, inspection.paArc, zip,
                    "2d/font/jillbtn.par", JILLBTN_OFFSET, JILLBTN_MEMBER_SIZE));

            String manifest = manifestJson(inspection, sourceName, files);
            zip.putNextEntry(new ZipEntry("manifest.json"));
            zip.write(manifest.getBytes(java.nio.charset.StandardCharsets.UTF_8));
            zip.closeEntry();
        }
    }

    private static Exported exportMember(Iso9660 iso, Iso9660.Entry paArc, ZipOutputStream zip,
                                         String path, long arcOffset, long memberSize) throws IOException {
        long payloadSize = memberSize - WRAPPER_SIZE;
        long absolute = paArc.extent * Iso9660.SECTOR_SIZE + arcOffset + WRAPPER_SIZE;
        MessageDigest digest = sha256();
        zip.putNextEntry(new ZipEntry(path));
        iso.copyRange(absolute, payloadSize, zip, digest);
        zip.closeEntry();
        return new Exported(path, payloadSize, hex(digest.digest()), arcOffset, memberSize);
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
                .append(", \"paArcSize\": ").append(i.paArc.size).append("},\n");
        s.append("  \"files\": [\n");
        for (int n = 0; n < files.size(); n++) {
            Exported f = files.get(n);
            s.append("    {\"path\": \"").append(json(f.path)).append("\", ")
                    .append("\"size\": ").append(f.payloadSize).append(", ")
                    .append("\"sha256\": \"").append(f.sha256).append("\", ")
                    .append("\"arcOffset\": \"0x").append(Long.toHexString(f.arcOffset).toUpperCase(Locale.ROOT)).append("\", ")
                    .append("\"memberSize\": ").append(f.memberSize).append(", ")
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

    private static String hex(byte[] data) {
        StringBuilder s = new StringBuilder(data.length * 2);
        for (byte b : data) s.append(String.format(Locale.ROOT, "%02x", b & 0xff));
        return s.toString();
    }
}
