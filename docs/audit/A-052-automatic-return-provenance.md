# A-052 — Automatic MIPS return-provenance walk

Date: 2026-08-30

## Purpose

The runtime breakpoint tracer path is suspended. Device-side capture cost was too high relative to the evidence obtained. This note records a new conservative static tool that automates the repetitive part of the reverse walk: following a function whose return value is an unchanged direct-callee `v0` into that callee until a real value construction or an evidence gap is reached.

Tool:

`cmd/return-provenance`

Typical invocation:

```text
go run ./cmd/return-provenance --input disassembly.txt --start 0x089E2B54
```

## Classification contract

The tool intentionally makes only narrow claims:

- `wrapper-passthrough`: a direct `jal` is the last resolved call before `jr ra`, and no supported instruction writes `v0` after the call's delay slot. The walk may descend to that callee.
- `value-transform`: `v0` is written after the last direct call. The automatic descent stops.
- `value-origin`: a leaf/no-direct-call function has a visible `v0` definition before return. The automatic descent stops at that definition.
- `uncertain`: the available text is insufficient for a conservative classification.
- missing function body: the requested callee is not present in the supplied evidence. The walk stops rather than inventing a next hop.

Partial audit snippets may omit isolated instructions. The parser therefore tolerates a small address gap (up to `0x20`) while looking for the return, but it never crosses a completed `jr ra` plus delay slot.

## Regression cases

Tests cover:

1. the A-049 `0x089E2B54 -> 0x089E2B00` passthrough shape, including the historical omitted `0x089E2B7C` line;
2. a post-call `v0` transformation, which must stop descent;
3. a leaf `lw v0,...` value origin, which must stop descent.

The first implementation failed the A-049 test because it incorrectly required every retained audit instruction to be exactly four bytes after the previous retained line. That was corrected rather than weakening the test fixture.

## Current evidence boundary

Existing checked-in audit evidence proves that `0x089E2B54` passes the return of `0x089E2B00` through unchanged (A-049). Subsequent static/runtime work established the chain through `0x089E2984`, which directly calls `0x08A23064` and saves its returned `v0` for later return/use.

A search of the previously uploaded 136-page runtime dump (`노그.docx`) recovered the `0x089E2984 -> 0x08A23064` call site and the immediate `move s0,v0`, but did not recover a function body whose function label is `z_un_08a23064` or an instruction stream starting at `0x08A23064`.

Therefore the automatic provenance walk currently reaches the same honest boundary as the manual investigation:

```text
0x0886C84C
  -> 0x089E2B54
  -> 0x089E2B00
  -> 0x089E2984
  -> 0x08A23064
       STOP: body not present in current preserved evidence corpus
```

This is an evidence-availability stop, not a claim that `0x08A23064` is the root cause or necessarily the final allocator primitive.

## Next static action

Search preserved repository history, uploaded dumps, or authenticated retail executable material for the actual `0x08A23064` body. Once that body is available as disassembly text, feed it to `cmd/return-provenance`; passthrough wrappers can then be descended automatically until the first visible `v0` construction/transformation or another explicit evidence gap.

No additional APK/device experiment is required merely to operate the provenance analyzer.
