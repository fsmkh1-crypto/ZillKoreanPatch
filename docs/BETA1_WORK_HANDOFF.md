# Beta1 finalization — authoritative work handoff

Date: 2026-09-06 UTC. Status: **INCOMPLETE — audit/review-proposal checkpoint only**.

## Verified checkpoint status addendum

- Checkpoint committed and pushed: **`43e5a09c81b498724357b223bb6f7805d2a011b3`**, `docs: checkpoint Beta1 census and section001 review proposals`.
- Exact active/target HEAD immediately before this status-only documentation commit: **`43e5a09c81b498724357b223bb6f7805d2a011b3`**. Local working tree clean at that commit; remote fast-forward verified. This addendum's containing commit records the status and does not change code/data.
- **CI `34033702959`: SUCCESS**, job `101487828263`. Step evidence verifies English-first audits, `go test ./...`, `go vet ./...`, Python tests, layout drift, `./zill korean-check`, and `./zill korean-font-check` all succeeded on checkpoint `43e5a09c`. `./zill check` also succeeded as a separate general check.
- **Korean data CI `34033702971`: SUCCESS** on the same checkpoint, including corpus integrity/glyph/layout/terminology/consistency/voice gates.
- Local Go dependency availability remains unresolved, but **remote CI gates are no longer PENDING for this checkpoint**. No new Android RC was triggered by these documentation/audit-only changes. Language edits are still unapplied; reviewed 176, proposed 29, applied 0, remaining contextual review 41,840.
- No new APK, semantic layouts, glyphs, or branch deletion. First next action remains the section-001 layout/consumer pre-flight and validated application of the saved proposals; CI may be used for actual Go validation if the next workspace still lacks dependencies.
- Both checkpoint and this status-only commit preserve the ongoing user authorization for Beta1. Do not re-request authorization solely because this is a new session.

## Repository and authority

- Repository: `fsmkh1-crypto/ZillKoreanPatch`
- Active/target branch: `milestone/Beta1`
- Exact audited remote HEAD and local HEAD before this documentation checkpoint: `34fc3e23c9c2350c8d27eb71f002e6fab23e4bdb`.
- This document's containing commit is a checkpoint descendant of that SHA; resolve it with `git log -1 --format=%H -- docs/BETA1_WORK_HANDOFF.md`. Never reset legitimate newer work to the audited SHA.
- Explicit task-local start: **YES**. The user's complete Beta1 handoff authorizes beginning/continuing finalization, recoverable batches, checkpoints and pushes. No further approval is needed for that scope. Beta2 is excluded.
- Protected milestones: `milestone/U7`, `milestone/U0-first-nonfreeze`; do not modify, delete, rewrite, or force-push them.
- U7 currently resolves to `18cfc6da5a4065144aa9c07b0f9ea4d1ae3247bb`; GitHub compare confirms Beta1 is 43 commits ahead, zero behind, with U7 as merge base. Compared 21 changed files. No branches deleted.

Beta1 exists to complete four goals: (1) full dialogue reflow stabilization, (2) actual comma/spacing/punctuation proofreading, (3) selected B-font readability, (4) translation/naturalness/terminology improvement. An APK or regex scan does not complete the language goals.

## Session pre-flight and acquisition

Read: `AGENTS.md`, `CONTRIBUTING.md`, `KOREAN_TRANSLATION_STYLE.md`, `KOREAN_DIALOGUE_QA_PROTOCOL.md`, `CLAUDE_FULL_REVIEW_DECISION_2026-09-06.md`, `CHANGELOG.md`, `BETA1_RELEASE_NOTES.md`. Their old U6 branch/Ferme instructions are superseded where the user explicitly decided otherwise.

Current work directory: `/workspace/scratch/fb80d7137d44/ZillKoreanPatch`.
Existing `/workspace/scratch/304526ffbc6b/ZillKoreanPatch` is on an old branch with 3 local commits and uncommitted work; it was inspected read-only and NOT reused or modified. Do not overwrite it.

Direct GitHub Git transport was unavailable in this environment. Used authenticated GitHub connector reads, reused only matching local Git blobs, fetched 36 different/new files at the exact Beta1 SHA, and checked all 1,136 tracked blobs against GitHub's recursive tree (zero mismatches). Reconstructed the exact unsigned baseline commit and tree `aabe83d3d542e6e44a2ef622425f6228f408c71d`; local checkout is shallow at the audited SHA. No fabricated baseline commit. Push checkpoint through GitHub Git-data connector API with expected parent and `force=false` if ordinary Git transport remains unavailable.

English parity target read directly from `HK47196/zill`, `master`, `internal/message/projection.go`, blob `5a86124c4b5330ee71f9e443cfde921e0b4ef0ae`: fixed nodes retain source controls; pure/caller movable substitutions remain in their source semantic fragment; ordinary line breaks are layout-owned. No engine modification made. Broader upstream storage/font/build comparison remains required before future changes in those areas.

## Verified baseline

- Source: **43,116** distinct IDs.
- Accepted Korean: **42,016** distinct IDs, across 278 sections; 323 physical overlay files contain 42,028 rows because 12 are identical aliases. Do not equate physical files/rows with sections/accepted IDs.
- Custom catalog: **1,308** glyphs; installed stock metric keys: 2,637. Repertoire audit: zero bad characters/records; before and after checkpoint identical. New proposal Hangul outside catalog: zero.
- Renderer and generated catalog agree: width/height 10×10, BearingX 1, advance 12, gamma 0.60 immediately before rounded 4bpp quantization, OpenType 10px/72 DPI/HintingNone. Identity: `opentype-10px-72dpi-hinting-none-origin-0,-2-alpha-gamma-0.60-round-4bpp-v2`. D (11×11/BearingX 0) remains rejected for Beta1.
- Baseline Android RC: `34031107941`, success, SHA `34fc3e23c9c2350c8d27eb71f002e6fab23e4bdb`, job `101480656628`. General CI `34031107998` also success. Prior RC `34030968735` failed; successor fixed the residual-scope assertion. Raster `34030968840` and Korean data `34030968861` succeeded on `2410568d1a887bc9f9fa8654ac5596c51f86e65c`.
- Baseline scanner census: `canonical=42016 checked=42016`; maximum conservative span 452, ID 2400002; 196 candidates at/above 0x100; exact gate CompileBankKorean remains asset-backed.
- Baseline **verified static dialogue** scope: **22,137**, derived layouts 13,337, residual overflow **0**. This denominator is not 42,016.
- Coverage: relevant/eligible 22,137, excluded 0, persisted layouts 113, static overflow 0; **47 runtime-unbounded inline rows remain PENDING**. Exact IDs and log excerpts: `docs/audit/beta1-baseline-ci-evidence.txt`.
- Whole Korean consumer census: 42,016 checked, English layouts 464, visual layouts 141, contracts/visual PASS, **7,572 advisory warnings**. No universal zero-warning or runtime-safety claim.
- Historical regression files inspected include 1980005, 560650, 950059, 280181 paths. Current baseline CI exercises Go tests; fresh local Go execution was blocked during dependency setup.

## Source reconciliation (completed as repository accounting)

Reproduce: `python3 tools/korean/audit-beta1-source-reconciliation.py --json docs/audit/beta1-source-reconciliation.json`.

| Disjoint category | Count |
| --- | ---: |
| Accepted Korean | 42016 |
| No visible text (empty/whitespace source) | 60 |
| Runtime/control-only after removing known controls | 249 |
| Punctuation/numerals/name-substitution punctuation passthrough | 691 |
| Sleep symbol | 1 |
| Synthetic unused marker `<未使用><end>` | 45 |
| Technical resource paths | 18 |
| Technical effect labels | 14 |
| Technical Dummy labels | 14 |
| Latin-only titles requiring REVIEW | 8 |
| Total | 43116 |

All IDs are accounted for through exact inclusive category ranges; every missing-overlay ID has its source/English/file/category in the JSON. No missing Japanese natural-language row was found by this accounting. **Do not claim zero genuinely untranslated user-visible content**: eight Latin titles remain unresolved for reachability/localization.

REVIEW IDs: `1940001`, `1940005`, `1940006`, `1940011`, `1960506`, `1960510`, `1960511`, `1960516`. They include Tug of Street, Lost Memory, Messenger/Messanger and Stranger titles in event-selection/debug-adjacent source. Source spelling must not be corrected. Determine actual surface/reachability before deciding translation versus documented technical exclusion.

Old packet filter, recomputed on current source: no-text 380 + passthrough 777 = 1,157, with **57 already accepted IDs overlapping those filter categories**. Thus missing overlays = 1,157 − 57 = 1,100. This explains the present count difference without asserting unknown historical intermediate counts. Sparse overlay builder preserves omitted records through the existing compiler path; no source record was added/deleted/translated automatically.

## Contextual copyedit progress — NOT APPLIED

Actually read all 176 records in `translations/korean/messages/msgsec001.toml`, IDs **10000–10175 inclusive**, together with Japanese and repository English. This includes initial character-creation questions and suggested player names. Section 000 and all other sections are not contextually reviewed in this session.

Durable per-record decisions and exact before/proposed text: `docs/audit/beta1-copyedit-batch001-proposals.json`.

- Reviewed: 176; KEEP: 147; proposed modifications: 29; **actually modified semantic records: 0**.
- Proposed overlapping categories: punctuation 27, translation 2, naturalness 2. No spacing-only/orthography-only corrections or terminology normalization applied.
- Examples: 10021 水平線 -> 수평선 (currently 지평선); 10035 strike a vital point -> 급소를 맞히는; 10022 more natural fulfillment-at-day's-end wording; sentence-boundary punctuation; three unnecessary dependent-clause commas removed in proposals. Vocative commas retained.
- Proposals are not accepted changes: source/control/consumer validation and effective-layout generation must precede acceptance. No new Hangul glyph needed by proposals.
- Layouts regenerated: **0**; persisted layouts unchanged. Before applying, resolve section-001 consumer/fixed-layout requirements using upstream/current layout engine. Do not simply delete authored special layout without proving regeneration behavior. Do not insert semantic `<line-break>` or manually copy old layout around new wording.
- No record claims are inferred from scanner coverage. Total contextual review this session is exactly 176/42016 (41,840 not yet reviewed), not full corpus.

## Canonical decisions to retain — migration NOT completed

| Japanese | English identity | Final Korean | Status |
| --- | --- | --- | --- |
| ロストール | Rostorl | 로스톨 | User-approved; migration pending; legacy 로스토올 |
| フェルム | Ferme | 페름 | User-approved; obsolete REVIEW superseded; legacy 펠름 |
| レムオン | Lemghon | 레무온 | User-approved; legacy 레몬 |
| アトレイア | Atleia | 아트레이아 | User-approved; audit 아트레아 variants |
| コーンス | Konsu | 콘스 | User-approved race name |
| ギア | Gea | 기어 | User-approved currency |
| 石化獣 | source entity | 석화수 | User-approved |
| ノクサ | Noxa | 녹사 | User-approved |
| フリント | source entity | 플린트 | User-approved |
| 小刀 | source category | 소도 | User-approved; do not globally replace with 단검 |
| 朱雀将軍 | source title | 주작장군 | User-approved |
| 玄武将軍 | Black Tortoise General | 현무장군 | User-approved; 玄 is correct, not 現; source immutable |
| パルシェン | Parshen | 파르셴 | User authorizes if no contradictory evidence; verify High Elf/leader identity, not falchion |
| 拳具 | fist weapons | 너클 | Use if source/equipment-category audit confirms consistency; no criminal handcuff meaning |

No term occurrence census/migration was completed. Existing 65-entry terminology scanner returned zero mismatch against the **old** canonical table; this does not certify new user decisions. Table still has 로스토올/레몬. Keep unknown new terms in REVIEW and continue safe work. Every future term batch must enumerate Japanese/English entity, all Korean surfaces and residual variants, then apply linked table+corpus changes.

## Actual checks

Local (unchanged semantic baseline):

| Command | Result |
| --- | --- |
| `python3 -m unittest discover -s tools/korean -p 'test_*.py'` | PASS, 47 tests |
| `python3 tools/korean/audit-korean-glyph-repertoire.py` | PASS, 1308 custom, 0 bad |
| `python3 tools/korean/qa-layout-drift.py` | PASS, 0 drift |
| `python3 tools/korean/qa-integrity.py --json ../baseline-integrity.json --max-examples 0` | 0 critical; 19 advisory = 12 aliases + 7 length ratios |
| `python3 tools/korean/qa-terminology.py --json ../baseline-terminology.json --max-examples 0` | 0 mismatch against legacy table, 8206 matches |
| `python3 tools/korean/qa-consistency.py --json ../baseline-consistency.json --max-examples 0` | 0 actionable; 14 contextual exception groups / 35 records |
| `python3 tools/korean/qa-consistency-triage.py --json ../baseline-triage.json --max-examples 0` | 2 lexical + 12 short-label groups, advisory |
| `python3 tools/korean/qa-voice.py --json ../baseline-voice.json --max-examples 0` | 0 actionable; 36 reviewed exception findings / 33 records |
| `python3 tools/korean/qa-text-sanity.py --json ../baseline-sanity.json --max-examples 0` | 0 findings; not proof of natural language quality |
| `python3 tools/korean/audit-beta1-source-reconciliation.py --json docs/audit/beta1-source-reconciliation.json` | Exact accounting PASS, 8 REVIEW |
| `PATH=/workspace/scratch/304526ffbc6b/_tools/go/bin:$PATH GOPROXY=off go test ./...` | FAILED SETUP: required modules not cached; dependency-independent packages passed, overall NOT PASS |
| `go vet ./...`, `./zill korean-check`, `./zill korean-font-check` | Not run locally: same unresolved dependencies |

No authenticated retail assets, emulator runtime or physical PSP tests executed. Baseline APK is not final, and its APK SHA256/artifact ID were not independently retrieved here. No new APK delivered. Checkpoint CI status must be independently fetched after pushing; do not substitute old success for current CI.

## Exact checkpoint file set

- `CHANGELOG.md`
- `docs/BETA1_RELEASE_NOTES.md`
- `docs/BETA1_WORK_HANDOFF.md`
- `docs/CLAUDE_FULL_REVIEW_DECISION_2026-09-06.md` (historical supersession note only)
- `docs/KOREAN_TRANSLATION_STYLE.md` (Ferme approval status only)
- `docs/audit/beta1-baseline-ci-evidence.txt`
- `docs/audit/beta1-copyedit-batch001-proposals.json`
- `docs/audit/beta1-source-reconciliation.json`
- `tools/korean/audit-beta1-source-reconciliation.py`

No translation, renderer, generated font/layout or workflow file changed. Local `tools/korean/__pycache__/` is disposable test output, excluded from commits. Scratch `batch001-edits.json` is redundant with the tracked proposal manifest. Preserve only the listed coherent checkpoint; no hidden partial semantic edit exists.

## Why checkpoint now and exact next actions

Long audit output was truncated; the user explicitly requires stopping new substantive batches when truncation/context reliability becomes an issue. No new batch should be inferred from this checkpoint. The local dependency blocker also prevents completing a validated layout/copyedit batch in this environment without arranging CI or a properly provisioned workspace.

1. Read this file first, then independently verify branch/HEAD/tree/recent commits and CI. Explain any newer work before continuing. Check for a newer checkpoint status addendum.
2. Read the mandatory pre-flight documents above, the proposal JSON, reconciliation JSON and baseline CI excerpts. Preserve all user decisions. Confirm the existing user authorization covers continued Beta1 finalization.
3. Restore an environment with Go 1.26.5 and the exact `go.sum` dependencies, or use current CI to execute the actual Go gates. Never weaken gates to sidestep dependencies.
4. Complete the section-001 consumer/layout pre-flight, apply the 29 reviewed proposals if still supported, preserve immutable source/IDs/fixed controls/substitutions, invalidate stale generated layout correctly, and derive effective layouts through the established engine. Run all required Korean gates, full consumer/reflow census, glyph audit, and meaningful checkpoint CI. Mark the batch accepted only on evidence.
5. Review section 000 (not yet read), then proceed section 002 onward in recoverable contextual batches; do not count proposals as applied fixes. Maintain every reviewed ID and per-batch baseline/commit statistics.
6. Enumerate/normalize approved terminology across all applicable surfaces with source evidence; resolve Parshen/拳具 and the eight Latin-title REVIEW rows. Do not let one ambiguous term stop safe batches.
7. Only after all 42,016 accepted records are actually reviewed and language edits complete: full reflow/glyph/data/font QA, updated docs, new fully successful Android Beta1 RC and artifact/hash verification. Keep the 47 runtime PENDING cases separate unless new evidence closes them.
8. Only then audit unique commits in old U branches, preserve U0/U7/Beta1, clean safe superseded branches or document exceptions. STOP at Beta1; no Beta2.

Useful commands in a provisioned checkout:

```sh
git status --short --branch
git rev-parse HEAD
git log -8 --oneline
go test ./...
go vet ./...
./zill korean-check
./zill korean-font-check
go test ./internal/release -run '^TestCurrentKoreanCorpusEnglishConsumerStorageContracts$' -count=1 -v
go test ./internal/release -run '^TestCurrentKoreanCorpusRetailScannerMaxSpanBelowInlineBoundary$' -count=1 -v
python3 tools/korean/audit-korean-glyph-repertoire.py
python3 tools/korean/qa-layout-drift.py
python3 tools/korean/audit-beta1-source-reconciliation.py
```

Use actual current workflow gates, not `./zill check` alone. Source count changes (if genuine untranslated rows are added) must be deliberately reflected in hardcoded 42,016 assertions and evidence; never merely edit counts to force PASS.
