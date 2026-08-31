package com.fsmkh1.zillfontdump;

import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertTrue;
import static org.junit.Assert.fail;

import org.junit.Test;

import java.io.File;
import java.io.FileOutputStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.security.MessageDigest;
import java.util.ArrayList;
import java.util.List;

public final class ProjectAssetIntegrityTest {
    private static final String[] PAYLOAD_PATHS = {
            "release/font/metrics.toml",
            "release/font/manifest.toml",
            "release/korean/font/glyphs.toml",
            "release/korean/strings/eboot.toml",
            "release/layout/consumer-map.toml",
            "release/layout/categories.toml",
            "docs/audit/fixtures/pr14-eboot-h0.toml",
            "docs/audit/fixtures/pr14-eboot-full.toml",
            "patches/executable/manifest.toml",
            "patches/system/param-sfo.toml",
            "payload-version.txt"
    };

    @Test
    public void markerAloneDoesNotCountAsCompletePayload() throws Exception {
        File root = Files.createTempDirectory("zillroot-partial").toFile();
        try {
            write(root, "payload-version.txt", "same-version\n");
            assertFalse(ProjectAssetIntegrity.isComplete(root));
            assertTrue(ProjectAssetIntegrity.missingFiles(root).contains("release/font/metrics.toml"));
            assertTrue(ProjectAssetIntegrity.missingFiles(root).contains(ProjectAssetIntegrity.MANIFEST_RELATIVE_PATH));
        } finally {
            deleteRecursively(root);
        }
    }

    @Test
    public void historicalH0FixtureIsRequired() throws Exception {
        File root = Files.createTempDirectory("zillroot-no-h0").toFile();
        try {
            writePayloadAndManifest(root);
            File h0 = new File(root, "docs/audit/fixtures/pr14-eboot-h0.toml");
            assertTrue(h0.delete());
            assertFalse(ProjectAssetIntegrity.isComplete(root));
            assertTrue(ProjectAssetIntegrity.missingFiles(root).contains("docs/audit/fixtures/pr14-eboot-h0.toml"));
        } finally {
            deleteRecursively(root);
        }
    }

    @Test
    public void manifestVerifiedPayloadPasses() throws Exception {
        File root = Files.createTempDirectory("zillroot-complete").toFile();
        try {
            writePayloadAndManifest(root);
            assertTrue(ProjectAssetIntegrity.isComplete(root));
            ProjectAssetIntegrity.verifyPayload(root);
        } finally {
            deleteRecursively(root);
        }
    }

    @Test
    public void changedPayloadByteFailsManifestVerification() throws Exception {
        File root = Files.createTempDirectory("zillroot-tampered").toFile();
        try {
            writePayloadAndManifest(root);
            write(root, "release/font/metrics.toml", "tampered\n");
            try {
                ProjectAssetIntegrity.verifyPayload(root);
                fail("tampered payload must fail verification");
            } catch (IllegalStateException expected) {
                assertTrue(expected.getMessage().contains("digest mismatch"));
            }
        } finally {
            deleteRecursively(root);
        }
    }

    @Test
    public void unexpectedPayloadFileFailsManifestVerification() throws Exception {
        File root = Files.createTempDirectory("zillroot-extra").toFile();
        try {
            writePayloadAndManifest(root);
            write(root, "translations/korean/messages/stale.toml", "stale\n");
            try {
                ProjectAssetIntegrity.verifyPayload(root);
                fail("unexpected payload member must fail verification");
            } catch (IllegalStateException expected) {
                assertTrue(expected.getMessage().contains("file set mismatch"));
            }
        } finally {
            deleteRecursively(root);
        }
    }

    private static void writePayloadAndManifest(File root) throws Exception {
        List<String> manifest = new ArrayList<>();
        for (String path : PAYLOAD_PATHS) {
            write(root, path, "content-for-" + path + "\n");
            manifest.add(sha256(new File(root, path)) + "  ./" + path);
        }
        write(root, ProjectAssetIntegrity.MANIFEST_RELATIVE_PATH, String.join("\n", manifest) + "\n");
    }

    private static String sha256(File file) throws Exception {
        byte[] data = Files.readAllBytes(file.toPath());
        byte[] digest = MessageDigest.getInstance("SHA-256").digest(data);
        StringBuilder out = new StringBuilder();
        for (byte b : digest) out.append(String.format("%02x", b & 0xff));
        return out.toString();
    }

    private static void write(File root, String relativePath, String content) throws Exception {
        File file = new File(root, relativePath);
        File parent = file.getParentFile();
        if (parent != null && !parent.isDirectory() && !parent.mkdirs()) {
            throw new IllegalStateException("mkdir failed: " + parent);
        }
        try (FileOutputStream out = new FileOutputStream(file)) {
            out.write(content.getBytes(StandardCharsets.UTF_8));
        }
    }

    private static void deleteRecursively(File file) {
        if (file == null || !file.exists()) return;
        if (file.isDirectory()) {
            File[] children = file.listFiles();
            if (children != null) {
                for (File child : children) deleteRecursively(child);
            }
        }
        file.delete();
    }
}
