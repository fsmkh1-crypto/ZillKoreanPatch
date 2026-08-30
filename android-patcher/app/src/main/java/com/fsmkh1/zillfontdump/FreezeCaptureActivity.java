package com.fsmkh1.zillfontdump;

import android.app.Activity;
import android.content.ClipData;
import android.content.ClipboardManager;
import android.content.Context;
import android.content.Intent;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.view.Gravity;
import android.widget.Button;
import android.widget.EditText;
import android.widget.LinearLayout;
import android.widget.TextView;

/** UI for rolling PPSSPP freeze tracing. */
public final class FreezeCaptureActivity extends Activity {
    private EditText portInput;
    private TextView status;
    private Button startButton;
    private final Handler handler = new Handler(Looper.getMainLooper());
    private final Runnable statusPoll = new Runnable() {
        @Override public void run() {
            refreshServiceState();
            handler.postDelayed(this, 500);
        }
    };

    @Override protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        int pad = (int) (24 * getResources().getDisplayMetrics().density);
        LinearLayout root = new LinearLayout(this);
        root.setOrientation(LinearLayout.VERTICAL);
        root.setGravity(Gravity.CENTER_HORIZONTAL);
        root.setPadding(pad, pad, pad, pad);

        TextView title = new TextView(this);
        title.setText("질올 PPSSPP 프리징 기록"); title.setTextSize(22);
        root.addView(title, new LinearLayout.LayoutParams(-1, -2));

        TextView info = new TextView(this);
        info.setText("1) 포트 34500 확인 후 '기록 시작'을 한 번만 누릅니다.\n"
                + "2) 아래 상태가 '연결됨 · 샘플링 중'이면 연결이 실제로 성립한 것입니다.\n"
                + "3) PPSSPP로 돌아가 문제 장면까지 평소처럼 플레이합니다.\n"
                + "4) 앱은 0.5초마다 PC/GPR/CPU 상태를 계속 기록합니다.\n"
                + "5) 프리징이 나면 이 앱으로 돌아와 '최근 기록 복사'를 누릅니다.\n\n"
                + "중요: 중복 시작은 차단됩니다. 연결 대기/재시도/샘플링 상태를 화면에서 직접 확인할 수 있습니다.");
        info.setTextSize(15);
        LinearLayout.LayoutParams infoParams = new LinearLayout.LayoutParams(-1, -2); infoParams.topMargin = pad / 2;
        root.addView(info, infoParams);

        portInput = new EditText(this); portInput.setHint("PPSSPP 로컬 서버 포트");
        portInput.setInputType(android.text.InputType.TYPE_CLASS_NUMBER); portInput.setText("34500");
        LinearLayout.LayoutParams portParams = new LinearLayout.LayoutParams(-1, -2); portParams.topMargin = pad;
        root.addView(portInput, portParams);

        startButton = new Button(this); startButton.setText("기록 시작"); startButton.setOnClickListener(v -> startTrace());
        LinearLayout.LayoutParams startParams = new LinearLayout.LayoutParams(-1, -2); startParams.topMargin = pad / 2;
        root.addView(startButton, startParams);
        Button stopButton = new Button(this); stopButton.setText("기록 중지"); stopButton.setOnClickListener(v -> stopTrace()); root.addView(stopButton, new LinearLayout.LayoutParams(-1, -2));
        Button copyButton = new Button(this); copyButton.setText("최근 기록 복사"); copyButton.setOnClickListener(v -> copyTrace()); root.addView(copyButton, new LinearLayout.LayoutParams(-1, -2));

        status = new TextView(this); status.setText("대기 중"); status.setTextSize(14); status.setTextIsSelectable(true);
        LinearLayout.LayoutParams statusParams = new LinearLayout.LayoutParams(-1, -2); statusParams.topMargin = pad;
        root.addView(status, statusParams);
        setContentView(root);
    }

    @Override protected void onResume() {
        super.onResume();
        handler.removeCallbacks(statusPoll);
        handler.post(statusPoll);
    }

    @Override protected void onPause() {
        handler.removeCallbacks(statusPoll);
        super.onPause();
    }

    private void refreshServiceState() {
        boolean running = FreezeTraceService.isTraceRunning();
        startButton.setEnabled(!running);
        String state = FreezeTraceService.getVisibleState();
        int port = FreezeTraceService.getVisiblePort();
        if (running && port > 0) status.setText(state + "\n현재 포트: " + port);
        else status.setText(state);
    }

    private int readPort() {
        try { int port = Integer.parseInt(portInput.getText().toString().trim()); if (port < 1 || port > 65535) throw new NumberFormatException(); return port; }
        catch (NumberFormatException e) { status.setText("포트는 1~65535 숫자여야 합니다."); return -1; }
    }

    private void startTrace() {
        if (FreezeTraceService.isTraceRunning()) { refreshServiceState(); return; }
        int port = readPort(); if (port < 0) return;
        Intent intent = new Intent(this, FreezeTraceService.class);
        intent.setAction(FreezeTraceService.ACTION_START);
        intent.putExtra(FreezeTraceService.EXTRA_PORT, port);
        startForegroundService(intent);
        status.setText("기록 시작 요청됨 · 연결 상태 확인 중");
        startButton.setEnabled(false);
    }

    private void stopTrace() {
        Intent intent = new Intent(this, FreezeTraceService.class);
        intent.setAction(FreezeTraceService.ACTION_STOP);
        startService(intent);
        status.setText("기록 중지를 요청했습니다. 마지막 로그는 보존됩니다.");
    }

    private void copyTrace() {
        try {
            String trace = FreezeTraceService.readLatestTrace(getFilesDir());
            if (trace.trim().isEmpty()) { status.setText("아직 저장된 기록이 없습니다."); return; }
            ClipboardManager clipboard = (ClipboardManager) getSystemService(Context.CLIPBOARD_SERVICE);
            if (clipboard == null) { status.setText("클립보드 서비스를 사용할 수 없습니다."); return; }
            clipboard.setPrimaryClip(ClipData.newPlainText("Zill PPSSPP rolling freeze trace", trace));
            status.setText("최근 프리징 기록을 복사했습니다. 그대로 채팅에 붙여 주세요.");
        } catch (Exception e) { status.setText("기록 읽기 실패: " + e.getMessage()); }
    }
}
