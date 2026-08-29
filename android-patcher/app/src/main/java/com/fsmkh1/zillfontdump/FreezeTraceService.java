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
 * PPSSPP pointer-boundary tracer.
 *
 * Breakpoints are used only to stop the CPU at exact addresses. The tracer does
 * not subscribe to or wait for breakpoint/stepping events. Instead it polls
 * cpu.status, whose PPSSPP contract explicitly says it is cheap to poll and that
 * pc is accurate while stepping, then reads registers only after stepping=true.
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
    private static final int STATUS_POLL_MS = 20;
    private static final int MAX_EVENTS = 16;

    // Raw runtime disassembly authority:
    // docs/audit/fixtures/runtime-20260829-freeze-disassembly.txt
    private static final long PRODUCER_CALL = 0x0886C938L;
    private static final long PRODUCER_POST_CALL = 0x0886C940L;
    private static final long PRODUCER_STORE = 0x0886C948L;
    private static final long BACKEND_ENTRY = 0x08A23064L;
    private static final long SECOND_SCANNER_CALL = 0x0886C9B4L;
    private static final long SECOND_SCANNER_SKIP = 0x0886C9C4L;
    private static final long SLOT_OFFSET = 0x3C0L;

    // Correlation key from repeated reproduced runaway scans. This is not a
    // claim that the bit pattern is inherently invalid in every context.
    private static final long OBSERVED_FAILING_FIELD_VALUE = 0x8C4C89A4L;

    private final ExecutorService worker = Executors.newSingleThreadExecutor();
    private final Deque<String> events = new ArrayDeque<>();
    private volatile String latestCandidate;
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

            try {
                // No event subscription. Arm the exact producer address while the
                // game is running, then detect the stop solely through cpu.status.
                addBreakpoint(writer, reader, requestId, activeBreakpoints, PRODUCER_CALL);
                writeEvent("armed", "producer_call", hex32(PRODUCER_CALL));
                updateNotification("추적 준비 완료 · 문제 장면까지 진행하세요");

                int candidateIndex = 0;
                while (!stopRequested) {
                    JSONObject producerRegs = waitForStoppedPc(writer, reader, requestId,
                            new long[]{PRODUCER_CALL});

                    long s4 = requireRegister(producerRegs, "s4", "producer call");
                    if (s4 != 0) {
                        // Keep the same breakpoint armed. cpu.resume uses PPSSPP's
                        // skip-first protection, so execution advances past this hit.
                        resume(writer, reader, requestId);
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
                    candidate.put("second_scanner_call", hex32(SECOND_SCANNER_CALL));

                    removeBreakpoint(writer, reader, requestId, activeBreakpoints, PRODUCER_CALL);
                    addBreakpoint(writer, reader, requestId, activeBreakpoints, BACKEND_ENTRY);
                    addBreakpoint(writer, reader, requestId, activeBreakpoints, PRODUCER_POST_CALL);
                    resume(writer, reader, requestId);

                    JSONObject firstBoundaryRegs = waitForStoppedPc(writer, reader, requestId,
                            new long[]{BACKEND_ENTRY, PRODUCER_POST_CALL});
                    long firstPc = requireRegister(firstBoundaryRegs, "pc", "first boundary");
                    boolean backendSeen = firstPc == BACKEND_ENTRY;
                    candidate.put("backend_seen", backendSeen);

                    if (backendSeen) {
                        candidate.put("backend_entry_gpr", selectedRegisters(firstBoundaryRegs));
                        long backendReturn = requireRegister(firstBoundaryRegs, "ra", "backend entry");
                        candidate.put("backend_return_pc", hex32(backendReturn));

                        removeBreakpoint(writer, reader, requestId, activeBreakpoints, BACKEND_ENTRY);
                        addBreakpoint(writer, reader, requestId, activeBreakpoints, backendReturn);
                        resume(writer, reader, requestId);

                        JSONObject returnRegs = waitForStoppedPc(writer, reader, requestId,
                                new long[]{backendReturn, PRODUCER_POST_CALL});
                        long returnPc = requireRegister(returnRegs, "pc", "backend return");
                        if (returnPc == backendReturn) {
                            candidate.put("backend_return_gpr", selectedRegisters(returnRegs));
                            removeBreakpoint(writer, reader, requestId,
                                    activeBreakpoints, backendReturn);
                            resume(writer, reader, requestId);
                            JSONObject postCallRegs = waitForStoppedPc(writer, reader, requestId,
                                    new long[]{PRODUCER_POST_CALL});
                            candidate.put("producer_post_call_gpr", selectedRegisters(postCallRegs));
                        } else {
                            candidate.put("backend_return_boundary_skipped", true);
                            candidate.put("producer_post_call_gpr", selectedRegisters(returnRegs));
                            removeBreakpointIfActive(writer, reader, requestId,
                                    activeBreakpoints, backendReturn);
                        }
                    } else {
                        candidate.put("backend_not_seen_before_wrapper_return", true);
                        candidate.put("producer_post_call_gpr", selectedRegisters(firstBoundaryRegs));
                        removeBreakpointIfActive(writer, reader, requestId,
                                activeBreakpoints, BACKEND_ENTRY);
                    }

                    removeBreakpoint(writer, reader, requestId,
                            activeBreakpoints, PRODUCER_POST_CALL);
                    addBreakpoint(writer, reader, requestId, activeBreakpoints, PRODUCER_STORE);
                    resume(writer, reader, requestId);

                    JSONObject storeRegs = waitForStoppedPc(writer, reader, requestId,
                            new long[]{PRODUCER_STORE});
                    long storeS4 = requireRegister(storeRegs, "s4", "producer store");
                    long storeBase = requireRegister(storeRegs, "a0", "producer store");
                    long storeValue = requireRegister(storeRegs, "v0", "producer store");
                    if (storeS4 != 0) {
                        throw new IllegalStateException(
                                "producer store에서 s4==0 invariant가 깨졌습니다: " + hex32(storeS4));
                    }

                    candidate.put("producer_store_gpr", selectedRegisters(storeRegs));
                    long destination = (storeBase + SLOT_OFFSET) & 0xffffffffL;
                    candidate.put("producer_store_destination", hex32(destination));
                    candidate.put("producer_store_v0", hex32(storeValue));
                    tryMemoryRead(writer, reader, requestId, candidate,
                            "producer_store_old_value", destination, 4);

                    removeBreakpoint(writer, reader, requestId,
                            activeBreakpoints, PRODUCER_STORE);
                    addBreakpoint(writer, reader, requestId,
                            activeBreakpoints, SECOND_SCANNER_CALL);
                    addBreakpoint(writer, reader, requestId,
                            activeBreakpoints, SECOND_SCANNER_SKIP);
                    resume(writer, reader, requestId);

                    JSONObject scannerRegs = waitForStoppedPc(writer, reader, requestId,
                            new long[]{SECOND_SCANNER_CALL, SECOND_SCANNER_SKIP});
                    long scannerPc = requireRegister(scannerRegs, "pc", "second scanner boundary");
                    long scannerA1 = requireRegister(scannerRegs, "a1", "second scanner boundary");
                    candidate.put("second_scanner_boundary", hex32(scannerPc));
                    candidate.put("second_scanner_gpr", selectedRegisters(scannerRegs));
                    candidate.put("second_scanner_a1", hex32(scannerA1));
                    candidate.put("store_equals_scanner_input", storeValue == scannerA1);

                    removeBreakpoint(writer, reader, requestId,
                            activeBreakpoints, SECOND_SCANNER_CALL);
                    removeBreakpoint(writer, reader, requestId,
                            activeBreakpoints, SECOND_SCANNER_SKIP);

                    // Exactly one rolling snapshot: a different address on another run
                    // still leaves the immediately preceding candidate available.
                    candidate.put("event", "pointer_boundary_candidate");
                    candidate.put("time_ms", System.currentTimeMillis());
                    saveLatestCandidate(candidate);

                    if (scannerPc == SECOND_SCANNER_CALL
                            && scannerA1 == OBSERVED_FAILING_FIELD_VALUE) {
                        candidate.put("event", "pointer_boundary_capture");
                        candidate.put("time_ms", System.currentTimeMillis());
                        candidate.put("matched_observed_failing_field_value", true);
                        candidate.put("observed_failing_field_value",
                                hex32(OBSERVED_FAILING_FIELD_VALUE));
                        clearLatestCandidate();
                        appendEvent(candidate.toString());
                        updateNotification("문제 경계 캡처 완료 · 로그 복사 가능");
                        stopRequested = true;
                        break;
                    }

                    addBreakpoint(writer, reader, requestId,
                            activeBreakpoints, PRODUCER_CALL);
                    resume(writer, reader, requestId);
                }
            } finally {
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

    /**
     * Wait for PPSSPP to stop at one of our exact breakpoints without consuming
     * any broadcast event. cpu.status is documented by PPSSPP as cheap to poll;
     * its pc is accurate whenever stepping=true.
     */
    private JSONObject waitForStoppedPc(BufferedWriter writer, BufferedReader reader,
                                        int[] requestId, long[] expected) throws Exception {
        while (!stopRequested) {
            JSONObject statusResponse = request(writer, reader, requestId[0]++,
                    rawCommand("cpu.status", new JSONObject()), RESPONSE_TIMEOUT_MS);
            JSONObject status = rawResponse(statusResponse);
            if (status.optBoolean("stepping", false)) {
                long pc = status.optLong("pc", -1) & 0xffffffffL;
                for (long target : expected) {
                    if (pc == (target & 0xffffffffL)) {
                        JSONObject regs = getAllRegs(writer, reader, requestId);
                        long regsPc = requireRegister(regs, "pc", "stopped register snapshot");
                        if (regsPc != pc) {
                            throw new IllegalStateException(
                                    "cpu.status PC와 register PC가 다릅니다: status="
                                            + hex32(pc) + " regs=" + hex32(regsPc));
                        }
                        return regs;
                    }
                }
                throw new IllegalStateException(
                        "예상하지 않은 stepping PC에서 CPU가 멈췄습니다: " + hex32(pc));
            }
            sleep(STATUS_POLL_MS);
        }
        throw new InterruptedException("stop requested");
    }

    private static JSONObject getAllRegs(BufferedWriter writer, BufferedReader reader,
                                         int[] requestId) throws Exception {
        return request(writer, reader, requestId[0]++,
                rawCommand("cpu.getAllRegs", new JSONObject()), RESPONSE_TIMEOUT_MS);
    }

    private static long requireRegister(JSONObject response, String name, String where) {
        long value = findRegister(response, name);
        if (value < 0) {
            throw new IllegalStateException(where + "에서 " + name + " 레지스터를 읽지 못했습니다");
        }
        return value & 0xffffffffL;
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

    private static void removeBreakpointIfActive(BufferedWriter writer, BufferedReader reader,
                                                 int[] requestId, Set<Long> active,
                                                 long address) throws Exception {
        if (active.contains(address & 0xffffffffL)) {
            removeBreakpoint(writer, reader, requestId, active, address);
        }
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

    private synchronized void saveLatestCandidate(JSONObject candidate) {
        latestCandidate = candidate.toString();
        persistEvents();
    }

    private synchronized void clearLatestCandidate() {
        latestCandidate = null;
        persistEvents();
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
            if (latestCandidate != null) {
                writer.write(latestCandidate);
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
        worker.shutdown();
        super.onDestroy();
    }

    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }
}
