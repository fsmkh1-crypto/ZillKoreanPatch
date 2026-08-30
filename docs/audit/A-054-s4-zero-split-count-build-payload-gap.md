# A-054 — `s4==0` narrows the captured freeze to a pre-split state

Date: 2026-08-30

Status: **strong forensic narrowing; root cause not yet promoted**

## Scope

This note re-evaluates the captured `z_un_0886c84c -> z_un_089661dc` freeze after the allocator-pair analysis in A-053. The purpose is to decide whether the observed `s0+0x3C0 = 0x8C4C89A4` can safely be treated as a value returned by the `0x100` allocation path.

## Preserved runtime facts

The freeze capture has:

- `ra = 0x0886C9C0`, identifying the second `z_un_089661dc` call;
- `s0 = 0x08ADF9C8`;
- `s4 = 0`;
- `s2 = 4`;
- `s7 = 8`;
- `s0+0x3C0 = 0x08ADFD88`;
- value loaded from that field: `0x8C4C89A4`;
- the scanner subsequently advances through `0x8D...` and `0x90...` without finding NUL.

The authoritative disassembly fixture records the writer path as:

```text
0x0886C938  jal   z_un_089e2b54
0x0886C93C  li    a0,0x100
0x0886C940  sll   a0,s4,0x2
0x0886C944  addu  a0,s0,a0
0x0886C948  sw    v0,0x3C0(a0)
0x0886C94C  addiu s4,s4,0x1
0x0886C950  move  a1,v0
0x0886C954  andi  s4,s4,0xFF
```

and the reader as:

```text
0x0886C9A4  lw    a1,0x3C0(s0)
0x0886C9AC  beq   a1,zero,0x0886C9C4
0x0886C9B4  jal   z_un_089661dc
```

## `s4` mutation evidence

The preserved function body initializes:

```text
0x0886C8B0  li s4,0
```

The split/allocation path increments it at `0x0886C94C` and masks the incremented value to one byte at `0x0886C954`.

Within the preserved path from initialization through the second scanner call, no reset of `s4` back to zero has been identified. Therefore the captured `s4==0` is strong evidence that this invocation reached `0x0886C9B4` without completing the `0x0886C938..0x0886C954` split/allocation path.

This interpretation is independently consistent with the earlier armed breakpoint tracer run that reached the target scene and froze without recording a `0x0886C938` hit. That tracer result was previously treated primarily as a tracer-design failure; it should now be retained as supporting, not standalone, evidence that the failing invocation may not have executed the allocation boundary.

## Consequence for A-053

A-053 correctly established that the `08A23064`-side return is used as a directly byte-addressable buffer when the split path executes. However, the captured `0x8C4C89A4` value must **not** currently be attributed to that allocator return: if `s4==0` faithfully represents this invocation's split count, the C938/C948 producer did not run.

The allocator pair remains relevant to split pages, but it is no longer the earliest supported origin for the specific captured bad `+0x3C0` field.

## Adjacent-object overwrite hypothesis

`z_un_0886c84c` begins copying into the object's internal text area before any heap split. The object region immediately preceding the first pointer slot is 0x100 bytes wide:

```text
s0 + 0x2C0 ... s0 + 0x3BF   initial text area
s0 + 0x3C0                  first secondary-page pointer slot
```

The loop writes bytes through its destination pointer (`sb a3,0(a1)`). Consequently, if an invocation copies more than 0x100 bytes before taking the split path, the next bytes naturally overwrite the pointer slot beginning at `s0+0x3C0`. The later unconditional reader then interprets those four bytes as a pointer and passes them to the NUL scanner.

This mechanism would explain, in one chain:

1. `s4==0`;
2. no observed C938 breakpoint hit;
3. a nonzero `+0x3C0` field despite no supported allocator-store event;
4. the text-like arbitrary 32-bit value `0x8C4C89A4`;
5. direct use of that value as `a1` by the second scanner;
6. the observed runaway NUL scan.

This is a **high-value hypothesis**, not yet root cause. Exact byte equality between the failing materialized payload at the 0x100 boundary and `A4 89 4C 8C` remains the strongest missing static proof.

## Current-build contradiction

The current branch forcibly assigns ID 210065 an eight-line layout in `internal/message/compile_korean.go`, and the compiler does materialize that `Layout` when non-empty. The authored four-line groups are approximately 126 and 107 encoded bytes respectively (Korean custom glyphs are two-byte renderer keys; ASCII spaces/punctuation are one byte where applicable, plus native line-break controls). Both are comfortably below 0x100.

Therefore the current eight-line build should not naturally produce a pre-split 0x100 overwrite if the runtime actually consumes those materialized bytes and if the inferred four-line split semantics are correct.

That creates a concrete evidence gap instead of a reason to force the hypothesis:

- either the captured freeze came from a build whose runtime 210065 payload did not contain the current layout;
- or the runtime input to this invocation is not the compiled 210065 bytes we think it is;
- or the split-condition interpretation still has a missing condition;
- or an unobserved `s4` mutation exists outside the preserved slice.

## Next static gate

Do not descend further into `08A23064` until this contradiction is resolved.

The next highest-value checks are:

1. enumerate every `s4` definition in the complete preserved `z_un_0886c84c` body and confirm no reset before C9B4;
2. identify the exact build/commit that produced the freeze dump and compare its 210065 compiler path with the current forced-layout path;
3. recover or deterministically materialize that build's 210065 bytes and compare offsets `0x100..0x103` with `A4 89 4C 8C`;
4. if the equality holds, promote internal-buffer overflow to root cause and add an encoded-page-capacity build gate rather than another ad-hoc layout-only workaround.

## Evidence policy

- `s4==0` + no known reset is strong path evidence, but not mathematical proof until the complete body is accounted for.
- A single failed breakpoint tracer is supporting evidence only.
- `0x8C4C89A4` must not be labeled an invalid address solely by numeric appearance.
- No root-cause promotion until the build/payload contradiction is resolved.
