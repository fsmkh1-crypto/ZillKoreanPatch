# A-026 — Retail focus-record CDC context capture

## Trigger

ID 10010 is in the retained C5 consumer set and contains inline `<value:$15>`, while `$15` has no repository-backed global encoded-byte maximum. Upstream classifies `$15` as a caller-supplied movable substitution, so corpus text alone cannot establish its runtime source or bound.

## Hypothesis

The authenticated retail CDC call sites that display message 10010 may expose the actual consumer context needed to narrow `$15` provenance: scene/member, C5 display mode, control evidence, association/addressee state, and other caller-local metadata.

## Verification

Added `cmd/zill/audit_focus_record_context.go` and wired it into `build-korean-iso` immediately after retail ISO extraction. The helper reuses the existing authenticated-retail `zill context --record 10010 --format json --verbose` analysis path, decodes the recovered `cdccontext.Result`, filters to entries whose `MessageID == 10010`, and emits compact `FORENSIC C5_FOCUS` records containing:

- scene ID and physical member
- source archive and CDC offset
- reachability
- display mode / portrait / name-label flags
- consumer evidence
- source controls
- possible addressee candidates

The diagnostic is non-blocking: a context-recovery failure is printed as forensic unavailability rather than treated as proof of a runtime defect.

## Result

The repository now has a deterministic path that will turn the next authenticated retail ISO build into asset-backed evidence about the actual 10010 CDC call context, without requiring another device run first.

No retail asset was present in CI at implementation time, so no claim is made yet about the recovered call site, `$15` source, `$15` maximum length, or destination buffer.

## Evidence grade

- **CONFIRMED** — instrumentation is wired into the authenticated retail ISO build path.
- **OPEN** — actual retail 10010 call-site result until an authenticated retail build executes it.

## What this excludes

Nothing about the runtime root cause is excluded yet. In particular, this does not prove that `$15` expands into the C5 256-byte page buffer, that it overflows, or that message 10010 causes the freeze.

## New question

When the authenticated retail build runs, what caller-local state and consumer evidence surround message 10010, and does that identify a bounded source/storage contract for `$15`?

## Commits

- `eae75a7e4c7550d636c48ed9997e5d78b20f2c80` — add compact retail focus-record context summarizer.
- `a7844a9ce883cb0384851b02d0e46805d1a5b913` — invoke the focus-record audit during authenticated retail ISO builds.
