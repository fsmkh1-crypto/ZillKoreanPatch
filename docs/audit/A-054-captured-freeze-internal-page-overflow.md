# A-054 — Captured freeze points to internal first-page overflow

Date: 2026-08-30

Status: **strong root-cause candidate for the captured freeze only**. The separate H1 eight-line runtime reproduction remains unresolved and must not be silently folded into this conclusion.

## Scope

This note reinterprets the preserved runtime register/disassembly evidence for the reproduced 210065 freeze. It specifically tests whether the bad `s0+0x3C0` word must have come from the allocator path, or whether the first inline page could have overwritten that field before any allocator call occurred.

## Object layout and page path

`z_un_0886C84C` starts with an inline destination page at `s0+0x2C0`. The next region begins at `s0+0x3C0`, exactly `0x100` bytes later. The latter is consumed as the first stored page-pointer slot by the later scanner path.

The producer block for additional pages is:

```text
0x0886C938  jal   z_un_089E2B54
0x0886C93C  li    a0,0x100
0x0886C940  sll   a0,s4,0x2
0x0886C944  addu  a0,s0,a0
0x0886C948  sw    v0,0x3C0(a0)
0x0886C94C  addiu s4,s4,1
0x0886C950  move  a1,v0
0x0886C954  andi  s4,s4,0xFF
```

Thus every executed additional-page allocation increments `s4` immediately after storing the returned pointer. The function initializes `s4` to zero. The same function has an eight-page guard (`s7=8`) before this producer path, so an 8-bit wrap from 256 allocations is not a plausible explanation for a later zero value.

## Runtime register evidence

During the captured freeze, repeated samples in `z_un_089661DC` showed:

```text
s0 = 0x08ADF9C8
s2 = 0x00000004
s3 = 0x00000113
s4 = 0x00000000
s5 = 0x00000113
s7 = 0x00000008
ra = 0x0886C9C0
```

The scanner was the second call from `0x0886C9B4`, after:

```text
0x0886C9A4  lw    a1,0x3C0(s0)
0x0886C9B4  jal   z_un_089661DC
```

and the observed slot value was:

```text
[s0+0x3C0] = 0x8C4C89A4
```

## Why `s4=0` materially changes the provenance interpretation

The preserved code initializes `s4=0` and increments it in the additional-page producer immediately after `C948`. With the eight-page guard, the captured `s4=0` strongly indicates that this invocation never successfully traversed the `C938/C948/C94C` allocation/store path before reaching the scanner.

This independently matches the later tracer result in which a breakpoint armed at `0x0886C938` was never observed before the same freeze. That tracer failure is not by itself proof of path absence, but it is directionally consistent with the register evidence rather than contradictory to it.

Therefore the captured `0x8C4C89A4` word should **not** currently be attributed to an `08A23064` allocator return. The allocator chain remains real code, but the captured invocation may never have reached it.

## `s3=0x113` and the 0x100-byte inline page

After the first scanner call at `0x0886C998`, the function executes:

```text
0x0886C9A8  move  s3,v0
```

`z_un_089661DC` counts line length / maximum line span while scanning a NUL-terminated byte string. At the later freeze sample, `s3=0x113` (275 decimal), which is the preserved result from scanning the inline first-page string before the second scanner is entered.

The inline page occupies exactly `0x100` bytes from `s0+0x2C0` through `s0+0x3BF`. A 275-byte unbroken first-page string exceeds that region by 19 bytes.

The immediately following address is `s0+0x3C0`, the exact word later loaded as the second scanner pointer.

This gives a direct corruption mechanism:

```text
long unbroken/materialized first page
  -> byte copy begins at s0+0x2C0
  -> no page-split path taken (`s4` remains 0)
  -> copy crosses the 0x100-byte inline-page boundary
  -> bytes at offsets 0x100.. overwrite s0+0x3C0 and following fields
  -> C9A4 interprets overwritten bytes as a pointer
  -> C9B4 passes that word directly to the NUL scanner
  -> scanner walks arbitrary memory looking for NUL
  -> observed runaway scan / freeze
```

This chain explains both the corrupted pointer slot and why the allocator breakpoint was not observed.

## Relationship to ID 210065

ID 210065 is the current visible freeze scene (the "광대한 대지..." opening text). The canonical Korean semantic text was historically stored without layout and was already identified by the consumer audit as a C22 contract violation. Current source code contains a forced eight-line diagnostic layout for 210065 and materializes `replacement.Layout` when present.

The captured runtime state, however, is not compatible with a safely split 20–33-byte-line page: `s3=0x113` shows an approximately 275-byte unbroken span in the inline page for that captured invocation.

Therefore one of the following must be true for that captured build:

1. it predates / did not contain the effective eight-line materialization;
2. the selected runtime record did not use that layout despite source intent; or
3. a different materialization/projection path reconstructed a long unbroken span.

The capture itself proves the runtime state; current source intent does not retroactively change it.

## Important limit: separate H1 eight-line reproduction

The user separately observed a freeze after an eight-line H1 experiment. That is strong failure evidence for that run, but no equivalent register/disassembly capture currently proves that H1 reached the same `s4=0`, `s3=0x113`, overwritten-`+0x3C0` state.

Do **not** claim that A-054 explains the H1 failure until the H1 artifact/build provenance or runtime bytes demonstrate that its line breaks were actually present in the executed 210065 record.

A same-screen freeze is not automatically the same machine-state cause.

## Evidence ranking

Strong for the captured run:

- inline page begins at `+0x2C0`; pointer slot begins exactly `0x100` bytes later at `+0x3C0`;
- second scanner directly loads `+0x3C0` as a raw address;
- captured `s4=0` despite `s4` being incremented by the additional-page producer;
- captured `s3=0x113` after the first inline-page scan;
- 0x113 bytes exceeds the 0x100 inline-page capacity by 19 bytes;
- observed `+0x3C0` contents are nonsensical as the required raw page pointer;
- `C938` was not observed by the later armed tracer.

Still unresolved:

- exact four bytes of materialized Korean data corresponding to the overwritten `0x8C4C89A4` word;
- exact build provenance of the captured ISO relative to the eight-line diagnostic commit;
- whether the separate H1 eight-line freeze shares this mechanism or represents an additional defect.

## Next cheapest gate

Do not descend into `08A23064` first. For the captured freeze, its path is no longer the earliest likely divergence.

Next verify the **actual runtime/materialized bytes of 210065 in the exact H1 eight-line build** or its reproducible build inputs. The decisive check is simple:

- if seven `0x0A` line breaks are present at the intended positions and every page span is <0x100, A-054 does not explain H1 and a second defect remains;
- if the runtime record is still a long unbroken span, the layout experiment never reached the consumer and the overflow chain becomes the unified explanation.
