# Zill O'll Infinite Plus Korean Patch — Audit Ledger

This document is the primary evidence ledger for the end-to-end technical audit of the Korean patch. It exists to prevent hypotheses, partial test results, and confirmed facts from being conflated as the project evolves.

## Evidence levels

- **CONFIRMED** — directly established by authenticated assets, byte-identical upstream code, deterministic tests, or reproducible runtime evidence.
- **STRONG** — supported by multiple independent observations or a narrow technical implication, but not yet proven end-to-end.
- **OPEN** — plausible and worth testing; not yet supported strongly enough to drive a production change.
- **REJECTED / SUPERSEDED** — contradicted, overclaimed, or replaced by stronger evidence.

## Interpretation rules

1. A runtime freeze/crash is strong negative evidence for that build/path.
2. One non-reproducing run is not proof of safety. Repeated runs increase confidence but do not prove absence of a bug.
3. CI success, APK packaging success, ISO build success, and runtime success are separate evidence levels.
4. Device testing is the expensive final gate. Static, corpus, compile, format, package and asset-backed checks should be exhausted first.
5. Upstream English-patch behavior is a strong reference, not automatic proof that every runtime consumer is complete or bug-free.
6. A diagnostic workaround must not be described as a root-cause fix until causality is established.
7. When evidence is unavailable in the repository, historical recollection alone is not upgraded to CONFIRMED.
8. Independent review findings are not accepted automatically: each is checked against current code/evidence, then accepted, narrowed, or rejected with the disposition retained in this ledger or its linked audit note.

---

## A-001 — English patch as architectural baseline

**Trigger**  
The Korean patch was originally reverse-engineered and implemented using `HK47196/zill` as the primary reference. Current runtime instability raised the possibility that assumptions inherited or reimplemented from that work were being treated as stronger than the evidence warranted.

**Question**  
Which parts of the Korean architecture are direct upstream inheritance, which are modifications, and which are Korean-only inventions?

**Checks performed**
- Compared executable patch manifest and supporting docs with upstream.
- Compared common release/message infrastructure.
- Identified Korean-only compiler/materializer/font/mobile paths.

**Result**  
Executable wide-message-offset/message-arena manifest content is byte-identical to upstream. Large parts of general release infrastructure are inherited. Korean semantic materialization, Korean slot allocation, mobile font planning, dynamic full-font repack, and Korean QA/storage logic are Korean-specific divergence points.

**Evidence** — **CONFIRMED**

**Meaning**  
A simple “bad port of the 12 wide-offset instructions” is less likely than before. Korean-only layers deserve higher priority. However, upstream completeness of the 12 sites is not independently reproven by identity alone.

**Next questions**
- Are all runtime bank consumers covered by upstream's 12 sites?
- Do Korean-only layers preserve all upstream runtime contracts?

---

## A-002 — 279-bank wide-format compilation and insertion

**Trigger**  
A possible explanation for freezes was a mixed retail/wide bank set or a bank omitted from replacement.

**Checks performed**
- Traced upstream release compilation and archive insertion.
- Traced Korean release compilation and insertion.
- Reviewed Korean `VerifyWideBank` checks.

**Result**  
Both flows require all 279 banks. Korean banks are compiled in wide format and verified before insertion. Common archive insertion requires the complete set.

**Evidence** — **CONFIRMED**

**Meaning**  
“Some banks accidentally remain retail uint16 while others are wide” and “some bank replacements are silently omitted” are substantially reduced as root-cause candidates.

**Still open**  
A separate runtime consumer not covered by upstream wide patches remains possible.

---

## A-003 — Korean bank binary format

**Trigger**  
The Korean compiler itself could have generated malformed bank headers or offsets.

**Checks performed**
- Compared upstream `CompileBank` and Korean `CompileBankKorean` table construction.
- Compared count, reserved field, table end, uint32 absolute offsets, sequential payload placement and per-section capacity logic.

**Result**  
Normal Korean bank structure matches the upstream wide-bank representation.

**Evidence** — **CONFIRMED** for normal compiler structure.

**Meaning**  
Malformed normal bank headers/tables moved down the candidate list.

**Caveat**  
Diagnostic experiment branches such as wide32 boundary probes intentionally alter placement and must not be treated as normal compiler behavior.

---

## A-004 — Korean control materialization

**Trigger**  
A Korean semantic path could reorder or lose source-owned controls such as `<value:$XX>`, fixed controls, and line breaks.

**Checks performed**
- Verified `Projection` topology is unchanged from upstream.
- Verified English/Korean materialization share the core `materializeValues()` control writer.
- Verified Korean control-contract validation requires source/Korean fixed runtime controls to preserve count/order.
- Added projection compatibility audit and synthetic regression tests for fixed-line-break/control adjacency and substitutions.

**Result**  
Control byte emission itself is shared and ordered. Korean source validation already rejects control-count/order drift. The audit code passes synthetic/CI tests.

**Evidence** — **CONFIRMED** for code-level contract; **OPEN** for full authenticated 43,116-record retail replay because CI lacks retail assets.

**Meaning**  
A generic “Korean materializer randomly drops `<value>` controls” theory is reduced.

**Next gate**  
Asset-backed full-corpus byte-identical replay of upstream vs Korean semantic path using the same source text.

---

## A-005 — Fixed line-break semantic divergence

**Trigger**  
Korean `splitSemanticWith()` permits source-fixed line breaks to be omitted from canonical Korean text, unlike upstream `SplitSemantic()`.

**Checks performed**
- Traced delimiter handling and fixed-node raw reinsertion.
- Added compatibility audit designed to detect byte differences.

**Result**  
Fixed line-break raw bytes are reinserted during materialization; they are not simply lost. However, delimiter/fragment assignment differs from upstream and still requires full asset-backed replay.

**Evidence** — **STRONG** that raw control preservation is correct; **OPEN** for all real fragment combinations.

**Meaning**  
This remains a narrow semantic-divergence audit target rather than a leading freeze hypothesis.

---

## A-006 — C5/C20/C22 and fixed-buffer storage contracts

**Trigger**  
Runtime freezes could come from text that is structurally valid but exceeds consumer-specific storage/page limits.

**Checks performed**
- Reconstructed English-patch consumer contracts.
- Audited Korean corpus against known contracts.
- Identified 488 records requiring attention outside the C5 precision set.
- Confirmed examples such as ID 210065 violate C22 without explicit layout.

**Result**  
Real storage/layout defects exist independently of the intermittent freeze. Correcting them is necessary but previous corrected builds still froze.

**Evidence** — **CONFIRMED**

**Meaning**  
Storage violations are real bugs, but they are not sufficient to explain the principal freeze by themselves.

**Important interpretation**  
A previous PASS of a record/build is only “failure not reproduced in that run”; it is not an allowlist or proof of safety.

---

## A-007 — ID 10010 `<value:$15>` adjacency

**Trigger**  
One isolation run suggested the freeze disappeared when ID 10010 was bypassed. Its Korean payload places custom Hangul immediately after `<value:$15>`.

**Checks performed**
- Confirmed materialized byte pattern is `02 15 [custom renderer-key bytes] ...`.
- Added/retained a separator experiment.
- Later quantified all inline-value adjacency across the current corpus in A-018.

**Result**  
The adjacency pattern is real, but it is not distinctive: the current corpus contains 5,742 inline-value/non-space adjacencies. Causality is not established and later builds still fail.

**Evidence** — **OPEN** for `$15` consumer/storage semantics; generic adjacency explanation **REDUCED**.

**Meaning**  
10010 remains a diagnostic clue for the `$15` runtime contract, not a proven root cause. The separator is an experiment, not a production fix.

---

## A-008 — PAF structure reverse engineering

**Trigger**  
Korean glyph reuse depends on accurate interpretation of the retail PAF table.

**Checks performed**
- Reviewed early parser history and authenticated-asset corrections.

**Result**  
An early hypothesis incorrectly treated the final record as truncated and used size `0x149c0`. Authenticated asset inspection corrected this to `0x149d0`, 2,637 complete records, stride `0x20`, root 1318, strict ascending keys and complete per-record fields.

**Evidence** — **CONFIRMED** for current parser contract; early interpretation **SUPERSEDED**.

**Meaning**  
The project history demonstrates why old reverse-engineering conclusions must carry provenance/evidence rather than be remembered as timeless facts.

---

## A-009 — Minimal one-glyph Korean font PoC

**Trigger**  
Need to establish what the earliest successful Korean rendering experiment actually proved.

**Checks performed**
- Recovered commit `a67da949...`.

**Result**  
The initial one-glyph PoC changed only 60 atlas bytes for an existing CP932 renderer key while preserving PAF key, metrics, BST, record count and every other font-member byte.

**Evidence** — **CONFIRMED**

**Meaning**  
At least that tested renderer path can display a Korean bitmap under an existing retail key without changing PAF metadata. It does **not** prove all keys use the same runtime path or that arbitrary slot reuse is safe.

---

## A-010 — “Unused CP932 key = safe Korean slot” assumption

**Trigger**  
Some Korean mappings using nominal `0x87xx` keys rendered game icons instead of Hangul.

**Checks performed**
- Reviewed slot planner assumptions and runtime observation.
- Separated observed symptom from unproven renderer mechanism.

**Result**  
The general assumption that a PAF-installed, text-unused CP932 key is automatically safe for Korean reuse is false.

**Evidence** — **CONFIRMED** that the assumption is false.

**Meaning**  
This is the clearest already-proven design flaw in the Korean-only font layer.

**Not proven**
- That all `0x87xx` keys are a renderer-private namespace.
- That executable code special-cases the 0x87 lead byte.
- That the icon symptom is the same root cause as the freeze.

---

## A-011 — Minimal87 diagnostic relocation

**Trigger**  
Observed icon substitution on selected 0x87 mappings.

**Checks performed**
- Relocated H0 mappings with 0x87 lead to spare round-trip Han keys while preserving unaffected mappings.

**Result**  
Minimal87 is a diagnostic avoidance strategy. Later PR #14 matrix evidence shows that correcting the known 0x87 visual corruption is not sufficient to eliminate the freeze when combined with Korean EBOOT strings.

**Evidence** — **CONFIRMED** as an experiment; root-cause interpretation **OPEN**.

**Meaning**  
Do not describe Minimal87 as proof that 0x87 is globally renderer-private or as the principal freeze fix.

---

## A-012 — Desktop vs mobile slot ownership safety

**Trigger**  
Runtime problems are being observed through the mobile-built patch path, and 0x87 exposed an ownership-assumption failure.

**Checks performed**
- Compared desktop and mobile planner inputs.
- Traced history of the exact-byte filter and later mobile literal-only decision.
- Added mobile exact-byte audit instrumentation.

**Result**  
Desktop slot planning can exclude a renderer key if its exact two-byte value appears anywhere in authenticated BOOT/EBOOT/BINDATA. Mobile deliberately uses CP932-literal scanning instead and does not apply the same raw exact-byte exclusion.

**Evidence** — **CONFIRMED**

**Meaning**  
Mobile has a weaker ownership proof. A key referenced as binary data/table/single constant can survive mobile filtering even though desktop would exclude it.

**Next gate**  
On authenticated retail assets, enumerate mobile-mapped keys that would fail desktop exact-byte ownership and classify every occurrence.

---

## A-013 — Dynamic full atlas + PAF repack

**Trigger**  
Unlike the English patch's frozen font transform, mobile Korean dynamically repacks all stock and Korean glyphs, changing stock atlas positions.

**Checks performed**
- Audited full-repack implementation.
- Added `VerifyFullRepackSemantics()` postcondition verification.

**Result**  
The repacker preserves keys/BST; stock metrics and raster content are intended to be preserved while Page/X/Y may change. Custom slots receive Korean raster/metrics. The new verifier reparses the result and checks all 2,637 records and stock/custom raster semantics.

**Evidence** — **CONFIRMED** for verifier design/CI; **OPEN** for actual retail-asset execution.

**Meaning**  
Internal repack consistency can now be proven, but that alone does not prove every runtime consumer follows the updated PAF Page/X/Y instead of using hidden direct atlas references.

---

## A-014 — Runtime glyph lookup / BST contract

**Trigger**  
To explain 0x87 icon substitution and judge whether arbitrary repacking is safe, the actual retail renderer lookup path must be established.

**Checks performed**
- Searched repository history for reproducible renderer-disassembly evidence.
- Recovered PAF file-format evidence but not a preserved runtime lookup proof.
- Added `tools/forensics/font-renderer-scan.go` to identify MIPS candidates using `*32` PAF-like access without hardcoding remembered addresses.

**Result**  
Repository evidence currently proves the PAF data structure, not the exact retail runtime function that walks it.

**Evidence** — **OPEN** for runtime lookup mechanics.

**Meaning**  
Historical recollections such as a specific `0x1F9F8C` lookup site are not treated as CONFIRMED until reproduced against authenticated EBOOT.

**Next gate**  
Asset-backed candidate scan, then verify key/left/right/page/metric field use and caller/dataflow.

---

## A-015 — Crash-address interpretation

**Trigger**  
PPSSPP reported Bad Execution Address `0x7922f770`, RA `0x08948dcc`.

**Checks performed**
- Added/recovered ELF `PT_LOAD` address mapper.
- Rejected raw subtraction or guessed file-offset mapping.

**Result**  
Runtime addresses must be translated via the actual ELF program headers and known module base. BSS-only addresses are rejected.

**Evidence** — **CONFIRMED** for mapping methodology; actual crash root **OPEN**.

**Meaning**  
The crash is strong evidence of pointer/control-state corruption or invalid dispatch somewhere in the path, but the screenshot alone does not prove the overwrite source or identify the RA routine.

---

## A-016 — Wide32 boundary experiment design

**Trigger**  
Need to determine whether a runtime consumer fails when a bank record offset crosses `0xFFFF`.

**Checks performed**
- Reviewed PR #16 implementation and build/test history.

**Result**  
The experiment does **not** move only ID 10007. Padding before 10007 shifts 10007 and the entire following suffix. It may also change the apparent extent of the preceding record if any consumer uses `next_offset-current_offset`. CI/APK packaging did not execute the real retail/corpus ISO path, and the actual patcher later failed before ISO creation.

**Evidence** — **CONFIRMED** that the earlier experiment description was overclaimed; experiment result **INCONCLUSIVE**.

**Meaning**  
PR #16 is not a clean one-variable runtime test and must not be used as proof for or against wide32 support.

**Next gate**  
Real-corpus preflight and record-termination semantics before any redesigned boundary test reaches device testing.

---

## A-017 — CI/package/runtime evidence separation

**Trigger**  
A repeated process mistake was treating successful CI/APK packaging as sufficient readiness for user runtime testing.

**Checks performed**
- Traced Android workflow vs `build-korean-iso` path.
- Compared failed actual builder behavior to prior successful Actions runs.

**Result**  
Android build success proves the app/payload can package; it does not prove the authenticated retail ISO can be patched or the generated game runs.

**Evidence** — **CONFIRMED**

**Meaning**  
Future device builds require an asset-backed or maximally equivalent preflight first. User testing is the final gate, not a substitute for build validation.

---

## A-018 — Runtime `<value>` roles and C5 dynamic-expansion boundary

**Trigger**  
The message-memory branch needed to distinguish rendered substitutions from control-flow reads and determine what static C5 validation actually proves.

**Checks performed**
- Inventoried all `<value:$XX>` uses in the current 42,028-record Korean corpus.
- Split inline/predicate/selector roles and added context/counterexample reporting.
- Read and directly regression-tested original `walkC5` substitution behavior.
- Added exact branch/page substitution-position accounting.
- Corrected the Korean C5 third-line-break page-byte off-by-one.
- Rechecked upstream provenance of the `$28` player-name bound.

**Result**
- 10,123 total `<value>` occurrences; 6,566 inline rendered values, 3,272 predicate reads, 285 selector reads.
- 2,548 C5 records contain inline runtime values.
- 5,742 inline values are immediately followed by non-space text; generic adjacency is therefore not distinctive to ID 10010.
- Original `walkC5` adds **zero bytes** for substitution tokens to page payload accounting; it only marks the leaf dynamic.
- Independent-review counterexample scan for comparison-RHS `<value>` forms found **0 current-corpus cases**.
- Upstream documents `$28` player-name expansion as max 16 encoded bytes under an eight-character input and 17-byte C-string storage contract.

**Evidence** — **CONFIRMED** for the static proof boundary, corpus counts, page-boundary bug/fix and documented player-name bound; actual runtime overflow causality **OPEN**.

**Meaning**  
Generic `<value>`/Hangul adjacency moves down. A more precise, testable question moves up: can actual runtime-expanded pages or intermediate staging buffers exceed their capacity even though static C5 validation passes?

**Detailed note**  
See `docs/audit/A-018-dynamic-value-runtime-boundary.md`.

---

## A-019 — EBOOT + renderer-mapping interaction matrix

**Trigger**  
Independent review identified that the two-branch synthesis failed to preserve PR #14's user-verified H0/A/B/combined/minimal runtime matrix.

**Checks performed**
- Recovered PR #14 issue comment `5453956902` and experiment commit IDs.
- Separated the explicit-layout H1 experiments from the H0 mapping/EBOOT matrix.

**Result**
- H0 baseline: PASS in the recorded run; known 0x87 icon corruption remains.
- H0 + Korean EBOOT UI only: PASS; icon corruption remains.
- H0 + full Han-safe remap only: PASS, but auto-wrap behavior regresses.
- Full Han-safe + Korean EBOOT UI: FAIL before 210065 subtitle.
- H0 mapping + Korean EBOOT UI + minimal 0x87 relocation: FAIL; visible 0x87 corruption is corrected.

**Evidence** — **CONFIRMED** as the recorded runtime outcome matrix; underlying interaction mechanism **OPEN**.

**Meaning**  
Neither “Korean EBOOT strings alone” nor “global remapping alone” explains the recorded matrix. Fixing the known 0x87 visual corruption is also insufficient. The matrix therefore remains a separate cross-cutting interaction branch until exact generated mapping/font/executable deltas explain it.

**Detailed note**  
See `docs/audit/A-019-eboot-mapping-interaction-matrix.md`.

---

## Independent review disposition — 2026-08-29

Claude's independent PR #17 review was treated as adversarial input rather than accepted wholesale.

**Accepted and fixed**
- C5 third-line-break boundary byte was omitted from both the Korean validator and first forensic helper. Korean accounting now counts the boundary byte before page transition and regression-tests the behavior.
- The new forensic `$33` select walker duplicated an unexplained `arms = 8` literal. The audit now centralizes that inherited assumption and explicitly leaves its retail-runtime derivation OPEN.
- PR #14's EBOOT/mapping interaction matrix was missing from the synthesis. Added as A-019.

**Accepted concern, then strengthened with repository evidence**
- Reviewer challenged `$28 = 16B` as potentially chronicle-local. Upstream documentation explicitly ties the bound to the player-name eight-character input and 17-byte C-string storage contract, so the current evidence supports a player-name-wide bound rather than a chronicle-only bound.

**High-value tests closed**
- Original `walkC5` substitution magnitude: direct regression test confirms the substitution contributes zero bytes to `leaf.data`, not the two-byte `02 XX` marker.
- Python classifier right-hand predicate-value blind spot: explicit current-corpus scan finds zero such counterexamples. The classifier remains a triage heuristic, not a parser proof.

**Still open after review**
- Independent retail-runtime proof of `$33` fixed eight-arm select semantics.
- `$15` character-creation-prompt source/max-length/staging-buffer contract.
- Whether C5 expands directly into the modeled 256-byte page buffer or through another scratch/staging buffer.

---

## Current root-cause map

### Higher-priority active branches

1. **Korean renderer-slot ownership / hidden glyph consumers — STRONG**
   - A foundational slot-safety assumption is already disproven by 0x87 icon behavior.
   - Mobile ownership filtering is weaker than desktop.
   - Exact collision classification and runtime lookup proof are still pending.
   - PR #14 stable-minimal failure shows fixing the visible 0x87 cases is not sufficient to eliminate the freeze.

2. **Runtime message consumer / dynamic expansion memory contract — STRONG**
   - Bad Execution Address is consistent with state/pointer corruption but does not identify its source.
   - 2,548 C5 records contain inline runtime substitutions whose eventual bytes are excluded entirely from ordinary static page accounting.
   - `$28` now has a documented 16-byte player-name maximum; the other eight inline opcode contracts remain incomplete.
   - The direct expansion destination/scratch-buffer topology remains unproven.

3. **EBOOT + mapping/font/executable interaction — STRONG interaction evidence / OPEN mechanism**
   - PR #14 matrix shows EBOOT-only and remap-only arms passed in their recorded runs while combined and stable-minimal+EBOOT arms failed.
   - This is not evidence that EBOOT runtime strings themselves are the cause; exact generated asset deltas must be compared first.
   - This branch may later collapse into slot ownership, message memory, or another mechanism, but current evidence does not justify doing so yet.

### Medium-priority

4. **Dynamic full-font repack compatibility — OPEN/STRONG**
   - Internal consistency can be verified, but hidden direct atlas-coordinate consumers have not been ruled out.

5. **Korean fixed-line-break semantic fragment assignment — OPEN**
   - Control bytes are preserved by design; full real-corpus replay still pending.

### Reduced / not sufficient alone

- Normal Korean wide-bank header/table corruption.
- Partial 279-bank replacement.
- Korean fixed EBOOT overlay overwriting manifest instructions.
- Known C22/character-choice storage defects as the sole explanation for the intermittent freeze.
- Generic `<value>` immediately followed by Hangul as a unique failure condition.
- ID 10010 separator as a proven fix.
- Minimal87 as a proven root-cause fix.
- Global Han-safe remapping alone as a sufficient explanation of the PR #14 matrix.

---

## Required asset-backed gates before the next device build

1. Run full projection compatibility replay over authenticated retail banks.
2. Run mobile exact-byte slot audit and classify mapped collisions in BOOT/EBOOT/BINDATA.
3. Run full-font semantic postcondition check across all 2,637 PAF glyphs.
4. Run the retail EBOOT font-renderer candidate scanner and establish or reject the PAF/BST lookup hypothesis.
5. Run corrected C5 known-expansion accounting on the authenticated Korean mapping and record known-overflow/unknown-substitution pages.
6. Compare exact mapping/font/PAF/EBOOT outputs across the PR #14 H0, B, A, combined and stable-minimal arms before choosing another runtime A/B.
7. Establish `$15` character-creation-prompt consumer/staging-buffer contract and the remaining inline-opcode maxima where possible.
8. Add/run ISO-free real-corpus Korean bank compile preflight with bank sizes/offsets/capacities recorded.
9. Establish record termination semantics before redesigning the `>0xFFFF` offset experiment.
10. Only after the above, build the authenticated retail ISO and then perform runtime repetitions.

---

## Claude independent review checklist

Review this ledger against the current audit branch and upstream `HK47196/zill`. Do not assume the ledger is correct. For every item:

1. Identify claims whose evidence level is too strong or too weak.
2. Identify missing alternative explanations.
3. Check whether a supposedly upstream-identical component actually diverges in a runtime-relevant way.
4. Check whether Korean-only code introduces a hidden contract not captured here.
5. Examine whether the current leading branches — renderer-slot ownership, runtime message/dynamic-expansion memory behavior, and the PR #14 EBOOT/mapping interaction — are framed appropriately and whether any can now be collapsed or rejected.
6. Specifically challenge the assumption that PAF key/BST preservation plus raster/metric verification is enough for runtime safety.
7. Specifically challenge whether the upstream 12 wide-message-offset patch sites prove complete consumer coverage.
8. Review the proposed asset-backed gates for false confidence, missing gates, or order-of-operations mistakes.
9. Separate findings into MUST-FIX, HIGH-VALUE TEST, and LOWER-PRIORITY.
10. Do not use prior-chat assumptions; cite code paths, commits, or repository artifacts for each substantive conclusion.

This ledger should be updated whenever a new test changes the evidence level of a hypothesis. Do not delete superseded history; mark it as superseded so the reasoning trail remains auditable.
