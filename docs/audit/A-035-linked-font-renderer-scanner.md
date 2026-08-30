# A-035 — Register-linked font renderer candidate scanner

## Trigger

The existing `tools/forensics/font-renderer-scan.go` treated any file-wide `sll ...,5` with a nearby load as a candidate. It only checked the ELF magic and then scanned every aligned word in the file, including non-executable regions. Nearby loads were not required to use the stride-derived address.

## Hypothesis

A useful authenticated-retail lead should at minimum show a local record-address shape rather than simple proximity: a 32-byte stride result should feed an address-add, and the resulting register should then be used for loads from offsets inside a 0x20-byte record.

## Verification

The scanner now:

- parses the input as ELF32 little-endian MIPS;
- scans executable `PT_LOAD` file ranges only;
- anchors on `sll ...,5`;
- requires a following `addu` that consumes the shifted register;
- requires at least one byte/halfword/word load whose base is the derived record pointer and whose signed offset is within `[0, 0x20)`;
- reports ELF virtual addresses as well as file offsets;
- ranks candidates by the number of linked in-record loads;
- retains the warning that zero candidates do not disprove a renderer that computes addresses differently.

Synthetic ELF regressions cover:

1. positive linked stride → address-add → two in-record loads;
2. rejection of unrelated nearby loads;
3. rejection of a load at offset `0x20`, outside the authenticated record extent;
4. rejection of the same positive shape in a non-executable `PT_LOAD`.

## Result

The scanner no longer treats arbitrary file data or unrelated load proximity as renderer evidence. A reported candidate now carries a concrete local register-flow relationship compatible with indexing a 0x20-byte record table.

This still does **not** prove that a candidate is the retail PAF/BST glyph lookup routine. It also does not prove the meaning of any field load.

## Evidence grade

- **CONFIRMED** for scanner behavior once CI passes.
- **STRONG DESIGN EVIDENCE** that the new candidate definition is more specific than the previous proximity-only heuristic.
- **OPEN** for authenticated retail candidates, actual PAF base provenance, BST traversal, key comparison, Page/X/Y consumption, hidden direct-atlas consumers, and freeze causality.

## Explicit blind spots

A zero-candidate result remains non-dispositive. The scanner can miss implementations that:

- multiply by 32 without `sll ...,5`;
- fold the stride into another addressing sequence;
- precompute record pointers in a caller;
- use a helper or table that separates the stride operation from field loads beyond the scan radius;
- access a different structure that also happens to use 0x20-byte records.

## New question

On authenticated retail EBOOT, do surviving candidates show additional PAF-specific evidence such as key comparison/BST child traversal and subsequent Page/X/Y field consumption, or do they collapse as unrelated 0x20-byte structures?
