# A-046 — Runtime invalid string pointer before `z_un_089661dc`

Date: 2026-08-29
Branch: `audit/projection-control-roundtrip`

## Observation

A PPSSPP runtime trace captured the freeze for roughly four minutes with no visible recovery. CPU ticks continued increasing while execution remained concentrated in `0x08966200..0x08966260`.

The hot function is `z_un_089661dc`. Its loop is a byte-oriented string/control scanner:

- loads the current byte from `0(a1)`
- advances `a1` by 1 or 2 bytes depending on control handling
- treats `0x1B`, CR (`0x0D`), and LF (`0x0A`) specially
- loops while the next loaded byte is nonzero

The back-edge at `0x08966260` is `bnel t2,zero,0x08966200`, so a NUL byte is the ordinary termination condition.

## Stronger caller evidence

The caller neighborhood shows the second call site as:

```asm
0x0886C9B0  lw    a1,0x3C0(s0)
0x0886C9B8  jal   z_un_089661dc
0x0886C9BC  addiu a0,sp,0x24
```

During the captured freeze:

- `s0 = 0x08ADF9C8`
- therefore the source field is `s0 + 0x3C0 = 0x08ADFD88`
- the value loaded from that field was observed as `0x8C4C89A4`
- the tracer's `memory.read` against the advancing `a1` region returned `Invalid address`

The scan subsequently advanced from approximately `0x8D1E5BC2` to `0x90E28B89` over about 199 seconds and continued without visible recovery for roughly four minutes total.

## Interpretation

This materially changes the leading hypothesis.

The strongest current explanation is no longer merely "a valid string lost its NUL terminator." Instead, the scanner appears to receive an already-invalid string pointer from the caller. The ordinary retail scanner then walks forward looking for a NUL and produces the user-visible freeze.

This does **not** yet prove where the pointer became invalid. The fault may still originate in translated data, substitution handling, projection/storage, or another upstream structure. The root cause is therefore not promoted beyond the evidence.

## `t2` clarification

Large `t2` samples are not evidence of one monotonic loop counter. Inside `z_un_089661dc`, the same register is reused for several roles, including:

- current input byte (`lb t2,0(a1)`)
- boolean results from `sltu`
- an output counter loaded from `0(a0)` and incremented

Therefore alternating values such as `0`, `0x0A`, and large integers are expected from register reuse and should not be interpreted as a single counter resetting.

## Next forensic target

The caller belongs to `z_un_0886c84c`. The captured window begins too late to explain the value producer for a nearby write of the form:

```asm
sw v0,0x3C0(...)
```

The next runtime capture should disassemble from the actual function start `0x0886C84C`, far enough through the `0x3C0` store and both calls to `z_un_089661dc`.

It should also preserve memory around the live caller object (`s0`, especially `s0+0x380..s0+0x500`) so the `+0x3C0` field and neighboring pointer/count fields can be correlated.

Success criterion for the next step: identify the instruction/dataflow that produces the value stored at `+0x3C0`, then follow that producer upstream before requesting another broad runtime experiment.

## Evidence policy

- The approximately four-minute non-recovery is strong evidence against a merely slow normal path.
- It is still not mathematical proof of an infinite loop.
- The invalid-pointer interpretation is stronger than a missing-sentinel-only interpretation because the caller-supplied value itself is outside the readable region observed by PPSSPP.
