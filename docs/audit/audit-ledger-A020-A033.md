# Zill O'll Infinite Plus Korean Patch — Audit Ledger Continuation A-020–A-033

This file continues `docs/audit/audit-ledger.md`. It is authoritative for A-020 onward and for supersessions introduced by those entries. Historical statements in the primary ledger are retained as history; where they conflict with this continuation, the newer evidence below controls.

## A-020 — Runtime repetition evidence policy

- One runtime non-reproduction is not proof of safety.
- H0/B/A PASS labels from the historical matrix mean only that the freeze was not reproduced in that recorded run.
- A reproduced freeze/crash is strong negative evidence for that build/path.
- See `A-020-runtime-repetition-evidence-policy.md`.

## A-021 — Failed-build common static conditions

- Compared known failed builds for common static conditions instead of treating passing matrix arms as allowlists.
- Result narrowed what is genuinely shared by failures but did not identify a unique root cause.
- See `A-021-failed-build-common-static-conditions.md`.

## A-022 — Expanded EBOOT is not necessary for the freeze

- Historical PR #14 expanded fixture contained 46 fixed fields.
- A reproduced freeze build used the H0 two-field EBOOT, not the expanded 46-field fixture.
- Therefore an expanded EBOOT string set is **not necessary** for the freeze.
- This supersedes any primary-ledger wording that treats “expanded EBOOT + mapping” as a required condition.
- The old matrix remains useful as recorded outcomes, but PASS arms remain one-shot non-reproductions and do not by themselves prove an interaction law.
- See `A-022-expanded-eboot-not-necessary.md`.

## A-023 — C5 contract provenance gap

- The retained C5 256-byte / three-line / nine-page contract entered public history already materialized; the original retail disassembly or generator establishing it is absent.
- Treat it as strong upstream-maintainer evidence, not independently reproduced retail-runtime proof.
- See `A-023-c5-contract-provenance-gap.md`.

## A-024 — C5 scanner synthetic ELF gate

- Added a heuristic executable-PT_LOAD MIPS scanner and synthetic ELF regression.
- Synthetic candidate detection is deterministic; candidate identity with the retail C5 handler remains OPEN.
- Zero retail candidates would not refute the contract.
- See `A-024-c5-scanner-synthetic-elf.md`.

## A-025 — Caller/substitution `$15` focus

- Narrowed the diagnostic focus from generic `<value>` adjacency to the unresolved `$15` runtime substitution path associated with ID 10010.
- No global semantic label for `$15` is justified by the corpus alone.
- See `A-025-caller-substitution-15.md`.

## A-026 — Retail focus-record context

- Preserved the exact focus-record/category/control context needed for later authenticated-asset checks while keeping corpus category labels distinct from runtime provenance.
- See `A-026-retail-focus-record-context.md`.

## A-027 — ID 10010 stored C5 headroom

- Current compiled ID 10010 payload before `<end>` is 69 stored bytes, including the two-byte `02 15` token.
- It is therefore far from the 256-byte stored boundary.
- This strongly weakens the naive theory that ordinary Korean text plus a short `$15` substitution simply fills a C5 256-byte buffer.
- It does **not** establish the runtime substitution destination, scratch-buffer size, downstream copy size, or `$15` source length.
- See `A-027-id10010-stored-c5-headroom.md`.

## A-028 — C5 authored command does not carry the `$15` string

- Authored C5 syntax carries display/entity/message-ID integers, not an explicit substitution string argument.
- Therefore the model “C5 directly passes the `$15` replacement string as an authored argument” is false.
- A shared formatter resolving `$15` from association/global state after message lookup remains plausible.
- See `A-028-c5-does-not-carry-value15.md`.

## A-029 — Register-linked substitution scanner

- Strengthened the `$15` scanner so prefix and opcode evidence must be register-linked, not merely colocated literals/loads.
- Literal 0x02/0x15 construction is accepted only from the zero register; linked source increments and scaled-index evidence are scored only when tied to the same loaded stream/opcode register.
- Scanner output remains heuristic; jump-table/table-driven implementations may evade it.
- See `A-029-register-linked-value-scanner.md`.

## A-030 — Substitution candidate disassembly context

- Real ISO-build logging now preserves compact decoded instruction neighborhoods with file and virtual addresses for the strongest substitution candidates.
- This improves reproducibility once an authenticated retail EBOOT is available but does not identify the dispatcher by itself.
- See the corresponding standalone A-030 note.

## A-031 — Substitution control-flow decoder fidelity

- Candidate-window decoding now retains useful direct control-flow and pointer-memory operations instead of opaque words where possible, including MIPS jump/call targets with PC high bits reconstructed.
- This is observability only; runtime semantics remain OPEN.
- See the corresponding standalone A-031 note.

## A-032 — Structured substitution candidate logging

- Real build logs classify candidate-window instructions as call/jump/branch/load/store/address-or-immediate/other without changing candidate scoring.
- This allows mechanical filtering around the strongest retail `$15` candidates.
- See `A-032-structured-substitution-candidate-log.md`.

## A-033 — Current evidence map after A-020–A-032

### Trigger

The primary ledger's old “Current root-cause map” predates A-020–A-032 and overstates what one-shot PASS arms establish. A-022, A-027 and A-028 materially changed the ranking and mechanism hypotheses.

### Result

The current active branches are deliberately kept independent:

1. **Renderer/font contract — active, high priority**
   - Proven design flaw: a PAF-installed/text-unused CP932 key is not automatically safe for Korean reuse (`0x87` icon behavior).
   - Mobile slot ownership filtering is weaker than desktop exact-byte filtering.
   - Full atlas/PAF repack internally verifies content, but hidden direct Page/X/Y or atlas consumers remain unexcluded.
   - Runtime PAF/BST lookup mechanics are still not independently reproduced from authenticated EBOOT.

2. **Runtime substitution/message memory contract — active, high priority**
   - ID 10010 remains a useful `$15` clue, but its stored payload is only 69 bytes.
   - C5 authored syntax does not directly carry a replacement string.
   - The next concrete target is the shared `0x02 <opcode>` formatting/dispatch path: `$15` source pointer -> copy/format routine -> destination -> capacity/downstream copy.
   - Known C5 storage accounting is not equivalent to proving runtime expansion safety.

3. **Historical EBOOT/mapping matrix — evidence retained, mechanism downgraded**
   - The recorded H0/B/A non-reproductions and combined/stable-minimal freezes remain historical evidence.
   - They do **not** prove a deterministic A+B interaction because the PASS arms were one-shot non-reproductions.
   - Expanded EBOOT content is not necessary for the freeze (A-022).
   - Keep the matrix as a comparison dataset, not as a standalone root-cause branch with STRONG causal status.

### Reduced hypotheses

- Simple ID10010/C5 256-byte stored overflow.
- Generic `<value>` directly followed by Hangul.
- Expanded 46-field EBOOT as a necessary freeze condition.
- C5 directly receiving the `$15` replacement string as an authored argument.
- Minimal87 as a complete freeze fix.

### Next asset-backed gates

1. Run the substitution scanner on the authenticated retail EBOOT and retain candidate/window logs.
2. For the strongest candidate(s), follow direct calls and pointer/data movement to identify `$15` source and destination capacity.
3. In parallel, run the renderer candidate scan on the same authenticated EBOOT and establish whether Page/X/Y are always consumed through PAF metadata or whether hidden direct atlas references exist.
4. Do not request another device runtime test until these static/asset-backed paths are exhausted enough to produce a narrower experiment.

### Evidence grade

- **CONFIRMED** for the static facts and supersessions cited above.
- **OPEN** for both final runtime mechanisms and freeze causality.

### Commit

This continuation ledger entry is created on `audit/projection-control-roundtrip` after A-032.