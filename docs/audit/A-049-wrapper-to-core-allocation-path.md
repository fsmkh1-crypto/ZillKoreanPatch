# A-049 — Wrapper-to-core allocation path at 0x089E2B54

## Runtime evidence

Captured from the 2026-08-29 freeze trace using tracer commit `82eb7030835fa789a9a12cbac5ae7c9c3e46decc`.

The caller in `z_un_0886c84c` executes:

- `0x0886C958 jal z_un_089E2B54`
- delay slot `0x0886C95C li a0,0x100`
- `0x0886C960 sll a0,s4,0x2`
- `0x0886C964 addu a0,s0,a0`
- `0x0886C968 sw v0,0x3C0(a0)`

For `s4 == 0`, the destination aliases `s0+0x3C0` exactly.

At the freeze, the later consumer loads `s0+0x3C0 = 0x8C4C89A4`, passes it as `a1` to `z_un_089661dc`, and the parser scans forward without encountering a terminating NUL for tens of MiB.

## New disassembly result

`z_un_089E2B54` is not the core allocator/producer. It is a thin wrapper around `z_un_089E2B00`.

Relevant instructions:

- `0x089E2B54 addiu sp,sp,-0x90`
- `0x089E2B58 sw s0,0x80(sp)`
- `0x089E2B5C lui s0,0x8A7`
- `0x089E2B60 lw a1,-0x53B4(s0)`
- `0x089E2B64 addiu a2,sp,0x10`
- `0x089E2B68 sw a1,0x10(sp)`
- `0x089E2B6C li a1,0x2`
- `0x089E2B70 sw a2,-0x53B4(s0)`
- `0x089E2B74 sb a1,0x14(sp)`
- `0x089E2B78 li a1,0x08A6A008`
- `0x089E2B80 sw ra,0x84(sp)`
- `0x089E2B84 jal z_un_089E2B00`
- `0x089E2B88 sw a1,0x18(sp)`
- `0x089E2B8C lw a0,0x10(sp)`
- `0x089E2B90 sw a0,-0x53B4(s0)`
- restore and return.

The wrapper does not modify `v0` after the call. Therefore the value stored by `0x0886C968` is the return value from `z_un_089E2B00` propagated through `z_un_089E2B54` unchanged.

The wrapper temporarily replaces a global/context pointer at `0x08A6AC4C` (`0x08A70000 - 0x53B4`) with a stack-resident record, tags it with value `2`, records constant `0x08A6A008`, invokes `z_un_089E2B00`, then restores the prior global/context pointer. This looks like diagnostic/allocation-context bookkeeping, but the exact semantics remain unconfirmed.

## Evidence level

### Confirmed

1. `z_un_089E2B54` is a wrapper around `z_un_089E2B00`.
2. The wrapper passes the original `a0` through to `z_un_089E2B00`; for the observed caller, that value is `0x100`.
3. `v0` is returned from `z_un_089E2B00` through the wrapper without transformation.
4. The caller stores that returned `v0` into the `s0+0x3C0+s4*4` table.

### Strongly supported

`z_un_089E2B00` is the immediate core producer of the pointer-like value later consumed as a string pointer.

### Still uncertain

- Whether `z_un_089E2B00` is a heap allocator, pool allocator, aligned allocator, arena allocator, or another resource constructor.
- Whether `0x8C4C89A4` is an actually invalid PSP CPU address, a mirrored/encoded address, or a value produced under a broken allocation/context contract.
- Whether the fault is in allocation itself, memory corruption after allocation, incorrect type/ownership, or Korean-patch-specific pressure/input.

## Exact next action

Capture disassembly starting before `0x089E2B00`, sufficiently wide to include its complete body and any immediate callees/return-value construction. Prefer a broad one-shot window so another APK cycle is not needed solely to chase one more hop. Then classify the returned `v0` and compare the same path against retail/English behavior before attributing the fault to Korean data.
