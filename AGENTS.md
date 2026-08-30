# ZillKoreanPatch Project Rules

## NON-NEGOTIABLE PROJECT PREMISE — ENGLISH PATCH FIRST

Where the existing English patch implements an engine-facing contract, that implementation is the primary reference for the Korean patch.

This premise applies project-wide, including but not limited to:

- message compilation and materialization
- consumer-specific storage limits
- line, page, and whole-record byte limits
- control-token topology and semantics
- dynamic substitutions such as `<value:$XX>`
- layout generation and page boundaries
- fixed-size fields and bounded labels
- bank/archive capacity and replacement semantics
- font/glyph mapping, renderer slots, atlas/PAF transforms
- static/precomputed font transforms
- BOOT.BIN / EBOOT.BIN / bindata handling
- ISO staging, authoring, and final payload provenance

The Korean path may diverge only where Korean text or the custom renderer genuinely requires different representation. Such a divergence must still preserve the underlying engine contract and must be documented with evidence.

A Korean-specific heuristic MUST NOT supersede an established English-patch contract unless there is evidence that the English contract is inapplicable to that consumer.

## Required order of work

When a Korean-path defect, freeze candidate, overflow, layout anomaly, or rendering issue is found:

1. Find the corresponding English-patch path and identify its consumer/engine contract first.
2. Check whether the Korean path preserves the same contract end-to-end.
3. Classify each comparison as one of:
   - `PASS`
   - `DIFFERENT-BY-DESIGN` — only with written evidence
   - `MISSING`
   - `UNKNOWN`
4. Fix `MISSING` contracts before inventing a new Korean-only rule.
5. Keep consumer-specific contracts scoped to the consumers that actually use them; do not generalize one consumer's scanner or buffer rule across unrelated record types.
6. Validate using the final encoded/materialized bytes where possible, not Unicode character count or source-text estimates.
7. Keep canonical Korean translation semantics separate from build-owned layout/projection whenever possible.
8. Fail closed when an established engine contract cannot be proven after materialization.

## Evidence discipline

- One successful runtime test means only that the freeze was not reproduced in that run.
- One freeze/crash is strong failure evidence.
- Do not treat PASS and FAIL symmetrically.
- Do not promote a hypothesis to universal root cause without sufficient evidence.
- Prefer static, compile-time, asset-backed, and exact-materialization gates before repeated runtime probing.
- A same-screen freeze does not by itself prove the same machine-state cause.

## Review requirement

Every change that touches Korean message/layout/rendering/storage/build paths must answer:

- `English patch parity checked: YES / N/A`
- `English reference/consumer contract:`
- `Korean divergence, if any:`
- `Evidence for divergence:`

Unexplained divergence from an established English-patch engine contract is a release blocker.
