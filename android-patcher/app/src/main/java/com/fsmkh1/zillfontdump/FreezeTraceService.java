package com.fsmkh1.zillfontdump;

import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.app.PendingIntent;
import android.app.Service;
import android.content.Intent;
import android.content.SharedPreferences;
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
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TimeoutException;

/** Persistent PPSSPP debugger connection with explicit user-triggered capture. */
public final class FreezeTraceService extends Service {
    public static final String ACTION_CONNECT = "com.fsmkh1.zillfontdump.CONNECT";
    public static final String ACTION_CAPTURE_NOW = "com.fsmkh1.zillfontdump.CAPTURE_NOW";
    public static final String ACTION_DISCONNECT = "com.fsmkh1.zillfontdump.DISCONNECT";
    public static final String EXTRA_PORT = "port";
    public static final String TRACE_FILE = "ppsspp-freeze-trace.jsonl";

    private static final String CHANNEL_ID = "freeze_trace";
    private static final int NOTIFICATION_ID = 21010;
    private static final int RESPONSE_TIMEOUT_MS = 6000;
    private static final long INLINE_PAGE_OFFSET = 0x2C0L;
    private static final int INLINE_PAGE_CAPACITY = 0x100;
    private static final int INLINE_DUMP_SIZE = 0x120;
    private static final int POINTER_SLOT_RELATIVE = 0x100;

    private static final String PREFS = "freeze_trace_state";
    private static final String PREF_WANT_CONNECTED = "want_connected";
    private static final String PREF_PORT = "port";
    private static final int DEFAULT_PORT = 34500;
    private static final int SUPERVISOR_PERIOD_SECONDS = 5;

    private final ExecutorService worker = Executors.newSingleThreadExecutor();
    private final ScheduledExecutorService supervisor = Executors.newSingleThreadScheduledExecutor();
    private volatile Process process;
    private BufferedReader reader;
    private BufferedWriter writer;
    private int nextRequestId = 1;
    private volatile boolean connected;
    private volatile int connectedPort = -1;

    @Override
    public void onCreate() {
        super.onCreate();
        NotificationManager manager = getSystemService(NotificationManager.class);
        if (manager != null) {
            manager.createNotificationChannel(new NotificationChannel(
                    CHANNEL_ID, "PPSSPP 수동 캡처", NotificationManager.IMPORTANCE_LOW));
        }
        supervisor.scheduleWithFixedDelay(this::superviseConnection,
                SUPERVISOR_PERIOD_SECONDS, SUPERVISOR_PERIOD_SECONDS, TimeUnit.SECONDS);
    }

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        SharedPreferences prefs = prefs();
        if (intent == null) {
            if (prefs.getBoolean(PREF_WANT_CONNECTED, false)) {
                int port = prefs.getInt(PREF_PORT, DEFAULT_PORT);
                startForeground(NOTIFICATION_ID, notification("PPSSPP 연결 복구 중 · " + port));
                worker.execute(() -> connectPersistent(port, "service_restart"));
                return START_STICKY;
            }
            stopSelf(startId);
            return START_NOT_STICKY;
        }

        String action = intent.getAction();
        int port = intent.getIntExtra(EXTRA_PORT, prefs.getInt(PREF_PORT, DEFAULT_PORT));

        if (ACTION_CONNECT.equals(action)) {
            prefs.edit().putBoolean(PREF_WANT_CONNECTED, true).putInt(PREF_PORT, port).apply();
            startForeground(NOTIFICATION_ID, notification("PPSSPP 연결 중 · " + port));
            worker.execute(() -> connectPersistent(port, "user_connect"));
            return START_STICKY;
        }
        if (ACTION_CAPTURE_NOW.equals(action)) {
            startForeground(NOTIFICATION_ID, notification("현재 상태 수동 캡처 중"));
            worker.execute(this::captureWithRecovery);
            return START_STICKY;
        }
        if (ACTION_DISCONNECT.equals(action)) {
            prefs.edit().putBoolean(PREF_WANT_CONNECTED, false).apply();
            worker.execute(() -> {
                closeConnection();
                updateNotification("연결 종료됨");
                stopForeground(STOP_FOREGROUND_REMOVE);
                stopSelf();
            });
            return START_NOT_STICKY;
        }
        return START_STICKY;
    }

    private SharedPreferences prefs() {
        return getSharedPreferences(PREFS, MODE_PRIVATE);
    }

    private synchronized boolean sessionUsable() {
        return connected && process != null && process.isAlive() && reader != null && writer != null;
    }

    private void superviseConnection() {
        SharedPreferences prefs = prefs();
        if (!prefs.getBoolean(PREF_WANT_CONNECTED, false)) return;
        int port = prefs.getInt(PREF_PORT, DEFAULT_PORT);
        synchronized (this) {
            if (!sessionUsable()) {
                connectPersistent(port, "supervisor_reconnect");
                return;
            }
            try {
                request(writer, reader, nextRequestId++, rawCommand("cpu.status", new JSONObject()), RESPONSE_TIMEOUT_MS);
            } catch (Exception e) {
                writeFailure("heartbeat_failed", e);
                closeConnection();
                updateNotification("연결 끊김 · 자동 재연결 대기");
            }
        }
    }

    private synchronized void connectPersistent(int port, String reason) {
        if (sessionUsable() && connectedPort == port) {
            updateNotification("연결 유지 중 · 프리징 후 수동 캡처");
            writeEvent("already_connected", "reason", reason);
            return;
        }
        closeConnection();
        writeEvent("connect_start", "reason", reason);
        try {
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
            process = builder.start();
            reader = new BufferedReader(new InputStreamReader(process.getInputStream(), StandardCharsets.UTF_8));
            writer = new BufferedWriter(new OutputStreamWriter(process.getOutputStream(), StandardCharsets.UTF_8));
            JSONObject ready = readObject(reader, "debugger handshake", 4000);
            if (!"ready".equals(ready.optString("event"))) {
                throw new IllegalStateException("unexpected handshake: " + ready);
            }
            connected = true;
            connectedPort = port;
            nextRequestId = 1;
            prefs().edit().putBoolean(PREF_WANT_CONNECTED, true).putInt(PREF_PORT, port).apply();
            writeEvent("connected", "reason", reason);
            updateNotification("연결 유지 중 · 프리징 후 수동 캡처");
        } catch (Exception e) {
            writeFailure("connect_failed", e);
            closeConnection();
            updateNotification("연결 실패 · 자동 재연결 대기");
        }
    }

    private synchronized void captureWithRecovery() {
        SharedPreferences prefs = prefs();
        int port = prefs.getInt(PREF_PORT, DEFAULT_PORT);
        boolean reused = sessionUsable();
        if (!reused) {
            if (!prefs.getBoolean(PREF_WANT_CONNECTED, false)) {
                writeFailure("manual_capture_failed",
                        new IllegalStateException("연결 유지가 시작되지 않았습니다"));
                updateNotification("캡처 실패 · 먼저 연결 유지 시작");
                return;
            }
            writeEvent("capture_reconnect_attempt", "port", port);
            connectPersistent(port, "capture_recovery");
        }
        if (!sessionUsable()) {
            writeFailure("manual_capture_failed",
                    new IllegalStateException("debugger 연결 복구 후에도 세션을 사용할 수 없습니다"));
            updateNotification("캡처 실패 · 연결 복구 실패");
            return;
        }

        try {
            JSONObject statusResponse = request(writer, reader, nextRequestId++,
                    rawCommand("cpu.status", new JSONObject()), RESPONSE_TIMEOUT_MS);
            JSONObject status = rawResponse(statusResponse);
            JSONObject regs = request(writer, reader, nextRequestId++,
                    rawCommand("cpu.getAllRegs", new JSONObject()), RESPONSE_TIMEOUT_MS);

            long s0 = requireRegister(regs, "s0");
            long s3 = requireRegister(regs, "s3");
            long s4 = requireRegister(regs, "s4");
            long a1 = requireRegister(regs, "a1");
            long pc = requireRegister(regs, "pc");
            long pageStart = (s0 + INLINE_PAGE_OFFSET) & 0xffffffffL;
            byte[] dump = readMemory(pageStart, INLINE_DUMP_SIZE);

            JSONObject capture = new JSONObject();
            capture.put("event", "manual_freeze_capture");
            capture.put("time_ms", System.currentTimeMillis());
            capture.put("target", "127.0.0.1:" + connectedPort);
            capture.put("connection_reused", reused);
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
            updateNotification("수동 캡처 완료 · 최근 로그 복사");
        } catch (Exception e) {
            writeFailure("manual_capture_failed", e);
            closeConnection();
            updateNotification("캡처 실패 · 자동 재연결 대기");
        }
    }

    private byte[] readMemory(long address, int size) throws Exception {
        JSONObject params = new JSONObject();
        params.put("address", address & 0xffffffffL);
        params.put("size", size);
        JSONObject response = request(writer, reader, nextRequestId++,
                rawCommand("memory.read", params), RESPONSE_TIMEOUT_MS);
        String encoded = rawResponse(response).optString("base64", "");
        if (encoded.isEmpty()) throw new IllegalStateException("memory.read 응답에 base64 데이터가 없습니다");
        byte[] decoded = Base64.decode(encoded, Base64.DEFAULT);
        if (decoded.length != size) {
            throw new IllegalStateException("memory.read 크기 불일치: got=" + decoded.length + " want=" + size);
        }
        return decoded;
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
        try (BufferedWriter fileWriter = new BufferedWriter(new FileWriter(out, false))) {
            fileWriter.write(line);
            fileWriter.newLine();
        } catch (Exception ignored) {}
    }

    private void writeEvent(String event, String key, Object value) {
        try {
            JSONObject obj = new JSONObject();
            obj.put("event", event);
            obj.put("time_ms", System.currentTimeMillis());
            obj.put("target", connectedPort > 0 ? "127.0.0.1:" + connectedPort : "not_connected");
            obj.put(key, value);
            writeCapture(obj.toString());
        } catch (Exception ignored) {}
    }

    private void writeFailure(String event, Exception error) {
        try {
            JSONObject obj = new JSONObject();
            obj.put("event", event);
            obj.put("time_ms", System.currentTimeMillis());
            obj.put("target", connectedPort > 0 ? "127.0.0.1:" + connectedPort : "not_connected");
            obj.put("wanted_connection", prefs().getBoolean(PREF_WANT_CONNECTED, false));
            obj.put("saved_port", prefs().getInt(PREF_PORT, DEFAULT_PORT));
            obj.put("process_alive", process != null && process.isAlive());
            obj.put("connected_flag", connected);
            obj.put("error", message(error));
            writeCapture(obj.toString());
        } catch (Exception ignored) {}
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

    private synchronized void closeConnection() {
        connected = false;
        connectedPort = -1;
        try { if (writer != null) writer.close(); } catch (Exception ignored) {}
        try { if (reader != null) reader.close(); } catch (Exception ignored) {}
        writer = null;
        reader = null;
        Process p = process;
        process = null;
        if (p != null) p.destroy();
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
                .setOngoing(prefs().getBoolean(PREF_WANT_CONNECTED, false))
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
        closeConnection();
        supervisor.shutdownNow();
        worker.shutdown();
        super.onDestroy();
    }

    @Override
    public IBinder onBind(Intent intent) { return null; }
}
