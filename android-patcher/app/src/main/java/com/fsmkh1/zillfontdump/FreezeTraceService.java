package com.fsmkh1.zillfontdump;

import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.app.PendingIntent;
import android.app.Service;
import android.content.Intent;
import android.os.IBinder;

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
import java.util.ArrayDeque;
import java.util.Deque;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeoutException;

/**
 * Minimal PPSSPP freeze tracer for the current unresolved allocator-backend
 * question. Earlier scanner/producer/wrapper disassembly is intentionally not
 * re-collected; that evidence is retained in docs/audit/A-046..A-049.
 */
public final class FreezeTraceService extends Service {
    public static final String ACTION_START = "com.fsmkh1.zillfontdump.START_FREEZE_TRACE";
    public static final String ACTION_STOP = "com.fsmkh1.zillfontdump.STOP_FREEZE_TRACE";
    public static final String EXTRA_PORT = "port";
    public static final String TRACE_FILE = "ppsspp-freeze-trace.jsonl";

    private static final String CHANNEL_ID = "freeze_trace";
    private static final int NOTIFICATION_ID = 21010;
    private static final int SAMPLE_INTERVAL_MS = 500;
    private static final int RESPONSE_TIMEOUT_MS = 1500;
    private static final int MAX_EVENTS = 24;

    private static final long HOT_PC_MIN = 0x08966200L;
    private static final long HOT_PC_MAX = 0x08966260L;
    private static final int HOT_LOOP_SAMPLES = 3;
    private static final long HOT_A1_MIN_ADVANCE = 0x1000L;

    // Current unresolved target only. Do not accumulate earlier disassembly.
    private static final long BACKEND_DISASM_START = 0x08A23064L;
    private static final int BACKEND_DISASM_COUNT = 256;

    // Small correlation read only; the full producer/object window is already archived.
    private static final long POINTER_FIELD_OFFSET = 0x3C0L;

    private final ExecutorService worker = Executors.newSingleThreadExecutor();
    private final Deque<String> events = new ArrayDeque<>();
    private volatile boolean stopRequested;
    private volatile Process process;

    @Override
    public void onCreate() {
        super.onCreate();
        NotificationManager manager = getSystemService(NotificationManager.class);
        if (manager != null) {
            manager.createNotificationChannel(new NotificationChannel(
                    CHANNEL_ID, "PPSSPP 프리징 기록", NotificationManager.IMPORTANCE_LOW));
        }
    }

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        if (intent != null && ACTION_STOP.equals(intent.getAction())) {
            writeEvent("stop_requested", "source", "ui");
            stopRequested = true;
            Process p = process;
            if (p != null) p.destroy();
            stopForeground(STOP_FOREGROUND_REMOVE);
            stopSelf();
            return START_NOT_STICKY;
        }

        final int port = intent == null ? 34500 : intent.getIntExtra(EXTRA_PORT, 34500);
        stopRequested = false;
        startForeground(NOTIFICATION_ID, notification("PPSSPP 연결 대기 중 · " + port));
        worker.execute(() -> runTrace(port));
        return START_NOT_STICKY;
    }

    private void runTrace(int port) {
        writeEvent("trace_start", "port", port);
        while (!stopRequested) {
            try {
                traceConnectedSession(port);
            } catch (Exception e) {
                if (!stopRequested) {
                    writeEvent("connection_lost", "error", message(e));
                    updateNotification("연결 끊김 · 재연결 대기");
                    sleep(1000);
                }
            }
        }
        writeEvent("trace_stop", "reason", "stop_requested_or_service_destroyed");
    }

    private void traceConnectedSession(int port) throws Exception {
        File executable = new File(getApplicationInfo().nativeLibraryDir, "libzill.so");
        if (!executable.isFile()) throw new IllegalStateException("내장 debugger 실행파일을 찾을 수 없습니다");

        ProcessBuilder builder = new ProcessBuilder(
                executable.getAbsolutePath(), "ppsspp-debugger",
                "--host", "127.0.0.1",
                "--port", Integer.toString(port),
                "--timeout", "6",
                "--connect-timeout", "3");
        builder.directory(getFilesDir());
        builder.redirectErrorStream(true);
        process = builder.start();

        try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream(), StandardCharsets.UTF_8));
             BufferedWriter writer = new BufferedWriter(new OutputStreamWriter(process.getOutputStream(), StandardCharsets.UTF_8))) {
            JSONObject ready = readObject(reader, "debugger handshake", 4000);
            if (!"ready".equals(ready.optString("event"))) {
                throw new IllegalStateException("unexpected handshake: " + ready);
            }
            writeEvent("connected", "target", "127.0.0.1:" + port);
            updateNotification("PPSSPP 기록 중 · allocator backend만 수집");

            int requestId = 1;
            int hotPcCount = 0;
            long hotA1Start = -1;
            boolean evidenceCaptured = false;

            while (!stopRequested) {
                JSONObject cpu = request(writer, reader, requestId++,
                        rawCommand("cpu.status", new JSONObject()), RESPONSE_TIMEOUT_MS);
                JSONObject regs = request(writer, reader, requestId++,
                        rawCommand("cpu.getAllRegs", new JSONObject()), RESPONSE_TIMEOUT_MS);

                JSONObject rawCpu = rawResponse(cpu);
                long pc = findRegister(regs, "pc");
                long a1 = findRegister(regs, "a1");
                long s0 = findRegister(regs, "s0");
                if (pc < 0) pc = rawCpu.optLong("pc", -1);
                if (pc >= 0) pc &= 0xffffffffL;

                if (isHotPc(pc)) {
                    if (hotPcCount == 0) hotA1Start = a1;
                    hotPcCount++;
                } else {
                    hotPcCount = 0;
                    hotA1Start = -1;
                }

                if (!evidenceCaptured
                        && hotPcCount >= HOT_LOOP_SAMPLES
                        && a1 >= 0 && hotA1Start >= 0
                        && unsignedAdvance(hotA1Start, a1) >= HOT_A1_MIN_ADVANCE) {
                    JSONObject evidence = new JSONObject();
                    evidence.put("event", "allocator_backend_capture");
                    evidence.put("time_ms", System.currentTimeMillis());
                    evidence.put("pc", hex32(pc));
                    evidence.put("a1_start", hex32(hotA1Start));
                    evidence.put("a1_now", hex32(a1));
                    evidence.put("a1_advance", hex32(unsignedAdvance(hotA1Start, a1)));
                    evidence.put("gpr", selectedRegisters(regs));

                    tryDisasm(writer, reader, requestId++, evidence,
                            "allocator_backend_disasm", BACKEND_DISASM_START, BACKEND_DISASM_COUNT);

                    if (s0 >= 0) {
                        long pointerField = (s0 + POINTER_FIELD_OFFSET) & 0xffffffffL;
                        evidence.put("s0", hex32(s0));
                        evidence.put("s0_pointer_field_address", hex32(pointerField));
                        tryMemoryRead(writer, reader, requestId++, evidence,
                                "s0_pointer_field", pointerField, 4);
                    }

                    appendEvent(evidence.toString());
                    evidenceCaptured = true;
                    updateNotification("allocator backend 캡처 완료 · 로그 복사 가능");
                }

                sleep(SAMPLE_INTERVAL_MS);
            }
        } finally {
            Process p = process;
            process = null;
            if (p != null) p.destroy();
        }
    }

    private static void tryDisasm(BufferedWriter writer, BufferedReader reader, int requestId,
                                  JSONObject evidence, String key, long address, int count) {
        try {
            JSONObject params = new JSONObject();
            params.put("address", address);
            params.put("count", count);
            params.put("displaySymbols", true);
            params.put("compact", false);
            JSONObject disasm = request(writer, reader, requestId,
                    rawCommand("memory.disasm", params), RESPONSE_TIMEOUT_MS);
            evidence.put(key, rawResponse(disasm));
        } catch (Exception error) {
            try { evidence.put(key + "_error", message(error)); } catch (Exception ignored) {}
        }
    }

    private static void tryMemoryRead(BufferedWriter writer, BufferedReader reader, int requestId,
                                      JSONObject evidence, String key, long address, int size) {
        try {
            JSONObject params = new JSONObject();
            params.put("address", address & 0xffffffffL);
            params.put("size", size);
            JSONObject memory = request(writer, reader, requestId,
                    rawCommand("memory.read", params), RESPONSE_TIMEOUT_MS);
            evidence.put(key + "_start", hex32(address));
            evidence.put(key, rawResponse(memory));
        } catch (Exception error) {
            try {
                evidence.put(key + "_start", hex32(address));
                evidence.put(key + "_error", message(error));
            } catch (Exception ignored) {}
        }
    }

    private synchronized void appendEvent(String line) {
        events.addLast(line);
        while (events.size() > MAX_EVENTS) events.removeFirst();
        persistEvents();
    }

    private synchronized void writeEvent(String event, String key, Object value) {
        try {
            JSONObject obj = new JSONObject();
            obj.put("event", event);
            obj.put("time_ms", System.currentTimeMillis());
            obj.put(key, value);
            appendEvent(obj.toString());
        } catch (Exception ignored) {
        }
    }

    private void persistEvents() {
        File out = new File(getFilesDir(), TRACE_FILE);
        try (BufferedWriter writer = new BufferedWriter(new FileWriter(out, false))) {
            for (String line : events) {
                writer.write(line);
                writer.newLine();
            }
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
                .setContentTitle("질올 PPSSPP 프리징 tracer")
                .setContentText(text)
                .setContentIntent(pending)
                .setOngoing(true)
                .build();
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
            throw new IllegalStateException(error == null ? response.toString()
                    : error.optString("message", response.toString()));
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

    private static JSONObject selectedRegisters(JSONObject response) throws Exception {
        JSONObject selected = new JSONObject();
        String[] names = new String[]{
                "v0", "v1", "a0", "a1", "a2", "a3",
                "t0", "t1", "t2", "t3",
                "s0", "s1", "s2", "s3", "s4", "s5", "s6", "s7",
                "sp", "fp", "ra", "pc"
        };
        for (String name : names) {
            long value = findRegister(response, name);
            if (value >= 0) selected.put(name, hex32(value));
        }
        return selected;
    }

    private static boolean isHotPc(long pc) {
        return pc >= HOT_PC_MIN && pc <= HOT_PC_MAX;
    }

    private static long unsignedAdvance(long start, long now) {
        return (now - start) & 0xffffffffL;
    }

    private static long findRegister(JSONObject response, String wanted) {
        JSONObject raw = rawResponse(response);
        JSONArray categories = raw.optJSONArray("categories");
        if (categories == null) return -1;
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
        return -1;
    }

    private static String hex32(long value) {
        return String.format("0x%08X", value & 0xffffffffL);
    }

    private static String message(Exception e) {
        String m = e.getMessage();
        return (m == null || m.trim().isEmpty()) ? e.getClass().getSimpleName() : m;
    }

    private static void sleep(long millis) {
        try {
            Thread.sleep(millis);
        } catch (InterruptedException ignored) {
            Thread.currentThread().interrupt();
        }
    }

    @Override
    public void onDestroy() {
        if (!stopRequested) writeEvent("service_destroyed", "source", "android");
        stopRequested = true;
        Process p = process;
        if (p != null) p.destroy();
        worker.shutdownNow();
        super.onDestroy();
    }

    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }
}
