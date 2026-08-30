# A-050 — Minimal tracer and workflow correction

## Correction

Commits `010c4a9c1eb639e767bd56e3c380ab693704d904` and `3291c95d87dd9552dd82611fa701cf39c0137397` changed only `.github/workflows/android-extractor.yml`; they did **not** persist the claimed wider runtime capture into `FreezeTraceService.java`.

The actual service source remained at the earlier `0x089E2B54` / 128-instruction allocator-candidate capture. Treat earlier statements that those commits had already changed the tracer source as incorrect.

## Current policy

Runtime evidence already established in A-046 through A-049 is not re-collected on every run. The tracer now records only the current unresolved backend question:

- disassembly start: `0x08A23064`
- disassembly count: 256 instructions
- one capture per connected session after the known runaway scanner signature is detected
- current registers and a four-byte correlation read of `s0+0x3C0`
- no repeated scanner disassembly
- no repeated producer disassembly
- no repeated wrapper-chain disassembly
- no full object-window capture
- no stack dump
- no persisted 500 ms sample stream

The evidence archive grows in `docs/audit/`; the runtime log should remain bounded to the current unresolved question.

## Android build policy

The Android workflow is now a tracer build, not a Korean ISO patcher build. It does not prepare or embed the Korean translation/ISO payload and does not run patcher payload inspection. It builds only the PPSSPP debugger bridge plus the Android tracer APK.

## Investigation target

The next runtime question is whether the body at `0x08A23064` establishes allocator/pool semantics and where the abnormal return lineage originates. If it is another wrapper, do not resume one-wrapper-per-APK chasing; switch to return-value/state instrumentation aimed at the first corrupted value instead.
