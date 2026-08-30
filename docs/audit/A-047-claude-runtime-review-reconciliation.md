# A-047 — Runtime review reconciliation and next producer trace

Date: 2026-08-29

## Scope

This note reconciles the independent Claude review of A-046 with the captured PPSSPP runtime evidence. It intentionally does not promote a root cause beyond the evidence.

## Evidence status after review

### Strongly supported

- The hot routine at `z_un_089661dc` behaves like a NUL-terminated string scanner/parser based on the captured runtime disassembly: byte loads from `a1`, special handling for `0x1B`, CR (`0x0D`) and LF (`0x0A`), variable `a1` advancement, and a nonzero-byte back-edge.
- `t2` is not one monotonic long-running counter. The routine reuses it for input bytes, comparison/boolean temporaries, and a counter loaded/stored through `a0`; this explains the observed alternation between small values and large values.
- The second call path loads `a1` from the object field at `s0+0x3C0` immediately before calling `z_un_089661dc`.
- During the failing run the scan continued for roughly four minutes with no visible recovery while CPU ticks advanced and `a1` moved through tens of megabytes.

### Not yet promoted to confirmed root cause

The runtime value associated with the `s0+0x3C0` field is suspicious and the debugger rejected `memory.read` in the later `a1` range as `Invalid address`, but that alone does not yet distinguish among:

1. a genuinely garbage pointer produced upstream;
2. a valid offset/length/handle being consumed as a pointer (type confusion or contract mismatch);
3. a pointer observed at an invalid lifecycle moment, before/after the target buffer is valid;
4. a PPSSPP debugger-addressability limitation rather than a CPU-visible invalid address.

The primary investigation therefore shifts to **the value/dataflow feeding `s0+0x3C0`**, while retaining these alternatives.

## Precision correction: `0x8C4C89A4` versus `0x8D1E5BC2`

These values are from different moments and must not be read as a discontinuity.

- `0x8C4C89A4` was observed as the value loaded from the `+0x3C0` field in the captured caller context.
- `0x8D1E5BC2` was the first value retained by the periodic hot-loop detector after the scan had already been running and advancing.

The detector samples every 500 ms and does not capture the exact cycle of function entry, so a gap between the initial caller-loaded value and the first retained hot-loop `a1_start` is expected.

## Reproducibility gap

The disassembly currently exists only inside runtime JSONL supplied from the Android tracer and quoted in PR discussion. A reviewer cannot regenerate those instructions from a checked-in raw executable because the retail executable is not committed to this repository.

The next capture must therefore preserve the raw debugger disassembly response as a first-class evidence artifact in the user-provided trace, and the relevant JSON should be checked in after capture (with no retail executable bytes added to the repository).

## Exact next runtime capture

Extend the hot-loop evidence capture with the following additional regions and values:

### Producer/caller disassembly

- Start: `0x0886C84C`
- Cover through at least `0x0886CA4C` (128 instructions minimum)
- Purpose: include the real caller-function entry, the computation of `v0`, the `sw v0,0x3C0(...)` store, both calls to `z_un_089661dc`, and nearby branch conditions.

### Object window

At the hot-loop evidence point, read the object based on the sampled `s0` register:

- start: `s0 + 0x380`
- size: `0x180` bytes
- covers: `s0+0x380 .. s0+0x4FF`

Also record separately:

- `s0`
- field address `s0+0x3C0`
- four raw bytes at `s0+0x3C0`
- decoded little-endian 32-bit value when readable

This allows the pointer-like field to be correlated with neighboring object state instead of interpreting the value in isolation.

### Preserve existing evidence

Continue retaining:

- hot-loop disassembly around `0x08966120..`
- current registers
- stack window
- `a1` progress
- memory-read success/failure for the current `a1` region
- elapsed hot-loop time

## Static follow-up after capture

From the expanded caller disassembly, perform backward dataflow from the `sw v0,0x3C0(...)` writer:

1. identify the exact destination base register and prove it aliases the later `s0` object;
2. identify the instruction(s) that define `v0` immediately before the store;
3. continue backward through calls/loads/arithmetic until the source is classified as pointer, offset, length, handle, or computed address;
4. compare the same path against the retail/English-patch behavior before attributing the corruption specifically to the Korean patch.

## Evidence policy

A successful runtime is only non-reproduction for that run. A freeze/crash is strong failure evidence. Do not call the scanner routine itself the root cause, and do not call `s0+0x3C0` a definitively corrupt pointer until the producer contract and PPSSPP address semantics are established.
