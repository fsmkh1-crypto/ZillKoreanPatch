# A-055 — 210065 materialization gate and runtime-byte discriminant

Date: 2026-08-30

Status: **static materialization path verified on current HEAD; H1 runtime provenance still unresolved**.

## What this gate establishes

Commit `3dd57e6cbde7ca3f26bd4cbca9523c436447ec62` adds a block-terminator-aware unit test for ID 210065. The synthetic source models the relevant runtime structure as ordinary text followed by the source-owned `05 05 05` block terminator and final NUL suffix.

The test drives `CompileBankKorean` through the same forced 210065 diagnostic layout used by the current Korean compiler and asserts:

- the fixed `<end>` / `05 05 05` control is preserved by projection/materialization;
- the materialized payload contains exactly seven `0x0A` line breaks;
- those breaks produce exactly eight text lines;
- every encoded line is at or below the audited C22 56-byte line contract;
- the complete materialized record is below `0x100` bytes in this isolated record fixture.

GitHub Actions CI run `33265884270` completed successfully, including `go test ./...`, `go vet ./...`, Python audit tests, `zill check`, `korean-check`, and `korean-font-check`.

This is static evidence about the current compiler path. It is **not** proof that a previously tested H1 ISO contained these exact bytes.

## Relationship to A-054

A-054 explains one captured freeze with a concrete machine-state chain:

- inline page begins at `s0+0x2C0`;
- the next pointer slot begins exactly `0x100` bytes later at `s0+0x3C0`;
- captured `s4=0` indicates no successful additional-page allocation was traversed in that invocation;
- captured `s3=0x113` records an approximately 275-byte first-page span;
- the later second scanner consumes the corrupted `+0x3C0` word as a pointer.

The current eight-line compiler gate is incompatible with a 275-byte unbroken first-page span **if those compiled bytes are actually the bytes consumed by the runtime**.

Therefore the remaining question is no longer whether the source code *intends* to split 210065. The decisive question is whether the exact runtime invocation that freezes receives the split bytes.

## Decisive next runtime capture

At the earliest reliable breakpoint after the first page has been populated but before the second scanner consumes `s0+0x3C0`, capture at least the following from the same invocation:

1. `s0`, `s3`, `s4`, and the scanner-boundary PC;
2. memory from `s0+0x2C0` through at least `s0+0x3DF` (the full 0x100-byte inline page plus the first 0x20 bytes of following fields);
3. the word at `s0+0x3C0` separately for easy correlation;
4. the number and offsets of `0x0A` bytes before the first NUL in the inline page;
5. the longest byte span between page-start / `0x0A` / NUL boundaries.

The raw memory response must be preserved in the JSONL trace. Derived counts are convenience fields only; they must not replace the raw bytes.

## Interpretation matrix

### Seven intended `0x0A` breaks are present, all spans are short, but the same freeze still occurs

A-054 does not explain that H1 reproduction. Keep A-054 as the explanation for the earlier captured run only and open a second runtime defect branch.

### The runtime page is still one long span or lacks the intended breaks

The eight-line source experiment did not reach this consumer as intended. Investigate build/ISO provenance, bank selection, replacement-map application, and any later reconstruction path before looking for a second runtime defect.

### The runtime page begins split but later bytes overwrite `+0x3C0`

The overflow is downstream of static message-bank materialization. Reconstruct the page-population/copy path between bank record consumption and `z_un_0886C84C` object layout.

## Evidence rule

A successful unit test proves only the current deterministic materializer behavior. A runtime freeze remains stronger evidence about the tested runtime artifact. Do not convert this gate into a claim that H1 was safe or that the prior freeze observation was mistaken.
