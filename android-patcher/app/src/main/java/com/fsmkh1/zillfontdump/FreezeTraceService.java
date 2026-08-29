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
 * One-shot PPSSPP boundary tracer.
 *
 * Rather than re-collecting the already established runaway scanner, this
 * tracer follows the single producer invocation that creates s0+0x3C0 and
 * records the value at three boundaries: backend entry/return, wrapper return,
 * and immediately before the slot store.  This avoids wrapper-by-wrapper APK
 * replays and does not require waiting for the later runaway scan.
 */
public final class FreezeTraceService extends Service {
    public static final String ACTION_START = "com.fsmkh1.zillfontdump.START_FREEZE_TRACE";
    public static final String ACTION_STOP = "com.fsmkh1.zillfontdump.STOP_FREEZE_TRACE";
    public static final String EXTRA_PORT = "port";
    public static final String TRACE_FILE = "ppsspp-freeze-trace.jsonl";

    private static final String CHANNEL_ID = "freeze_trace";
    private static final int NOTIFICATION_ID = 21010;
    private static final int RESPONSE_TIMEOUT_MS = 6000;
    private static final int BREAK_WAIT_SECONDS = 300;
    private static final int BREAK_WAIT_TIMEOUT_MS = 305000;
    private static final int MAX_EVENTS = 16;

    // Raw runtime disassembly authority: docs/audit/fixtures/runtime-20260829-freeze-disassembly.txt.
    private static final long PRODUCER_CALL = 0x0886C938L;
    private static final long PRODUCER_POST_CALL = 0x0886C940L;
    private static final long PRODUCER_STORE = 0x0886C948L;
    private static final long BACKEND_ENTRY = 0x08A23064L;
    private static final long SLOT_OFFSET = 0x3C0L;

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
                    updateNotification("연결/캡처 실패 · 재연결 대기");
                    sleep(1000);
                }
            }
        }
        writeEvent("trace_stop", "reason", "stop_requested_or_capture_complete");
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
            updateNotification("경계 캡처 대기 · 게임을 진행하세요");

            int[] requestId = new int[]{1};
            JSONObject capture = new JSONObject();
            capture.put("event", "pointer_boundary_capture");
            capture.put("time_ms", System.currentTimeMillis());
            capture.put("producer_call", hex32(PRODUCER_CALL));
            capture.put("backend_entry", hex32(BACKEND_ENTRY));
            capture.put("producer_post_call", hex32(PRODUCER_POST_CALL));
            capture.put("producer_store", hex32(PRODUCER_STORE));

            // Scope the entire trace to the exact producer invocation first.
            addBreakpoint(writer, reader, requestId, PRODUCER_CALL);
            JSONObject producerRegs = waitForExpectedPc(writer, reader, requestId,
                    new long[]{PRODUCER_CALL}, capture, "producer_call_hit");
            capture.put("producer_call_gpr", selectedRegisters(producerRegs));
            removeBreakpointQuiet(writer, reader, requestId, PRODUCER_CALL);

            // Arm both boundaries before resuming.  If the backend is bypassed,
            // the post-call breakpoint wins and we still learn that fact in one run.
            addBreakpoint(writer, reader, requestId, BACKEND_ENTRY);
            addBreakpoint(writer, reader, requestId, PRODUCER_POST_CALL);
            resume(writer, reader, requestId);

            JSONObject firstBoundaryRegs = waitForExpectedPc(writer, reader, requestId,
                    new long[]{BACKEND_ENTRY, PRODUCER_POST_CALL}, capture, "first_boundary_hit");
            long firstPc = findRegister(firstBoundaryRegs, "pc");
            boolean backendSeen = firstPc == BACKEND_ENTRY;
            capture.put("backend_seen", backendSeen);

            if (backendSeen) {
                capture.put("backend_entry_gpr", selectedRegisters(firstBoundaryRegs));
                long backendReturn = findRegister(firstBoundaryRegs, "ra");
                if (backendReturn < 0) throw new IllegalStateException("backend entry에서 ra를 읽지 못했습니다");
                backendReturn &= 0xffffffffL;
                capture.put("backend_return_pc", hex32(backendReturn));

                removeBreakpointQuiet(writer, reader, requestId, BACKEND_ENTRY);
                addBreakpoint(writer, reader, requestId, backendReturn);
                resume(writer, reader, requestId);

                JSONObject backendReturnRegs = waitForExpectedPc(writer, reader, requestId,
                        new long[]{backendReturn, PRODUCER_POST_CALL}, capture, "backend_return_hit");
                long returnHitPc = findRegister(backendReturnRegs, "pc");
                if (returnHitPc == backendReturn) {
                    capture.put("backend_return_gpr", selectedRegisters(backendReturnRegs));
                    removeBreakpointQuiet(writer, reader, requestId, backendReturn);
                    resume(writer, reader, requestId);
                    JSONObject postCallRegs = waitForExpectedPc(writer, reader, requestId,
                            new long[]{PRODUCER_POST_CALL}, capture, "producer_post_call_hit");
                    capture.put("producer_post_call_gpr", selectedRegisters(postCallRegs));
                } else {
                    // Defensive fallback if the backend's ra aliases/skips directly
                    // to the already-armed producer boundary.
                    capture.put("backend_return_boundary_skipped", true);
                    capture.put("producer_post_call_gpr", selectedRegisters(backendReturnRegs));
                }
            } else {
                capture.put("backend_not_seen_before_wrapper_return", true);
                capture.put("producer_post_call_gpr", selectedRegisters(firstBoundaryRegs));
                removeBreakpointQuiet(writer, reader, requestId, BACKEND_ENTRY);
            }

            removeBreakpointQuiet(writer, reader, requestId, PRODUCER_POST_CALL);

            // The raw log proves 0x0886C948 is the actual sw v0,0x3C0(a0).
            // At this breakpoint the preceding sll/addu have already computed a0.
            addBreakpoint(writer, reader, requestId, PRODUCER_STORE);
            resume(writer, reader, requestId);
            JSONObject storeRegs = waitForExpectedPc(writer, reader, requestId,
                    new long[]{PRODUCER_STORE}, capture, "producer_store_hit");
            capture.put("producer_store_gpr", selectedRegisters(storeRegs));

            long storeBase = findRegister(storeRegs, "a0");
            long storeValue = findRegister(storeRegs, "v0");
            if (storeBase >= 0) {
                long destination = (storeBase + SLOT_OFFSET) & 0xffffffffL;
                capture.put("producer_store_destination", hex32(destination));
                tryMemoryRead(writer, reader, requestId, capture,
                        "producer_store_old_value", destination, 4);
            }
            if (storeValue >= 0) capture.put("producer_store_v0", hex32(storeValue));

            removeBreakpointQuiet(writer, reader, requestId, PRODUCER_STORE);
            resumeQuiet(writer, reader, requestId);

            appendEvent(capture.toString());
            updateNotification("경계 캡처 완료 · 로그 복사 가능");
            stopRequested = true;
        } finally {
            Process p = process;
            process = null;
            if (p != null) p.destroy();
        }
    }

    private static JSONObject waitForExpectedPc(BufferedWriter writer, BufferedReader reader,
                                                int[] requestId, long[] expected,
                                                JSONObject capture, String label) throws Exception {
        for (int attempt = 0; attempt < 8; attempt++) {
            JSONObject wait = bridgeWait(writer, reader, requestId);
            JSONObject regs = getAllRegs(writer, reader, requestId);
            long pc = findRegister(regs, "pc");
            if (pc >= 0) pc &= 0xffffffffL;
            for (long target : expected) {
                if (pc == target) {
                    capture.put(label, hex32(pc));
                    capture.put(label + "_stepping", wait.optJSONObject("result"));
                    return regs;
                }
            }
            capture.put(label + "_unexpected_pc_" + attempt, pc < 0 ? "unknown" : hex32(pc));
            resume(writer, reader, requestId);
        }
        throw new IllegalStateException(label + " expected breakpoint를 찾지 못했습니다");
    }

    private static JSONObject getAllRegs(BufferedWriter writer, BufferedReader reader,
                                         int[] requestId) throws Exception {
        return request(writer, reader, requestId[0]++,
                rawCommand("cpu.getAllRegs", new JSONObject()), RESPONSE_TIMEOUT_MS);
    }

    private static void addBreakpoint(BufferedWriter writer, BufferedReader reader,
                                      int[] requestId, long address) throws Exception {
        JSONObject params = new JSONObject();
        params.put("address", address & 0xffffffffL);
        params.put("enabled", true);
        request(writer, reader, requestId[0]++,
                rawCommand("cpu.breakpoint.add", params), RESPONSE_TIMEOUT_MS);
    }

    private static void removeBreakpointQuiet(BufferedWriter writer, BufferedReader reader,
                                              int[] requestId, long address) {
        try {
            JSONObject params = new JSONObject();
            params.put("address", address & 0xffffffffL);
            request(writer, reader, requestId[0]++,
                    rawCommand("cpu.breakpoint.remove", params), RESPONSE_TIMEOUT_MS);
        } catch (Exception ignored) {
        }
    }

    private static JSONObject bridgeWait(BufferedWriter writer, BufferedReader reader,
                                         int[] requestId) throws Exception {
        JSONObject command = new JSONObject();
        command.put("command", "wait");
        command.put("event", "cpu.stepping");
        command.put("buffered", true);
        command.put("timeout", BREAK_WAIT_SECONDS);
        return request(writer, reader, requestId[0]++, command, BREAK_WAIT_TIMEOUT_MS);
    }

    private static void resume(BufferedWriter writer, BufferedReader reader,
                               int[] requestId) throws Exception {
        JSONObject command = new JSONObject();
        command.put("command", "resume");
        command.put("timeout", 6);
        request(writer, reader, requestId[0]++, command, RESPONSE_TIMEOUT_MS);
    }

    private static void resumeQuiet(BufferedWriter writer, BufferedReader reader,
                                    int[] requestId) {
        try {
            resume(writer, reader, requestId);
        } catch (Exception ignored) {
        }
    }

    private static void tryMemoryRead(BufferedWriter writer, BufferedReader reader,
                                      int[] requestId, JSONObject evidence,
                                      String key, long address, int size) {
        try {
            JSONObject params = new JSONObject();
            params.put("address", address & 0xffffffffL);
            params.put("size", size);
            JSONObject memory = request(writer, reader, requestId[0]++,
                    rawCommand("memory.read", params), RESPONSE_TIMEOUT_MS);
            evidence.put(key + "_start", hex32(address));
            evidence.put(key, rawResponse(memory));
        } catch (Exception error) {
            try {
                evidence.put(key + "_start", hex32(address));
                evidence.put(key + "_error", message(error));
            } catch (Exception ignored) {
            }
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
                .setContentTitle("질올 PPSSPP 경계 tracer")
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
