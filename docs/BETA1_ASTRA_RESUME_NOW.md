# Beta1 Astra resume-now handoff

## Exact stop point: Batch 020 COMPLETE / Batch 021 next

Date: 2026-09-07 UTC. Repository: `fsmkh1-crypto/ZillKoreanPatch`.
Working branch: `milestone/Beta1`. **Beta1 INCOMPLETE. No Beta2.**

Remote HEAD immediately before this documentation checkpoint:
`84c30b504afbb167551466ab3c4d0e59da2a2ed5` (generated scope coverage).
This document's commit is above that SHA: fetch actual remote HEAD before mutation.
Starting baseline: `c9c1f8e0660adbe3a603391ee5ab18ea1a53b6d5`.
Legitimate structural-refresh commit `4634f4fd145838a886d35f090d7dd37fa8fb2c3e`
was preserved by the heavy queue's normal rebase. No reset/force-push.

Read first, in order:

1. This file.
2. `docs/BETA1_REVIEW_PIPELINE_V3_CHECKPOINT.md`.
3. `docs/BETA1_REVIEW_LEDGER_POLICY.md`.
4. `docs/audit/beta1-review-coverage.md`.
5. `docs/audit/beta1-review-distribution.md`.
6. AGENTS, CONTRIBUTING, Korean style/QA and `docs/audit/review/README.md` as required.
   Older handoffs are historical only. Do not resume006 or reapply018/019/020.

## Batch 020 work actually completed

Direct source-order packet read, all states, JP meaning/adjacent context → pinned
EN handling and reason → KO accuracy/naturalness:

- `translations/korean/messages/msgsec010-part99.toml`:100000–100018.
- `translations/korean/messages/msgsec011-part99.toml`:110000–110033.
- `translations/korean/messages/msgsec012-part99.toml`:120000–120005.
- `translations/korean/messages/msgsec013-part99.toml`:130000–130024.

**84 directly read IDs / approximately281 visible state segments / KEEP53 /
EDIT27 / internal marker exclusions4 / contextual80 / propagated0.**
Excluded IDs:100018,110033,120005,130024. These have source section-ending markers
and blank EN; exclusion from review credit does not delete accepted corpus rows.
No scanner-only KEEP was credited.

EDIT IDs:
100002,100003,100004,100005,100006,100009,110001,110002,110004,110006,110011,
110013,110014,110016,110019,110020,110023,110024,110026,120004,130000,130001,
130003,130004,130005,130006,130009.

Overlapping categories: naturalness15, punctuation5, terminology8, grammar6,
translation5. No newly classified spacing-only edit. Significant fixes:

-110011: child is the target of the ordered killing, not the person ordered to obey.
-110004/130005: Japanese 充実 means fulfillment/vigor in context, not Korean 충실.
-120004: puzzlement over Ladras's collapse, not delighted fascination.
-130006: 罰は当たらない is no-harm idiom, not literal punishment.
-100006/110001/110016/110023/130009: predicate/particle/relative-clause repairs.
-Seven directly read Dyneskal IDs normalize 딩갈→딘갈; existing canonical
 ディンガル士官=딘갈 장교 and prior70002 agree. No global propagation.
-130003: 지하도로→지하가도 matches directly read guide/choice labels130010–130013.

Evidence files:

- `docs/audit/beta1-contextual-copyedit-020-reviewed.json`: exact before/after,
  JP/EN, pinned English SHA and individual reasons.
- `docs/audit/review/scope-020.json`: exact ID sets and deterministic packet SHA.
- `docs/audit/review/notes-020.md`: preflight, direct KEEP context notes and safety.
- `docs/audit/review/throughput-020.json`: coverage/workload/timing and CI evidence.
- `docs/audit/beta1-copyedit-queue.json`: revision21, completed020.

## Commits and successful validation

-Queue:`cc03b778d2d471fa4e1e7139851da5352ee2c9ef`.
-Semantic:`1867f6229a4d398bbfecb15327fbf0f8642139b0`.
-Review basis:`d37698270c84a8478e8cd46fff00fab43f8f503d`.
-Scope/notes:`4b91398087a99106debe4a7266fa74f5ca7b3af4`.
-Generated coverage:`84c30b504afbb167551466ab3c4d0e59da2a2ed5`.

Heavy EDIT workflow **34081021753 SUCCESS**:
exact-before, reviewed apply, stale-layout invalidation postcondition, zero drift,
allowed-file restriction, glyph/Korean/font/integrity/terminology/consistency/
text-sanity, pinned-English consumer/storage/effective-layout contract, semantic
push and review-basis registration all passed.

Local exact-before/after, immutable JP, all control and numeric literal sequences,
TOML and absence of semantic line-breaks passed. Remote semantic contents matched
the locally verified changes; scratch copies alone had an extra trailing blank line.

Scope workflow **34081135516 SUCCESS**, first attempt, independently regenerated
packet from Git history and pinned English and rebuilt generated ledger/coverage.
No placeholder hash, failing hash-oracle run, or validation bypass was used.
Packet SHA256:`b200f0d7ede5ed844ea3ab12bb900ef936e4eb1e187685d0474e8161686175e2`.
Locally this was rendered using the unchanged official render() over SHA-pinned
fetched TOML. Do not claim a full local dual-Git checkout; CI supplied that proof.
Structural refresh34081021738 and general queue CI34081021772 also SUCCESS.
No batch validation failure. No engine/font/parser/workflow code was changed.

Measured GitHub job execution windows: heavy44s, scope20s, sum64s.
These are CI windows, not measured local idle time or full production throughput.
Reading/evidence/wall-clock and incident seconds were not separately instrumented
and are null. Never invent time or extrapolate corpus ETA from raw IDs/hour.

## Current generated coverage and evidence limits

Verified at generated commit84c30b50:

- Accepted42016; valid contextual **623 (1.483%)**, up80 from543.
- Legacy full_read176; scope_full_read103; manifest_edit344; propagated0.
- CONTEXT_STALE0; LAYOUT_RECHECK0; pending review-basis batches none.
- UNREVIEWED41393.
- RUNTIME_PENDING full population47 (5 among reviewed ledger rows).

Glyph QA: installed/custom required **1308**, bad characters0/bad records0;
raster catalog covers all1308. No new glyph requirement or font-profile change.
Persisted layout population141; drift0; invalidated0. All84 scope rows lack
persisted layout. Effective-layout storage test passed; ordinary wrapping remains
build-owned. Non-verbose Go output did not expose a fresh numeric whole-corpus
overflow census; do not invent one. Final full-corpus/static overflow proof still
required after final language edit.

Repository/static and CI evidence only. No new authenticated retail-asset,
emulator or real-PSP evidence. Runtime-unbounded47 stays PENDING. No new APK,
release artifact or APK hash. No branches deleted; U0-first-nonfreeze/U7 preserved.
Whole corpus review, source-anomaly dispositions, second-pass KEEP accuracy audit,
final full QA and post-copyedit Android RC are still unfinished.

## Exact next operation: Batch 021

No021 edits or scope assertions exist from this session. Fetch actual remote HEAD,
recent commits, queue and CI again. Do not assume revision21 if concurrent work
advanced it. If unchanged, next queue revision22 / manifest021.

Next file observed in baseline tree after completed013 is
`translations/korean/messages/msgsec014-part99.toml`; re-enumerate the current
tree and read actual files. Then select a coherent workload of roughly200–300
ordinary visible-dialogue equivalents; physical IDs may be fewer for state-heavy
records. Read every selected JP/EN/KO row and all states in source order.
Do not credit a scanner result or an unread duplicate as KEEP.

Pinned English:`HK47196/zill@a98d9ce29f361d666ec23da0dcfd351f24537ffd`.
Japanese is semantic authority. Inspect EN's handling/reason, not EN→KO backtranslation.
Maintain semantic/layout separation and the existing general consumer contract.
Use the same apply tool and one heavy queue per coherent EDIT batch, then the light
scope verifier. Recheck actual remote before each write-sensitive phase.

Useful established commands (at a verified checkout):

```sh
git status --short
git rev-parse HEAD
git log -5 --oneline
git ls-remote origin refs/heads/milestone/Beta1
python3 tools/korean/build-beta1-review-packet.py --english-root ../english --scope docs/audit/review/scope-021.json --verify
python3 tools/korean/apply-multifile-reviewed-copyedit.py --manifest docs/audit/beta1-contextual-copyedit-021-reviewed.json
```

Do not run021 commands before its reviewed manifest/scope exists. Queue owns apply,
layout invalidation, glyph/data/font/storage checks and semantic registration;
scope workflow owns generated ledger/coverage. Never hand-edit those outputs.

Scratch `/workspace/scratch/fb80d7137d44/batch020` is a four-file review snapshot,
not a complete Git checkout. Its semantic edits are already published. No unpublished
semantic work remains. Earlier `Beta1-current`/other scratch repositories are stale;
do not use their HEAD as current or reset them blindly. Reconstruct from remote.

## Locked goals and decisions

Four goals: whole-dialogue reflow; comma/spacing/punctuation proofreading;
B-font readability; translation/naturalness/names/terminology quality.
B font:`10x10 / BearingX1 / advance12 / gamma0.60 / 10px / 72dpi / HintingNone`.
Keep 로스톨,페름,레무온,아트레이아,콘스,기어,석화수,녹사,파르셴,플린트,소도,
주작장군,현무장군. Do not substitute D or undo prior terminology decisions.
Review-policy v3 remains locked, including a different reviewer/model for final
KEEP accuracy sampling. No force-push, no unrelated branches, no Beta2.
