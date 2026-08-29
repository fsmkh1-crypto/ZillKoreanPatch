# A-053 — 08A23064 / 08A230B0 paired memory-contract evidence

Date: 2026-08-30

Status: evidence narrowing only; not yet root-cause proof.

## Scope

This note compares the preserved PPSSPP disassembly around the two neighboring backend functions used by the message-page path:

- `z_un_08A23064`, reached through `z_un_089E2984`
- `z_un_08A230B0`, reached through `z_un_089E2BA4`

The goal is to determine whether the `0x8C4C89A4` value observed in `s0+0x3C0` could reasonably be a normal opaque handle, or whether the surrounding contract expects an ordinary CPU-dereferenceable buffer pointer.

## Acquire-side evidence

`z_un_089E2984` calls `z_un_08A23064` and immediately preserves the return value:

```text
0x089E2A0C  jal   z_un_08A23064
0x089E2A10  move  a1,a2        ; delay slot
0x089E2A14  move  s0,v0
0x089E2A18  bne   s0,zero,...
```

The caller therefore treats `v0 == 0` as failure and a nonzero return as the acquired object/buffer value.

In the message-page path, that same returned value is passed through the wrapper chain and stored directly:

```text
0x0886C938  jal   z_un_089E2B54
0x0886C93C  li    a0,0x100
0x0886C948  sw    v0,0x3C0(a0)
0x0886C950  move  a1,v0
```

The later consumer loads the exact stored word and supplies it directly to the NUL-string scanner:

```text
0x0886C9A4  lw    a1,0x3C0(s0)
0x0886C9B4  jal   z_un_089661DC
```

`z_un_089661DC` performs byte loads from `a1` and increments it as a normal address. Therefore, whatever `08A23064` ultimately returns on this path must be usable as a CPU-visible byte-addressable buffer pointer by the time it reaches `C9B4`; no handle-decoding step exists between the store and the scanner.

## Release-side evidence

The neighboring wrapper `z_un_089E2BA4` calls `z_un_08A230B0` while preserving the incoming `a0`. It does not consume the backend return value and restores only diagnostic/context state afterward:

```text
...                       ; diagnostic/context setup
0x089E2BCC  jal   z_un_08A230B0
0x089E2BD0  sw    zero,0x18(sp)  ; delay slot
...                       ; restore diagnostic/context state
0x089E2BE4  jr    ra
```

A second wrapper, `z_un_089E2C3C`, delegates to `z_un_089E2BA4` in the same no-return-consumption pattern.

This acquire/nonzero-return versus release/no-return-consumption asymmetry, together with the 0x4C-byte adjacency of `08A23064` and `08A230B0`, strongly supports a paired acquire/release memory API interpretation. This is still an inference because the backend bodies are not preserved in the current evidence set.

## Consequence for the freeze value

During the reproduced freeze:

```text
s0+0x3C0 = 0x08ADFD88
[s0+0x3C0] = 0x8C4C89A4
```

The second scanner then walks forward from the value loaded out of that slot and fails to encounter a terminating NUL, advancing by megabytes while CPU ticks continue.

Because there is no handle-to-pointer conversion between `C9A4` and `C9B4`, the `+0x3C0` value is required by this caller contract to be directly dereferenceable as the page buffer. Therefore the prior possibility that `0x8C4C89A4` is merely a normal opaque/tagged handle is now substantially weakened.

This does **not** yet prove whether:

1. `08A23064` itself returned the bad value;
2. a normal returned pointer was later overwritten in `+0x3C0`;
3. the pointer was valid but the backing allocation/lifetime became invalid; or
4. the pointed page lacked a terminating zero due to allocation/initialization behavior.

## Current narrowed chain

```text
210065 page split
  -> request 0x100 bytes
  -> 089E2B54
  -> 089E2B00
  -> 089E2984
  -> 08A23064  [nonzero acquired value]
  -> C948 stores v0 in +0x3C0
  -> C9A4 loads the same slot as a raw address
  -> C9B4 / 089661DC byte-dereferences it as a NUL string
  -> runaway scan observed
```

## Evidence ranking

Strong:
- the `+0x3C0` word is consumed as a raw pointer by `089661DC`;
- no conversion step exists between load and dereference;
- `089E2984` treats `08A23064` return as nullable acquired state;
- the neighbor `08A230B0` is wrapped in a release-like, return-ignored pattern.

Still unresolved:
- exact implementation of `08A23064` / `08A230B0`;
- whether the bad word originates at allocation return or is corrupted afterward;
- whether normal Japanese execution receives zero-filled allocations while the failing path sees stale data.

## Next static gate

Do not build a new tracer yet. The next cheapest evidence is to inspect preserved callers of the acquire wrapper and determine whether any caller writes through the returned value immediately (`sb/sw ... 0(v0)` or equivalent). That would independently confirm ordinary pointer semantics and may expose expected initialization behavior. If no preserved caller provides that evidence, the allocator body remains the hard boundary.
