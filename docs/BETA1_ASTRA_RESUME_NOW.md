# Beta1 Astra resume-now handoff

Date: 2026-09-07 KST
Status: **PAUSED FOR ASTRA TAKEOVER / BETA1 INCOMPLETE**
Repository: `fsmkh1-crypto/ZillKoreanPatch`
Branch: `milestone/Beta1`

## Authoritative current snapshot

At handoff completion, remote `milestone/Beta1` HEAD is:

`a372f89f07d6825b413124aedaabfb4c14ee0827`

Commit message:

`docs: refresh Astra handoff at contextual copyedit 005`

Its parent is the successful batch-005 semantic commit:

`452e8acf323583866f6ccf9353498d0e55fdfeb2`

Commit message:

`copyedit: repair grammar and residual concatenation`

Batch-005 workflow:

- workflow: `Beta1 reviewed copyedit queue`
- run ID: `34063116001`
- conclusion: **SUCCESS**
- manifest: `docs/audit/beta1-contextual-copyedit-005-reviewed.json`
- records: 14

Do not reapply batch 005.

Always verify actual remote HEAD again before mutation because later automation may have advanced it legitimately.

## What to read first

1. `docs/BETA1_ASTRA_RESUME_NOW.md` — this file, latest exact stop point.
2. `docs/BETA1_WORK_HANDOFF_CURRENT.md` — detailed method/history; note its batch-005 section was written while 005 was still running, and this file supersedes that one point: **005 succeeded**.
3. `docs/BETA1_WORK_HANDOFF.md` — older long-form background only if more history is needed.

## User's Beta1 goals

Beta1 must complete all four, not merely build an APK:

1. Whole-dialogue Korean reflow/line-break stabilization.
2. Copyedit including excessive commas, spacing and punctuation.
3. Readability using locked B font profile.
4. Translation/naturalness/terminology cleanup.

Do not begin Beta2.

## English-patch-first rule

For storage, controls, consumer behavior, reflow, layout, font metrics and runtime/build behavior, inspect `HK47196/zill` first and understand how/why the English patch handles the same area before inventing Korean-specific behavior.

For names: English identifies entity -> Japanese supplies pronunciation/context -> natural Korean is final.

## Current completed contextual-copyedit checkpoints

- Section001 calibration: 29 edits accepted/applied; 27 persisted layouts synchronized; fixed-buffer regression discovered at ID 10022 and solved with `하루를 마친 뿌듯함`.
- Terminology migration: 1,431 records / 1,498 replacements, source-anchored, checkpoint `d0296a92950162d02e5e3f83adaba80cea4f4f57`.
- High-confidence particle/spacing cleanup: 69 findings / 64 records applied.
- Contextual batch 002: bot commit `8b7b22310b2cf97c5831faca35d29ed1722667e4`.
- Contextual batch 003: 32 records, run `34062754635` SUCCESS, bot commit `86a7bf0712cbd2c656337315f0a26a13961e7430`.
- Contextual batch 004: 10 records, run `34062953941` SUCCESS, bot commit `310ebbf9f9a00da61b3cdcac39731ec7bae55905`.
- Contextual batch 005: 14 records, run `34063116001` SUCCESS, bot commit `452e8acf323583866f6ccf9353498d0e55fdfeb2`.

All 003/004/005 ordinary gates passed: exact before-values, apply, persisted layout synchronization, zero layout drift, changed-file restriction, Korean/glyph/data gates, English-consumer storage contract, commit/push.

## Fresh aligned corpus evidence used for current phase

Post-002 aligned JP/EN/KO export:

- workflow run: `34062586889` SUCCESS
- exported head: `80ebf4631b9c8ef041a9629b83a6a840d69d2c47`
- artifact ID: `9997936211`
- artifact name: `beta1-copyedit-corpus`
- digest: `sha256:4e06c594b7a1c85f8e92c999ab179085ca025a098624cdc606e8938c9e518621`

Because 003-005 have since changed the corpus, use exact-current repository values for new manifests. Refresh the aligned corpus export when enough new edits accumulate or stale exact values become inconvenient; do not refresh after every tiny batch merely for ritual.

## Preferred ordinary copyedit mechanism

Tool:
`tools/korean/apply-multifile-reviewed-copyedit.py`

Queue workflow:
`.github/workflows/beta1-reviewed-copyedit-queue.yml`

Queue control:
`docs/audit/beta1-copyedit-queue.json`

Current queue revision: 5, pointing at already-completed batch 005. For the next batch, create a new reviewed manifest and increment/update the queue once.

The workflow already enforces:

- exact before-value matching;
- control topology preservation;
- semantic-only reviewed changes;
- persisted layout lexical sync;
- zero layout drift;
- manifest path restriction;
- glyph/Korean integrity/terminology/consistency/text-sanity checks;
- English-consumer storage contract;
- commit only on success.

Do not bypass it for speed.

## Validation cadence

Do not repeat expensive full proof after every ordinary language batch.

Always for ordinary batches: exact before/after, controls, layout sync/drift, glyph/data/Korean gates, relevant consumer storage.

Only when related inputs/code changed: broad renderer/font/parser/engine proof.

At meaningful checkpoints/new risk classes/final release: whole repository/corpus tests.

Keep narrow regression tests that found real failures.

## Exact next work

Continue contextual proofreading from the current repository state, without artificial section boundaries.

Prioritize indisputable defects first:

- words glued together where Japanese `<line-break>` removal collapsed spacing;
- objectively wrong Korean particles;
- accidental Japanese-style comma carryover;
- obvious translation-state placeholders/anomalies;
- clear literal/translationese constructions where Japanese and English agree on meaning.

Then deepen into actual sentence-level naturalness/translation review:

- unnecessary/excessive commas;
- awkward punctuation;
- word order;
- redundant wording;
- unnatural endings;
- speaker register/honorific consistency;
- mistranslation/omission/addition;
- terminology inconsistencies.

Use Japanese + pinned English + surrounding context before semantic rewriting. Prefer minimal meaning-preserving edits when the defect is obvious. Ask the user only if the evidence still leaves a genuine choice.

Do not stop merely because one section or ordinary batch is complete.

## Locked technical facts

B font remains locked: 10x10, BearingX 1, advance 12, gamma 0.60, 10px/72dpi/no hinting. Do not switch to D/11x11.

Reflow architecture already covers source-aware controls, expression operands, Japanese-guided control boundaries, bounded `$28`, stale/generated layouts, static/runtime-pending split and whole-dialogue audit. Historical praying-girl records remain regression anchors, not a special manual target.

Known broad status before ongoing copyedit: verified static dialogue 22,137; residual static overflow 0; runtime-unbounded PENDING 47. `STATIC PASS != RUNTIME PASS`.

## Final Beta1 gate later

After contextual translation/copyedit reaches a reasonable whole-corpus completion point:

1. resolve genuine REVIEW items;
2. final whole-corpus reflow/consumer/glyph/data/font QA;
3. update CHANGELOG and Beta1 release notes;
4. run a new post-copyedit Android Beta1 RC;
5. record final SHA/run/artifact/APK SHA-256;
6. only then audit/delete obsolete U-series branches, preserving at least U0-first-nonfreeze, U7 and Beta1 and ensuring no useful unique commits are stranded.

Do not ship the old pre-copyedit APK as final Beta1.

## Astra context-limit rule

If token/context capacity becomes unreliable, or an abnormal failure/user decision appears:

1. stop starting new substantive work;
2. finish current atomic operation if safe, otherwise mark it partial;
3. update this resume-now file or the current handoff with actual HEAD, workflow status, completed changes and exact next action;
4. commit the documentation to `milestone/Beta1`;
5. leave enough evidence for ChatGPT/Astra to resume without relying on chat memory;
6. never falsely mark Beta1 complete.
