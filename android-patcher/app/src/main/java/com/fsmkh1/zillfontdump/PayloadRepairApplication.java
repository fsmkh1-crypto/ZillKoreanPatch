package com.fsmkh1.zillfontdump;

import android.app.Application;

import java.io.File;

public final class PayloadRepairApplication extends Application {
    @Override
    public void onCreate() {
        super.onCreate();
        File root = new File(getFilesDir(), "zillroot-beta-current");
        if (root.exists() && !ProjectAssetIntegrity.isComplete(root)) {
            deleteRecursively(root);
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
        //noinspection ResultOfMethodCallIgnored
        file.delete();
    }
}
