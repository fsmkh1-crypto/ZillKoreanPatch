package com.fsmkh1.zillfontdump;

import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertTrue;

import org.junit.Test;

import java.io.File;
import java.io.FileOutputStream;
import java.nio.file.Files;

public final class ProjectAssetIntegrityTest {
    @Test
    public void markerAloneDoesNotCountAsCompletePayload() throws Exception {
        File root = Files.createTempDirectory("zillroot-partial").toFile();
        try {
            write(root, "payload-version.txt", "same-version\n");
            assertFalse(ProjectAssetIntegrity.isComplete(root));
            assertTrue(ProjectAssetIntegrity.missingFiles(root).contains("release/font/metrics.toml"));
        } finally {
            deleteRecursively(root);
        }
    }

    @Test
    public void requiredPayloadFilesCountAsComplete() throws Exception {
        File root = Files.createTempDirectory("zillroot-complete").toFile();
        try {
            String[] paths = {
                    "release/font/metrics.toml",
                    "release/font/manifest.toml",
                    "release/korean/font/glyphs.toml",
                    "release/korean/strings/eboot.toml",
                    "release/layout/consumer-map.toml",
                    "release/layout/categories.toml",
                    "docs/audit/fixtures/pr14-eboot-full.toml",
                    "patches/executable/manifest.toml",
                    "patches/system/param-sfo.toml",
                    "payload-version.txt"
            };
            for (String path : paths) {
                write(root, path, "x\n");
            }
            assertTrue(ProjectAssetIntegrity.isComplete(root));
        } finally {
            deleteRecursively(root);
        }
    }

    private static void write(File root, String relativePath, String content) throws Exception {
        File file = new File(root, relativePath);
        File parent = file.getParentFile();
        if (parent != null && !parent.isDirectory() && !parent.mkdirs()) {
            throw new IllegalStateException("mkdir failed: " + parent);
        }
        try (FileOutputStream out = new FileOutputStream(file)) {
            out.write(content.getBytes("UTF-8"));
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
