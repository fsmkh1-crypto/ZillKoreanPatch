# A-019 — EBOOT + renderer-mapping runtime matrix

## Trigger

Independent review of PR #17 identified that the audit needed to preserve PR #14's user-observed runtime matrix. A subsequent evidence-discipline correction is equally important: the H0, A and B PASS observations were each single runs, so none is evidence that the corresponding build was safe.

## Source evidence

PR #14 runtime update, issue comment `5453956902`, records the following observed builds:

| Build | Mapping/font change | Korean EBOOT fixed strings | Observed runtime result | Evidence interpretation |
| --- | --- | --- | --- | --- |
| H0 historical baseline `44a0bd7f...` | H0 mapping; known 0x87 bad slots remain | No added Korean UI set | Reached world map; icon corruption remains | **single-run non-reproduction only** |
| B `4d503f65...` | Same H0 mapping; no Han-safe remap | Yes | Reached tested point; icon corruption remains | **single-run non-reproduction only** |
| A `02402373...` | Full 1,308-rune Han-safe remap | No | Reached tested point; auto-wrap behavior changes materially | **single-run non-reproduction only** plus a real rendering regression |
| Combined runtime commit `4c04e523...` | Full Han-safe remap | Yes | Freeze before ID 210065 subtitle appears; glyphs visually correct | **strong positive failure evidence for that configuration** |
| Stable-minimal `52b9a8aa...` | Preserve H0 mapping except relocate 0x87-lead Korean assignments | Yes | Freeze before ID 210065; glyphs visually correct | **strong positive failure evidence for that configuration** |

Separate H1 explicit-layout experiments changed the line-break layout of ID 210065 and are not part of this matrix.

## Critical interpretation rule

Runtime PASS and runtime FAIL are intentionally asymmetric evidence in this audit.

- A single freeze/crash is strong evidence that the tested configuration can fail.
- A single run without a freeze proves only that the failure was **not reproduced in that run**.
- Therefore H0, A and B must not be called safe, stable, negative controls, or evidence that their respective changes cannot cause the intermittent failure.
- Repetition can strengthen non-reproduction evidence, but even repeated PASS does not prove absence of an intermittent bug.

This rule is especially important here because later project testing established that the freeze is intermittent/non-deterministic enough that one successful playthrough cannot be treated as a clean negative result.

## What is actually established by the matrix

### 1. Combined and stable-minimal are known failing configurations

Each produced an observed freeze. This is strong positive failure evidence for those exact generated/runtime configurations.

### 2. Correcting the visible 0x87 glyph corruption is not sufficient to guarantee no freeze

Stable-minimal corrected the known visible `게`/`깃` icon substitution but still produced a freeze. This is useful because it directly rejects the narrow claim that the visible 0x87 symptom, by itself, accounts for the whole freeze behavior.

### 3. A, B and H0 do not establish safety

Their recorded successful runs are compatible with either:

- a genuinely safer configuration, or
- the same intermittent defect simply not reproducing in that run.

The historical matrix does not discriminate between those possibilities.

## What the matrix does NOT prove

- It does **not** prove that Korean EBOOT strings alone are safe.
- It does **not** prove that full Han-safe remapping alone is safe.
- It does **not** prove that H0 itself is safe.
- It does **not** prove that EBOOT + mapping changes must both be present for the freeze.
- It does **not** prove a causal EBOOT/mapping interaction.
- It does not prove that EBOOT fixed-string execution at runtime directly corrupts state.
- It does not prove that renderer-slot ownership is unrelated to the freeze.
- It does not prove that dynamic message expansion is unrelated to the freeze.

## Build-isolation quality

The historical branches are still valuable for static comparison even though their runtime PASS arms are weak evidence.

Recovered commit history shows:

- H0 → B changes only `release/korean/strings/eboot.toml` (+51 lines).
- H0 → A changes only `cmd/zill/build_korean_mobile_plan.go`.
- The deleted combined branch is recoverable from Actions history: `e84f4f20...` applies the Han-only renderer-slot change on the H0 parent, and child `4c04e523...` then adds exactly the EBOOT UI string set.

Thus the source-level experimental arms were relatively well isolated. That improves their value for generated-asset comparison, but **does not upgrade the single-run PASS observations into causal runtime evidence**.

## Evidence assessment

### CONFIRMED

- The source-level changes for H0/B/A/combined can be reconstructed from commits/Actions history.
- Combined and stable-minimal each produced an observed freeze.
- Stable-minimal corrected the known visible 0x87 glyph corruption and still froze.
- H0, A and B each have only the recorded non-reproducing run in this matrix unless additional runtime repetitions are independently recovered.

### OPEN

- Whether H0, A or B would also freeze under repeated testing.
- Whether failure probability differs materially between arms.
- Whether an EBOOT/mapping interaction exists at all.
- Whether the decisive generated delta is mapping identity, newly required glyphs, PAF geometry, atlas layout, executable fixed-string bytes, message expansion, or another shared state.
- Whether any of these configurations share the mechanism behind the later Bad Execution Address.

## Relationship to the current root-cause map

This matrix is now treated as **cross-cutting configuration evidence**, not as STRONG evidence of a causal interaction.

The active technical branches remain:

1. renderer-slot ownership / hidden glyph consumers;
2. runtime message consumer / dynamic expansion memory contract;
3. build/runtime configuration coupling suggested by the failing combined/stable-minimal configurations — **OPEN mechanism and OPEN interaction**.

Branch 3 is a bookkeeping category for unexplained configuration-sensitive observations. It must not be promoted to a causal interaction unless repeated or static evidence isolates that interaction.

## Highest-value next gates

Before another device A/B:

1. Compare exact generated mapping digests and required-rune sets for H0, B, A, combined and stable-minimal.
2. Compare final PAF/atlas semantic digests and changed glyph records.
3. Compare patched EBOOT bytes/regions between corresponding arms.
4. Determine whether adding EBOOT Korean strings changes the custom-rune requirement set and therefore indirectly changes mappings/font output.
5. Recover any additional historical runtime repetitions, if they exist, and record run counts separately from build counts.
6. Do not choose a new runtime matrix until static/asset-backed deltas have been exhausted.
7. If another runtime matrix is eventually justified, require repeated runs per arm and report the raw run counts (`freeze / total`) rather than PASS/FAIL labels.
