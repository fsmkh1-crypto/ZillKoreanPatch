package com.fsmkh1.zillfontdump;

import android.app.Activity;
import android.content.ClipData;
import android.content.ClipboardManager;
import android.content.Context;
import android.os.Bundle;
import android.view.Gravity;
import android.widget.Button;
import android.widget.EditText;
import android.widget.LinearLayout;
import android.widget.TextView;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.File;
import java.io.InputStreamReader;
import java.io.OutputStreamWriter;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * One-shot on-device PPSSPP CPU capture for the reproducible Korean-patch freeze.
 *
 * PPSSPP's WebSocket debugger is exposed on the same port as Remote ISO sharing.
 * Set a fixed Local Server Port in PPSSPP (34500 by default here), enable
 * Settings > Tools > Developer Tools > Allow remote debugger, reproduce the
 * freeze, switch to this activity, and capture. Android may UI-pause PPSSPP
 * during the app switch; that stopped state is valid evidence and is read as-is.
 */
public final class FreezeCaptureActivity extends Activity {
    private final ExecutorService worker = Executors.newSingleThreadExecutor();
    private EditText portInput;
    private Button captureButton;
    private Button copyButton;
    private TextView status;
    private String lastSnapshot = "";

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        int pad = (int) (24 * getResources().getDisplayMetrics().density);
        LinearLayout root = new LinearLayout(this);
        root.setOrientation(LinearLayout.VERTICAL);
        root.setGravity(Gravity.CENTER_HORIZONTAL);
        root.setPadding(pad, pad, pad, pad);

        TextView title = new TextView(this);
        title.setText("질올 PPSSPP 프리징 캡처");
        title.setTextSize(22);
        root.addView(title, new LinearLayout.LayoutParams(-1, -2));

        TextView info = new TextView(this);
        info.setText("1) PPSSPP에서 설정 > 도구 > 개발자 도구 > '원격 디버거 허용'을 켭니다.\n"
                + "2) PPSSPP의 로컬 서버 포트를 아래 값과 같게 고정합니다. (기본 34500)\n"
                + "3) 한국어 패치 ISO에서 프리징을 재현합니다.\n"
                + "4) 프리징 상태 그대로 이 앱으로 전환해 캡처합니다.\n\n"
                + "캡처는 PSP PC/레지스터/PC 주변 명령/스택을 읽습니다. 앱 전환 때문에 PPSSPP가 이미 UI-pause된 경우 그 상태를 그대로 보존하고 읽습니다.");
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

        captureButton = new Button(this);
        captureButton.setText("프리징 CPU 캡처");
        captureButton.setOnClickListener(v -> captureFreeze());
        LinearLayout.LayoutParams captureParams = new LinearLayout.LayoutParams(-1, -2);
        captureParams.topMargin = pad / 2;
        root.addView(captureButton, captureParams);

        copyButton = new Button(this);
        copyButton.setText("프리징 로그 복사");
        copyButton.setEnabled(false);
        copyButton.setOnClickListener(v -> copySnapshot());
        LinearLayout.LayoutParams copyParams = new LinearLayout.LayoutParams(-1, -2);
        copyParams.topMargin = pad / 2;
        root.addView(copyButton, copyParams);

        status = new TextView(this);
        status.setText("대기 중");
        status.setTextSize(14);
        status.setTextIsSelectable(true);
        LinearLayout.LayoutParams statusParams = new LinearLayout.LayoutParams(-1, -2);
        statusParams.topMargin = pad;
        root.addView(status, statusParams);

        setContentView(root);
    }

    private void captureFreeze() {
        final int port;
        try {
            port = Integer.parseInt(portInput.getText().toString().trim());
            if (port < 1 || port > 65535) throw new NumberFormatException();
        } catch (NumberFormatException e) {
            status.setText("포트는 1~65535 숫자여야 합니다.");
            return;
        }

        captureButton.setEnabled(false);
        copyButton.setEnabled(false);
        lastSnapshot = "";
        status.setText("127.0.0.1:" + port + "의 PPSSPP 디버거에 연결 중…");

        worker.execute(() -> {
            Process process = null;
            try {
                File executable = new File(getApplicationInfo().nativeLibraryDir, "libzill.so");
                if (!executable.isFile()) {
                    throw new IllegalStateException("내장 zill 실행파일을 찾을 수 없습니다: " + executable);
                }
                ProcessBuilder builder = new ProcessBuilder(
                        executable.getAbsolutePath(),
                        "ppsspp-debugger",
                        "--host", "127.0.0.1",
                        "--port", Integer.toString(port),
                        "--timeout", "8",
                        "--connect-timeout", "4");
                builder.directory(getFilesDir());
                builder.redirectErrorStream(true);
                process = builder.start();

                try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()));
                     BufferedWriter writer = new BufferedWriter(new OutputStreamWriter(process.getOutputStream()))) {
                    JSONObject snapshot = new JSONObject();
                    snapshot.put("format", "zill-android-ppsspp-freeze-snapshot-v1");
                    snapshot.put("target", "127.0.0.1:" + port);

                    JSONObject ready = readObject(reader, "debugger handshake");
                    if (!"ready".equals(ready.optString("event"))) {
                        throw new IllegalStateException("PPSSPP 디버거 연결 실패: " + ready);
                    }
                    snapshot.put("ready", ready);

                    JSONObject before = request(writer, reader, 1, command("status"));
                    snapshot.put("status_before", before);
                    JSONObject beforeCpu = before.optJSONObject("result") == null
                            ? null : before.optJSONObject("result").optJSONObject("cpu");
                    boolean alreadyPaused = beforeCpu != null && beforeCpu.optBoolean("paused", false);
                    boolean alreadyStepping = beforeCpu != null && beforeCpu.optBoolean("stepping", false);

                    if (alreadyPaused || alreadyStepping) {
                        JSONObject preserved = new JSONObject();
                        preserved.put("changed", false);
                        preserved.put("reason", alreadyPaused ? "already_ui_paused" : "already_debugger_stopped");
                        snapshot.put("pause", preserved);
                    } else {
                        snapshot.put("pause", request(writer, reader, 2, command("pause")));
                    }

                    JSONObject cpu = request(writer, reader, 3,
                            rawCommand("cpu.status", new JSONObject()));
                    snapshot.put("cpu", cpu);

                    JSONObject regs = request(writer, reader, 4,
                            rawCommand("cpu.getAllRegs", new JSONObject()));
                    snapshot.put("registers", regs);

                    long pcValue = findRegister(regs, "pc");
                    if (pcValue < 0) pcValue = rawResponse(cpu).optLong("pc", -1);
                    long spValue = findRegister(regs, "sp");
                    if (pcValue < 0) throw new IllegalStateException("캡처 응답에서 PC를 찾지 못했습니다.");
                    final long pc = pcValue;
                    final long sp = spValue;
                    snapshot.put("pc", pc);
                    snapshot.put("pc_hex", hex32(pc));
                    if (sp >= 0) {
                        snapshot.put("sp", sp);
                        snapshot.put("sp_hex", hex32(sp));
                    }

                    long disasmStart = Math.max(0L, (pc & 0xffffffffL) - 64L);
                    JSONObject disasmParams = new JSONObject();
                    disasmParams.put("address", disasmStart);
                    disasmParams.put("count", 48);
                    disasmParams.put("displaySymbols", true);
                    JSONObject disasm = request(writer, reader, 5,
                            rawCommand("memory.disasm", disasmParams));
                    snapshot.put("disassembly", disasm);

                    if (sp >= 0) {
                        JSONObject stackCmd = command("read_memory");
                        stackCmd.put("address", sp & 0xffffffffL);
                        stackCmd.put("size", 512);
                        stackCmd.put("replacements", true);
                        try {
                            snapshot.put("stack", request(writer, reader, 6, stackCmd));
                        } catch (Exception stackError) {
                            snapshot.put("stack_error", message(stackError));
                        }
                    }

                    snapshot.put("left_stopped", true);
                    snapshot.put("stop_origin", alreadyPaused ? "android_ui_pause" : (alreadyStepping ? "preexisting_debugger_stop" : "zill_debugger_stop"));
                    try {
                        request(writer, reader, 7, command("quit"));
                    } catch (Exception ignored) {
                    }

                    final String result = snapshot.toString(2);
                    runOnUiThread(() -> {
                        lastSnapshot = result;
                        copyButton.setEnabled(true);
                        captureButton.setEnabled(true);
                        status.setText("캡처 완료: PC=" + hex32(pc) + (sp >= 0 ? "  SP=" + hex32(sp) : "")
                                + "\n'프리징 로그 복사'를 눌러 그대로 전달해 주세요.");
                    });
                }
            } catch (Exception e) {
                final String error = message(e);
                runOnUiThread(() -> {
                    captureButton.setEnabled(true);
                    copyButton.setEnabled(!lastSnapshot.isEmpty());
                    status.setText("캡처 실패: " + error + "\nPPSSPP의 '원격 디버거 허용'과 로컬 서버 포트가 같은지 확인하세요.");
                });
            } finally {
                if (process != null) process.destroy();
            }
        });
    }

    private static JSONObject command(String name) throws Exception {
        JSONObject out = new JSONObject();
        out.put("command", name);
        return out;
    }

    private static JSONObject rawCommand(String event, JSONObject params) throws Exception {
        JSONObject out = command("raw");
        out.put("event", event);
        out.put("params", params);
        return out;
    }

    private static JSONObject request(BufferedWriter writer, BufferedReader reader, int id, JSONObject command) throws Exception {
        command.put("id", id);
        writer.write(command.toString());
        writer.newLine();
        writer.flush();
        JSONObject response = readObject(reader, "command " + id);
        while (!response.has("id")) {
            response = readObject(reader, "command " + id);
        }
        if (!response.optBoolean("ok", false)) {
            JSONObject error = response.optJSONObject("error");
            throw new IllegalStateException(error == null ? response.toString() : error.optString("message", response.toString()));
        }
        return response;
    }

    private static JSONObject readObject(BufferedReader reader, String what) throws Exception {
        String line = reader.readLine();
        if (line == null) throw new IllegalStateException(what + " 응답 전에 디버거 연결이 종료됐습니다.");
        return new JSONObject(line);
    }

    private static JSONObject rawResponse(JSONObject commandResponse) {
        JSONObject result = commandResponse.optJSONObject("result");
        if (result == null) return new JSONObject();
        JSONObject response = result.optJSONObject("response");
        return response == null ? new JSONObject() : response;
    }

    private static long findRegister(JSONObject commandResponse, String wanted) {
        JSONObject response = rawResponse(commandResponse);
        JSONArray categories = response.optJSONArray("categories");
        if (categories == null) return -1;
        for (int i = 0; i < categories.length(); i++) {
            JSONObject category = categories.optJSONObject(i);
            if (category == null) continue;
            JSONArray names = category.optJSONArray("registerNames");
            JSONArray values = category.optJSONArray("uintValues");
            if (names == null || values == null) continue;
            int count = Math.min(names.length(), values.length());
            for (int j = 0; j < count; j++) {
                if (wanted.equals(names.optString(j))) {
                    return values.optLong(j, -1) & 0xffffffffL;
                }
            }
        }
        return -1;
    }

    private void copySnapshot() {
        if (lastSnapshot.trim().isEmpty()) {
            status.setText("복사할 프리징 로그가 없습니다.");
            return;
        }
        ClipboardManager clipboard = (ClipboardManager) getSystemService(Context.CLIPBOARD_SERVICE);
        if (clipboard == null) {
            status.setText("클립보드 서비스를 사용할 수 없습니다.");
            return;
        }
        clipboard.setPrimaryClip(ClipData.newPlainText("Zill PPSSPP freeze snapshot", lastSnapshot));
        status.setText("프리징 로그를 클립보드에 복사했습니다.");
    }

    private static String hex32(long value) {
        return String.format("0x%08X", value & 0xffffffffL);
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
