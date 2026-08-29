# A-057 — A-054 release acceptance criterion

Date: 2026-08-30

Status: release-candidate decision rule for `release/a054-inline-span-rc`.

## Decision rule

Treat A-054 (runaway scanner caused by an overlong materialized first-line span crossing the 0x100 inline-region boundary) as the accepted release root cause if the A-054 release-candidate ISO passes the historically reproducing freeze scene three consecutive times without a freeze.

This is a pragmatic release decision rule, not a claim of mathematical proof. Under the project's evidence policy, each successful run is individually only non-reproduction; the three-run streak is used here to decide whether to promote the highest-probability root-cause hypothesis into the release branch.

## If the three-run criterion passes

Immediately perform a full-corpus audit using the exact production materialization path and retail `z_un_089661DC` max-line-span semantics. Any accepted Korean record capable of producing a scanner span at or above 0x100 bytes must be treated as a release blocker and receive a build-owned safe layout or other consumer-specific correction before final release.

The current static gate `TestCurrentKoreanCorpusRetailScannerMaxSpanBelowInlineBoundary` already exercises all materializable Korean entries through the same record-local compiler materializer and fails on any max span >= 0x100. Keep this as a permanent regression gate.

## If any of the three runs freezes

A single freeze is strong failure evidence. Stop the A-054 release promotion immediately and return to the unresolved second-defect / provenance branches; do not average the failed run away against prior non-reproductions.
