# A-019 — EBOOT + renderer-mapping interaction matrix

## Trigger

Independent review of PR #17 identified that the current two-branch root-cause map did not explicitly preserve the user-verified runtime matrix from PR #14. That matrix contains interaction evidence which is not reducible to either renderer-slot ownership alone or C5/dynamic-expansion behavior alone.

## Source evidence

PR #14 runtime update, issue comment `5453956902`, records the following user-verified builds:

| Build | Mapping/font change | Korean EBOOT fixed strings | Runtime result |
| --- | --- | --- | --- |
| H0 historical baseline `44a0bd7f...` | H0 mapping; known 0x87 bad slots remain | No added Korean UI set | PASS to world map; icon corruption remains |
| B `4d503f65...` | Same H0 mapping; no Han-safe remap | Yes | PASS; icon corruption remains |
| A `02402373...` | Full 1,308-rune Han-safe remap | No | PASS, but auto-wrap behavior changes materially |
| Combined | Full Han-safe remap | Yes | FAIL before ID 210065 subtitle appears; glyphs visually correct |
| Stable-minimal `52b9a8aa...` | Preserve H0 mapping except relocate 0x87-lead Korean assignments | Yes | FAIL before ID 210065; glyphs visually correct |

Separate H1 explicit-layout experiments changed the line-break layout of ID 210065 and are not part of this matrix.

## Confirmed implications

### 1. EBOOT Korean UI strings alone are not sufficient for the observed freeze

B passes with the EBOOT Korean UI set while retaining H0 mapping.

### 2. Full Han-safe remapping alone is not sufficient for the observed freeze

A reaches the world map, although its auto-wrap behavior is visibly different and therefore it is not behaviorally equivalent to H0.

### 3. Removing the visible 0x87 icon corruption is not sufficient to remove the freeze

The stable-minimal experiment fixes the known `게`/`깃` visual corruption while retaining most of H0 mapping, yet the EBOOT-combined build still freezes.

### 4. The failure is interaction-sensitive

The observed outcome depends on a combination of changes. The matrix therefore cannot support a simple statement such as “the freeze is caused by global remapping” or “the freeze is caused by Korean EBOOT strings” in isolation.

## What the matrix does NOT prove

- It does not prove that EBOOT fixed-string execution at runtime directly corrupts state.
- It does not prove that the font planner changes only the explicitly intended slots between matrix arms; the exact generated mapping/repack/executable bytes must be compared.
- It does not prove that renderer-slot ownership is unrelated to the freeze; hidden slot/resource interactions may remain after 0x87 relocation.
- It does not prove that dynamic message expansion is unrelated; the failure can surface later from state established by an earlier subsystem.
- A PASS arm is not a proof of safety; it means the failure was not reproduced in that tested run/path.

## Evidence assessment

### CONFIRMED

- The runtime outcome matrix above is the recorded user-observed evidence for the cited experiment builds.
- B and A individually did not reproduce the freeze in those runs.
- Combined and stable-minimal did reproduce the freeze.
- Stable-minimal corrected the known visible 0x87 glyph corruption but did not prevent the freeze.

### STRONG

- A single-factor explanation based solely on “global remapping” is weakened by stable-minimal failure.
- A single-factor explanation based solely on “presence of Korean EBOOT strings” is weakened by B passing.
- Build-time or runtime interaction between EBOOT fixed strings, mapping/font generation, executable bytes, or another shared state deserves an explicit branch in the root-cause map.

### OPEN

- Whether the interaction is caused at build time by different candidate allocation/font repack output.
- Whether the interaction is caused at runtime by executing/rendering one of the Korean EBOOT strings.
- Whether the relevant delta is mapping identity, PAF geometry, glyph metrics, executable string bytes, allocator/layout state, or another coupled effect.
- Whether the same mechanism explains the Bad Execution Address seen in later testing.

## Relationship to current leading branches

This matrix should be treated as a **cross-cutting interaction branch**, not forced into either existing branch:

1. renderer-slot ownership / hidden glyph consumers;
2. runtime message consumer / dynamic expansion memory contract;
3. **EBOOT + mapping/font/executable interaction — STRONG evidence of interaction, OPEN mechanism**.

The third branch may eventually collapse into branch 1, branch 2, or another mechanism after exact generated-asset comparison, but the current evidence does not justify that collapse yet.

## Highest-value next static gates

Before another device A/B:

1. Reconstruct or regenerate each PR #14 arm and compare exact mapping digests.
2. Compare final PAF/atlas semantic digests and changed glyph records between H0, B, A, combined, and stable-minimal.
3. Compare patched EBOOT bytes/regions between B and stable-minimal, and between A and combined.
4. Determine whether adding EBOOT strings changes the set of required custom runes or merely uses mappings already required by runtime messages.
5. Identify whether full/minimal remapping changes any runtime-message bytes unrelated to the visibly corrected 0x87 slots.

Only after those build-time deltas are isolated should a new device experiment be chosen.
