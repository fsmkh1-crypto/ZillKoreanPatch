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

/** UI for the one-shot PPSSPP bad-pointer boundary capture. */
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
        title.setText("질올 PPSSPP 경계 캡처");
        title.setTextSize(22);
        root.addView(title, new LinearLayout.LayoutParams(-1, -2));

        TextView info = new TextView(this);
        info.setText("1) PPSSPP에서 '원격 디버거 허용'을 켜고 로컬 서버 포트를 34500으로 맞춥니다.\n"
                + "2) 여기서 '캡처 시작'을 누릅니다. PPSSPP가 늦게 켜져도 자동 재연결합니다.\n"
                + "3) PPSSPP로 돌아가 문제가 발생하던 장면까지 평소처럼 진행합니다.\n"
                + "4) 문제 메시지 생성 경로가 실행되면 producer → backend → store 경계를 한 번만 자동 캡처합니다.\n"
                + "5) 프리징까지 기다릴 필요가 없습니다. '경계 캡처 완료' 알림이 뜨면 이 앱으로 돌아와 로그를 복사합니다.\n\n"
                + "이미 확인된 스캐너/producer 디스어셈블은 다시 기록하지 않습니다.");
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
        startButton.setText("캡처 시작");
        startButton.setOnClickListener(v -> startTrace());
        LinearLayout.LayoutParams startParams = new LinearLayout.LayoutParams(-1, -2);
        startParams.topMargin = pad / 2;
        root.addView(startButton, startParams);

        Button stopButton = new Button(this);
        stopButton.setText("캡처 중지");
        stopButton.setOnClickListener(v -> stopTrace());
        root.addView(stopButton, new LinearLayout.LayoutParams(-1, -2));

        Button copyButton = new Button(this);
        copyButton.setText("최근 로그 복사");
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
        status.setText("경계 캡처를 시작했습니다. 이제 PPSSPP로 돌아가 문제 장면까지 진행하세요.");
    }

    private void stopTrace() {
        Intent intent = new Intent(this, FreezeTraceService.class);
        intent.setAction(FreezeTraceService.ACTION_STOP);
        startService(intent);
        status.setText("캡처 중지를 요청했습니다. 저장된 로그는 유지됩니다.");
    }

    private void copyTrace() {
        try {
            String trace = FreezeTraceService.readLatestTrace(getFilesDir());
            if (trace.trim().isEmpty()) {
                status.setText("아직 저장된 로그가 없습니다. '캡처 시작' 후 PPSSPP에서 문제 장면까지 진행하세요.");
                return;
            }
            ClipboardManager clipboard = (ClipboardManager) getSystemService(Context.CLIPBOARD_SERVICE);
            if (clipboard == null) {
                status.setText("클립보드 서비스를 사용할 수 없습니다.");
                return;
            }
            clipboard.setPrimaryClip(ClipData.newPlainText("Zill PPSSPP pointer boundary capture", trace));
            status.setText("경계 캡처 로그를 복사했습니다. 그대로 채팅에 붙여 주세요.");
        } catch (Exception e) {
            status.setText("로그 읽기 실패: " + e.getMessage());
        }
    }
}
