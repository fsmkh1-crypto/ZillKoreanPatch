# A-022 — Expanded PR #14 EBOOT bundle is not a necessary freeze condition

## Trigger

A-019/A-021 preserved the fact that both historical PR #14 observed-freeze builds (Combined and Stable-minimal) used the same expanded Korean EBOOT fixed-string table. After adopting the asymmetric runtime-evidence rule in A-020, that historical commonality still did not establish causality. A later observed-freeze build provides a stronger discriminant.

## Static reconstruction

### Build `08458f8e9924177e6e80c4950ae0653c4e9b4d39`

The repository at this commit contains only the two H0-era Korean EBOOT fields:

- `0x243d28 = 신께 영혼을 맡긴다`
- `0x243d3c = 영혼을 증명한다`

The expanded 46-field title/character-creation/options/system-menu bundle from PR #14 is absent.

The mobile planner explicitly states that it starts from the H0 plan and relocates only custom mappings whose encoded lead byte is `0x87`; its diagnostic string says `new_eboot_bundle=false`.

Relative to H0 baseline `44a0bd7ffcec83ed47234214ca8639059e86fd92`, Git reports only two changed files:

1. `cmd/zill/build_korean_mobile_plan.go`
   - Minimal87 relocation policy.
2. `internal/message/compile_korean.go`
   - seven bounded character-creation choice shortenings;
   - retained ID 10010 separator diagnostic;
   - authored eight-line C22 layout for ID 210065.

No expanded EBOOT table is part of this delta.

## Runtime observation

The user observed the same opening freeze on this build in the recorded debugging sequence. This is a positive failure observation, not a safety inference from a PASS.

The runtime observation originated in the active debugging conversation rather than a dedicated repeated-run GitHub test record. It therefore establishes that this configuration **can fail**, but does not establish a failure rate.

## Evidence consequence

### CONFIRMED static fact

The expanded 46-field PR #14 EBOOT bundle is absent from `08458f8e...`.

### STRONG runtime implication

Because a freeze was observed with only the two H0-era EBOOT fields, the expanded PR #14 EBOOT bundle is **not a necessary condition** for the freeze phenotype.

This directly weakens the earlier temptation to frame `expanded EBOOT + mapping/font` as an independent leading root-cause branch.

## What this does NOT prove

- It does not prove that EBOOT fixed-string rendering can never contribute to a failure.
- It does not prove that the two H0-era EBOOT fields are irrelevant.
- It does not identify whether Minimal87, the message diagnostics, full font repack, runtime message expansion, or another shared Korean-only condition is causal.
- It does not turn the historical H0 single-run non-reproduction into evidence that either Minimal87 or the compiler diagnostics caused the later failure.

## Updated synthesis

The root-cause map should return to two principal technical branches, with EBOOT behavior treated as a possible submechanism rather than a standalone leading branch:

1. **Renderer/font runtime contract / slot ownership**
   - proven unsafe-slot assumption from 0x87 icon interception;
   - mobile ownership proof remains weaker than desktop;
   - dynamic full-repack runtime coordinate contract is not independently proven.

2. **Runtime message consumer / dynamic expansion memory contract**
   - C5 static accounting omits substitution expansion entirely;
   - 2,548 current C5 records contain inline runtime values;
   - Bad Execution Address remains compatible with state/pointer corruption without identifying the source.

The PR #14 EBOOT/mapping matrix remains useful historical evidence, but after this counterexample it should be classified as **OPEN contextual interaction evidence**, not a third co-equal root-cause branch.

## Next discriminating gates

- Complete the current-corpus PR #14 policy replay to quantify mapping and EBOOT-encoding deltas without relying on historical PASS arms.
- Run the asset-backed exact-byte slot audit and full-font semantic/fingerprint gates.
- Continue reconstruction of actual runtime glyph lookup and `$15`/C5 expansion destinations.
