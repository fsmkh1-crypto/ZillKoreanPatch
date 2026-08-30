# A-071 — English-first parity closure matrix

## Purpose

This record consolidates the full English-first contract audit after the mobile-path, archive, font, executable, Android payload, and output-provenance gaps were closed.

The project premise is not that Korean must copy every English implementation detail. The rule is:

1. engine-invariant contracts must be preserved;
2. English implementation policy may differ only where Korean encoding/font mechanics require it;
3. Korean-only machinery may add stricter guards but may not weaken the underlying engine contract.

Status vocabulary:

- **PASS** — the Korean path enforces the same engine boundary, or a stricter Korean-specific equivalent, on the bytes actually consumed downstream.
- **DIFFERENT-BY-DESIGN** — implementation differs for a demonstrated Korean-specific reason while preserving the engine boundary.
- **MISSING** — known engine contract has no equivalent enforcement.
- **UNKNOWN** — evidence is insufficient to classify the boundary.

## Closure matrix

| Surface | Class | Status | Korean enforcement / evidence |
|---|---|---|---|
| Canonical source → Korean projection | ENGINE-INVARIANT | PASS | canonical source binding plus `BuildKoreanBetaProject`; compiler/materializer semantic preservation gates remain release-blocking |
| Control-token / projection semantics | ENGINE-INVARIANT | PASS | shared semantic traversal/materialization plus `message.PreservesLayoutSemantics` |
| C5 branch/page lowering | ENGINE-INVARIANT | PASS | exact Korean lowering and runtime-storage validation before compile; desktop/mobile/preflight all share the gate |
| C20 aggregate/group buffer | ENGINE-INVARIANT | PASS | upstream English consumer membership and shared capacity constant mirrored in Korean validator |
| C22 total/page/line/page-count | ENGINE-INVARIANT | PASS | upstream English limits mirrored with Korean renderer-byte measurement; safe layout derivation precedes validation |
| Bounded labels | ENGINE-INVARIANT | PASS | upstream fixed-width consumer contract mirrored with Korean renderer-byte measurement |
| Guild client / region / posting | ENGINE-INVARIANT | PASS | upstream fixed-width and posting-role/value maxima mirrored |
| Trap text | ENGINE-INVARIANT | PASS | upstream buffer and dynamic-value allowance mirrored |
| Character-creation choices | ENGINE-INVARIANT | PASS | upstream fixed-width capacity mirrored |
| Equipment feedback | ENGINE-INVARIANT | PASS | upstream fixed-width capacity mirrored |
| Chronicle entries | ENGINE-INVARIANT | PASS | upstream payload/player-name dynamic allowance mirrored |
| Dynamic `<value>` maxima in known fixed consumers | ENGINE-INVARIANT | PASS | same upstream maxima/policies used by Korean consumer validator |
| Message bank capacity / table writes | ENGINE-INVARIANT | PASS | Korean compiler uses the shared runtime bank-capacity/table contract before archive insertion |
| Retail BINDATA authentication/layout | ENGINE-INVARIANT | PASS | retail SHA/layout guarded; 132-record equipment structure validated before renderer-slot planning |
| BINDATA equipment translation coverage | ENGLISH-POLICY | DIFFERENT-BY-DESIGN | English translates equipment names; Korean currently has no equipment translation source, so authenticated retail bytes are intentionally retained |
| BOOT/EBOOT source authentication | ENGINE-INVARIANT | PASS | authenticated retail BOOT/EBOOT required by both Korean planners |
| Executable patch manifest chain | ENGINE-INVARIANT | PASS | shared manifest applied first; patched ELF fingerprint authenticated; localized overlay is post-verified against manifest spans |
| Fixed EBOOT string coverage | ENGLISH-POLICY | DIFFERENT-BY-DESIGN | English has a complete fixed-string table; Korean uses a reviewed sparse overlay but keeps source guards, fixed capacity, NUL, overlap, and clear-then-copy invariants |
| Renderer-slot ownership | KOREAN-ADDITIONAL | PASS | BOOT/EBOOT/BINDATA exact-byte references excluded before allocation; mobile mapping mutation is re-audited afterwards |
| Desktop font slot universe | KOREAN-ADDITIONAL | DIFFERENT-BY-DESIGN | atlas-only desktop path restricts allocation to geometry-compatible installed slots and does not mutate the chosen mapping after `BuildPlan` |
| Mobile font slot universe | KOREAN-ADDITIONAL | DIFFERENT-BY-DESIGN | full authenticated atlas+PAF repack can use installed double-byte slots; private `0x87` relocation is followed by exact-byte ownership re-audit |
| Retail font source identity | ENGINE-INVARIANT | PASS | Korean path pins the same atlas/PAF retail source SHA-256 fingerprints as the upstream English font manifest |
| Korean font result boundary | KOREAN-ADDITIONAL | PASS | full-repack verifier allows only modeled atlas payload and PAF geometry/metric mutation; immutable container bytes are fail-closed |
| PAA member replacement selection | ENGINE-INVARIANT | PASS | shared rebuilder rejects duplicate selection of one member |
| PAA rebuilt payload identity | ENGINE-INVARIANT | PASS | shared rebuilder reopens output and compares every replaced member payload exactly |
| ISO staged PSP_GAME provenance | ENGINE-INVARIANT | PASS | shared authoring helper reopens authored ISO and exact-compares staged PSP_GAME files |
| Desktop release entry point | ENGINE-INVARIANT | PASS | English-first storage validation executes before Korean bank compilation; authenticated slot ownership inputs are passed to `BuildPlan` |
| Mobile ISO release entry point | ENGINE-INVARIANT | PASS | same storage gates execute before compile; bound mobile planner feeds the release path |
| Mobile preflight entry point | ENGINE-INVARIANT | PASS | same English-first validation chain runs without writing an output ISO |
| Android APK embedded project payload | KOREAN-ADDITIONAL | PASS | content-addressed manifest verified during packaging, after APK extraction, and after app-private cache materialization |
| Android launcher routing | KOREAN-ADDITIONAL | PASS | `MainActivity` is the sole launcher and routes to `build-korean-iso`; forensic freeze activity is non-exported and not a launcher |
| Android final user-selected ISO copy | KOREAN-ADDITIONAL | PASS | copied URI is reopened and checked for length + SHA-256 against the already verified temporary ISO before success is reported |
| PPSSPP freeze tracer | FORENSIC-UTILITY | DIFFERENT-BY-DESIGN | diagnostic-only activity/service; it is not a patch-production entry point and is kept outside the production launcher chain |

## Closure result

For the **known upstream English contract surface and the Korean production/build transport surface audited above**, there are currently:

- **MISSING: 0**
- **UNKNOWN: 0**

This is a static/build-contract closure statement, not a claim that runtime freezing is universally solved.

A later runtime freeze remains strong failure evidence. Dynamic substitutions whose concrete values are supplied only at runtime, scene/state-dependent consumer behavior not represented by the known English contract map, or a newly discovered engine consumer can still create a new audit item. Such a finding must reopen this matrix as **UNKNOWN** or **MISSING** until classified and gated.

A successful build or playthrough remains non-reproduction evidence only.

## CI enforcement

Two fail-closed audits now cover this closure:

- `tools/korean/audit-english-first-full-parity.py` — consumer capacities, materialization, compiler, executable, BINDATA, font, archive, ISO, and Android provenance.
- `tools/korean/audit-english-first-entrypoint-closure.py` — desktop/mobile authenticated slot inputs, post-allocation mutation discipline, command routing, mobile build/preflight binding, and sole Android launcher routing.

Both generic CI and the Android Korean release workflow execute the entrypoint closure audit. The rolling freeze-tracer workflow separately validates its diagnostic APK identity so a stale tracer string assertion cannot masquerade as a Korean patch-build failure.
