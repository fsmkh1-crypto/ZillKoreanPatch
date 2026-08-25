package com.fsmkh1.zillfontdump;

import android.app.Activity;
import android.content.Intent;
import android.database.Cursor;
import android.net.Uri;
import android.os.Bundle;
import android.os.ParcelFileDescriptor;
import android.provider.DocumentsContract;
import android.provider.OpenableColumns;
import android.view.Gravity;
import android.widget.Button;
import android.widget.LinearLayout;
import android.widget.TextView;

import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.nio.channels.FileChannel;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

public final class MainActivity extends Activity {
    private static final int PICK_ISO = 1001;
    private static final int CREATE_AUDIT_ZIP = 1002;
    private static final int CREATE_PATCHED_ISO = 1003;

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
        title.setText("Zill Infinite Plus – Korean Patch Audit");
        title.setTextSize(22);
        root.addView(title, new LinearLayout.LayoutParams(-1, -2));

        TextView info = new TextView(this);
        info.setText("대상: ULJM-05410 v1.03\n원본 ISO는 절대 수정하지 않습니다.\nISO 선택 후 폰트 + EBOOT + bindata 감사 자료 ZIP을 먼저 저장합니다.\n이 ZIP만 전달하면 재사용 가능한 한글 renderer slot 후보를 좁힐 수 있습니다.\n기존 8글자 smoke ISO 생성 기능도 아래 버튼으로 남겨둡니다.");
        info.setTextSize(15);
        LinearLayout.LayoutParams infoParams = new LinearLayout.LayoutParams(-1, -2);
        infoParams.topMargin = pad / 2;
        root.addView(info, infoParams);

        chooseButton = new Button(this);
        chooseButton.setText("원본 ISO 선택 → 감사 ZIP 추출");
        chooseButton.setOnClickListener(v -> pickIso());
        LinearLayout.LayoutParams buttonParams = new LinearLayout.LayoutParams(-1, -2);
        buttonParams.topMargin = pad;
        root.addView(chooseButton, buttonParams);

        patchButton = new Button(this);
        patchButton.setText("기존 8글자 smoke ISO 만들기");
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
        } else if (requestCode == CREATE_AUDIT_ZIP) {
            exportAuditZip(data.getData());
        } else if (requestCode == CREATE_PATCHED_ISO) {
            createPatchedIso(data.getData());
        }
    }

    private void inspectSelectedIso() {
        inspection = null;
        setBusy(true, "원본 ISO, EBOOT, 폰트, bindata 해시 확인 중…");
        Uri uri = sourceUri;
        worker.execute(() -> {
            try (ParcelFileDescriptor pfd = getContentResolver().openFileDescriptor(uri, "r");
                 FileInputStream in = pfd == null ? null : new FileInputStream(pfd.getFileDescriptor());
                 FileChannel channel = in == null ? null : in.getChannel()) {
                if (channel == null) throw new IllegalStateException("ISO를 읽기 전용으로 열 수 없습니다.");
                FontExtractor.Inspection checked = FontExtractor.inspect(channel);
                inspection = checked;
                runOnUiThread(() -> {
                    setBusy(false, "원본 검증 완료. 감사 자료 ZIP 저장 위치를 선택하세요.");
                    Intent out = new Intent(Intent.ACTION_CREATE_DOCUMENT);
                    out.addCategory(Intent.CATEGORY_OPENABLE);
                    out.setType("application/zip");
                    out.putExtra(Intent.EXTRA_TITLE, "zill-slot-audit-assets-ULJM05410-v103.zip");
                    startActivityForResult(out, CREATE_AUDIT_ZIP);
                });
            } catch (Exception e) {
                inspection = null;
                runOnUiThread(() -> setBusy(false, "검증 실패: " + message(e)));
            }
        });
    }

    private void exportAuditZip(Uri outputUri) {
        if (sourceUri == null || inspection == null) {
            status.setText("먼저 지원되는 원본 ISO를 선택해야 합니다.");
            return;
        }
        if (sourceUri.equals(outputUri)) {
            status.setText("원본 ISO와 같은 문서에는 쓸 수 없습니다.");
            return;
        }
        setBusy(true, "감사 자료 추출 중… ISO는 읽기 전용입니다.");
        Uri inputUri = sourceUri;
        FontExtractor.Inspection checked = inspection;
        String sourceName = displayName(inputUri);
        worker.execute(() -> {
            boolean success = false;
            try (ParcelFileDescriptor inPfd = getContentResolver().openFileDescriptor(inputUri, "r");
                 ParcelFileDescriptor outPfd = getContentResolver().openFileDescriptor(outputUri, "rwt");
                 FileInputStream in = inPfd == null ? null : new FileInputStream(inPfd.getFileDescriptor());
                 FileOutputStream out = outPfd == null ? null : new FileOutputStream(outPfd.getFileDescriptor());
                 FileChannel channel = in == null ? null : in.getChannel()) {
                if (channel == null || out == null) throw new IllegalStateException("입출력 파일을 열 수 없습니다.");
                FontExtractor.Inspection fresh = FontExtractor.inspect(channel);
                if (!fresh.discId.equals(checked.discId) || !fresh.version.equals(checked.version) ||
                        fresh.isoSize != checked.isoSize) {
                    throw new IllegalStateException("선택된 ISO가 검증 후 변경되었습니다.");
                }
                FontExtractor.export(channel, fresh, out, sourceName);
                out.flush();
                success = true;
                runOnUiThread(() -> setBusy(false,
                        "감사 ZIP 추출 완료. 이 ZIP만 GPT에 전달하면 됩니다. 원본 ISO는 변경되지 않았습니다."));
            } catch (Exception e) {
                final String error = message(e);
                runOnUiThread(() -> setBusy(false, "감사 자료 추출 실패: " + error));
            } finally {
                if (!success) deleteQuietly(outputUri);
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
        out.putExtra(Intent.EXTRA_TITLE, "Zill_Oll_Infinite_Plus_Korean_8Glyph_Smoke.iso");
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
        setBusy(true, "새 ISO 생성 중… 원본 전체를 스트리밍 복사합니다.");
        Uri inputUri = sourceUri;
        FontExtractor.Inspection checked = inspection;
        worker.execute(() -> {
            boolean success = false;
            try (ParcelFileDescriptor inPfd = getContentResolver().openFileDescriptor(inputUri, "r");
                 ParcelFileDescriptor outPfd = getContentResolver().openFileDescriptor(outputUri, "rwt");
                 FileInputStream in = inPfd == null ? null : new FileInputStream(inPfd.getFileDescriptor());
                 FileOutputStream out = outPfd == null ? null : new FileOutputStream(outPfd.getFileDescriptor());
                 FileChannel channel = in == null ? null : in.getChannel()) {
                if (channel == null || out == null) throw new IllegalStateException("입출력 파일을 열 수 없습니다.");
                FontExtractor.Inspection fresh = FontExtractor.inspect(channel);
                if (!fresh.discId.equals(checked.discId) || !fresh.version.equals(checked.version) ||
                        fresh.isoSize != checked.isoSize) {
                    throw new IllegalStateException("선택된 ISO가 검증 후 변경되었습니다.");
                }
                channel.position(0);
                PoCPatcher.copyAndPatch(in, out, fresh);
                out.flush();
                if (out.getChannel().size() != fresh.isoSize) {
                    throw new IllegalStateException("결과 ISO 크기가 원본과 다릅니다.");
                }
                success = true;
                runOnUiThread(() -> setBusy(false,
                        "완료. PPSSPP에서 신규 게임을 시작해 가/나/다/라/마/바/사/아 8자를 확인하세요."));
            } catch (Exception e) {
                final String error = message(e);
                runOnUiThread(() -> setBusy(false, "패치 실패: " + error));
            } finally {
                if (!success) deleteQuietly(outputUri);
            }
        });
    }

    private void deleteQuietly(Uri uri) {
        try {
            DocumentsContract.deleteDocument(getContentResolver(), uri);
        } catch (Exception ignored) {
        }
    }

    private void setBusy(boolean busy, String text) {
        chooseButton.setEnabled(!busy);
        patchButton.setEnabled(!busy && inspection != null);
        status.setText(text);
    }

    private String displayName(Uri uri) {
        try (Cursor c = getContentResolver().query(uri,
                new String[]{OpenableColumns.DISPLAY_NAME}, null, null, null)) {
            if (c != null && c.moveToFirst()) {
                String value = c.getString(0);
                if (value != null && !value.trim().isEmpty()) return value;
            }
        } catch (Exception ignored) {
        }
        return "source.iso";
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
