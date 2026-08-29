# Current forensic synthesis

This is the compact current-state index for the Korean runtime audit. It supersedes older root-cause rankings inside `audit-ledger.md` when a later numbered audit note changed the interpretation. The full reasoning trail remains in the ledger and A-xxx notes; do not delete superseded history.

## Runtime evidence rule

See A-020.

- A freeze/crash is positive evidence that the tested configuration **can fail**.
- A single non-freezing run proves only non-reproduction in that run.
- Historical H0/B/A `PASS` observations from PR #14 were single-run non-reproductions and are not safety evidence.
- Repeated non-reproduction may increase confidence but never proves absence of an intermittent bug.
- CI success, Android/APK packaging, authenticated ISO build, and runtime behavior are separate evidence levels.

## Current principal root-cause branches

### 1. Renderer/font runtime contract and renderer-key ownership — active / strong evidence of real defects, causality open

Established facts:

- The assumption `PAF-installed + text-unused CP932 key => safe Korean slot` is false. Observed 0x87 mappings rendered UI/icon graphics instead of the staged Korean raster.
- Mobile slot ownership filtering is weaker than the desktop exact-byte ownership audit.
- Full mobile repack changes Page/X/Y geometry across the 2,637-glyph atlas while preserving key/BST and intended raster/metric semantics.
- A postcondition verifier can prove internal PAF/atlas consistency, but there is not yet an independently reconstructed proof that every retail runtime consumer follows the updated PAF coordinates.
- The public repository does not preserve the remembered retail glyph-lookup disassembly as reproducible evidence.

Open discriminants:

- authenticated exact-byte mapped-key collisions in BOOT/EBOOT/BINDATA;
- authenticated full-font semantic audit and output fingerprints;
- retail glyph lookup/BST/coordinate consumer reconstruction;
- classification of hidden icon/direct-atlas consumers.

### 2. Runtime message consumer / dynamic-substitution memory contract — active / strong static proof gap, causality open

Established facts:

- Current corpus: 42,028 accepted Korean records.
- 10,123 `<value:$XX>` occurrences split into 6,566 inline rendered values, 3,272 predicates, and 285 selectors.
- 2,548 C5 records contain inline runtime values.
- Original `walkC5` contributes zero bytes for substitution tokens to page-payload accounting; it only marks the branch dynamic.
- Generic inline-value/non-space adjacency occurs 5,742 times, so ID 10010's immediate `$15`/Hangul adjacency is not distinctive.
- `$28` player name has a documented maximum of 16 encoded bytes from the eight-character / 17-byte C-string input-storage contract.
- Other inline substitution maxima, including `$15`, remain incomplete.

Important provenance boundary (A-023):

- Upstream intentionally documents/enforces a 256-byte C5 page buffer and treats it as a verified runtime contract.
- However, the public repository does not preserve the original executable address/disassembly/generator proving that contract; `consumer-map.toml` was already present in the initial public Go-project commit.
- Therefore the current audit has not independently reconstructed where substitution expansion happens relative to that 256-byte page buffer or any earlier/later scratch buffer.

Open discriminants:

- authenticated C5 runtime candidate scan and subsequent data/control-flow reconstruction;
- `$15` source/max-length/staging-buffer contract;
- known-max expansion headroom in real authenticated C5 pages;
- formatter/copy order and overflow/check semantics.

## Demoted / superseded interpretations

### Expanded PR #14 EBOOT as a required freeze condition — rejected as necessary condition

See A-022.

Observed-freeze build `08458f8e9924177e6e80c4950ae0653c4e9b4d39` contains only the two H0-era Korean EBOOT fields and explicitly uses `new_eboot_bundle=false`. Therefore the historical 46-field PR #14 EBOOT bundle is not necessary for the freeze phenotype.

EBOOT/UI rendering can still participate as a submechanism, but `expanded EBOOT + mapping` is no longer a co-equal root-cause branch.

### Generic `$15` followed immediately by Hangul — reduced

The byte pattern exists, but thousands of other inline substitutions are directly adjacent to following text. The useful remaining question is the `$15` runtime value and its consumer/buffer contract, not generic adjacency.

### Visible 0x87 icon corruption as the complete freeze cause — insufficient

Relocating visible 0x87 assignments can correct the icon symptom while a freeze is still observed. The unsafe-slot defect is real, but it is not sufficient by itself to explain all runtime failures.

### Known C22/character-choice storage defects as the sole cause — insufficient

Those are real independent defects and must stay fixed, but corrected builds have still exhibited the freeze.

## Historical PR #14 replay status

The current production EBOOT input has 2 H0-era fields. The historical expanded PR #14 table has 46 fields and is preserved only as `docs/audit/fixtures/pr14-eboot-full.toml`.

The authenticated ISO preflight now replays H0/B/A/Combined/Stable-minimal planner policies against the **current** corpus and explicitly labels the output `current_corpus_replay=true`. This is deterministic static comparison, not a claim to reproduce the historical runtime ISOs byte-for-byte.

## Required asset-backed evidence before another device experiment

1. PR #14 current-corpus policy replay: mapping SHA/deltas, EBOOT encoding digests, exact-byte mapped-key hits.
2. Mobile exact-byte ownership audit on authenticated BOOT/EBOOT/BINDATA.
3. Full-font 2,637-glyph semantic postcondition verification and retail/patched PAF+atlas fingerprints.
4. Heuristic C5 retail-executable candidate scan, followed by manual/dataflow validation; zero candidates is not a contract disproof.
5. Font-renderer 0x20-stride candidate scan, followed by actual key/BST/Page/X/Y dataflow validation.
6. Corrected C5 known-expansion report on authenticated Korean output.
7. Full projection compatibility replay on authenticated retail banks.
8. Only after the above, choose a one-variable runtime experiment and record repeated outcomes as failure-count / run-count rather than `PASS` shorthand.

## Evidence hygiene

When a later A-xxx note contradicts an older synthesis, retain the old entry as historical reasoning but use this current synthesis and the later note for present decisions. A diagnostic workaround is never a root-cause fix without causal evidence.
