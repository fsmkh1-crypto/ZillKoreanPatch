# Beta 1 Release Notes

Date: 2026-09-06  
Release branch: `milestone/Beta1`  
Preserved predecessor: `milestone/U7`  
Earliest preserved U baseline: `milestone/U0-first-nonfreeze`

## 1. Purpose of Beta 1

Beta 1 is the first named beta baseline after the U-series development milestones. It consolidates the current Korean dialogue reflow-parity work and freezes the selected B font profile as the release font baseline.

The release is intentionally documented as an implementation baseline, not as a claim that every dialogue path has completed full PSP hardware runtime QA.

## 2. Project authority: English patch first

For engine, consumer, layout, storage, control, and build behavior, the English patch is the first reference because it already demonstrates how the retail game path was made to consume modified text safely.

The required reasoning order is:

1. identify the English consumer/path and its contract;
2. determine why the English implementation separates or preserves specific controls/substitutions;
3. make Korean follow that contract where the consumer is shared;
4. add Korean-only behavior only where the Korean renderer or Korean data has separately proven requirements;
5. validate the final Korean output without weakening the English-derived safety contract.

English surface wording is not a translation authority. For proper names, English spelling is identity evidence, Japanese source text is pronunciation evidence, and natural Korean is the final localization target.

## 3. B font profile adopted for Beta 1

### Final profile

| Property | Beta 1 value | Implementation role |
| --- | ---: | --- |
| Source raster | 10x10 | Existing deterministic Korean raster geometry |
| BearingX | 1 | Existing proven runtime placement metric |
| Advance | 12 | Existing proven Korean renderer advance |
| Alpha gamma | 0.60 | Selected B-profile visual-weight adjustment |
| Quantization | 4bpp, rounded | Existing PSP font-atlas format |

### How it is implemented

`internal/koreanfont/render.go` keeps the existing 10px OpenType render path, 10x10 canvas, origin, DPI, and no-hinting rule. Immediately before conversion to 4bpp, normalized alpha is transformed as:

```text
a = alpha / 255
corrected = pow(a, 0.60)
value4bpp = round(corrected * 15)
```

The deterministic render-rule identifier is:

```text
opentype-10px-72dpi-hinting-none-origin-0,-2-alpha-gamma-0.60-round-4bpp-v2
```

`internal/zillfont/korean_metrics.go` retains the established 10x10 / BearingX 1 / Advance 12 geometry and placement contract.

### Why this implementation was chosen

The user compared the font candidates visually and selected B because it appeared cleaner. The adjustment is deliberately restricted to alpha intensity before 4bpp quantization. It does not change raster dimensions, placement, or advance, so choosing the cleaner visual weight does not introduce an unrelated reflow-metric change.

### Reproducibility

The Korean raster workflow regenerates the repository-owned glyph catalog from the renderer rather than relying on hand-edited raster data. After the Beta 1 gamma change, the `Korean raster catalog` workflow completed successfully and GitHub Actions committed the regenerated catalog.

A unit test locks the gamma value, render-rule identifier, and representative quantization results to prevent accidental drift.

## 4. Dialogue reflow parity consolidated into Beta 1

### Problem class

Earlier Korean reflow logic could disagree with the actual English consumer projection in records containing runtime value substitutions. The important distinction is not simply “control tag versus text.” It is whether the consumer treats a node as fixed control flow or as a movable substitution embedded in a semantic fragment.

### English behavior used as authority

The English projection keeps movable substitutions inside the semantic text fragment while fixed control-flow nodes delimit cases/fragments. This distinction is necessary because splitting at every value tag changes the fragment presented to the reflow consumer and can produce layout decisions that do not correspond to the actual runtime branch.

### Korean implementation direction carried by Beta 1

The Beta 1 line therefore keeps the following rules:

- fixed control-flow nodes delimit reflow fragments according to the English consumer contract;
- movable runtime substitutions remain inside their semantic fragment when the English projection does so;
- bounded substitutions such as the player-name path may participate in reflow using the proven renderer reservation rather than being treated as automatically unknowable;
- source-aware fragment handling is used instead of flattening the entire conditional record into one artificial string;
- final post-reflow width is still checked and fails closed when a fragment cannot be proven to fit;
- a bounded substitution being eligible for analysis does not mean every sentence containing it is automatically safe;
- regression anchors preserve previously reproduced defect classes, including the player-name/C5 cases investigated during this work;
- final-layout auditing is kept distinct from derivation-eligibility auditing so “not selected for reflow” cannot be mistaken for “visually proven safe.”

### Why this is safer than one-off manual wrapping

A manual line break can hide one observed overflow while bypassing the shared consumer contract. Following the English projection rules fixes the class of error at the point where fragments are defined and then lets width checks decide whether each resulting Korean fragment is safe.

## 5. U7 relationship and Beta 1 lineage

The temporary `fix/u6-korean-reflow-parity` line and `milestone/U7` diverged during parallel work. U7 contained three later commits not present as ancestors of the temporary branch.

Before Beta 1 finalization those U7-only changes were inspected rather than blindly replayed:

- U7's fixed-control `$20` C5 behavior is already present in the working line in a broader English-parity implementation;
- U7's record 280181 regression coverage is already present in the working line in expanded form;
- U7's Android release-candidate branch trigger is superseded by retargeting the workflow to `milestone/Beta1`.

Beta 1 should therefore preserve `milestone/U7` as an explicit parent/ancestor while using the final working tree, avoiding both loss of U7 history and rollback of the newer reflow implementation.

## 6. Release workflow

`.github/workflows/android-korean-a054-rc.yml` is retargeted from the temporary reflow branch to `milestone/Beta1`.

The workflow continues to perform English-first parity audits, Go tests, Korean full-corpus scanner/consumer checks, Android unit tests, APK construction, embedded payload verification, and artifact publication when triggered on the Beta 1 branch.

The temporary working branch is not intended to remain as a release branch after Beta 1 is established.

## 7. Validation vocabulary for Beta 1

The project must state exactly what evidence exists.

### Completed during Beta 1 preparation

- deterministic B-profile renderer implementation;
- quantization regression test for gamma 0.60;
- generated raster catalog regeneration by GitHub Actions;
- successful Korean raster catalog workflow for the B renderer change;
- existing repository reflow and consumer regression coverage carried by the working line.

### What those checks prove

They prove that the checked repository generation/test paths agree with the recorded Beta 1 font and reflow contracts.

### What they do not prove

They do not, on their own, prove every retail-asset branch or every PSP/emulator visual scene at runtime. Asset-backed checks, emulator samples, actual PSP hardware samples, and full runtime QA must be reported separately when performed.

## 8. Preserved branch policy after Beta 1

After Beta 1 is created and verified, the intended retained milestone set is:

- `milestone/U0-first-nonfreeze` — earliest preserved U baseline;
- `milestone/U7` — immediate pre-Beta preserved baseline;
- `milestone/Beta1` — current beta baseline.

Other U-series development branches are historical intermediates and are intended for deletion once Beta 1 is safely established.

## 9. Required change record for future work

For every material change after Beta 1, add an entry to `CHANGELOG.md` and, for release-level or technically nontrivial changes, a detailed note containing:

```text
Date:
Affected release/branch:
What changed:
How it was implemented:
Why it was changed:
English-patch/consumer reference:
Korean-only divergence, if any, and evidence:
Validation actually completed:
Runtime/asset evidence actually completed:
Known limits or pending checks:
Relevant commits/files/tests:
```

Do not omit the reason or validation boundary merely because the code change is small.
