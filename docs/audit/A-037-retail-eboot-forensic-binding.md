# A-037 — Bind runtime forensic output to the authenticated retail EBOOT

## Trigger

The C5, `$15` substitution, and font-renderer candidate scanners are now invoked automatically from the authenticated `build-korean-iso` preflight. Their output records file offsets and virtual addresses, but the log did not identify the exact EBOOT byte sequence that produced those candidates.

## Hypothesis

Recording a cryptographic digest of the authenticated retail EBOOT before any heuristic scanner output will make candidate reports reproducible and prevent results from different retail dumps/build inputs from being conflated.

## Verification

- `auditC5RuntimeCandidates()` already receives the exact byte slice returned by `loadAuthenticatedRetailEBOOT(gameDir)`.
- Added a SHA-256 digest over that byte slice before running C5, `$15`, or font scans.
- The ISO preflight now emits:
  - `FORENSIC RETAIL_EBOOT_BINDING`
  - SHA-256 digest
  - exact byte length
  - `authenticated=true`
- All three scanner families then run against that same in-memory byte slice.

## Result

A future authenticated-retail build log can bind every C5, `$15`, and font-renderer candidate to one exact EBOOT input without relying on filenames, remembered provenance, or guessed executable revisions.

This improves evidence provenance only. It does not increase the semantic strength of any heuristic candidate.

## Evidence grade

- **CONFIRMED** for deterministic digest binding once CI passes.
- **OPEN** for authenticated-retail scanner results because the CI environment still does not contain the retail game asset.

## What this excludes

- It excludes accidentally comparing candidate offsets from different EBOOT byte sequences as though they came from the same executable.
- It does not prove that an authenticated candidate is the real C5 handler, substitution dispatcher, or renderer.

## New question

Once the authenticated retail build is run, what candidate set is produced under the bound EBOOT SHA-256, and can either the `$15` path or PAF-record path be followed to a concrete copy/lookup contract without another gameplay test?
