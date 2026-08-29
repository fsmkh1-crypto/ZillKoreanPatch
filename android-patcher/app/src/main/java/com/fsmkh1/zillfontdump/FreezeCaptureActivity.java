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

/**
 * UI for rolling PPSSPP freeze tracing. The foreground service repeatedly tries
 * to connect while the user returns to PPSSPP, then keeps a short pre-freeze
 * history instead of attempting a first connection after the freeze occurred.
 */
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
        title.setText("질올 PPSSPP 프리징 기록");
        title.setTextSize(22);
        root.addView(title, new LinearLayout.LayoutParams(-1, -2));

        TextView info = new TextView(this);
        info.setText("1) PPSSPP에서 '원격 디버거 허용'을 켜고 로컬 서버 포트를 34500으로 맞춥니다.\n"
                + "2) 여기서 '기록 시작'을 누릅니다. 처음에는 PPSSPP 연결을 계속 재시도합니다.\n"
                + "3) PPSSPP로 돌아가 한국어 패치 ISO를 플레이합니다.\n"
                + "4) 연결되면 최근 약 30초의 PC/SP/CPU 상태를 계속 보존합니다.\n"
                + "5) 프리징 또는 연결 끊김 후 이 앱으로 돌아와 '최근 기록 복사'를 누릅니다.\n\n"
                + "중요: 기록 시작 후에는 프리징이 날 때까지 이 앱으로 다시 전환할 필요가 없습니다.");
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

        Button startButton = new Button(this);
        startButton.setText("기록 시작");
        startButton.setOnClickListener(v -> startTrace());
        LinearLayout.LayoutParams startParams = new LinearLayout.LayoutParams(-1, -2);
        startParams.topMargin = pad / 2;
        root.addView(startButton, startParams);

        Button stopButton = new Button(this);
        stopButton.setText("기록 중지");
        stopButton.setOnClickListener(v -> stopTrace());
        root.addView(stopButton, new LinearLayout.LayoutParams(-1, -2));

        Button copyButton = new Button(this);
        copyButton.setText("최근 기록 복사");
        copyButton.setOnClickListener(v -> copyTrace());
        root.addView(copyButton, new LinearLayout.LayoutParams(-1, -2));

        status = new TextView(this);
        status.setText("대기 중");
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

    private void startTrace() {
        int port = readPort();
        if (port < 0) return;
        Intent intent = new Intent(this, FreezeTraceService.class);
        intent.setAction(FreezeTraceService.ACTION_START);
        intent.putExtra(FreezeTraceService.EXTRA_PORT, port);
        startForegroundService(intent);
        status.setText("기록 서비스를 시작했습니다. 이제 PPSSPP로 돌아가세요. 연결될 때까지 자동 재시도합니다.");
    }

    private void stopTrace() {
        Intent intent = new Intent(this, FreezeTraceService.class);
        intent.setAction(FreezeTraceService.ACTION_STOP);
        startService(intent);
        status.setText("기록 중지를 요청했습니다. 마지막 링 버퍼는 보존됩니다.");
    }

    private void copyTrace() {
        try {
            String trace = FreezeTraceService.readLatestTrace(getFilesDir());
            if (trace.trim().isEmpty()) {
                status.setText("아직 저장된 기록이 없습니다. '기록 시작' 후 PPSSPP로 돌아가 연결될 때까지 기다리세요.");
                return;
            }
            ClipboardManager clipboard = (ClipboardManager) getSystemService(Context.CLIPBOARD_SERVICE);
            if (clipboard == null) {
                status.setText("클립보드 서비스를 사용할 수 없습니다.");
                return;
            }
            clipboard.setPrimaryClip(ClipData.newPlainText("Zill PPSSPP rolling freeze trace", trace));
            status.setText("최근 프리징 기록을 클립보드에 복사했습니다. 그대로 채팅에 붙여 주세요.");
        } catch (Exception e) {
            status.setText("기록 읽기 실패: " + e.getMessage());
        }
    }
}
