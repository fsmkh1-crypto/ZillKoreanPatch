# Current forensic synthesis

This is the compact current-state index for the Korean runtime audit. It supersedes older root-cause rankings inside `audit-ledger.md` when a later numbered audit note changed the interpretation. The full reasoning trail remains in the ledger and A-xxx notes; do not delete superseded history.

## Runtime evidence rule

See A-020.

- A freeze/crash is positive evidence that the tested configuration **can fail**.
- A single non-freezing run proves only non-reproduction in that run.
- Historical H0/B/A `PASS` observations from PR #14 were single-run non-reproductions and are not safety evidence.
- Repeated non-reproduction may increase confidence but never proves absence of an intermittent bug.
- CI success, Android/APK packaging, authenticated ISO build, and runtime behavior are separate evidence levels.

## A-054 historical captured-run mechanism — retained, but rejected as the current sole cause

A-054 remains a strong explanation for the preserved historical freeze capture, but A-058 supersedes the earlier ranking that treated this mechanism as the leading blocker for the current HEAD.

Established for the historical captured run:

- `z_un_0886c84c` has an inline text region at `s0+0x2C0 .. s0+0x3BF`; the first secondary-page pointer slot begins immediately at `s0+0x3C0`.
- The captured second scanner call has `s4=0`, while the known split/allocation producer increments `s4` immediately after storing an allocated page pointer. No reset before the second scanner was identified.
- The captured first scanner result is `s3=0x113` (275 bytes), exceeding the 0x100-byte inline region by 19 bytes.
- The later load from `s0+0x3C0` yields `0x8C4C89A4`, which is then passed directly to the NUL scanner and produces the observed runaway scan.
- A separately armed breakpoint at the additional-page allocation boundary was not observed before that freeze. This remains supporting evidence only.

Historical captured-run interpretation:

```text
long first-page/materialized span
  -> copy crosses the inline 0x100-byte region
  -> overwrites the first secondary-page pointer slot
  -> later code interprets overwritten text bytes as a pointer
  -> NUL scanner walks arbitrary memory
  -> freeze
```

A-055 then added a compiler/materialization gate for ID 210065, and the current branch's full-corpus static gate verifies that accepted Korean materializations do not reach the observed `0x100` scanner-span boundary for this mechanism.

A-058 records the decisive runtime result: a patched ISO carrying the current A-055-safe forensic payload still froze at the 210065 scene once. Under A-020, that single freeze is strong failure evidence and is sufficient to reject A-054 as the **sole** explanation for the current HEAD.

Therefore:

- do **not** infer that the historical `s3=0x113`, `s4=0`, or overwritten `s0+0x3C0` state recurred merely because the visible freeze location is the same;
- do **not** spend further tracer work on proving a scanner-window condition unless new independent evidence points back to it;
- do **not** mass-fix current records for the >=0x100 scanner-span mechanism: the current static corpus gate has no offender to fix;
- retain A-054 as a valid historical captured-run mechanism, not as the active sole-cause hypothesis for the current freeze.

The next candidate must independently explain a freeze that survives the current safe materialization and full-corpus scanner-span gate.

## Current principal root-cause branches

### 1. Renderer/font runtime contract and renderer-key ownership — active / strongest independent branch after A-058

Established facts:

- The assumption `PAF-installed + text-unused CP932 key => safe Korean slot` is false. Observed 0x87 mappings rendered UI/icon graphics instead of the staged Korean raster.
- Mobile slot ownership filtering is weaker than the desktop exact-byte ownership audit.
- Full mobile repack changes Page/X/Y geometry across the 2,637-glyph atlas while preserving key/BST and intended raster/metric semantics.
- A postcondition verifier can prove internal PAF/atlas consistency, but there is not yet an independently reconstructed proof that every retail runtime consumer follows the updated PAF coordinates.
- The public repository does not preserve the remembered retail glyph-lookup disassembly as reproducible evidence.
- A-058 removes the current >=0x100 scanner-span mechanism as a sufficient explanation, increasing the relative priority of renderer/key/coordinate ownership work without proving it causal.

Open discriminants:

- authenticated exact-byte mapped-key collisions in BOOT/EBOOT/BINDATA;
- authenticated full-font semantic audit and output fingerprints;
- retail glyph lookup/BST/coordinate consumer reconstruction;
- classification of hidden icon/direct-atlas consumers;
- proof that all consumers of a relocated glyph resolve Page/X/Y through the patched PAF rather than through fixed/direct atlas geometry.

### 2. Runtime message consumer / dynamic-substitution memory contract — active, but `$15` itself reduced

Established facts:

- Current corpus: 42,028 accepted Korean records.
- 10,123 `<value:$XX>` occurrences split into 6,566 inline rendered values, 3,272 predicates, and 285 selectors.
- 2,548 C5 records contain inline runtime values.
- Original `walkC5` contributes zero bytes for substitution tokens to page-payload accounting; it only marks the branch dynamic.
- Generic inline-value/non-space adjacency occurs 5,742 times, so ID 10010's immediate `$15`/Hangul adjacency is not distinctive.
- `$28` player name has a documented maximum of 16 encoded bytes from the eight-character / 17-byte C-string input-storage contract.
- A-025 reconstructs caller index 4 and shows that the relevant formatter's `d: %d ` output for the observed value is exactly `d: 15 `; the Korean fixed literal is the same seven-byte text at the shared runtime boundary. That makes “the numeric value 15 itself” a weak suspect.
- Other inline substitution maxima remain incomplete.

Important provenance boundary (A-023):

- Upstream intentionally documents/enforces a 256-byte C5 page buffer and treats it as a verified runtime contract.
- However, the public repository does not preserve the original executable address/disassembly/generator proving that contract; `consumer-map.toml` was already present in the initial public Go-project commit.
- Therefore the current audit has not independently reconstructed where substitution expansion happens relative to that 256-byte page buffer or any earlier/later scratch buffer.

Open discriminants:

- authenticated C5 runtime candidate scan and subsequent data/control-flow reconstruction;
- known-max expansion headroom in real authenticated C5 pages;
- formatter/copy order and overflow/check semantics for substitutions other than the already reduced caller-4 `$15` case;
- any separate shared-runtime ownership or termination condition that survives the A-055/A-058 scanner-span result.

## Demoted / superseded interpretations

### A-054 >=0x100 scanner span as the sole current blocker — rejected

See A-058.

The historical captured invocation remains explained well by A-054, but the current A-055-safe payload and full-corpus <0x100 scanner-span gate still produced a runtime freeze once. The mechanism is therefore not sufficient to explain the current HEAD by itself.

### Expanded PR #14 EBOOT as a required freeze condition — rejected as necessary condition

See A-022.

Observed-freeze build `08458f8e9924177e6e80c4950ae0653c4e9b4d39` contains only the two H0-era Korean EBOOT fields and explicitly uses `new_eboot_bundle=false`. Therefore the historical 46-field PR #14 EBOOT bundle is not necessary for the freeze phenotype.

EBOOT/UI rendering can still participate as a submechanism, but `expanded EBOOT + mapping` is no longer a co-equal root-cause branch.

### Generic `$15` followed immediately by Hangul — reduced

The byte pattern exists, but thousands of other inline substitutions are directly adjacent to following text. A-025 additionally reconstructs the relevant caller and shows semantic/effective text equivalence for the observed `15` case. Remaining useful questions concern shared runtime ownership and other substitution contracts, not generic adjacency or the integer 15 by itself.

### Visible 0x87 icon corruption as the complete freeze cause — insufficient

Relocating visible 0x87 assignments can correct the icon symptom while a freeze is still observed. The unsafe-slot defect is real, but it is not sufficient by itself to explain all runtime failures.

### Known C22/character-choice storage defects as the sole cause — insufficient

Those are real independent defects and must stay fixed, but corrected builds have still exhibited the freeze.

## Historical PR #14 replay status

The current production EBOOT input has 2 H0-era fields. The historical expanded PR #14 table has 46 fields and is preserved only as `docs/audit/fixtures/pr14-eboot-full.toml`.

The authenticated ISO preflight now replays H0/B/A/Combined/Stable-minimal planner policies against the **current** corpus and explicitly labels the output `current_corpus_replay=true`. This is deterministic static comparison, not a claim to reproduce the historical runtime ISOs byte-for-byte.

## Required asset-backed/static evidence before another device experiment

1. Authenticated exact-byte mapped-key ownership audit across BOOT/EBOOT/BINDATA, using the exact encoded byte pairs actually assigned to Korean glyphs.
2. Full-font 2,637-glyph semantic postcondition verification plus retail/patched PAF+atlas fingerprints.
3. Reconstruct the retail glyph lookup/BST/Page/X/Y consumer and prove whether every relocated glyph follows patched PAF coordinates or whether any consumer derives/directly owns atlas geometry.
4. Classify hidden icon/direct-atlas consumers sufficiently to distinguish an unsafe key collision from a pure glyph-raster/metric issue.
5. Heuristic C5 retail-executable candidate scan, followed by manual/dataflow validation; zero candidates is not a contract disproof.
6. Corrected C5 known-expansion report on authenticated Korean output for substitution cases not already reduced by A-025.
7. Full projection compatibility replay on authenticated retail banks.
8. Only after an independent discriminant survives these gates, choose a one-variable runtime experiment and record repeated outcomes as failure-count / run-count rather than `PASS` shorthand.

Do not rerun the same 210065 device scene merely to accumulate non-freezes after A-058. A new runtime experiment must discriminate a newly supported mechanism.

## Evidence hygiene

When a later A-xxx note contradicts an older synthesis, retain the old entry as historical reasoning but use this current synthesis and the later note for present decisions. A diagnostic workaround is never a root-cause fix without causal evidence.
