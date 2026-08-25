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
    private static final long ZILLFONT_MEMBER_SIZE = 0x80470L;
    private static final String ZILLFONT_SOURCE_SHA256 = "0d3d6d2648870e87a01636cdfc7cc7af8100ea40b71e5ed05f82ac197606584a";
    private static final long JILLBTN_MEMBER_SIZE = 0x18E60L;
    private static final String JILLBTN_SOURCE_SHA256 = "95b48379092db4db72f890d5a221ba8c4094dd438cb4c4eba98eb5520c7b17aa";
    private static final String EBOOT_SOURCE_SHA256 = "2a52012be00c07512dcde932ff6e9eb9b96912c59dd5a25c7c26ef821c124d68";
    private static final String BINDATA_SOURCE_SHA256 = "3241fc000f3d52fe8522baaa985fd866e29d64d3a0f23ac4e28b66dee957de3e";

    static final class Inspection {
        final String discId;
        final String version;
        final long isoSize;
        final Iso9660.Entry eboot;
        final Iso9660.Entry paBin;
        final Iso9660.Entry paArc;
        final PaaIndex.Member zillfont;
        final PaaIndex.Member jillbtn;
        final Iso9660.Entry bindataArc;
        final PaaIndex.Member bindata;
        final String bindataArchive;

        Inspection(String discId, String version, long isoSize,
                   Iso9660.Entry eboot, Iso9660.Entry paBin, Iso9660.Entry paArc,
                   PaaIndex.Member zillfont, PaaIndex.Member jillbtn,
                   Iso9660.Entry bindataArc, PaaIndex.Member bindata, String bindataArchive) {
            this.discId = discId;
            this.version = version;
            this.isoSize = isoSize;
            this.eboot = eboot;
            this.paBin = paBin;
            this.paArc = paArc;
            this.zillfont = zillfont;
            this.jillbtn = jillbtn;
            this.bindataArc = bindataArc;
            this.bindata = bindata;
            this.bindataArchive = bindataArchive;
        }
    }

    private static final class Exported {
        final String path;
        final long payloadSize;
        final String sha256;
        final PaaIndex.Member member;
        final String archive;

        Exported(String path, long payloadSize, String sha256,
                 PaaIndex.Member member, String archive) {
            this.path = path;
            this.payloadSize = payloadSize;
            this.sha256 = sha256;
            this.member = member;
            this.archive = archive;
        }
    }

    private static final class LocatedMember {
        final String archive;
        final Iso9660.Entry arc;
        final PaaIndex.Member member;

        LocatedMember(String archive, Iso9660.Entry arc, PaaIndex.Member member) {
            this.archive = archive;
            this.arc = arc;
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

        Iso9660.Entry eboot = require(iso, "PSP_GAME/SYSDIR/EBOOT.BIN");
        validateIsoEntryHash(iso, eboot, "EBOOT.BIN", EBOOT_SOURCE_SHA256);

        Iso9660.Entry paBin = require(iso, "PSP_GAME/USRDIR/pa.bin");
        Iso9660.Entry paArc = require(iso, "PSP_GAME/USRDIR/pa.arc");
        PaaIndex paIndex = PaaIndex.parse(iso.read(paBin));
        if (paIndex.count() != EXPECTED_MEMBER_COUNT) {
            throw new IOException("PAA member count=" + paIndex.count() + " (expected " + EXPECTED_MEMBER_COUNT + ")");
        }

        PaaIndex.Member zillfont = paIndex.member(ZILLFONT_INDEX);
        PaaIndex.Member jillbtn = paIndex.member(JILLBTN_INDEX);
        validateMember(zillfont, "font/zillfont.par", ZILLFONT_MEMBER_SIZE);
        validateMember(jillbtn, "2d/font/jillbtn.par", JILLBTN_MEMBER_SIZE);

        long required = Math.max(zillfont.offset + zillfont.size, jillbtn.offset + jillbtn.size);
        if (paArc.size < required) {
            throw new IOException("pa.arc is smaller than the PAA index requires");
        }
        validateSourceHash(iso, paArc, zillfont, ZILLFONT_SOURCE_SHA256);
        validateSourceHash(iso, paArc, jillbtn, JILLBTN_SOURCE_SHA256);

        LocatedMember bindata = locateMember(iso, paIndex, paArc, "pa", "data/bindata.dat");
        if (bindata == null) {
            Iso9660.Entry pamiBin = require(iso, "PSP_GAME/USRDIR/pami.bin");
            Iso9660.Entry pamiArc = require(iso, "PSP_GAME/USRDIR/pami.arc");
            PaaIndex pamiIndex = PaaIndex.parse(iso.read(pamiBin));
            bindata = locateMember(iso, pamiIndex, pamiArc, "pami", "data/bindata.dat");
        }
        if (bindata == null) throw new IOException("Retail archives do not contain data/bindata.dat");
        validateSourceHash(iso, bindata.arc, bindata.member, BINDATA_SOURCE_SHA256);

        return new Inspection(discId, version, channel.size(), eboot, paBin, paArc,
                zillfont, jillbtn, bindata.arc, bindata.member, bindata.archive);
    }

    static void export(FileChannel channel, Inspection inspection, OutputStream output,
                       String sourceName) throws IOException {
        Iso9660 iso = new Iso9660(channel);
        try (ZipOutputStream zip = new ZipOutputStream(output)) {
            List<Exported> files = new ArrayList<>();
            files.add(exportMember(iso, inspection.paArc, zip,
                    "font/zillfont.par", inspection.zillfont, "pa"));
            files.add(exportMember(iso, inspection.paArc, zip,
                    "2d/font/jillbtn.par", inspection.jillbtn, "pa"));
            files.add(exportIsoEntry(iso, zip, "SYSDIR/EBOOT.BIN", inspection.eboot));
            files.add(exportMember(iso, inspection.bindataArc, zip,
                    "data/bindata.dat", inspection.bindata, inspection.bindataArchive));

            String manifest = manifestJson(inspection, sourceName, files);
            zip.putNextEntry(new ZipEntry("manifest.json"));
            zip.write(manifest.getBytes(java.nio.charset.StandardCharsets.UTF_8));
            zip.closeEntry();
        }
    }

    private static LocatedMember locateMember(Iso9660 iso, PaaIndex index, Iso9660.Entry arc,
                                               String archive, String wanted) throws IOException {
        PaaIndex.Member found = null;
        for (int i = 0; i < index.count(); i++) {
            PaaIndex.Member member = index.member(i);
            if (!wanted.equals(member.name)) continue;
            if (found != null) throw new IOException(wanted + " is duplicated in " + archive + " archive");
            if (member.offset + member.size > arc.size) {
                throw new IOException(wanted + " extends past " + archive + ".arc");
            }
            found = member;
        }
        return found == null ? null : new LocatedMember(archive, arc, found);
    }

    private static Exported exportIsoEntry(Iso9660 iso, ZipOutputStream zip,
                                            String path, Iso9660.Entry entry) throws IOException {
        MessageDigest digest = sha256();
        zip.putNextEntry(new ZipEntry(path));
        long absolute = entry.extent * Iso9660.SECTOR_SIZE;
        iso.copyRange(absolute, entry.size, zip, digest);
        zip.closeEntry();
        return new Exported(path, entry.size, hex(digest.digest()), null, "iso");
    }

    private static Exported exportMember(Iso9660 iso, Iso9660.Entry arc, ZipOutputStream zip,
                                         String path, PaaIndex.Member member,
                                         String archive) throws IOException {
        long absolute = arc.extent * Iso9660.SECTOR_SIZE + member.offset;
        MessageDigest digest = sha256();
        zip.putNextEntry(new ZipEntry(path));
        iso.copyRange(absolute, member.size, zip, digest);
        zip.closeEntry();
        return new Exported(path, member.size, hex(digest.digest()), member, archive);
    }

    private static void validateIsoEntryHash(Iso9660 iso, Iso9660.Entry entry,
                                             String label, String expected) throws IOException {
        MessageDigest digest = sha256();
        iso.copyRange(entry.extent * Iso9660.SECTOR_SIZE, entry.size, nullSink(), digest);
        String actual = hex(digest.digest());
        if (!expected.equals(actual)) {
            throw new IOException(label + " SHA-256=" + actual + " (unsupported source file)");
        }
    }

    private static void validateSourceHash(Iso9660 iso, Iso9660.Entry arc,
                                           PaaIndex.Member member, String expected) throws IOException {
        MessageDigest digest = sha256();
        long absolute = arc.extent * Iso9660.SECTOR_SIZE + member.offset;
        iso.copyRange(absolute, member.size, nullSink(), digest);
        String actual = hex(digest.digest());
        if (!expected.equals(actual)) {
            throw new IOException(member.name + " SHA-256=" + actual + " (unsupported source member)");
        }
    }

    private static OutputStream nullSink() {
        return new OutputStream() {
            @Override public void write(int value) {}
            @Override public void write(byte[] bytes, int offset, int length) {}
        };
    }

    private static void validateMember(PaaIndex.Member member, String expectedName,
                                       long expectedSize) throws IOException {
        if (!expectedName.equals(member.name)) {
            throw new IOException("PAA member " + member.index + " is " + member.name +
                    " (expected " + expectedName + ")");
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
        s.append("  \"format\": \"zill-slot-audit-extract-v3\",\n");
        s.append("  \"extractorVersion\": \"0.4.0\",\n");
        s.append("  \"target\": \"ULJM-05410 v1.03\",\n");
        s.append("  \"discId\": \"").append(json(i.discId)).append("\",\n");
        s.append("  \"discVersion\": \"").append(json(i.version)).append("\",\n");
        s.append("  \"sourceIso\": {\"name\": \"").append(json(sourceName))
                .append("\", \"size\": ").append(i.isoSize).append("},\n");
        s.append("  \"files\": [\n");
        for (int n = 0; n < files.size(); n++) {
            Exported f = files.get(n);
            s.append("    {\"path\": \"").append(json(f.path)).append("\", ")
                    .append("\"size\": ").append(f.payloadSize).append(", ")
                    .append("\"sha256\": \"").append(f.sha256).append("\", ")
                    .append("\"source\": \"").append(json(f.archive)).append("\"");
            if (f.member != null) {
                s.append(", \"paaIndex\": ").append(f.member.index)
                        .append(", \"arcOffset\": \"").append(hex0x(f.member.offset)).append("\"")
                        .append(", \"memberSize\": ").append(f.member.size);
            }
            s.append('}');
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
