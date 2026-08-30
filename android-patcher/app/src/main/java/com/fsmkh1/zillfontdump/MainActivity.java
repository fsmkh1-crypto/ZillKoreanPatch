package com.fsmkh1.zillfontdump;

import android.app.Activity;
import android.content.ClipData;
import android.content.ClipboardManager;
import android.content.Context;
import android.content.Intent;
import android.net.Uri;
import android.os.Bundle;
import android.os.ParcelFileDescriptor;
import android.provider.DocumentsContract;
import android.view.Gravity;
import android.widget.Button;
import android.widget.LinearLayout;
import android.widget.TextView;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.io.OutputStream;
import java.nio.channels.FileChannel;
import java.util.ArrayDeque;
import java.util.Deque;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

public final class MainActivity extends Activity {
    private static final int PICK_ISO = 1001;
    private static final int CREATE_PATCHED_ISO = 1002;

    private final ExecutorService worker = Executors.newSingleThreadExecutor();
    private TextView status;
    private Button chooseButton;
    private Button preflightButton;
    private Button patchButton;
    private Button copyLogButton;
    private Uri sourceUri;
    private FontExtractor.Inspection inspection;
    private String lastForensicLog = "";

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        int pad = (int) (24 * getResources().getDisplayMetrics().density);
        LinearLayout root = new LinearLayout(this);
        root.setOrientation(LinearLayout.VERTICAL);
        root.setGravity(Gravity.CENTER_HORIZONTAL);
        root.setPadding(pad, pad, pad, pad);

        TextView title = new TextView(this);
        title.setText("질올 인피니트 플러스 한국어 패치 Beta");
        title.setTextSize(22);
        root.addView(title, new LinearLayout.LayoutParams(-1, -2));

        TextView info = new TextView(this);
        info.setText("대상: 일본판 ULJM-05410 v1.03\n검수된 한국어 정본과 현재 한글 폰트/실행파일 패치를 사용합니다.\n'RETAIL 진단만 실행'은 결과 ISO를 만들지 않고 인증·은행 바인딩·슬롯/충돌·C5/PR14·폰트·실행파일 정적 검증까지만 수행합니다.\n실제 패치는 원본 ISO를 읽기 전용으로만 사용하며 새 ISO를 별도로 생성합니다.\n앱 업데이트 시 내장 한국어 데이터 버전을 확인하여 변경된 데이터는 자동으로 다시 준비합니다.");
        info.setTextSize(15);
        LinearLayout.LayoutParams infoParams = new LinearLayout.LayoutParams(-1, -2);
        infoParams.topMargin = pad / 2;
        root.addView(info, infoParams);

        chooseButton = new Button(this);
        chooseButton.setText("원본 ISO 선택");
        chooseButton.setOnClickListener(v -> pickIso());
        LinearLayout.LayoutParams buttonParams = new LinearLayout.LayoutParams(-1, -2);
        buttonParams.topMargin = pad;
        root.addView(chooseButton, buttonParams);

        preflightButton = new Button(this);
        preflightButton.setText("RETAIL 진단만 실행");
        preflightButton.setEnabled(false);
        preflightButton.setOnClickListener(v -> runForensicPreflight());
        LinearLayout.LayoutParams preflightParams = new LinearLayout.LayoutParams(-1, -2);
        preflightParams.topMargin = pad / 2;
        root.addView(preflightButton, preflightParams);

        patchButton = new Button(this);
        patchButton.setText("한국어 BETA ISO 만들기");
        patchButton.setEnabled(false);
        patchButton.setOnClickListener(v -> choosePatchedIsoDestination());
        LinearLayout.LayoutParams patchParams = new LinearLayout.LayoutParams(-1, -2);
        patchParams.topMargin = pad / 2;
        root.addView(patchButton, patchParams);

        copyLogButton = new Button(this);
        copyLogButton.setText("진단 로그 복사");
        copyLogButton.setEnabled(false);
        copyLogButton.setOnClickListener(v -> copyForensicLog());
        LinearLayout.LayoutParams copyLogParams = new LinearLayout.LayoutParams(-1, -2);
        copyLogParams.topMargin = pad / 2;
        root.addView(copyLogButton, copyLogParams);

        status = new TextView(this);
        status.setText("대기 중");
        status.setTextSize(15);
        status.setTextIsSelectable(true);
        LinearLayout.LayoutParams statusParams = new LinearLayout.LayoutParams(-1, -2);
        statusParams.topMargin = pad;
        root.addView(status, statusParams);

        setContentView(root);
    }

    private void pickIso() {
        Intent i = new Intent(Intent.ACTION_OPEN_DOCUMENT);
        i.addCategory(Intent.CATEGORY_OPENABLE);
        i.setType("*/*");
        startActivityForResult(i, PICK_ISO);
    }

    @Override
    protected void onActivityResult(int requestCode, int resultCode, Intent data) {
        super.onActivityResult(requestCode, resultCode, data);
        if (resultCode != RESULT_OK || data == null || data.getData() == null) return;
        if (requestCode == PICK_ISO) {
            sourceUri = data.getData();
            inspectSelectedIso();
        } else if (requestCode == CREATE_PATCHED_ISO) {
            createPatchedIso(data.getData());
        }
    }

    private void inspectSelectedIso() {
        inspection = null;
        setBusy(true, "원본 ISO 검증 중…");
        Uri uri = sourceUri;
        worker.execute(() -> {
            try (ParcelFileDescriptor pfd = getContentResolver().openFileDescriptor(uri, "r");
                 FileInputStream in = pfd == null ? null : new FileInputStream(pfd.getFileDescriptor());
                 FileChannel channel = in == null ? null : in.getChannel()) {
                if (channel == null) throw new IllegalStateException("ISO를 읽기 전용으로 열 수 없습니다.");
                FontExtractor.Inspection checked = FontExtractor.inspect(channel);
                inspection = checked;
                runOnUiThread(() -> setBusy(false,
                        "검증 완료: " + checked.discId + " v" + checked.version + ". 먼저 RETAIL 진단만 실행할 수 있습니다."));
            } catch (Exception e) {
                inspection = null;
                runOnUiThread(() -> setBusy(false, "검증 실패: " + message(e)));
            }
        });
    }

    private void runForensicPreflight() {
        if (sourceUri == null || inspection == null) {
            status.setText("먼저 지원되는 원본 ISO를 선택하고 검증해야 합니다.");
            return;
        }

        lastForensicLog = "";
        copyLogButton.setEnabled(false);
        setBusy(true, "RETAIL asset-backed 진단 준비 중…");
        Uri inputUri = sourceUri;
        FontExtractor.Inspection checked = inspection;
        worker.execute(() -> {
            StringBuilder forensic = new StringBuilder();
            File session = new File(getCacheDir(), "korean-preflight-" + System.nanoTime());
            try {
                if (!session.mkdirs()) throw new IllegalStateException("임시 작업 폴더를 만들 수 없습니다.");
                File rootDir = ensureProjectAssets();
                File source = new File(session, "source.iso");
                File work = new File(session, "work");

                updateStatus("원본 ISO를 진단 작업공간으로 복사 중…");
                copyUriToFile(inputUri, source);
                if (source.length() != checked.isoSize) {
                    throw new IllegalStateException("복사된 ISO 크기가 검증 값과 다릅니다.");
                }

                File executable = new File(getApplicationInfo().nativeLibraryDir, "libzill.so");
                if (!executable.isFile()) {
                    throw new IllegalStateException("내장 Korean builder를 찾을 수 없습니다: " + executable);
                }

                updateStatus("인증된 retail 정적 preflight 실행 중…");
                ProcessBuilder builder = new ProcessBuilder(
                        executable.getAbsolutePath(),
                        "build-korean-iso",
                        "--iso", source.getAbsolutePath(),
                        "--work-dir", work.getAbsolutePath(),
                        "--version", "mobile-beta-0.9.8",
                        "--preflight-only");
                builder.directory(rootDir);
                builder.redirectErrorStream(true);
                Process process = builder.start();
                Deque<String> tail = new ArrayDeque<>();
                try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
                    String line;
                    while ((line = reader.readLine()) != null) {
                        if (tail.size() == 12) tail.removeFirst();
                        tail.addLast(line);
                        if (line.startsWith("FORENSIC")) {
                            forensic.append(line).append('\n');
                        }
                        updateStatus(line);
                    }
                }
                int exit = process.waitFor();
                if (exit != 0) {
                    throw new IllegalStateException("Retail preflight 실패(" + exit + "): " + String.join(" | ", tail));
                }

                final String captured = forensic.toString().trim();
                runOnUiThread(() -> {
                    lastForensicLog = captured;
                    copyLogButton.setEnabled(!captured.isEmpty());
                    String logStatus = captured.isEmpty()
                            ? " 진단 로그는 생성되지 않았습니다."
                            : " '진단 로그 복사' 버튼으로 결과를 복사할 수 있습니다.";
                    setBusy(false, "RETAIL asset-backed preflight 완료. 결과 ISO는 생성하지 않았습니다." + logStatus);
                });
            } catch (Exception e) {
                final String error = message(e);
                final String captured = forensic.toString().trim();
                runOnUiThread(() -> {
                    lastForensicLog = captured;
                    copyLogButton.setEnabled(!captured.isEmpty());
                    String logStatus = captured.isEmpty() ? "" : " 생성된 진단 로그는 복사할 수 있습니다.";
                    setBusy(false, "RETAIL 진단 실패: " + error + logStatus);
                });
            } finally {
                deleteRecursively(session);
            }
        });
    }

    private void choosePatchedIsoDestination() {
        if (sourceUri == null || inspection == null) {
            status.setText("먼저 지원되는 원본 ISO를 선택하고 검증해야 합니다.");
            return;
        }
        Intent out = new Intent(Intent.ACTION_CREATE_DOCUMENT);
        out.addCategory(Intent.CATEGORY_OPENABLE);
        out.setType("application/octet-stream");
        out.putExtra(Intent.EXTRA_TITLE, "Zill_Oll_Infinite_Plus_Korean_Beta.iso");
        startActivityForResult(out, CREATE_PATCHED_ISO);
    }

    private void createPatchedIso(Uri outputUri) {
        if (sourceUri == null || inspection == null) {
            status.setText("먼저 지원되는 원본 ISO를 선택해야 합니다.");
            return;
        }
        if (sourceUri.equals(outputUri)) {
            status.setText("원본 ISO와 같은 문서에는 쓸 수 없습니다.");
            return;
        }

        lastForensicLog = "";
        copyLogButton.setEnabled(false);
        setBusy(true, "한국어 Beta 빌드 준비 중…");
        Uri inputUri = sourceUri;
        FontExtractor.Inspection checked = inspection;
        worker.execute(() -> {
            boolean success = false;
            StringBuilder forensic = new StringBuilder();
            File session = new File(getCacheDir(), "korean-beta-" + System.nanoTime());
            try {
                if (!session.mkdirs()) throw new IllegalStateException("임시 작업 폴더를 만들 수 없습니다.");
                File rootDir = ensureProjectAssets();
                File source = new File(session, "source.iso");
                File output = new File(session, "korean-beta.iso");
                File work = new File(session, "work");

                updateStatus("원본 ISO를 앱 작업공간으로 복사 중…");
                copyUriToFile(inputUri, source);
                if (source.length() != checked.isoSize) {
                    throw new IllegalStateException("복사된 ISO 크기가 검증 값과 다릅니다.");
                }

                File executable = new File(getApplicationInfo().nativeLibraryDir, "libzill.so");
                if (!executable.isFile()) {
                    throw new IllegalStateException("내장 Korean builder를 찾을 수 없습니다: " + executable);
                }

                updateStatus("게임 데이터 추출 및 한글 메시지/폰트 빌드 시작…");
                ProcessBuilder builder = new ProcessBuilder(
                        executable.getAbsolutePath(),
                        "build-korean-iso",
                        "--iso", source.getAbsolutePath(),
                        "--out", output.getAbsolutePath(),
                        "--work-dir", work.getAbsolutePath(),
                        "--version", "mobile-beta-0.9.8");
                builder.directory(rootDir);
                builder.redirectErrorStream(true);
                Process process = builder.start();
                Deque<String> tail = new ArrayDeque<>();
                try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
                    String line;
                    while ((line = reader.readLine()) != null) {
                        if (tail.size() == 12) tail.removeFirst();
                        tail.addLast(line);
                        if (line.startsWith("FORENSIC")) {
                            forensic.append(line).append('\n');
                        }
                        updateStatus(line);
                    }
                }
                int exit = process.waitFor();
                if (exit != 0) {
                    throw new IllegalStateException("Korean builder 실패(" + exit + "): " + String.join(" | ", tail));
                }
                if (!output.isFile() || output.length() == 0) {
                    throw new IllegalStateException("Korean builder가 결과 ISO를 만들지 못했습니다.");
                }

                updateStatus("완성 ISO를 선택한 위치로 저장 중…");
                copyFileToUri(output, outputUri);
                success = true;
                final String captured = forensic.toString().trim();
                runOnUiThread(() -> {
                    lastForensicLog = captured;
                    copyLogButton.setEnabled(!captured.isEmpty());
                    String logStatus = captured.isEmpty()
                            ? " 진단 로그는 생성되지 않았습니다."
                            : " '진단 로그 복사' 버튼으로 retail preflight 결과를 복사할 수 있습니다.";
                    setBusy(false, "완료. 생성된 Korean Beta ISO를 저장했습니다." + logStatus);
                });
            } catch (Exception e) {
                final String error = message(e);
                final String captured = forensic.toString().trim();
                runOnUiThread(() -> {
                    lastForensicLog = captured;
                    copyLogButton.setEnabled(!captured.isEmpty());
                    String logStatus = captured.isEmpty() ? "" : " 생성된 진단 로그는 복사할 수 있습니다.";
                    setBusy(false, "패치 실패: " + error + logStatus);
                });
            } finally {
                deleteRecursively(session);
                if (!success) deleteQuietly(outputUri);
            }
        });
    }

    private File ensureProjectAssets() throws Exception {
        final String assetMarkerPath = "zillroot/payload-version.txt";
        String packagedVersion = readAssetText(assetMarkerPath).trim();
        if (packagedVersion.isEmpty()) {
            throw new IllegalStateException("내장 데이터 버전 표식이 비어 있습니다.");
        }

        File root = new File(getFilesDir(), "zillroot-beta-current");
        File marker = new File(root, "payload-version.txt");
        if (marker.isFile()) {
            String installedVersion = readFileText(marker).trim();
            if (packagedVersion.equals(installedVersion)) {
                return root;
            }
        }

        deleteRecursively(root);
        if (!root.mkdirs()) throw new IllegalStateException("내장 한글 데이터 폴더를 만들 수 없습니다.");
        copyAssetTree("zillroot", root);
        if (!marker.isFile()) {
            throw new IllegalStateException("내장 데이터 버전 표식이 복사되지 않았습니다.");
        }
        String copiedVersion = readFileText(marker).trim();
        if (!packagedVersion.equals(copiedVersion)) {
            throw new IllegalStateException("내장 데이터 버전이 APK payload와 일치하지 않습니다.");
        }
        return root;
    }

    private String readAssetText(String assetPath) throws Exception {
        try (InputStream in = getAssets().open(assetPath);
             BufferedReader reader = new BufferedReader(new InputStreamReader(in))) {
            StringBuilder out = new StringBuilder();
            String line;
            while ((line = reader.readLine()) != null) {
                out.append(line).append('\n');
            }
            return out.toString();
        }
    }

    private static String readFileText(File file) throws Exception {
        try (InputStream in = new FileInputStream(file);
             BufferedReader reader = new BufferedReader(new InputStreamReader(in))) {
            StringBuilder out = new StringBuilder();
            String line;
            while ((line = reader.readLine()) != null) {
                out.append(line).append('\n');
            }
            return out.toString();
        }
    }

    private void copyAssetTree(String assetPath, File destination) throws Exception {
        String[] children = getAssets().list(assetPath);
        if (children == null) throw new IllegalStateException("assets 목록을 읽을 수 없습니다: " + assetPath);
        if (children.length == 0) {
            File parent = destination.getParentFile();
            if (parent != null && !parent.isDirectory() && !parent.mkdirs()) {
                throw new IllegalStateException("assets 출력 폴더 생성 실패");
            }
            try (InputStream in = getAssets().open(assetPath);
                 OutputStream out = new FileOutputStream(destination)) {
                copy(in, out);
            }
            return;
        }
        if (!destination.isDirectory() && !destination.mkdirs()) {
            throw new IllegalStateException("assets 폴더 생성 실패: " + destination);
        }
        for (String child : children) {
            copyAssetTree(assetPath + "/" + child, new File(destination, child));
        }
    }

    private void copyUriToFile(Uri uri, File destination) throws Exception {
        try (ParcelFileDescriptor pfd = getContentResolver().openFileDescriptor(uri, "r");
             InputStream in = pfd == null ? null : new FileInputStream(pfd.getFileDescriptor());
             OutputStream out = new FileOutputStream(destination)) {
            if (in == null) throw new IllegalStateException("원본 ISO를 열 수 없습니다.");
            copy(in, out);
        }
    }

    private void copyFileToUri(File source, Uri uri) throws Exception {
        try (InputStream in = new FileInputStream(source);
             ParcelFileDescriptor pfd = getContentResolver().openFileDescriptor(uri, "rwt");
             OutputStream out = pfd == null ? null : new FileOutputStream(pfd.getFileDescriptor())) {
            if (out == null) throw new IllegalStateException("결과 ISO 저장 위치를 열 수 없습니다.");
            copy(in, out);
            out.flush();
        }
    }

    private static void copy(InputStream in, OutputStream out) throws Exception {
        byte[] buffer = new byte[1024 * 1024];
        int read;
        while ((read = in.read(buffer)) >= 0) {
            if (read == 0) continue;
            out.write(buffer, 0, read);
        }
    }

    private void copyForensicLog() {
        if (lastForensicLog == null || lastForensicLog.trim().isEmpty()) {
            status.setText("복사할 진단 로그가 없습니다.");
            return;
        }
        ClipboardManager clipboard = (ClipboardManager) getSystemService(Context.CLIPBOARD_SERVICE);
        if (clipboard == null) {
            status.setText("클립보드 서비스를 사용할 수 없습니다.");
            return;
        }
        clipboard.setPrimaryClip(ClipData.newPlainText("Zill Korean forensic log", lastForensicLog));
        status.setText("진단 로그를 클립보드에 복사했습니다.");
    }

    private void updateStatus(String text) {
        runOnUiThread(() -> status.setText(text));
    }

    private void deleteQuietly(Uri uri) {
        try {
            DocumentsContract.deleteDocument(getContentResolver(), uri);
        } catch (Exception ignored) {
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

    private void setBusy(boolean busy, String text) {
        chooseButton.setEnabled(!busy);
        preflightButton.setEnabled(!busy && inspection != null);
        patchButton.setEnabled(!busy && inspection != null);
        status.setText(text);
    }

    private static String message(Exception e) {
        String m = e.getMessage();
        return (m == null || m.trim().isEmpty()) ? e.getClass().getSimpleName() : m;
    }

    @Override
    protected void onDestroy() {
        worker.shutdownNow();
        super.onDestroy();
    }
}
