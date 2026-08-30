# A-062 — English-first full parity restart

Status: ACTIVE
Branch: `audit/english-first-full-parity-restart`
Baseline: `7ef420526bce6d6ec11bc863a6814bf11f3985fb`

## Purpose

Restart the Korean freeze/storage audit from the beginning under the repository-wide `AGENTS.md` premise: an established upstream English engine-facing contract is the primary reference for the Korean path. Previous forensic conclusions remain evidence, but no previous PASS is inherited automatically into this audit.

## Classification

Every row is re-established as one of `PASS`, `DIFFERENT-BY-DESIGN`, `MISSING`, or `UNKNOWN` from current source. `DIFFERENT-BY-DESIGN` requires explicit evidence. Runtime success is not a safety proof.

## Audit order

| # | Contract surface | English reference | Korean path | Restart status |
|---|---|---|---|---|
| 1 | consumer membership and fixed-storage limits | `release/layout/consumer-map.toml`, `internal/layout/rules.go`, `Engine.Validate` | `ValidateKoreanEnglishConsumerContracts` + `ValidateKoreanC5` | PASS (source-structural; guarded by parity script) |
| 2 | semantic/control token traversal | `Projection.SplitSemantic` / shared traversal | `SplitSemanticKorean` | PASS — both use `splitSemanticWith`; Korean differs only in natural-text encoder validation |
| 3 | materialization/control lowering | `Projection.Materialize` | `Projection.MaterializeKorean` | PASS — both use shared `materializeValues`; Korean supplies renderer-slot encoder |
| 4 | bank table format and runtime slot capacity | `CompileBank` | `CompileBankKorean` | PASS — same uint32 offset table and shared `RuntimeBankCapacity(section)` |
| 5 | generated-layout semantic preservation | `preservesSemantics` in English compiler | `PreservesLayoutSemantics` used by Korean derivation | PASS for current derivation call sites; continue end-to-end audit |
| 6 | C22 line/page/whole-record limits | English `c22Violation` | Korean byte-measured mirror | PASS source parity; exact retail build still required |
| 7 | C5 branch/page/page-count limits | English `c5Violation` | Korean `c5ViolationKorean` / `ValidateKoreanC5` | PASS source parity; dynamic expansion remains separate UNKNOWN boundary |
| 8 | fixed labels / guild / trap / character choice / chronicle / C20 | English `Engine.Validate` | Korean English-contract validator | PASS source parity |
| 9 | dynamic substitutions | English consumer-specific bounds | Korean shared bounds + C5 known `$28` bound | PARTIAL: known bounds PASS; unknown runtime maxima remain UNKNOWN, not guessed |
| 10 | bank/archive replacement integrity | English release path | Korean release path | UNKNOWN in restart audit |
| 11 | BOOT/EBOOT/BINDATA reservation and transforms | English release path | Korean custom renderer path | UNKNOWN in restart audit |
| 12 | font/renderer slot allocation | English/retail renderer assumptions | Korean slot planner + font replacement | UNKNOWN in restart audit |
| 13 | atlas/PAF transform semantics | English/static transform | Korean full repack | UNKNOWN in restart audit |
| 14 | mobile APK ISO build parity | desktop/release Korean path as contract carrier | Android/native build entrypoint(s) | UNKNOWN in restart audit |
| 15 | staged PSP_GAME -> authored ISO exact provenance | release authoring | Korean final ISO | UNKNOWN in restart audit |

## Restart findings — phase 1

### Consumer map is shared, not copied

Both English and Korean validators are instantiated from the same layout engine and therefore the same generated `release/layout/consumer-map.toml`. Korean consumer membership is not maintained as a second hand-written list. This is the desired structural parity model.

### Limits are shared constants

The Korean English-contract validator references the same constants in `internal/layout/rules.go` as the English validator. Korean does not own a second set of C22/C20/fixed-buffer numeric limits. Its deliberate difference is byte measurement through the authenticated Korean renderer mapping.

### C5 is intentionally split, not missing

`Engine.Validate` contains C5 in the English path. The Korean English-contract validator omits C5 because exact Korean C5 checking is implemented separately by `ValidateKoreanC5`, using Korean materialization and the same shared C5 constants/consumer membership. This is classified `DIFFERENT-BY-DESIGN`, with the separate Korean validator required in every production Korean build path.

### Materialization has the preferred shared-core shape

`SplitSemanticKorean` delegates to the same `splitSemanticWith` traversal as English. `MaterializeKorean` delegates to the same `materializeValues` lowering as English. The Korean-only code is the renderer-slot encoder/validation. This materially reduces the chance of token/control drift.

## New mechanical gate

`tools/korean/audit-english-first-full-parity.py` is the restart audit gate. It compares English vs Korean consumer references and shared capacity constants, verifies the shared materialization core, verifies bank-capacity/table anchors, and checks that the production release source contains both the English-consumer validator and separate Korean C5 validator. The only current consumer-reference exception is the documented C5 split.

## Next work

Restart phase 2 from the release boundary rather than from record 210065:

1. re-audit every Korean production/mobile build entrypoint and prove both validators are invoked before bank compilation;
2. re-audit archive replacement and final ISO provenance from current source;
3. re-audit BOOT/EBOOT/BINDATA slot reservation against the English/retail path;
4. re-audit font/PAF transforms;
5. only after static parity is exhausted, return to runtime-only unknowns.
