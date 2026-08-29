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
import java.util.HashSet;
import java.util.Set;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeoutException;

/**
 * One-shot PPSSPP pointer-boundary tracer.
 *
 * It does not assume the first producer hit is the failing message. It considers
 * only s4==0 (+0x3C0) producer candidates, follows each candidate through the
 * allocator-backend boundary, and commits a trace only when the actual pre-store
 * v0 equals the repeatedly observed failing field value 0x8C4C89A4.
 *
 * Established scanner/producer disassembly stays in docs/audit fixtures and is
 * intentionally not re-collected here.
 */
public final class FreezeTraceService extends Service {
    public static final String ACTION_START = "com.fsmkh1.zillfontdump.START_FREEZE_TRACE";
    public static final String ACTION_STOP = "com.fsmkh1.zillfontdump.STOP_FREEZE_TRACE";
    public static final String EXTRA_PORT = "port";
    public static final String TRACE_FILE = "ppsspp-freeze-trace.jsonl";

    private static final String CHANNEL_ID = "freeze_trace";
    private static final int NOTIFICATION_ID = 21010;
    private static final int RESPONSE_TIMEOUT_MS = 6000;
    private static final int BREAK_POLL_SECONDS = 2;
    private static final int BREAK_POLL_TIMEOUT_MS = 3500;
    private static final int MAX_EVENTS = 16;

    // Raw runtime disassembly authority:
    // docs/audit/fixtures/runtime-20260829-freeze-disassembly.txt
    private static final long PRODUCER_CALL = 0x0886C938L;
    private static final long PRODUCER_POST_CALL = 0x0886C940L;
    private static final long PRODUCER_STORE = 0x0886C948L;
    private static final long BACKEND_ENTRY = 0x08A23064L;
    private static final long SLOT_OFFSET = 0x3C0L;

    // Repeatedly observed value at *(s0+0x3C0) during the reproduced runaway scan.
    // This is a correlation key, not a claim that the value is inherently invalid.
    private static final long OBSERVED_FAILING_FIELD_VALUE = 0x8C4C89A4L;

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
            updateNotification("안전 종료 중 · breakpoint 정리 대기");
            // Do not destroy the bridge here. The worker owns breakpoint cleanup and
            // CPU resume. Break waits are short-polled so stop is noticed promptly.
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
        try {
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
        } finally {
            writeEvent("trace_stop", "reason", "stop_requested_or_capture_complete");
            stopForeground(STOP_FOREGROUND_REMOVE);
            stopSelf();
        }
    }

    private void traceConnectedSession(int port) throws Exception {
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

        Set<Long> activeBreakpoints = new HashSet<>();
        int[] requestId = new int[]{1};

        try (BufferedReader reader = new BufferedReader(new InputStreamReader(
                     process.getInputStream(), StandardCharsets.UTF_8));
             BufferedWriter writer = new BufferedWriter(new OutputStreamWriter(
                     process.getOutputStream(), StandardCharsets.UTF_8))) {

            JSONObject ready = readObject(reader, "debugger handshake", 4000);
            if (!"ready".equals(ready.optString("event"))) {
                throw new IllegalStateException("unexpected handshake: " + ready);
            }
            writeEvent("connected", "target", "127.0.0.1:" + port);
            updateNotification("후보 경계 추적 중 · 문제 장면까지 진행하세요");

            try {
                // Own the stepping state from a known point before installing breakpoints.
                pause(writer, reader, requestId);
                drainBridgeEvents(writer, reader, requestId);

                addBreakpoint(writer, reader, requestId, activeBreakpoints, PRODUCER_CALL);

                int candidateIndex = 0;
                while (!stopRequested) {
                    drainBridgeEvents(writer, reader, requestId);
                    resume(writer, reader, requestId);
                    JSONObject producerRegs = waitForPc(writer, reader, requestId,
                            new long[]{PRODUCER_CALL});

                    long s4 = findRegister(producerRegs, "s4");
                    if (s4 != 0) {
                        // Only s4==0 writes the base object's +0x3C0 slot.
                        continue;
                    }

                    candidateIndex++;
                    JSONObject candidate = new JSONObject();
                    candidate.put("candidate_index", candidateIndex);
                    candidate.put("producer_call", hex32(PRODUCER_CALL));
                    candidate.put("producer_call_gpr", selectedRegisters(producerRegs));
                    candidate.put("backend_entry", hex32(BACKEND_ENTRY));
                    candidate.put("producer_post_call", hex32(PRODUCER_POST_CALL));
                    candidate.put("producer_store", hex32(PRODUCER_STORE));

                    removeBreakpoint(writer, reader, requestId, activeBreakpoints, PRODUCER_CALL);
                    addBreakpoint(writer, reader, requestId, activeBreakpoints, BACKEND_ENTRY);
                    addBreakpoint(writer, reader, requestId, activeBreakpoints, PRODUCER_POST_CALL);

                    drainBridgeEvents(writer, reader, requestId);
                    resume(writer, reader, requestId);
                    JSONObject firstBoundaryRegs = waitForPc(writer, reader, requestId,
                            new long[]{BACKEND_ENTRY, PRODUCER_POST_CALL});
                    long firstPc = findRegister(firstBoundaryRegs, "pc");
                    boolean backendSeen = firstPc == BACKEND_ENTRY;
                    candidate.put("backend_seen", backendSeen);

                    if (backendSeen) {
                        candidate.put("backend_entry_gpr", selectedRegisters(firstBoundaryRegs));
                        long backendReturn = findRegister(firstBoundaryRegs, "ra");
                        if (backendReturn < 0) {
                            throw new IllegalStateException("backend entry에서 ra를 읽지 못했습니다");
                        }
                        backendReturn &= 0xffffffffL;
                        candidate.put("backend_return_pc", hex32(backendReturn));

                        removeBreakpoint(writer, reader, requestId, activeBreakpoints, BACKEND_ENTRY);
                        addBreakpoint(writer, reader, requestId, activeBreakpoints, backendReturn);

                        drainBridgeEvents(writer, reader, requestId);
                        resume(writer, reader, requestId);
                        JSONObject returnRegs = waitForPc(writer, reader, requestId,
                                new long[]{backendReturn, PRODUCER_POST_CALL});
                        long returnPc = findRegister(returnRegs, "pc");

                        if (returnPc == backendReturn) {
                            candidate.put("backend_return_gpr", selectedRegisters(returnRegs));
                            removeBreakpoint(writer, reader, requestId,
                                    activeBreakpoints, backendReturn);

                            drainBridgeEvents(writer, reader, requestId);
                            resume(writer, reader, requestId);
                            JSONObject postCallRegs = waitForPc(writer, reader, requestId,
                                    new long[]{PRODUCER_POST_CALL});
                            candidate.put("producer_post_call_gpr",
                                    selectedRegisters(postCallRegs));
                        } else {
                            candidate.put("backend_return_boundary_skipped", true);
                            candidate.put("producer_post_call_gpr",
                                    selectedRegisters(returnRegs));
                        }
                    } else {
                        candidate.put("backend_not_seen_before_wrapper_return", true);
                        candidate.put("producer_post_call_gpr",
                                selectedRegisters(firstBoundaryRegs));
                        removeBreakpoint(writer, reader, requestId,
                                activeBreakpoints, BACKEND_ENTRY);
                    }

                    removeBreakpoint(writer, reader, requestId,
                            activeBreakpoints, PRODUCER_POST_CALL);

                    addBreakpoint(writer, reader, requestId, activeBreakpoints, PRODUCER_STORE);
                    drainBridgeEvents(writer, reader, requestId);
                    resume(writer, reader, requestId);
                    JSONObject storeRegs = waitForPc(writer, reader, requestId,
                            new long[]{PRODUCER_STORE});

                    long storeS4 = findRegister(storeRegs, "s4");
                    long storeBase = findRegister(storeRegs, "a0");
                    long storeValue = findRegister(storeRegs, "v0");

                    // C948 is before the store and before C94C increments s4. If this
                    // invariant fails, do not silently correlate the candidate.
                    if (storeS4 != 0) {
                        throw new IllegalStateException(
                                "producer store에서 s4==0 invariant가 깨졌습니다: " + hex32(storeS4));
                    }

                    candidate.put("producer_store_gpr", selectedRegisters(storeRegs));
                    if (storeBase >= 0) {
                        long destination = (storeBase + SLOT_OFFSET) & 0xffffffffL;
                        candidate.put("producer_store_destination", hex32(destination));
                        tryMemoryRead(writer, reader, requestId, candidate,
                                "producer_store_old_value", destination, 4);
                    }
                    if (storeValue >= 0) {
                        candidate.put("producer_store_v0", hex32(storeValue));
                    }

                    removeBreakpoint(writer, reader, requestId,
                            activeBreakpoints, PRODUCER_STORE);

                    if (storeValue == OBSERVED_FAILING_FIELD_VALUE) {
                        candidate.put("event", "pointer_boundary_capture");
                        candidate.put("time_ms", System.currentTimeMillis());
                        candidate.put("matched_observed_failing_field_value", true);
                        candidate.put("observed_failing_field_value",
                                hex32(OBSERVED_FAILING_FIELD_VALUE));
                        appendEvent(candidate.toString());
                        updateNotification("문제 경계 캡처 완료 · 로그 복사 가능");
                        stopRequested = true;
                        break;
                    }

                    // Normal/nonmatching +0x3C0 candidate: discard it rather than
                    // growing the log, then continue to the next producer invocation.
                    addBreakpoint(writer, reader, requestId,
                            activeBreakpoints, PRODUCER_CALL);
                }
            } finally {
                // Best-effort cleanup on normal completion, timeout, user stop,
                // or analysis exception, while the same bridge streams are alive.
                cleanupBreakpoints(writer, reader, requestId, activeBreakpoints);
                resumeQuiet(writer, reader, requestId);
            }
        } finally {
            Process p = process;
            process = null;
            if (p != null) {
                p.destroy();
            }
        }
    }

    private JSONObject waitForPc(BufferedWriter writer, BufferedReader reader,
                                 int[] requestId, long[] expected) throws Exception {
        while (!stopRequested) {
            try {
                bridgeWait(writer, reader, requestId);
            } catch (Exception e) {
                if (isWaitTimeout(e)) {
                    continue;
                }
                throw e;
            }
            JSONObject regs = getAllRegs(writer, reader, requestId);
            long pc = findRegister(regs, "pc");
            if (pc >= 0) {
                pc &= 0xffffffffL;
            }
            for (long target : expected) {
                if (pc == target) {
                    return regs;
                }
            }

            // A non-target stepping event may be a UI/debugger transition. Drain
            // it, resume, and keep waiting without logging noisy samples.
            drainBridgeEvents(writer, reader, requestId);
            resume(writer, reader, requestId);
        }
        throw new InterruptedException("stop requested");
    }

    private static boolean isWaitTimeout(Exception e) {
        String m = message(e);
        return m.contains("timed out waiting for") || m.contains("응답이");
    }

    private static JSONObject getAllRegs(BufferedWriter writer, BufferedReader reader,
                                         int[] requestId) throws Exception {
        return request(writer, reader, requestId[0]++,
                rawCommand("cpu.getAllRegs", new JSONObject()), RESPONSE_TIMEOUT_MS);
    }

    private static void pause(BufferedWriter writer, BufferedReader reader,
                              int[] requestId) throws Exception {
        JSONObject command = new JSONObject();
        command.put("command", "pause");
        command.put("timeout", 6);
        request(writer, reader, requestId[0]++, command, RESPONSE_TIMEOUT_MS);
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

    private static void addBreakpoint(BufferedWriter writer, BufferedReader reader,
                                      int[] requestId, Set<Long> active,
                                      long address) throws Exception {
        JSONObject params = new JSONObject();
        params.put("address", address & 0xffffffffL);
        params.put("enabled", true);
        request(writer, reader, requestId[0]++,
                rawCommand("cpu.breakpoint.add", params), RESPONSE_TIMEOUT_MS);
        active.add(address & 0xffffffffL);
    }

    private static void removeBreakpoint(BufferedWriter writer, BufferedReader reader,
                                         int[] requestId, Set<Long> active,
                                         long address) throws Exception {
        JSONObject params = new JSONObject();
        params.put("address", address & 0xffffffffL);
        request(writer, reader, requestId[0]++,
                rawCommand("cpu.breakpoint.remove", params), RESPONSE_TIMEOUT_MS);
        active.remove(address & 0xffffffffL);
    }

    private static void cleanupBreakpoints(BufferedWriter writer, BufferedReader reader,
                                           int[] requestId, Set<Long> active) {
        Long[] snapshot = active.toArray(new Long[0]);
        for (long address : snapshot) {
            try {
                JSONObject params = new JSONObject();
                params.put("address", address & 0xffffffffL);
                request(writer, reader, requestId[0]++,
                        rawCommand("cpu.breakpoint.remove", params), RESPONSE_TIMEOUT_MS);
                active.remove(address & 0xffffffffL);
            } catch (Exception ignored) {
            }
        }
    }

    private static void drainBridgeEvents(BufferedWriter writer, BufferedReader reader,
                                          int[] requestId) throws Exception {
        JSONObject command = new JSONObject();
        command.put("command", "drain");
        command.put("limit", 1024);
        request(writer, reader, requestId[0]++, command, RESPONSE_TIMEOUT_MS);
    }

    private static void bridgeWait(BufferedWriter writer, BufferedReader reader,
                                   int[] requestId) throws Exception {
        JSONObject command = new JSONObject();
        command.put("command", "wait");
        command.put("event", "cpu.stepping");
        command.put("buffered", true);
        command.put("timeout", BREAK_POLL_SECONDS);
        request(writer, reader, requestId[0]++, command, BREAK_POLL_TIMEOUT_MS);
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
        while (events.size() > MAX_EVENTS) {
            events.removeFirst();
        }
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
        if (!trace.isFile()) {
            return "";
        }
        StringBuilder out = new StringBuilder();
        try (BufferedReader reader = new BufferedReader(new InputStreamReader(
                new FileInputStream(trace), StandardCharsets.UTF_8))) {
            String line;
            while ((line = reader.readLine()) != null) {
                out.append(line).append('\n');
            }
        }
        return out.toString();
    }

    private void updateNotification(String text) {
        NotificationManager manager = getSystemService(NotificationManager.class);
        if (manager != null) {
            manager.notify(NOTIFICATION_ID, notification(text));
        }
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
        while (!response.has("id")) {
            response = readObject(reader, "command " + id, timeoutMs);
        }
        if (!response.optBoolean("ok", false)) {
            JSONObject error = response.optJSONObject("error");
            throw new IllegalStateException(error == null
                    ? response.toString()
                    : error.optString("message", response.toString()));
        }
        return response;
    }

    private static JSONObject readObject(BufferedReader reader, String what,
                                         int timeoutMs) throws Exception {
        long deadline = System.currentTimeMillis() + timeoutMs;
        while (System.currentTimeMillis() < deadline) {
            if (reader.ready()) {
                String line = reader.readLine();
                if (line == null) {
                    throw new IllegalStateException(what + " 응답 전에 연결이 종료됐습니다");
                }
                return new JSONObject(line);
            }
            Thread.sleep(20);
        }
        throw new TimeoutException(what + " 응답이 " + timeoutMs + "ms 동안 없습니다");
    }

    private static JSONObject rawResponse(JSONObject response) {
        JSONObject result = response.optJSONObject("result");
        if (result == null) {
            return new JSONObject();
        }
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
            if (value >= 0) {
                selected.put(name, hex32(value));
            }
        }
        return selected;
    }

    private static long findRegister(JSONObject response, String wanted) {
        JSONObject raw = rawResponse(response);
        JSONArray categories = raw.optJSONArray("categories");
        if (categories == null) {
            return -1;
        }
        for (int i = 0; i < categories.length(); i++) {
            JSONObject category = categories.optJSONObject(i);
            if (category == null) {
                continue;
            }
            JSONArray names = category.optJSONArray("registerNames");
            JSONArray values = category.optJSONArray("uintValues");
            if (names == null || values == null) {
                continue;
            }
            int count = Math.min(names.length(), values.length());
            for (int j = 0; j < count; j++) {
                if (wanted.equals(names.optString(j))) {
                    return values.optLong(j, -1) & 0xffffffffL;
                }
            }
        }
        return -1;
    }

    private static String hex32(long value) {
        return String.format("0x%08X", value & 0xffffffffL);
    }

    private static String message(Exception e) {
        String m = e.getMessage();
        return (m == null || m.trim().isEmpty())
                ? e.getClass().getSimpleName()
                : m;
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
        stopRequested = true;
        // Do not destroy the bridge here before the worker has had a chance to
        // remove its breakpoints. runTrace/traceConnectedSession own shutdown.
        worker.shutdown();
        super.onDestroy();
    }

    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }
}
