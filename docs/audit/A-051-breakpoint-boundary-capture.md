# A-051 — Breakpoint boundary capture and producer-store address correction

## Status

Forensic instrumentation change. This note does not declare a root cause.

## Why the trace strategy changed

A-046 through A-049 already established the downstream runaway scanner and the producer/wrapper lineage. Re-collecting those disassemblies on every runtime reproduction was redundant and caused runtime logs to grow as the investigation narrowed.

The current objective is instead to observe the earliest boundary at which the value later consumed from `s0+0x3C0` becomes unsuitable for the NUL-terminated scanner.

The PPSSPP WebSocket debugger exposes CPU instruction breakpoints and stepping/resume events, and the repository's `internal/ppssppdebug` bridge already exposes raw debugger events plus `wait` and `resume`. Therefore the Android tracer can observe the value at execution boundaries directly instead of following one wrapper per APK replay.

## Raw-address correction

The uploaded 2026-08-29 `memory.disasm` payload was re-read as the authority and copied into `docs/audit/fixtures/runtime-20260829-freeze-disassembly.txt`.

The exact producer block is:

```text
0x0886C938  encoding=0x0E278AD5  jal   z_un_089e2b54
0x0886C93C  encoding=0x34040100  li    a0,0x100
0x0886C940  encoding=0x00142080  sll   a0,s4,0x2
0x0886C944  encoding=0x02042021  addu  a0,s0,a0
0x0886C948  encoding=0xAC8203C0  sw    v0,0x3C0(a0)
0x0886C94C  encoding=0x26940001  addiu s4,s4,0x1
0x0886C950  encoding=0x00402825  move  a1,v0
0x0886C954  encoding=0x329400FF  andi  s4,s4,0xFF
0x0886C958  encoding=0x34040000  li    a0,0x0
```

This corrects earlier prose transcriptions that identified the writer as `0x0886C94C` (and, earlier still, other incorrect offsets). The actual store is `0x0886C948`. `0x0886C94C` is the following `addiu s4,s4,1` instruction.

Future address-level claims for this block must cite the raw fixture rather than a prose note or remembered address.

## New one-shot boundary capture

`FreezeTraceService` now scopes capture to the exact producer invocation using breakpoints:

1. break at `0x0886C938` (`jal z_un_089e2b54`),
2. after that exact call is reached, arm both `0x08A23064` and `0x0886C940`,
3. if `0x08A23064` is entered, capture its arguments and dynamic `ra`,
4. break at that dynamic return address and capture returned `v0`,
5. capture `v0` again at `0x0886C940` after the wrapper chain returns,
6. break at the actual store `0x0886C948`, capture `v0`, computed `a0`, destination `a0+0x3C0`, and the previous four destination bytes,
7. remove the tracer breakpoints and resume execution.

The tracer emits one compact `pointer_boundary_capture` record and does not re-collect scanner, caller, wrapper, stack, or broad object-window disassembly.

## Evidence interpretation

The new capture can distinguish several cases in one reproduction:

- bad `v0` already at the backend return: the earliest observed failure boundary is at or below the backend call;
- backend return normal but wrapper/post-call `v0` bad: corruption/transformation lies between those boundaries;
- post-call `v0` normal but pre-store `v0` bad: corruption lies in the short producer sequence;
- pre-store `v0` normal but the later reader is bad: the slot is overwritten/corrupted after this store;
- backend not reached before the producer post-call: the assumed backend lineage is incorrect for this exact invocation and must be revised.

None of those outcomes alone should be described as the Korean patch root cause. Once the first bad runtime boundary is identified, the investigation still has to connect that boundary to the Korean patch's earliest incorrect input, storage contract, or state transition.
