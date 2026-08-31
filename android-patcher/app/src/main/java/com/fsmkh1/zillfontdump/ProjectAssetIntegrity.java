package com.fsmkh1.zillfontdump;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileReader;
import java.security.MessageDigest;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;

final class ProjectAssetIntegrity {
    static final String MANIFEST_RELATIVE_PATH = "payload-manifest.sha256";

    private static final String[] REQUIRED_RELATIVE_PATHS = {
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
            "payload-version.txt",
            MANIFEST_RELATIVE_PATH
    };

    private ProjectAssetIntegrity() {
    }

    static boolean isComplete(File root) {
        return missingFiles(root).isEmpty();
    }

    static List<String> missingFiles(File root) {
        if (root == null || !root.isDirectory()) {
            return Collections.singletonList("<project-root>");
        }
        List<String> missing = new ArrayList<>();
        for (String relativePath : REQUIRED_RELATIVE_PATHS) {
            if (!new File(root, relativePath).isFile()) {
                missing.add(relativePath);
            }
        }
        return missing;
    }

    static void verifyPayload(File root) throws Exception {
        List<String> missing = missingFiles(root);
        if (!missing.isEmpty()) {
            throw new IllegalStateException("embedded project payload is incomplete: " + missing);
        }

        File manifestFile = new File(root, MANIFEST_RELATIVE_PATH);
        Map<String, String> expected = readManifest(manifestFile);
        if (expected.isEmpty()) {
            throw new IllegalStateException("embedded project payload manifest is empty");
        }

        Set<String> required = new HashSet<>();
        for (String relativePath : REQUIRED_RELATIVE_PATHS) {
            if (!MANIFEST_RELATIVE_PATH.equals(relativePath)) {
                required.add(relativePath);
            }
        }
        if (!expected.keySet().containsAll(required)) {
            Set<String> absent = new HashSet<>(required);
            absent.removeAll(expected.keySet());
            throw new IllegalStateException("payload manifest omits required files: " + absent);
        }

        for (Map.Entry<String, String> entry : expected.entrySet()) {
            File file = new File(root, entry.getKey());
            if (!file.isFile()) {
                throw new IllegalStateException("payload manifest member is missing: " + entry.getKey());
            }
            String actual = sha256(file);
            if (!entry.getValue().equals(actual)) {
                throw new IllegalStateException("payload manifest digest mismatch: " + entry.getKey());
            }
        }

        Set<String> actualFiles = new HashSet<>();
        collectFiles(root, root, actualFiles);
        Set<String> permitted = new HashSet<>(expected.keySet());
        permitted.add(MANIFEST_RELATIVE_PATH);
        if (!actualFiles.equals(permitted)) {
            Set<String> unexpected = new HashSet<>(actualFiles);
            unexpected.removeAll(permitted);
            Set<String> absent = new HashSet<>(permitted);
            absent.removeAll(actualFiles);
            throw new IllegalStateException("payload file set mismatch: unexpected=" + unexpected + " missing=" + absent);
        }
    }

    private static Map<String, String> readManifest(File manifestFile) throws Exception {
        Map<String, String> out = new HashMap<>();
        try (BufferedReader reader = new BufferedReader(new FileReader(manifestFile))) {
            String line;
            int lineNumber = 0;
            while ((line = reader.readLine()) != null) {
                lineNumber++;
                if (line.trim().isEmpty()) continue;
                if (line.length() < 67) {
                    throw new IllegalStateException("invalid payload manifest line " + lineNumber);
                }
                String digest = line.substring(0, 64).toLowerCase();
                if (!digest.matches("[0-9a-f]{64}")) {
                    throw new IllegalStateException("invalid payload manifest digest at line " + lineNumber);
                }
                String path = line.substring(64).trim();
                if (path.startsWith("*")) path = path.substring(1);
                if (path.startsWith("./")) path = path.substring(2);
                path = path.replace('\\', '/');
                if (path.isEmpty() || path.startsWith("/") || path.contains("../") || path.equals("..") || MANIFEST_RELATIVE_PATH.equals(path)) {
                    throw new IllegalStateException("invalid payload manifest path at line " + lineNumber + ": " + path);
                }
                if (out.put(path, digest) != null) {
                    throw new IllegalStateException("duplicate payload manifest path: " + path);
                }
            }
        }
        return out;
    }

    private static String sha256(File file) throws Exception {
        MessageDigest digest = MessageDigest.getInstance("SHA-256");
        byte[] buffer = new byte[64 * 1024];
        try (FileInputStream in = new FileInputStream(file)) {
            int read;
            while ((read = in.read(buffer)) >= 0) {
                if (read == 0) continue;
                digest.update(buffer, 0, read);
            }
        }
        StringBuilder out = new StringBuilder(64);
        for (byte b : digest.digest()) {
            out.append(String.format("%02x", b & 0xff));
        }
        return out.toString();
    }

    private static void collectFiles(File root, File current, Set<String> out) throws Exception {
        File[] children = current.listFiles();
        if (children == null) {
            throw new IllegalStateException("cannot list payload directory: " + current);
        }
        for (File child : children) {
            if (child.isDirectory()) {
                collectFiles(root, child, out);
            } else if (child.isFile()) {
                String relative = root.toURI().relativize(child.toURI()).getPath();
                if (relative.isEmpty()) {
                    throw new IllegalStateException("cannot relativize payload file: " + child);
                }
                out.add(relative);
            } else {
                throw new IllegalStateException("payload contains non-file entry: " + child);
            }
        }
    }
}
