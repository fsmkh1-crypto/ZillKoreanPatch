package com.fsmkh1.zillfontdump;

import java.io.File;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

final class ProjectAssetIntegrity {
    private static final String[] REQUIRED_RELATIVE_PATHS = {
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
}
