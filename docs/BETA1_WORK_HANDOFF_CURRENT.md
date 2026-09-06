# Beta1 finalization — current active handoff for Astra

Date: 2026-09-06
Status: **ACTIVE / INCOMPLETE**
Repository: `fsmkh1-crypto/ZillKoreanPatch`
Branch: `milestone/Beta1`

## Resume rule

This file is the short current-state handoff. Read it before the older, longer `docs/BETA1_WORK_HANDOFF.md`.

At the time this handoff was written, remote `milestone/Beta1` HEAD was:

`515fec109db5f98b0d1132be721c4ef9ce7675f0`

Commit message:

`qa: refine canonical-name particle candidate scan`

Always verify the actual remote HEAD before continuing. Do not reset legitimate newer work to this SHA.

The user has already authorized continued Beta1 finalization. Do not ask for a new start signal merely because Astra/Work has opened a new context.

Do not begin Beta2.

---

## Beta1 definition of done

Beta1 exists to finish all four user goals:

1. Whole-dialogue Korean line wrapping/reflow stabilization.
2. Actual Korean copyedit: excessive commas, spacing, punctuation and similar writing defects.
3. Korean readability with the user-selected B font profile.
4. Translation/naturalness/terminology cleanup.

An APK build or scanner pass alone does not complete goals 2 or 4.

---

## English-patch-first rule

For storage, controls, consumer behavior, reflow, layout, font metrics and build/runtime behavior, inspect how the English patch (`HK47196/zill`) handles the same area and why before inventing Korean-only behavior.

For names:

1. English spelling identifies the entity.
2. Japanese source supplies pronunciation/context.
3. Natural Korean is the final form.

---

## Important change in working cadence

**Do not stop or create a checkpoint merely because one section or one ordinary copyedit batch has finished.**

Continue through adjacent safe copyedit work without artificial section boundaries.

Create/update a durable checkpoint only when one of these is true:

1. Context/token capacity is becoming unreliable or a system limit warning appears.
2. A new failure type, regression, tool defect or abnormal result appears.
3. A user decision is required because translation/terminology cannot be resolved safely from Japanese + English + context.
4. A materially new validation mechanism or high-risk migration has just been proven and should be preserved.
5. Final Beta1 release/APK checkpoint.

Routine successful section transitions are not handoff triggers.

---

## Validation cadence — do not repeat expensive proof unnecessarily

The repository now uses risk-based validation rather than full-suite repetition after every small language batch.

### Always / ordinary language batch

Keep checks that can actually be affected by the text change:

- exact reviewed before -> after values;
- immutable Japanese preservation;
- control-token/topology preservation;
- existing layout lexical synchronization / zero layout drift;
- glyph repertoire delta;
- Korean contract checks;
- Korean corpus integrity/terminology/consistency/text-sanity checks as applicable;
- relevant English-consumer storage contract.

### Change-triggered only

Re-run broad renderer/font/parser/engine proof only if the corresponding implementation or inputs changed.

Examples:

- full font/raster proof only if renderer/glyph inputs changed materially;
- engine/parser structural investigation only if control/reflow behavior changed;
- full English-patch structural comparison only when entering a new engine-facing area.

### Checkpoint / release

Use expensive whole-repository / whole-corpus checks at meaningful checkpoints, new risk classes and final Beta1 release rather than after every ordinary copyedit batch.

Do not remove a narrow regression test after it has found a real failure. Convert the failure class into a cheap permanent gate and stop repeating the larger investigation that originally discovered it.

---

## Completed technical work inherited from earlier Beta1

### Reflow

Current architecture already includes source-aware Korean reflow with:

- fixed vs movable controls;
- expression operands excluded from visible width;
- Japanese/source-guided control boundaries;
- bounded `$28` handling;
- stale/generated layout regeneration behavior;
- static/runtime-pending separation;
- whole-dialogue coverage audit.

Known regression anchors include `1980005`, `560650`, `950059`, `280181`.

Do not special-case the historical praying-girl dialogue. Whole-population review/reflow is the target; existing regression tests remain safety nets.

Latest known broad static status before the current copyedit work:

- accepted Korean IDs: about 42,016;
- verified static dialogue scope: 22,137;
- residual static overflow: 0;
- runtime-unbounded PENDING: 47.

Keep `STATIC PASS != RUNTIME PASS` explicit.

### B font

Beta1 B profile remains locked:

- 10x10 raster;
- BearingX 1;
- advance 12;
- gamma 0.60;
- 10 px / 72 DPI / no hinting;
- gamma immediately before 4bpp quantization.

Do not switch to D/11x11.

---

## Source reconciliation — completed

The old apparent `43,116 source vs ~42,016 Korean` gap is not 1,100 missing translations.

The reconciliation audit accounted for all 43,116 source IDs through accepted Korean plus intentional categories such as no-visible-text, runtime/control-only, punctuation/numeral/name-substitution passthrough and technical labels.

Eight Latin-only title IDs remain REVIEW for reachability/localization:

- `1940001`
- `1940005`
- `1940006`
- `1940011`
- `1960506`
- `1960510`
- `1960511`
- `1960516`

Do not automatically translate skipped records.

Audit source:

`docs/audit/beta1-source-reconciliation.json`

---

## Actual contextual copyedit already applied

### Section001 calibration batch

Astra initially reviewed IDs `10000..10175` (176 records) and saved 29 proposed changes.

That batch has since been **actually applied and validated**.

Important details:

- 29 semantic edits accepted;
- 27 existing persisted layouts synchronized;
- Japanese/control data preserved;
- glyph/Korean QA passed;
- English-consumer storage contract passed;
- first calibration included full Go testing;
- copyedit application was pushed to Beta1.

The initial wording proposed for ID `10022` exceeded the English-consumer character-creation choice buffer by one byte (31 > 30). It was shortened to:

`하루를 마친 뿌듯함`

This was a useful new failure class, not a reason to weaken the consumer contract.

The fixed-buffer contract must remain a cheap targeted gate for similar records; full `go test ./...` need not be repeated solely to rediscover it in every ordinary batch.

The section001 semantic commit is in Beta1 history (known checkpoint `d47f3adf...`; verify exact ancestry from current HEAD if needed).

---

## Layout synchronization defect found and fixed

During the first real copyedit application, `tools/korean/qa-layout-drift.py --fix` was found capable of moving newly introduced punctuation to the beginning of the next persisted line, e.g. conceptually:

`문장<line-break>.다음`

The tool was repaired and a regression test added.

Korean data CI for that regression fix passed (`34035092951`).

Do not re-investigate this bug unless its regression test fails or the tool is changed again.

---

## User-approved terminology normalization — applied

User decisions used for Beta1 include:

- `ロストール / Rostorl` -> `로스톨`
- `フェルム / Ferme` -> `페름`
- `レムオン / Lemghon` -> `레무온`
- `アトレイア / Atleia` -> `아트레이아`
- `コーンス / Konsu` -> `콘스`
- `ギア / Gea` -> `기어`
- `石化獣` -> `석화수`
- `ノクサ / Noxa` -> `녹사`
- `パルシェン / Parshen` -> `파르셴` (named High Elf, not falchion sword)
- `フリント` -> `플린트`
- `小刀` -> `소도`
- `朱雀将軍` -> `주작장군`
- `玄武将軍` -> `현무장군` (`玄`, not `現`)

The migration was source-anchored rather than blind global replacement.

Known successful migration checkpoint:

`d0296a92950162d02e5e3f83adaba80cea4f4f57`

Migration result:

- 1,431 records changed;
- 1,498 terminology replacements;
- canonical terminology table updated together with corpus;
- known legacy/unrecognized variants reduced to zero for the migrated set;
- layouts synchronized;
- Korean QA passed;
- full Go tests passed for this new high-risk migration class.

Do not repeat the full terminology migration audit on every ordinary copyedit batch unless terminology inputs change.

### `拳具`

`拳具` means the game's fist/knuckle weapon category, not handcuffs. English patch uses `fist weapons`. `너클` was recommended but was deliberately not included in the large automatic terminology migration without a final category-consistency decision. Re-check current repository decision before changing it.

---

## Current candidate-driven copyedit method

A candidate scanner was added to accelerate full-corpus review while preventing blind automatic correction:

`tools/korean/beta1-copyedit-candidate-scan.py`

Workflow:

`.github/workflows/beta1-copyedit-candidate-scan.yml`

The scanner separates:

- high-confidence mechanical candidates;
- broad contextual candidates that require human/source review.

Never auto-fix the broad contextual population merely because a regex matched it.

The first broad scan found thousands of `space_before_common_particle` candidates; this rule is intentionally treated as high-recall/context-review because it has too many false positives.

### Refined scan at current HEAD

Current HEAD at this handoff:

`515fec109db5f98b0d1132be721c4ef9ce7675f0`

Refined scan workflow run:

`34036795838` — SUCCESS

Artifact:

- name: `beta1-copyedit-candidates`
- artifact ID: `9990414618`
- artifact digest: `sha256:5421f16287ac45716c0b2f10d922e7119bea01ea9233b5bfbd902e8ff460627a`

The refined scanner additionally checks Korean final consonant behavior for approved names, so it can catch forms such as:

- `로스톨가` -> `로스톨이`
- `로스톨를` -> `로스톨을`
- `로스톨와...` -> `로스톨과...`
- `발로르을` -> `발로르를`
- `발로르이` -> `발로르가`
- `소도으로` -> `소도로`
- `모양 이지만` -> `모양이지만`

At the latest analysis before this handoff, the refined high-confidence population was approximately:

- 69 findings;
- 64 distinct records.

This figure must be recomputed from artifact `9990414618` before mutation rather than trusted blindly.

The next action is to extract the high-confidence exact rows, review them for source/context safety, and apply only confirmed errors through exact-before manifests. Do not auto-apply the broad contextual candidates.

---

## Reviewed-copyedit application path

Reusable workflow:

`.github/workflows/beta1-reviewed-copyedit-apply.yml`

It applies an exact reviewed manifest, synchronizes existing layout lexical content, verifies exact values, runs per-batch glyph/Korean QA and the English-consumer storage contract, uploads evidence and commits only if all required gates pass.

This is the preferred ordinary-language-batch mechanism when its single-target-file assumption matches the batch.

If a future batch spans many files, adapt the mechanism carefully rather than bypassing its exact-before/control/layout principles.

---

## What remains

The dominant remaining Beta1 work is **actual contextual proofreading of the accepted Korean corpus**.

Do not confuse scanner coverage with contextual review.

Continue reviewing Japanese + English + Korean for:

- unnecessary/excessive commas;
- spacing/orthography;
- awkward punctuation;
- unnatural particles/word order;
- Japanese/English translationese;
- redundant wording;
- awkward sentence endings;
- speaker register/honorific problems;
- obvious mistranslation/omission/addition;
- newly discovered terminology inconsistency.

Use candidate scanners to prioritize obvious errors, but eventually ensure the actual user-visible corpus has been contextually reviewed rather than only regex-scanned.

After semantic changes, keep layout/control/glyph/storage validation aligned with the risk-based cadence above.

Before final Beta1 acceptance:

1. finish actual copyedit/translation review;
2. resolve remaining genuine REVIEW items, including the eight Latin title IDs and any newly discovered ambiguous terms;
3. run final whole-corpus reflow/consumer/glyph/data/font QA;
4. update `CHANGELOG.md` and `docs/BETA1_RELEASE_NOTES.md` to reflect the real final four-goal scope;
5. run a **new post-copyedit Android Beta1 RC**;
6. record final Beta1 SHA, workflow run, artifact and APK SHA-256;
7. only then inspect/delete old U-series branches while preserving `milestone/U0-first-nonfreeze`, `milestone/U7`, and `milestone/Beta1` and ensuring no useful unique work is stranded.

Do not deliver the old pre-copyedit APK as final Beta1.

---

## When to hand off again

If context/token capacity becomes unreliable, a new abnormal failure appears, or a user decision is required:

1. stop starting new substantive work;
2. finish or clearly mark the current atomic operation as partial;
3. update this file with actual current remote HEAD, exact changes, tests/workflows and exact next action;
4. preserve material progress in GitHub;
5. do not falsely mark Beta1 complete.

The next Astra/Work session should start from this file and the actual remote HEAD, not from chat memory alone.
