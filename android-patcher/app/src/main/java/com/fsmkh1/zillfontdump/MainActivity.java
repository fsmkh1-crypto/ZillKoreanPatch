package com.fsmkh1.zillfontdump;

import android.app.Activity;
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
    private Button patchButton;
    private Uri sourceUri;
    private FontExtractor.Inspection inspection;

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
        info.setText("대상: 일본판 ULJM-05410 v1.03\n검수된 한국어 정본 42,016건과 현재 한글 폰트/실행파일 패치를 사용합니다.\n원본 ISO는 읽기 전용으로만 사용하며 새 ISO를 별도로 생성합니다.\n작업 중 내부 임시 추출과 ISO 재생성이 필요하므로 여유 공간 3GB 이상을 권장합니다.\n첫 실행할 때는 내장 한국어 데이터 파일을 앱 내부 작업공간에 준비합니다.");
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

        patchButton = new Button(this);
        patchButton.setText("한국어 BETA ISO 만들기");
        patchButton.setEnabled(false);
        patchButton.setOnClickListener(v -> choosePatchedIsoDestination());
        LinearLayout.LayoutParams patchParams = new LinearLayout.LayoutParams(-1, -2);
        patchParams.topMargin = pad / 2;
        root.addView(patchButton, patchParams);

        status = new TextView(this);
        status.setText("대기 중");
        status.setTextSize(15);
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
                        "검증 완료: " + checked.discId + " v" + checked.version + ". 아래 버튼으로 Beta ISO를 만드세요."));
            } catch (Exception e) {
                inspection = null;
                runOnUiThread(() -> setBusy(false, "검증 실패: " + message(e)));
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

        setBusy(true, "한국어 Beta 빌드 준비 중…");
        Uri inputUri = sourceUri;
        FontExtractor.Inspection checked = inspection;
        worker.execute(() -> {
            boolean success = false;
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
                        "--version", "mobile-beta-0.9.4");
                builder.directory(rootDir);
                builder.redirectErrorStream(true);
                Process process = builder.start();
                Deque<String> tail = new ArrayDeque<>();
                try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
                    String line;
                    while ((line = reader.readLine()) != null) {
                        if (tail.size() == 12) tail.removeFirst();
                        tail.addLast(line);
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
                runOnUiThread(() -> setBusy(false,
                        "완료. 생성된 Korean Beta ISO를 PPSSPP에서 실행해 주세요. 실제 화면의 줄바꿈·잘림·폰트는 이번 베타에서 확인합니다."));
            } catch (Exception e) {
                final String error = message(e);
                runOnUiThread(() -> setBusy(false, "패치 실패: " + error));
            } finally {
                deleteRecursively(session);
                if (!success) deleteQuietly(outputUri);
            }
        });
    }

    private File ensureProjectAssets() throws Exception {
        File root = new File(getFilesDir(), "zillroot-beta-v3");
        File marker = new File(root, ".ready");
        if (marker.isFile()) return root;
        deleteRecursively(root);
        if (!root.mkdirs()) throw new IllegalStateException("내장 한글 데이터 폴더를 만들 수 없습니다.");
        copyAssetTree("zillroot", root);
        if (!marker.createNewFile()) throw new IllegalStateException("내장 한글 데이터 준비를 완료할 수 없습니다.");
        return root;
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