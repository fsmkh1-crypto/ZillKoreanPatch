package com.fsmkh1.zillfontdump;

import android.app.Activity;
import android.content.ClipData;
import android.content.ClipboardManager;
import android.content.Context;
import android.content.Intent;
import android.os.Bundle;
import android.view.Gravity;
import android.widget.Button;
import android.widget.EditText;
import android.widget.LinearLayout;
import android.widget.TextView;

/** UI for explicit user-triggered PPSSPP freeze-state capture. */
public final class FreezeCaptureActivity extends Activity {
    private EditText portInput;
    private TextView status;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        int pad = (int) (24 * getResources().getDisplayMetrics().density);
        LinearLayout root = new LinearLayout(this);
        root.setOrientation(LinearLayout.VERTICAL);
        root.setGravity(Gravity.CENTER_HORIZONTAL);
        root.setPadding(pad, pad, pad, pad);

        TextView title = new TextView(this);
        title.setText("질올 PPSSPP 수동 프리징 캡처");
        title.setTextSize(22);
        root.addView(title, new LinearLayout.LayoutParams(-1, -2));

        TextView info = new TextView(this);
        info.setText("자동 감지나 breakpoint를 사용하지 않습니다.\n\n"
                + "1) PPSSPP 원격 디버거를 켜고 포트를 34500으로 맞춥니다.\n"
                + "2) 이 앱은 건드리지 말고 PPSSPP에서 평소처럼 프리징을 재현합니다.\n"
                + "3) 화면이 실제로 멈춘 뒤 이 앱으로 돌아옵니다.\n"
                + "4) 아래 '지금 상태 수동 캡처'를 딱 한 번 누릅니다.\n"
                + "5) 완료 알림이 뜨면 '최근 로그 복사'를 눌러 채팅에 붙입니다.\n\n"
                + "캡처 버튼을 누른 순간의 PC/GPR과 s0+0x2C0 메모리 0x120바이트를 그대로 기록합니다.");
        info.setTextSize(15);
        LinearLayout.LayoutParams infoParams = new LinearLayout.LayoutParams(-1, -2);
        infoParams.topMargin = pad / 2;
        root.addView(info, infoParams);

        portInput = new EditText(this);
        portInput.setHint("PPSSPP 로컬 서버 포트");
        portInput.setInputType(android.text.InputType.TYPE_CLASS_NUMBER);
        portInput.setText("34500");
        LinearLayout.LayoutParams portParams = new LinearLayout.LayoutParams(-1, -2);
        portParams.topMargin = pad;
        root.addView(portInput, portParams);

        Button captureButton = new Button(this);
        captureButton.setText("지금 상태 수동 캡처");
        captureButton.setOnClickListener(v -> captureNow());
        LinearLayout.LayoutParams captureParams = new LinearLayout.LayoutParams(-1, -2);
        captureParams.topMargin = pad / 2;
        root.addView(captureButton, captureParams);

        Button copyButton = new Button(this);
        copyButton.setText("최근 로그 복사");
        copyButton.setOnClickListener(v -> copyTrace());
        root.addView(copyButton, new LinearLayout.LayoutParams(-1, -2));

        status = new TextView(this);
        status.setText("PPSSPP에서 프리징을 먼저 재현한 뒤 캡처 버튼을 누르세요.");
        status.setTextSize(14);
        status.setTextIsSelectable(true);
        LinearLayout.LayoutParams statusParams = new LinearLayout.LayoutParams(-1, -2);
        statusParams.topMargin = pad;
        root.addView(status, statusParams);

        setContentView(root);
    }

    private int readPort() {
        try {
            int port = Integer.parseInt(portInput.getText().toString().trim());
            if (port < 1 || port > 65535) throw new NumberFormatException();
            return port;
        } catch (NumberFormatException e) {
            status.setText("포트는 1~65535 숫자여야 합니다.");
            return -1;
        }
    }

    private void captureNow() {
        int port = readPort();
        if (port < 0) return;
        Intent intent = new Intent(this, FreezeTraceService.class);
        intent.setAction(FreezeTraceService.ACTION_CAPTURE_NOW);
        intent.putExtra(FreezeTraceService.EXTRA_PORT, port);
        startForegroundService(intent);
        status.setText("현재 프리징 상태 캡처를 요청했습니다. 완료 알림 후 로그를 복사하세요.");
    }

    private void copyTrace() {
        try {
            String trace = FreezeTraceService.readLatestTrace(getFilesDir());
            if (trace.trim().isEmpty()) {
                status.setText("저장된 로그가 없습니다. 프리징 재현 후 '지금 상태 수동 캡처'를 누르세요.");
                return;
            }
            ClipboardManager clipboard = (ClipboardManager) getSystemService(Context.CLIPBOARD_SERVICE);
            if (clipboard == null) {
                status.setText("클립보드 서비스를 사용할 수 없습니다.");
                return;
            }
            clipboard.setPrimaryClip(ClipData.newPlainText("Zill PPSSPP manual freeze capture", trace));
            status.setText("수동 캡처 로그를 복사했습니다. 그대로 채팅에 붙여 주세요.");
        } catch (Exception e) {
            status.setText("로그 읽기 실패: " + e.getMessage());
        }
    }
}
