# A-020 — Runtime repetition and evidence asymmetry policy

## Trigger

Historical debugging repeatedly used labels such as `PASS` and `FAIL` for one-off PPSSPP runs. Once the freeze was recognized as intermittent, a one-run PASS could no longer be treated as a reliable negative control. This matters directly to PR #14's H0/A/B/combined matrix and to earlier ID 10010 / ID 210065 isolation experiments.

## Policy

### Positive failure observation

One observed freeze, crash, Bad Execution Address, malformed renderer output, or other concrete runtime defect is strong positive evidence that the exact tested configuration **can fail**.

It does not, by itself, identify the root cause.

### Negative runtime observation

One run that reaches the target point without the defect is recorded only as:

> `non-reproduction: 1/1 run`

It must not be called:

- safe;
- stable;
- fixed;
- a passing negative control;
- proof that the changed factor is unrelated to the defect.

### Repeated runtime observations

When device testing is eventually justified, report raw counts rather than binary labels:

- `freeze_count / total_runs`
- exact build/commit/artifact identity
- same save/input path and test endpoint where practicable
- whether the test reached the same scene/consumer

Suggested interpretation for this project:

- 1 run: non-reproduction only; negligible negative evidence.
- 3 repeated non-reproductions: modest negative evidence, still insufficient for safety claims.
- 5 repeated non-reproductions: useful comparative evidence against an arm that reproducibly fails, but still not proof of absence.
- any observed freeze/crash: strong positive failure evidence for that configuration.

These thresholds are audit practice, not statistical proof. If the underlying failure probability is low, even five non-reproductions may be weak.

## Consequence for historical evidence

All historical one-run `PASS` labels must be interpreted retroactively as `failure not reproduced in the recorded run` unless additional repetitions are recovered.

This includes, where only one run is documented:

- H0 historical baseline;
- PR #14 B / EBOOT-only arm;
- PR #14 A / Han-safe-only arm;
- one-off ID 10010 bypass/delimiter experiments;
- one-off ID 210065 layout/bypass experiments;
- other single-playthrough diagnostic branches.

By contrast, an observed freeze remains valid evidence that the corresponding configuration can fail.

## Impact on causal matrices

A matrix of:

- A: one non-reproducing run
- B: one non-reproducing run
- A+B: one observed failure

is **not sufficient** to establish an A×B interaction.

It is compatible with an intermittent defect shared by all three arms that happened to reproduce only in A+B. A causal interaction requires stronger support, such as:

- repeated, materially different failure rates across isolated arms; or
- static/asset-backed evidence showing that combining A and B creates a unique invalid state/contract violation.

## Device-test gate

Runtime repetitions remain the final, expensive gate. Before requesting them, exhaust static and asset-backed evidence such as:

- authenticated bank compilation/format checks;
- consumer-storage and dynamic-expansion analysis;
- exact mapping/slot ownership analysis;
- PAF/atlas semantic verification;
- executable overlay verification;
- generated-asset comparison between experiment arms.

The user should not be asked to compensate for missing static proof by repeatedly testing builds that can be discriminated offline first.
