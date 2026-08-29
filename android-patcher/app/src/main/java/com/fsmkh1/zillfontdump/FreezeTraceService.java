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
import java.util.ArrayDeque;
import java.util.ArrayList;
import java.util.Deque;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeoutException;

/**
 * PPSSPP inline-page provenance tracer.
 *
 * This stage deliberately stops at the second scanner call, where the preserved
 * runtime evidence showed the bad +0x3C0 field being consumed.  At that exact
 * boundary it records s0/s3/s4/a1 and dumps s0+0x2C0 through +0x3DF so we can
 * distinguish an actually materialized eight-line H1 page from the previously
 * captured 0x113-byte unbroken overflow.
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
    private static final int MAX_EVENTS = 24;

    private static final long SECOND_SCANNER_CALL = 0x0886C9B4L;
    private static final long INLINE_PAGE_OFFSET = 0x2C0L;
    private static final int INLINE_PAGE_CAPACITY = 0x100;
    private static final int INLINE_DUMP_SIZE = 0x120;
    private static final int POINTER_SLOT_RELATIVE = 0x100;
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
                addBreakpoint(writer, reader, requestId, activeBreakpoints, SECOND_SCANNER_CALL);
                writeEvent("armed", "second_scanner_call", hex32(SECOND_SCANNER_CALL));
                updateNotification("추적 준비 완료 · 문제 장면까지 진행하세요");

                int candidateIndex = 0;
                while (!stopRequested) {
                    JSONObject regs = waitForStoppedPc(writer, reader, requestId, SECOND_SCANNER_CALL);
                    candidateIndex++;

                    long s0 = requireRegister(regs, "s0", "second scanner call");
                    long s3 = requireRegister(regs, "s3", "second scanner call");
                    long s4 = requireRegister(regs, "s4", "second scanner call");
                    long a1 = requireRegister(regs, "a1", "second scanner call");
                    long pageStart = (s0 + INLINE_PAGE_OFFSET) & 0xffffffffL;

                    JSONObject candidate = new JSONObject();
                    candidate.put("event", "inline_page_candidate");
                    candidate.put("time_ms", System.currentTimeMillis());
                    candidate.put("candidate_index", candidateIndex);
                    candidate.put("boundary_pc", hex32(SECOND_SCANNER_CALL));
                    candidate.put("gpr", selectedRegisters(regs));
                    candidate.put("inline_page_start", hex32(pageStart));
                    candidate.put("inline_page_capacity", INLINE_PAGE_CAPACITY);
                    candidate.put("scanner_a1", hex32(a1));

                    byte[] dump = readMemory(writer, reader, requestId, pageStart, INLINE_DUMP_SIZE);
                    candidate.put("inline_dump_size", dump.length);
                    candidate.put("inline_dump_base64", Base64.encodeToString(dump, Base64.NO_WRAP));
                    appendInlineAnalysis(candidate, dump);
                    saveLatestCandidate(candidate);

                    boolean suspicious = a1 == OBSERVED_FAILING_FIELD_VALUE
                            || s3 >= INLINE_PAGE_CAPACITY
                            || (s4 == 0 && candidate.optLong("first_nul_offset", -1) >= INLINE_PAGE_CAPACITY);
                    if (suspicious) {
                        candidate.put("event", "inline_page_capture");
                        candidate.put("matched_observed_failing_field_value",
                                a1 == OBSERVED_FAILING_FIELD_VALUE);
                        candidate.put("s3_exceeds_inline_capacity", s3 >= INLINE_PAGE_CAPACITY);
                        candidate.put("s4_zero", s4 == 0);
                        clearLatestCandidate();
                        appendEvent(candidate.toString());
                        updateNotification("inline page 캡처 완료 · 로그 복사 가능");
                        stopRequested = true;
                        break;
                    }

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

    private static void appendInlineAnalysis(JSONObject out, byte[] dump) throws Exception {
        int firstNul = -1;
        List<Integer> lfPositions = new ArrayList<>();
        int maxSpan = 0;
        int span = 0;
        int analysisEnd = Math.min(dump.length, INLINE_PAGE_CAPACITY);

        for (int i = 0; i < dump.length; i++) {
            int value = dump[i] & 0xff;
            if (firstNul < 0 && value == 0) {
                firstNul = i;
            }
            if (i >= analysisEnd || (firstNul >= 0 && i >= firstNul)) {
                continue;
            }
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
        out.put("lf_count_before_nul", lfPositions.size());
        out.put("lf_positions", lf);
        out.put("first_nul_offset", firstNul);
        out.put("max_non_lf_span_before_nul", maxSpan);
        out.put("has_nul_within_inline_page", firstNul >= 0 && firstNul < INLINE_PAGE_CAPACITY);

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
        if (encoded.isEmpty() && size != 0) {
            throw new IllegalStateException("memory.read 응답에 base64 데이터가 없습니다");
        }
        byte[] decoded = Base64.decode(encoded, Base64.DEFAULT);
        if (decoded.length != size) {
            throw new IllegalStateException(
                    "memory.read 크기 불일치: got=" + decoded.length + " want=" + size);
        }
        return decoded;
    }

    private static JSONObject waitForStoppedPc(BufferedWriter writer, BufferedReader reader,
                                                int[] requestId, long expected) throws Exception {
        while (true) {
            JSONObject statusResponse = request(writer, reader, requestId[0]++,
                    rawCommand("cpu.status", new JSONObject()), RESPONSE_TIMEOUT_MS);
            JSONObject status = rawResponse(statusResponse);
            if (status.optBoolean("stepping", false)) {
                long pc = status.optLong("pc", -1) & 0xffffffffL;
                if (pc != (expected & 0xffffffffL)) {
                    throw new IllegalStateException(
                            "예상하지 않은 stepping PC에서 CPU가 멈췄습니다: " + hex32(pc));
                }
                JSONObject regs = getAllRegs(writer, reader, requestId);
                long regsPc = requireRegister(regs, "pc", "stopped register snapshot");
                if (regsPc != pc) {
                    throw new IllegalStateException(
                            "cpu.status PC와 register PC가 다릅니다: status="
                                    + hex32(pc) + " regs=" + hex32(regsPc));
                }
                return regs;
            }
            sleep(STATUS_POLL_MS);
        }
    }

    private static JSONObject getAllRegs(BufferedWriter writer, BufferedReader reader,
                                         int[] requestId) throws Exception {
        return request(writer, reader, requestId[0]++,
                rawCommand("cpu.getAllRegs", new JSONObject()), RESPONSE_TIMEOUT_MS);
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
                .setContentTitle("질올 PPSSPP inline page tracer")
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
        if (result == null) return new JSONObject();
        JSONObject raw = result.optJSONObject("response");
        return raw == null ? new JSONObject() : raw;
    }

    private static JSONObject selectedRegisters(JSONObject response) throws Exception {
        JSONObject selected = new JSONObject();
        String[] names = new String[]{
                "v0", "v1", "a0", "a1", "a2", "a3",
                "s0", "s1", "s2", "s3", "s4", "s5", "s6", "s7",
                "sp", "fp", "ra", "pc"
        };
        for (String name : names) {
            long value = findRegister(response, name);
            if (value >= 0) selected.put(name, hex32(value));
        }
        return selected;
    }

    private static long requireRegister(JSONObject response, String name, String where) {
        long value = findRegister(response, name);
        if (value < 0) {
            throw new IllegalStateException(where + "에서 " + name + " 레지스터를 읽지 못했습니다");
        }
        return value & 0xffffffffL;
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
        stopRequested = true;
        worker.shutdown();
        super.onDestroy();
    }

    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }
}
