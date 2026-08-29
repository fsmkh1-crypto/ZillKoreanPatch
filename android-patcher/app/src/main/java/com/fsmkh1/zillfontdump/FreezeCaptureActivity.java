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

/** UI for persistent-connection, explicit PPSSPP freeze-state capture. */
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
                + "2) 게임이 정상 동작 중일 때 '연결 유지 시작'을 누릅니다.\n"
                + "3) 연결 완료 알림을 확인한 뒤 PPSSPP로 돌아가 프리징을 재현합니다.\n"
                + "4) 실제로 프리징되면 이 앱으로 돌아와 '지금 상태 수동 캡처'를 누릅니다.\n"
                + "5) 완료 알림이 뜨면 '최근 로그 복사'를 눌러 채팅에 붙입니다.\n\n"
                + "프리징 뒤 새 handshake를 만들지 않고 이미 열린 연결을 재사용합니다.");
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

        Button connectButton = new Button(this);
        connectButton.setText("연결 유지 시작");
        connectButton.setOnClickListener(v -> connectPersistent());
        LinearLayout.LayoutParams connectParams = new LinearLayout.LayoutParams(-1, -2);
        connectParams.topMargin = pad / 2;
        root.addView(connectButton, connectParams);

        Button captureButton = new Button(this);
        captureButton.setText("지금 상태 수동 캡처");
        captureButton.setOnClickListener(v -> captureNow());
        root.addView(captureButton, new LinearLayout.LayoutParams(-1, -2));

        Button disconnectButton = new Button(this);
        disconnectButton.setText("연결 종료");
        disconnectButton.setOnClickListener(v -> disconnect());
        root.addView(disconnectButton, new LinearLayout.LayoutParams(-1, -2));

        Button copyButton = new Button(this);
        copyButton.setText("최근 로그 복사");
        copyButton.setOnClickListener(v -> copyTrace());
        root.addView(copyButton, new LinearLayout.LayoutParams(-1, -2));

        status = new TextView(this);
        status.setText("정상 상태에서 먼저 '연결 유지 시작'을 누르세요.");
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

    private void connectPersistent() {
        int port = readPort();
        if (port < 0) return;
        Intent intent = new Intent(this, FreezeTraceService.class);
        intent.setAction(FreezeTraceService.ACTION_CONNECT);
        intent.putExtra(FreezeTraceService.EXTRA_PORT, port);
        startForegroundService(intent);
        status.setText("연결을 시작했습니다. 연결 완료 알림 후 PPSSPP에서 프리징을 재현하세요.");
    }

    private void captureNow() {
        Intent intent = new Intent(this, FreezeTraceService.class);
        intent.setAction(FreezeTraceService.ACTION_CAPTURE_NOW);
        startForegroundService(intent);
        status.setText("유지 중인 연결로 현재 상태 캡처를 요청했습니다. 완료 알림 후 로그를 복사하세요.");
    }

    private void disconnect() {
        Intent intent = new Intent(this, FreezeTraceService.class);
        intent.setAction(FreezeTraceService.ACTION_DISCONNECT);
        startService(intent);
        status.setText("debugger 연결 종료를 요청했습니다.");
    }

    private void copyTrace() {
        try {
            String trace = FreezeTraceService.readLatestTrace(getFilesDir());
            if (trace.trim().isEmpty()) {
                status.setText("저장된 로그가 없습니다.");
                return;
            }
            ClipboardManager clipboard = (ClipboardManager) getSystemService(Context.CLIPBOARD_SERVICE);
            if (clipboard == null) {
                status.setText("클립보드 서비스를 사용할 수 없습니다.");
                return;
            }
            clipboard.setPrimaryClip(ClipData.newPlainText("Zill PPSSPP persistent manual freeze capture", trace));
            status.setText("수동 캡처 로그를 복사했습니다. 그대로 채팅에 붙여 주세요.");
        } catch (Exception e) {
            status.setText("로그 읽기 실패: " + e.getMessage());
        }
    }
}
