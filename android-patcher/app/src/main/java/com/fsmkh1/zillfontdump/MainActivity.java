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
import android.view.View;
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
    private static final int CREATE_ZIP = 1002;

    private final ExecutorService worker = Executors.newSingleThreadExecutor();
    private TextView status;
    private Button chooseButton;
    private Uri sourceUri;
    private String sourceName = "source.iso";
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
        title.setText("Zill Infinite Plus – Font Extractor");
        title.setTextSize(22);
        root.addView(title, new LinearLayout.LayoutParams(-1, -2));

        TextView info = new TextView(this);
        info.setText("대상: ULJM-05410 v1.03\n원본 ISO는 읽기 전용으로 열며 수정하지 않습니다.\n두 폰트 PAR와 manifest.json만 ZIP으로 내보냅니다.");
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
            sourceName = displayName(sourceUri);
            inspectSelectedIso();
        } else if (requestCode == CREATE_ZIP) {
            exportZip(data.getData());
        }
    }

    private void inspectSelectedIso() {
        setBusy(true, "ISO 확인 중…");
        Uri uri = sourceUri;
        worker.execute(() -> {
            try (ParcelFileDescriptor pfd = getContentResolver().openFileDescriptor(uri, "r");
                 FileInputStream in = pfd == null ? null : new FileInputStream(pfd.getFileDescriptor());
                 FileChannel channel = in == null ? null : in.getChannel()) {
                if (channel == null) throw new IllegalStateException("ISO를 읽기 전용으로 열 수 없습니다.");
                FontExtractor.Inspection checked = FontExtractor.inspect(channel);
                inspection = checked;
                runOnUiThread(() -> {
                    setBusy(false, "확인 완료: ULJM-05410 v" + checked.version + "\nZIP 저장 위치를 선택하세요.");
                    Intent out = new Intent(Intent.ACTION_CREATE_DOCUMENT);
                    out.addCategory(Intent.CATEGORY_OPENABLE);
                    out.setType("application/zip");
                    out.putExtra(Intent.EXTRA_TITLE, "zill-font-resources-ULJM05410-v103.zip");
                    startActivityForResult(out, CREATE_ZIP);
                });
            } catch (Exception e) {
                inspection = null;
                runOnUiThread(() -> setBusy(false, "검증 실패: " + message(e)));
            }
        });
    }

    private void exportZip(Uri outputUri) {
        if (sourceUri == null || inspection == null) {
            status.setText("먼저 지원되는 ISO를 선택해야 합니다.");
            return;
        }
        setBusy(true, "폰트 리소스 추출 중…");
        Uri inputUri = sourceUri;
        FontExtractor.Inspection checked = inspection;
        String inputName = sourceName;
        worker.execute(() -> {
            boolean success = false;
            try (ParcelFileDescriptor inPfd = getContentResolver().openFileDescriptor(inputUri, "r");
                 ParcelFileDescriptor outPfd = getContentResolver().openFileDescriptor(outputUri, "rwt");
                 FileInputStream in = inPfd == null ? null : new FileInputStream(inPfd.getFileDescriptor());
                 FileOutputStream out = outPfd == null ? null : new FileOutputStream(outPfd.getFileDescriptor());
                 FileChannel channel = in == null ? null : in.getChannel()) {
                if (channel == null || out == null) throw new IllegalStateException("입출력 파일을 열 수 없습니다.");
                // Revalidate the source on the second open before fixed-offset extraction.
                FontExtractor.Inspection fresh = FontExtractor.inspect(channel);
                if (!fresh.discId.equals(checked.discId) || !fresh.version.equals(checked.version) ||
                        fresh.isoSize != checked.isoSize) {
                    throw new IllegalStateException("선택된 ISO가 검증 후 변경되었습니다.");
                }
                FontExtractor.export(channel, fresh, out, inputName);
                out.flush();
                success = true;
                runOnUiThread(() -> setBusy(false,
                        "완료. ZIP에는 font/zillfont.par, 2d/font/jillbtn.par, manifest.json만 포함됩니다."));
            } catch (Exception e) {
                final String error = message(e);
                runOnUiThread(() -> setBusy(false, "추출 실패: " + error));
            } finally {
                if (!success) {
                    try {
                        DocumentsContract.deleteDocument(getContentResolver(), outputUri);
                    } catch (Exception ignored) {
                        // Some document providers do not permit deletion. The UI already marks the output as failed.
                    }
                }
            }
        });
    }

    private void setBusy(boolean busy, String text) {
        chooseButton.setEnabled(!busy);
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
