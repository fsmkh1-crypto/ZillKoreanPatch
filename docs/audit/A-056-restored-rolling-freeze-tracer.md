# A-056 — Restored rolling PPSSPP freeze tracer

## Status

Evidence/tooling update only. This note does not promote a new root cause and does not merge the separate H1 reproduction into the A-054 captured-run interpretation.

## Why the rolling tracer is restored

The breakpoint/manual-only capture experiments were useful for discriminating specific states, but they introduced a practical failure mode: after PPSSPP entered the freeze state, a newly initiated debugger handshake could fail or the app/service could lose the session before the requested state was captured.

The restored tracer instead keeps attempting the debugger connection while the game runs and continuously preserves the most recent successful samples. Connection loss and command timeouts are recorded as evidence rather than discarding the preceding state.

## Current capture behavior

At branch HEAD `dc5bf45076336fbadb1bb85203b4adb234f66dbf` the Android tracer:

- samples `cpu.status` and GPR state every 500 ms;
- keeps a rolling ring of the most recent 60 samples (about 30 seconds at the nominal interval);
- records `connection_lost` and `sample_timeout` events;
- watches the known runaway scanner window `0x08966200..0x08966260`;
- requires repeated samples in that window plus meaningful unsigned `a1` advancement before emitting `hot_loop_detected`;
- on hot-loop evidence, captures disassembly near `0x089661E0`, memory near the live `a1`, and `0x120` bytes beginning at `s0+0x2C0` when `s0` is available;
- keeps the trace file available for copy from the Android UI.

No breakpoint is required for this rolling mode.

## Relationship to A-054/A-055

A-054 remains the strongest explanation for the preserved ID 210065 captured invocation:

- `s4 = 0`;
- `s3 = 0x113` (275);
- the inline region at `s0+0x2C0` is `0x100` bytes;
- the first secondary-page pointer slot is at `s0+0x3C0`.

That state is consistent with an unbroken first-page span crossing the inline boundary and corrupting the pointer slot subsequently consumed by the second scanner.

A-055 independently establishes that the current compiler can materialize the intended ID 210065 diagnostic as 7 LF / 8 lines, with every line <=56 bytes and the whole record below `0x100` bytes while preserving the `05 05 05 00` ending. It does not establish that the historical H1 artifact contained those exact bytes.

The restored rolling tracer is therefore intended to provide a fresh runtime bridge between those two evidence sets: live object/materialization state versus the preserved A-054 failure state and the current compiler-side compact record.

## CI / artifact

For HEAD `dc5bf45076336fbadb1bb85203b4adb234f66dbf`:

- CI: success (`33268087041`)
- Korean data CI: success (`33268087044`)
- Android PPSSPP rolling freeze tracer: success (`33268087045`)
- Android artifact: `zill-rolling-freeze-tracer-debug`
- artifact id: `9719269325`
- artifact digest: `sha256:370ba80b00c07825fec771cb71a89857e0b35e69f3bd43cf869eacb1b5346f20`
- extracted APK SHA-256: `bfb40818c632d4cc922545732d0886a128ff5dc41f1a2f5e0c382b77ac26403d`

## Next discriminant

Use a fresh rolling trace from the ID 210065 freeze scene and inspect whether the live evidence again shows the A-054 signature (`s4==0`, oversized first-page span / equivalent object corruption, bad `+0x3C0` state), or instead shows a different state consistent with the separate H1 reproduction.

Until that evidence exists, same-screen freezes must remain potentially distinct causes.