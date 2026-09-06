# Beta1 finalization — current active handoff for Astra

Date: 2026-09-07 KST
Status: **PAUSED FOR ASTRA HANDOFF / BETA1 INCOMPLETE**
Repository: `fsmkh1-crypto/ZillKoreanPatch`
Branch: `milestone/Beta1`

## FIRST ACTION — DO THIS BEFORE ANY NEW EDIT

The user explicitly asked ChatGPT to stop here so Astra can temporarily take over.

At handoff-writing time, remote `milestone/Beta1` was at queue commit:

`6e17dea0a3f04a66a9a4a376e6774e498631c5c3`

Commit message:

`queue: apply Beta1 contextual copyedit 005`

The corresponding reviewed-copyedit workflow was **still in progress**:

- workflow: `Beta1 reviewed copyedit queue`
- run ID: **34063116001**
- manifest: `docs/audit/beta1-contextual-copyedit-005-reviewed.json`
- intended semantic records: **14**
- intended final bot commit message: `copyedit: repair grammar and residual concatenation`

Therefore Astra MUST begin by:

1. Verify the actual current `milestone/Beta1` HEAD.
2. Check run `34063116001` and determine SUCCESS/FAILURE.
3. If SUCCESS, identify the resulting bot semantic commit and use that as the new baseline.
4. If FAILURE, inspect the failed step/log and repair only that failure before starting more copyedit.
5. Do not reapply batch 005 blindly.
6. Do not reset legitimate newer work.

Do not begin Beta2.

---

## Beta1 definition of done

Beta1 is not merely an APK build. The user's four required goals are:

1. Whole-dialogue Korean line wrapping/reflow stabilization.
2. Korean copyedit: excessive commas, spacing, punctuation and similar defects.
3. Readability using the user-selected **B font profile**.
4. Translation/naturalness/terminology cleanup.

The final APK is accepted only after those four goals are substantially completed and final whole-corpus/release QA passes.

---

## Mandatory English-patch-first rule

For storage, controls, consumer behavior, reflow, layout, font metrics and build/runtime behavior, first inspect how the English patch `HK47196/zill` handles the same area and why.

For names:

1. English spelling establishes entity identity.
2. Japanese source supplies pronunciation/context.
3. Natural Korean is the final canonical form.

Do not invent Korean-only engine behavior without evidence.

---

## Current working cadence

Do **not** stop merely because a section or ordinary copyedit batch ends.

Continue through adjacent safe copyedit work and create/update a checkpoint only when:

- context/token capacity becomes unreliable;
- a new failure type, regression, tool defect or abnormal result appears;
- a user decision is genuinely required;
- a materially new high-risk mechanism has just been proven;
- final Beta1 release/APK checkpoint is reached.

The user explicitly asked for this cadence.

---

## Validation cadence — avoid needless repeated proof

### Ordinary language batch — always keep

- exact reviewed before -> after values;
- Japanese/source immutability;
- control-token/topology preservation;
- persisted layout lexical synchronization and zero layout drift;
- glyph repertoire check;
- Korean contract/data checks;
- integrity / terminology / consistency / text-sanity checks as applicable;
- relevant English-consumer storage contract.

### Change-triggered only

Re-run broad renderer/font/parser/engine proof only if corresponding implementation or inputs changed.

Do not repeatedly run `go test ./...` solely for ordinary language edits after the failure class has already been converted into a cheap targeted gate.

### Checkpoint/release

Use expensive whole-repository / whole-corpus checks at meaningful checkpoints, new risk classes and final Beta1 release.

Never delete a narrow regression test that caught a real failure.

---

## Completed foundational Beta1 work

### Reflow

Current implementation already includes:

- source-aware Korean reflow;
- fixed vs movable control classification;
- expression operands excluded from visible width;
- Japanese/source-guided control boundaries;
- bounded `$28` handling;
- stale/generated layout regeneration behavior;
- static vs runtime-pending separation;
- whole-dialogue static coverage audit.

Known regression anchors include `1980005`, `560650`, `950059`, `280181`.

Do not special-case the historical praying-girl dialogue; whole-population validation is the actual goal.

Latest known broad state before ongoing copyedit:

- accepted Korean IDs: about 42,016;
- verified static dialogue scope: 22,137;
- residual static overflow: 0;
- runtime-unbounded PENDING: 47.

`STATIC PASS != RUNTIME PASS` must remain explicit.

### B font profile — locked

- raster: 10x10
- BearingX: 1
- advance: 12
- gamma: 0.60
- 10 px / 72 DPI / no hinting
- gamma immediately before 4bpp quantization

Do not switch to D / 11x11 unless the user explicitly changes direction.

---

## Source reconciliation — completed

The apparent `43,116 source vs ~42,016 Korean` gap was accounted for. It is not 1,100 missing translations; intentional categories include no-visible-text, runtime/control-only, punctuation/numeral/name-substitution passthrough and technical labels.

Eight Latin-only title IDs remain REVIEW for reachability/localization:

`1940001, 1940005, 1940006, 1940011, 1960506, 1960510, 1960511, 1960516`

Do not auto-translate skipped records.

See `docs/audit/beta1-source-reconciliation.json`.

---

## Completed terminology normalization

User-approved Beta1 forms include:

- `ロストール / Rostorl` -> `로스톨`
- `フェルム / Ferme` -> `페름`
- `レムオン / Lemghon` -> `레무온`
- `アトレイア / Atleia` -> `아트레이아`
- `コーンス / Konsu` -> `콘스`
- `ギア / Gea` -> `기어`
- `石化獣` -> `석화수`
- `ノクサ / Noxa` -> `녹사`
- `パルシェン / Parshen` -> `파르셴`
- `フリント` -> `플린트`
- `小刀` -> `소도`
- `朱雀将軍` -> `주작장군`
- `玄武将軍` -> `현무장군`

Large source-anchored migration checkpoint:

`d0296a92950162d02e5e3f83adaba80cea4f4f57`

Result:

- 1,431 records changed;
- 1,498 terminology replacements;
- canonical terminology table updated;
- known migrated legacy/unrecognized variants reduced to zero;
- layouts synchronized;
- Korean QA passed;
- full Go tests passed for this high-risk migration class.

`拳具` is the game's fist/knuckle weapon category, not handcuffs. English patch uses `fist weapons`. `너클` was recommended but deliberately not folded into the large migration without final category-consistency confirmation. Re-check current repo/user decision before changing it.

---

## Copyedit work completed before this handoff

### Section001 calibration

IDs `10000..10175` were context-reviewed; **29 semantic edits** were accepted and applied.

- 27 persisted layouts synchronized;
- Japanese/control data preserved;
- glyph/Korean QA passed;
- English-consumer storage contract passed;
- calibration included full Go test.

ID `10022` originally exceeded the English character-creation choice fixed buffer: 31 bytes > 30. It was shortened to:

`하루를 마친 뿌듯함`

This fixed-buffer contract is now a targeted safety requirement; do not weaken it.

### Layout-sync bug fixed

`qa-layout-drift.py --fix` could once move newly introduced punctuation to the beginning of the next persisted line. It was repaired and regression-tested. CI `34035092951` passed. Do not re-investigate unless the test or tool changes.

### High-confidence mechanical particle/spacing batch

A refined scanner identified canonical-name particle and obvious spacing problems. **69 findings / 64 records** were applied safely before the later contextual work.

Examples included:

- `로스톨가` -> `로스톨이`
- `로스톨를` -> `로스톨을`
- `로스톨와...` -> `로스톨과...`
- `발로르을` -> `발로르를`
- `발로르이` -> `발로르가`
- `소도으로` -> `소도로`
- `모양 이지만` -> `모양이지만`

Known semantic checkpoint from that phase: `9e2e6480...`.

### Contextual copyedit 002 — completed

Manifest:
`docs/audit/beta1-contextual-copyedit-002-reviewed.json`

Result bot commit:
`8b7b22310b2cf97c5831faca35d29ed1722667e4`

Commit message:
`copyedit: clean contextual spacing and Japanese comma carryover`

This batch used Japanese + pinned English patch + Korean context, and fixed collapsed spacing, Japanese comma carryover and clear translationese/naturalness defects.

### Fresh aligned corpus export after 002

Workflow run:
`34062586889` — SUCCESS

Head exported:
`80ebf4631b9c8ef041a9629b83a6a840d69d2c47`

Artifact:
- name: `beta1-copyedit-corpus`
- artifact ID: `9997936211`
- digest: `sha256:4e06c594b7a1c85f8e92c999ab179085ca025a098624cdc606e8938c9e518621`

This is the aligned JP/EN/KO corpus used to identify batches 003–005. If Astra substantially advances the corpus, refresh the export before relying on exact old Korean values.

### Contextual copyedit 003 — completed

Manifest:
`docs/audit/beta1-contextual-copyedit-003-reviewed.json`

Records: **32**

Workflow run:
`34062754635` — SUCCESS

Result bot commit:
`86a7bf0712cbd2c656337315f0a26a13961e7430`

Commit message:
`copyedit: repair residual collapsed spacing and naturalness`

All steps passed:
- exact before-values;
- apply;
- persisted layout sync;
- zero layout drift;
- changed-file restriction;
- glyph/Korean gates;
- English-consumer storage contract;
- commit/push.

### Contextual copyedit 004 — completed

Manifest:
`docs/audit/beta1-contextual-copyedit-004-reviewed.json`

Records: **10**

Workflow run:
`34062953941` — SUCCESS

Result bot commit:
`310ebbf9f9a00da61b3cdcac39731ec7bae55905`

Commit message:
`copyedit: repair additional collapsed spacing`

All ordinary batch gates passed.

### Contextual copyedit 005 — PENDING AT HANDOFF

Manifest:
`docs/audit/beta1-contextual-copyedit-005-reviewed.json`

Records: **14**

Queue commit:
`6e17dea0a3f04a66a9a4a376e6774e498631c5c3`

Workflow run:
`34063116001`

State when this file was written: **IN PROGRESS / NO CONCLUSION YET**.

Examples include clear errors such as:

- `발로르이라는 남자가` -> `발로르라는 남자가`
- `함께라면분명` -> `함께라면 분명`
- `이건아들이 / 이건딸이` -> `이건 아들이 / 이건 딸이`
- `가면을지키기에` -> `가면을 지키기에`
- `일도그녀가` -> `일도 그녀가`
- `생각으로여기` -> `생각으로 여기`
- `직접상대해` -> `직접 상대해`
- `님을함정에` -> `님을 함정에`

Do not trust that these are committed until run `34063116001` is checked.

---

## Current application mechanism

Preferred multifile copyedit tool:

`tools/korean/apply-multifile-reviewed-copyedit.py`

Preferred queue workflow:

`.github/workflows/beta1-reviewed-copyedit-queue.yml`

Queue control file:

`docs/audit/beta1-copyedit-queue.json`

Current queue revision at handoff: **5**, pointing to batch 005.

The workflow:

1. reads the reviewed manifest;
2. verifies exact before-values;
3. preserves control topology;
4. applies only reviewed Korean values;
5. runs `qa-layout-drift.py --fix`;
6. verifies exact after-values and zero layout drift;
7. requires changed overlays stay within manifest paths;
8. runs glyph/Korean integrity/terminology/consistency/text-sanity gates;
9. runs `TestCurrentKoreanCorpusEnglishConsumerStorageContracts`;
10. commits/pushes only if all gates succeed.

Do not bypass this merely to go faster.

Broad regex candidates are prioritization aids only. Do not blindly auto-fix broad candidate populations.

---

## How to continue contextual proofreading

The dominant remaining work is **actual contextual proofreading of the accepted Korean corpus**, not reflow architecture work.

Continue comparing Japanese + pinned English patch + current Korean for:

- collapsed spacing from removed `<line-break>` boundaries;
- unnecessary/excessive commas;
- awkward punctuation;
- incorrect particles;
- unnatural word order;
- Japanese/English translationese;
- redundant wording;
- awkward sentence endings;
- speaker register/honorific inconsistencies;
- obvious mistranslation, omission or addition;
- newly discovered terminology inconsistency.

Prefer minimal meaning-preserving corrections when the defect is obvious.

For ambiguous semantic changes, inspect the Japanese source, English patch and surrounding speaker/context before editing. Ask the user only when the evidence still does not resolve the choice.

Do not artificially stop at section boundaries.

---

## Important current strategy

Prioritize indisputable defects first:

- words glued together across old line-break boundaries;
- objectively wrong particles;
- accidental Japanese-style comma carryover;
- obvious placeholder/translation-state anomalies;
- clearly unnatural literal constructions where JP+EN agree on meaning.

Leave optional style choices such as permissible auxiliary-verb spacing for later consistency passes unless a global style decision has been established.

After enough obvious defects are cleared, move deeper into sentence-level naturalness/translation review rather than only regex-scanning.

---

## Before final Beta1 acceptance

1. Finish contextual copyedit/translation review to a reasonable whole-corpus completion point.
2. Resolve genuine remaining REVIEW items, including the eight Latin title IDs and any new ambiguous terms.
3. Run final whole-corpus reflow/consumer/glyph/data/font QA.
4. Update `CHANGELOG.md` and `docs/BETA1_RELEASE_NOTES.md` with the actual completed four-goal scope.
5. Run a **new post-copyedit Android Beta1 RC**.
6. Record exact final Beta1 SHA, workflow run ID, artifact and APK SHA-256.
7. Only after Beta1 succeeds, audit/delete obsolete U-series branches while preserving at minimum:
   - `milestone/U0-first-nonfreeze`
   - `milestone/U7`
   - `milestone/Beta1`
   and ensuring no useful unique commits are stranded.

Do not deliver the old pre-copyedit APK as final Beta1.

---

## Astra stop/handoff rule

If Astra approaches its token/context limit, encounters an abnormal failure, or needs a user decision:

1. stop starting new substantive batches;
2. finish the current atomic operation if safely possible, otherwise mark it explicitly partial/in-progress;
3. update **this file** with actual remote HEAD, workflow state, completed batches and exact next action;
4. commit that documentation to `milestone/Beta1`;
5. leave enough information for ChatGPT or another Astra session to resume without reconstructing state from chat history;
6. do not falsely mark Beta1 complete.

The user specifically requested this behavior.
