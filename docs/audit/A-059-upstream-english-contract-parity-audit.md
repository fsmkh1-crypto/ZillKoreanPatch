# A-059 — Upstream English contract parity audit

Date: 2026-08-30

Status: active release-hardening baseline

## Rationale

The Korean freeze investigation has repeatedly found defects where the Korean build path diverged from engine-level contracts already enforced by the upstream English patcher: C5 page capacity, C22 line/page/total capacity, build-owned layout materialization, fixed storage limits, and archive/runtime capacity constraints. These are not English-language rules; they are contracts imposed by the supported PSP engine and its fixed buffers.

Therefore the Korean release path now treats upstream English engine contracts as the baseline. Korean-specific behavior may differ only where the representation genuinely differs (renderer-slot encoding, Korean font/PAF ownership, glyph mapping). Such differences must preserve the same engine contract or add stricter Korean-only checks.

## Classification

Every upstream rule is classified as one of:

- ENGINE_INVARIANT — must hold identically for Korean.
- LANGUAGE_POLICY — presentation/layout policy may differ, but must remain inside the engine invariant.
- KOREAN_EXTENSION — no English analogue; must be checked in addition to upstream invariants.

## Current parity matrix

| Contract | Class | English path | Korean path | Status |
| --- | --- | --- | --- | --- |
| control / projection topology preservation | ENGINE_INVARIANT | `message.Project`, fragment/control validation | `Project` + `MaterializeKorean` | PASS, shared projection |
| generated layout may not alter semantic/control text | ENGINE_INVARIANT | compiler `preservesSemantics` | `message.PreservesLayoutSemantics` | PASS after `76c0962`; local whitespace approximation removed |
| C22 max line 56 bytes | ENGINE_INVARIANT | `layout.Validate` | `ValidateKoreanEnglishConsumerContracts` using renderer-slot bytes | RELEASE-BLOCKING |
| C22 page payload < 256 bytes | ENGINE_INVARIANT | `layout.Validate` | same Korean validator | RELEASE-BLOCKING |
| C22 total payload < 512 bytes | ENGINE_INVARIANT | `layout.Validate` | same Korean validator | RELEASE-BLOCKING |
| C22 max pages 9 | ENGINE_INVARIANT | `layout.Validate` | same Korean validator | RELEASE-BLOCKING |
| C5 branch-local page payload < 256 bytes | ENGINE_INVARIANT | C5 validator | `ValidateKoreanC5` with exact Korean bytes | RELEASE-BLOCKING |
| C5 page-count / single-page membership | ENGINE_INVARIANT | consumer map | same consumer map | RELEASE-BLOCKING |
| bounded label buffer | ENGINE_INVARIANT | fixed-storage validator | Korean English-contract validator | RELEASE-BLOCKING |
| guild client buffer | ENGINE_INVARIANT | fixed-storage validator | Korean English-contract validator | RELEASE-BLOCKING |
| guild region buffer | ENGINE_INVARIANT | fixed-storage validator | Korean English-contract validator | RELEASE-BLOCKING |
| guild posting buffer + bound-role maxima | ENGINE_INVARIANT | posting validator | Korean posting validator | RELEASE-BLOCKING |
| trap buffer including dynamic value allowance | ENGINE_INVARIANT | fixed-storage validator | Korean English-contract validator | RELEASE-BLOCKING |
| character-creation choice buffer | ENGINE_INVARIANT | category validator | Korean English-contract validator | RELEASE-BLOCKING |
| equipment feedback buffer | ENGINE_INVARIANT | category validator | Korean English-contract validator | RELEASE-BLOCKING |
| chronicle payload incl. player-name bound | ENGINE_INVARIANT | category validator | Korean English-contract validator | RELEASE-BLOCKING |
| C20 grouped buffer | ENGINE_INVARIANT | grouped validator | Korean English-contract validator | RELEASE-BLOCKING |
| bank runtime slot capacity / widened offset table | ENGINE_INVARIANT | compiler + `VerifyWideBank` | same compile/release path | RELEASE-BLOCKING |
| BOOT/EBOOT fixed-field capacity | ENGINE_INVARIANT | sparse fixed-field application | Korean overlay rejects encoded replacements larger than source capacity | RELEASE-BLOCKING |
| archive replacement ownership / rebuild | ENGINE_INVARIANT | release archive path | same archive rebuild path | SHARED PATH; asset-backed provenance still worth expanding |
| NUL scanner `z_un_089661DC` 0x100 inline-span contract | CONSUMER-SPECIFIC ENGINE_INVARIANT | not universal | Korean A-054 hardening only where consumer evidence supports it | SCOPED; universal application explicitly rejected |
| line wrapping width / visual flow | LANGUAGE_POLICY | English font metrics and rules | build-owned Korean layout using Korean encoding/metrics constraints | DIFFERENT BY DESIGN |
| renderer-slot mapping / two-byte custom glyph keys | KOREAN_EXTENSION | none | `koreanslots` mapping + round-trip / ownership gates | KOREAN-ONLY |
| PAF / atlas Korean glyph placement | KOREAN_EXTENSION | stock/English path differs | Korean font repack and mapping validation | KOREAN-ONLY |

## Audit rule

Do not introduce a Korean-only compiler/materializer shortcut without first answering:

1. What upstream English engine contract applies at this point?
2. Is the Korean path using the same consumer membership and the same capacity/topology limit?
3. Is the only difference the Korean byte encoder / renderer mapping?
4. Does the final compiled/archive output still prove the contract, rather than only the source intent?

Any UNKNOWN or MISSING answer blocks promotion until either implemented or documented as a deliberate non-applicable rule with evidence.

## Remaining high-value parity work

1. Asset-backed archive provenance: prove rebuilt BINDATA records equal the exact compiled Korean bank bytes for selected/high-risk records and representative sections.
2. BOOT/BINDATA reservation parity: enumerate upstream reserved regions / replacement ownership assumptions and verify Korean release code shares them rather than duplicating them.
3. Consumer-map parity: ensure Korean validators do not maintain shadow membership lists separate from `release/layout/consumer-map.toml`.
4. Dynamic substitution contracts: retain upstream proven bounds where known; unknown substitutions remain runtime-QA rather than guessed static sizes.
5. Font/rendering parity: compare English renderer/font safety gates with Korean PAF/atlas/mapping gates and identify any upstream invariant that is not yet enforced in Korean output.

## Evidence policy

Static parity passing proves the enumerated build/engine contracts for the generated artifact. It does not prove universal absence of runtime freezes. A runtime freeze remains strong rejection evidence and triggers a search for another missing or incorrectly scoped contract.
