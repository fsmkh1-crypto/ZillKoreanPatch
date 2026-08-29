# A-048 — Producer return path for the runaway string-scan input

Date: 2026-08-29

## Scope

This note records the expanded PPSSPP runtime disassembly captured after A-047. It narrows the source of the value later consumed as the second `z_un_089661dc` string input. It does not yet promote a final root cause.

## Confirmed runtime dataflow

Within `z_un_0886c84c`, the expanded disassembly shows the following sequence:

```text
0x0886C958  jal   z_un_089E2B54
0x0886C95C  li    a0,0x100        # delay slot
0x0886C960  sll   a0,s4,0x2
0x0886C964  addu  a0,s0,a0
0x0886C968  sw    v0,0x3C0(a0)
0x0886C96C  addiu s4,s4,0x1
0x0886C970  move  a1,v0
0x0886C974  andi  s4,s4,0xFF
```

The later second scanner call is:

```text
0x0886C9A4  sw    zero,0x24(sp)
0x0886C9A8  lw    a1,0x3C0(s0)
0x0886C9AC  move  s3,v0
0x0886C9B0  beq   a1,zero,0x0886C9C4
0x0886C9B4  li    a0,0x0
0x0886C9B8  jal   z_un_089661dc
0x0886C9BC  addiu a0,sp,0x24
```

Therefore the value consumed as a string pointer by the second scanner is not an arbitrary pre-existing field read. It is sourced from values returned by `z_un_089E2B54` and stored into the object table at `s0 + 0x3C0 + (s4 * 4)`.

For the failing run, the runtime object state was:

- `s0 = 0x08ADF9C8`
- `s0 + 0x3C0 = 0x08ADFD88`
- raw field bytes: `A4 89 4C 8C`
- little-endian field value: `0x8C4C89A4`
- second scanner return address: `ra = 0x0886C9C0`

The hot scanner then advanced `a1` continuously through the `0x8D...` range while CPU ticks continued to increase. Debugger `memory.read` rejected those later `a1` addresses as `Invalid address`.

## Strongly supported interpretation

The immediate producer chain is now:

```text
z_un_089E2B54(a0 = 0x100)
    -> v0
    -> sw v0, s0+0x3C0+(s4*4)
    -> later lw a1, s0+0x3C0
    -> z_un_089661dc(a1)
    -> runaway NUL-terminated scan
```

This is stronger than the previous generic "upstream field corruption" hypothesis because it identifies the direct return-value source of the table entries.

## Still uncertain

The semantics of `z_un_089E2B54` are not yet established from checked-in disassembly. The `a0 = 0x100` call shape makes an allocation/block-creation routine a plausible hypothesis, but this must not yet be labeled as confirmed allocator behavior.

Open alternatives include:

1. allocation-like routine returning a pointer;
2. handle/object constructor returning a non-pointer value later misused as a pointer;
3. address transformation/mapping helper returning a value with a different contract;
4. a valid pointer under a PPSSPP/PSP memory-map mode not yet modeled correctly by the debugger API;
5. corruption inside or immediately after `z_un_089E2B54` before the table entry is consumed.

## Exact next target

Capture and preserve raw disassembly for `z_un_089E2B54`, including enough instructions to identify:

- whether it calls a known allocator or heap primitive;
- what `a0=0x100` means;
- how `v0` is produced on the failing path;
- whether the function returns raw memory, a handle, an offset, or a transformed address;
- its callers and any relevant global heap state.

Prefer capturing the function body and one surrounding caller/callee layer before introducing a runtime patch. Compare the same contract against retail/English-patch behavior before attributing the bad return specifically to Korean patch data.

## Evidence policy

A successful run remains only non-reproduction. The failing runtime chain above is strong failure evidence. Do not call `z_un_089E2B54` the root cause until its return contract and the reason for `0x8C4C89A4` are established.
