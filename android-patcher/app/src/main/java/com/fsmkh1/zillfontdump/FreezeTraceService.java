package com.fsmkh1.zillfontdump;

import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.app.PendingIntent;
import android.app.Service;
import android.content.Intent;
import android.os.IBinder;
import android.util.Base64;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileWriter;
import java.io.InputStreamReader;
import java.io.OutputStreamWriter;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeoutException;

/**
 * Manual PPSSPP freeze-state capture.
 *
 * No breakpoint is installed and no automatic problem-scene detection is
 * attempted. The user reproduces the freeze first, returns to this app, and
 * explicitly presses the capture button. At that moment we snapshot cpu.status,
 * GPRs, and the inline message-object region derived from the live s0 register.
 */
public final class FreezeTraceService extends Service {
    public static final String ACTION_CAPTURE_NOW = "com.fsmkh1.zillfontdump.CAPTURE_NOW";
    public static final String EXTRA_PORT = "port";
    public static final String TRACE_FILE = "ppsspp-freeze-trace.jsonl";

    private static final String CHANNEL_ID = "freeze_trace";
    private static final int NOTIFICATION_ID = 21010;
    private static final int RESPONSE_TIMEOUT_MS = 6000;
    private static final long INLINE_PAGE_OFFSET = 0x2C0L;
    private static final int INLINE_PAGE_CAPACITY = 0x100;
    private static final int INLINE_DUMP_SIZE = 0x120;
    private static final int POINTER_SLOT_RELATIVE = 0x100;

    private final ExecutorService worker = Executors.newSingleThreadExecutor();

    @Override
    public void onCreate() {
        super.onCreate();
        NotificationManager manager = getSystemService(NotificationManager.class);
        if (manager != null) {
            manager.createNotificationChannel(new NotificationChannel(
                    CHANNEL_ID, "PPSSPP 수동 캡처", NotificationManager.IMPORTANCE_LOW));
        }
    }

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        if (intent == null || !ACTION_CAPTURE_NOW.equals(intent.getAction())) {
            stopSelf(startId);
            return START_NOT_STICKY;
        }
        final int port = intent.getIntExtra(EXTRA_PORT, 34500);
        startForeground(NOTIFICATION_ID, notification("현재 프리징 상태 캡처 중 · " + port));
        worker.execute(() -> {
            try {
                captureNow(port);
                updateNotification("수동 캡처 완료 · 최근 로그 복사");
            } catch (Exception e) {
                writeFailure(port, e);
                updateNotification("수동 캡처 실패 · 로그 확인");
            } finally {
                stopForeground(STOP_FOREGROUND_DETACH);
                stopSelf(startId);
            }
        });
        return START_NOT_STICKY;
    }

    private void captureNow(int port) throws Exception {
        File executable = new File(getApplicationInfo().nativeLibraryDir, "libzill.so");
        if (!executable.isFile()) {
            throw new IllegalStateException("내장 debugger 실행파일을 찾을 수 없습니다");
        }

        ProcessBuilder builder = new ProcessBuilder(
                executable.getAbsolutePath(), "ppsspp-debugger",
                "--host", "127.0.0.1",
                "--port", Integer.toString(port),
                "--timeout", "6",
                "--connect-timeout", "3");
        builder.directory(getFilesDir());
        builder.redirectErrorStream(true);
        Process process = builder.start();
        int[] requestId = new int[]{1};

        try (BufferedReader reader = new BufferedReader(new InputStreamReader(
                     process.getInputStream(), StandardCharsets.UTF_8));
             BufferedWriter writer = new BufferedWriter(new OutputStreamWriter(
                     process.getOutputStream(), StandardCharsets.UTF_8))) {
            JSONObject ready = readObject(reader, "debugger handshake", 4000);
            if (!"ready".equals(ready.optString("event"))) {
                throw new IllegalStateException("unexpected handshake: " + ready);
            }

            JSONObject statusResponse = request(writer, reader, requestId[0]++,
                    rawCommand("cpu.status", new JSONObject()), RESPONSE_TIMEOUT_MS);
            JSONObject status = rawResponse(statusResponse);
            JSONObject regs = request(writer, reader, requestId[0]++,
                    rawCommand("cpu.getAllRegs", new JSONObject()), RESPONSE_TIMEOUT_MS);

            long s0 = requireRegister(regs, "s0");
            long s3 = requireRegister(regs, "s3");
            long s4 = requireRegister(regs, "s4");
            long a1 = requireRegister(regs, "a1");
            long pc = requireRegister(regs, "pc");
            long pageStart = (s0 + INLINE_PAGE_OFFSET) & 0xffffffffL;
            byte[] dump = readMemory(writer, reader, requestId, pageStart, INLINE_DUMP_SIZE);

            JSONObject capture = new JSONObject();
            capture.put("event", "manual_freeze_capture");
            capture.put("time_ms", System.currentTimeMillis());
            capture.put("target", "127.0.0.1:" + port);
            capture.put("cpu_status", status);
            capture.put("pc", hex32(pc));
            capture.put("s0", hex32(s0));
            capture.put("s3", hex32(s3));
            capture.put("s4", hex32(s4));
            capture.put("a1", hex32(a1));
            capture.put("inline_page_start", hex32(pageStart));
            capture.put("inline_page_capacity", INLINE_PAGE_CAPACITY);
            capture.put("inline_dump_size", dump.length);
            capture.put("inline_dump_base64", Base64.encodeToString(dump, Base64.NO_WRAP));
            appendInlineAnalysis(capture, dump);
            writeCapture(capture.toString());
        } finally {
            process.destroy();
        }
    }

    private static void appendInlineAnalysis(JSONObject out, byte[] dump) throws Exception {
        int firstNul = -1;
        List<Integer> lfPositions = new ArrayList<>();
        int maxSpan = 0;
        int span = 0;
        int analysisEnd = Math.min(dump.length, INLINE_PAGE_CAPACITY);

        for (int i = 0; i < dump.length; i++) {
            int value = dump[i] & 0xff;
            if (firstNul < 0 && value == 0) firstNul = i;
            if (i >= analysisEnd || (firstNul >= 0 && i >= firstNul)) continue;
            if (value == 0x0A) {
                lfPositions.add(i);
                if (span > maxSpan) maxSpan = span;
                span = 0;
            } else {
                span++;
            }
        }
        if (span > maxSpan) maxSpan = span;

        JSONArray lf = new JSONArray();
        for (int position : lfPositions) lf.put(position);
        out.put("first_nul_offset", firstNul);
        out.put("has_nul_within_inline_page", firstNul >= 0 && firstNul < INLINE_PAGE_CAPACITY);
        out.put("lf_count_before_nul", lfPositions.size());
        out.put("lf_positions", lf);
        out.put("max_non_lf_span_before_nul", maxSpan);

        if (dump.length >= POINTER_SLOT_RELATIVE + 4) {
            long slot = ((long) dump[POINTER_SLOT_RELATIVE] & 0xff)
                    | (((long) dump[POINTER_SLOT_RELATIVE + 1] & 0xff) << 8)
                    | (((long) dump[POINTER_SLOT_RELATIVE + 2] & 0xff) << 16)
                    | (((long) dump[POINTER_SLOT_RELATIVE + 3] & 0xff) << 24);
            out.put("slot_plus_3c0_word", hex32(slot));
        }
    }

    private static byte[] readMemory(BufferedWriter writer, BufferedReader reader,
                                     int[] requestId, long address, int size) throws Exception {
        JSONObject params = new JSONObject();
        params.put("address", address & 0xffffffffL);
        params.put("size", size);
        JSONObject response = request(writer, reader, requestId[0]++,
                rawCommand("memory.read", params), RESPONSE_TIMEOUT_MS);
        String encoded = rawResponse(response).optString("base64", "");
        if (encoded.isEmpty()) {
            throw new IllegalStateException("memory.read 응답에 base64 데이터가 없습니다");
        }
        byte[] decoded = Base64.decode(encoded, Base64.DEFAULT);
        if (decoded.length != size) {
            throw new IllegalStateException("memory.read 크기 불일치: got=" + decoded.length + " want=" + size);
        }
        return decoded;
    }

    private static JSONObject rawCommand(String event, JSONObject params) throws Exception {
        JSONObject out = new JSONObject();
        out.put("command", "raw");
        out.put("event", event);
        out.put("params", params);
        return out;
    }

    private static JSONObject request(BufferedWriter writer, BufferedReader reader, int id,
                                      JSONObject command, int timeoutMs) throws Exception {
        command.put("id", id);
        writer.write(command.toString());
        writer.newLine();
        writer.flush();
        JSONObject response = readObject(reader, "command " + id, timeoutMs);
        while (!response.has("id")) response = readObject(reader, "command " + id, timeoutMs);
        if (!response.optBoolean("ok", false)) {
            JSONObject error = response.optJSONObject("error");
            throw new IllegalStateException(error == null
                    ? response.toString() : error.optString("message", response.toString()));
        }
        return response;
    }

    private static JSONObject readObject(BufferedReader reader, String what, int timeoutMs) throws Exception {
        long deadline = System.currentTimeMillis() + timeoutMs;
        while (System.currentTimeMillis() < deadline) {
            if (reader.ready()) {
                String line = reader.readLine();
                if (line == null) throw new IllegalStateException(what + " 응답 전에 연결이 종료됐습니다");
                return new JSONObject(line);
            }
            Thread.sleep(20);
        }
        throw new TimeoutException(what + " 응답이 " + timeoutMs + "ms 동안 없습니다");
    }

    private static JSONObject rawResponse(JSONObject response) {
        JSONObject result = response.optJSONObject("result");
        if (result == null) return new JSONObject();
        JSONObject raw = result.optJSONObject("response");
        return raw == null ? new JSONObject() : raw;
    }

    private static long requireRegister(JSONObject response, String wanted) {
        JSONObject raw = rawResponse(response);
        JSONArray categories = raw.optJSONArray("categories");
        if (categories == null) throw new IllegalStateException(wanted + " 레지스터를 읽지 못했습니다");
        for (int i = 0; i < categories.length(); i++) {
            JSONObject category = categories.optJSONObject(i);
            if (category == null) continue;
            JSONArray names = category.optJSONArray("registerNames");
            JSONArray values = category.optJSONArray("uintValues");
            if (names == null || values == null) continue;
            int count = Math.min(names.length(), values.length());
            for (int j = 0; j < count; j++) {
                if (wanted.equals(names.optString(j))) return values.optLong(j, -1) & 0xffffffffL;
            }
        }
        throw new IllegalStateException(wanted + " 레지스터를 읽지 못했습니다");
    }

    private synchronized void writeCapture(String line) {
        File out = new File(getFilesDir(), TRACE_FILE);
        try (BufferedWriter writer = new BufferedWriter(new FileWriter(out, false))) {
            writer.write(line);
            writer.newLine();
        } catch (Exception ignored) {
        }
    }

    private void writeFailure(int port, Exception error) {
        try {
            JSONObject obj = new JSONObject();
            obj.put("event", "manual_capture_failed");
            obj.put("time_ms", System.currentTimeMillis());
            obj.put("target", "127.0.0.1:" + port);
            obj.put("error", message(error));
            writeCapture(obj.toString());
        } catch (Exception ignored) {
        }
    }

    public static String readLatestTrace(File filesDir) throws Exception {
        File trace = new File(filesDir, TRACE_FILE);
        if (!trace.isFile()) return "";
        StringBuilder out = new StringBuilder();
        try (BufferedReader reader = new BufferedReader(new InputStreamReader(
                new FileInputStream(trace), StandardCharsets.UTF_8))) {
            String line;
            while ((line = reader.readLine()) != null) out.append(line).append('\n');
        }
        return out.toString();
    }

    private void updateNotification(String text) {
        NotificationManager manager = getSystemService(NotificationManager.class);
        if (manager != null) manager.notify(NOTIFICATION_ID, notification(text));
    }

    private Notification notification(String text) {
        Intent open = new Intent(this, FreezeCaptureActivity.class);
        PendingIntent pending = PendingIntent.getActivity(this, 0, open,
                PendingIntent.FLAG_UPDATE_CURRENT | PendingIntent.FLAG_IMMUTABLE);
        return new Notification.Builder(this, CHANNEL_ID)
                .setSmallIcon(android.R.drawable.stat_notify_sync)
                .setContentTitle("질올 PPSSPP 수동 캡처")
                .setContentText(text)
                .setContentIntent(pending)
                .build();
    }

    private static String hex32(long value) {
        return String.format("0x%08X", value & 0xffffffffL);
    }

    private static String message(Exception e) {
        String m = e.getMessage();
        return (m == null || m.trim().isEmpty()) ? e.getClass().getSimpleName() : m;
    }

    @Override
    public void onDestroy() {
        worker.shutdown();
        super.onDestroy();
    }

    @Override
    public IBinder onBind(Intent intent) { return null; }
}
