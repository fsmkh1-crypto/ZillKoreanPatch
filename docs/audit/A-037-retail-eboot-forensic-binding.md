# A-037 — Bind runtime forensic output to the retail EBOOT byte sequence

## Trigger

The C5, `$15` substitution, and font-renderer candidate scanners are now invoked automatically from the `build-korean-iso` preflight. Their output records file offsets and virtual addresses, but the log did not identify the exact EBOOT byte sequence that produced those candidates.

## Hypothesis

Recording a cryptographic digest of the EBOOT byte slice used by the preflight before any heuristic scanner output will make candidate reports reproducible and prevent results from different executable inputs from being conflated.

## Verification

- `auditC5RuntimeCandidates()` receives the exact byte slice returned by `loadAuthenticatedRetailEBOOT(gameDir)`.
- Added a SHA-256 digest over that byte slice before running C5, `$15`, or font scans.
- The ISO preflight now emits:
  - `FORENSIC RETAIL_EBOOT_BINDING`
  - SHA-256 digest
  - exact byte length
  - `retail_preflight_input=true`
  - `authentication_contract_unverified_here=true`
- All three scanner families then run against that same in-memory byte slice.

## Evidence-discipline correction

The first version of this note and log emitted `authenticated=true` solely because the loader function is named `loadAuthenticatedRetailEBOOT`. During follow-up review, the exact loader authentication contract could not be independently reproduced from the current PR diff/code-search path. Under this audit's evidence policy, a function name is not sufficient proof of an authentication guarantee.

The log therefore no longer asserts `authenticated=true`. It records only what is directly established here: the exact byte sequence supplied by the retail preflight loader is cryptographically bound to the scanner output. If the loader's revision/hash/size authentication contract is later independently recovered, that can be documented separately and the evidence grade upgraded.

## Result

A future retail-preflight build log can bind every C5, `$15`, and font-renderer candidate to one exact EBOOT input without relying on filenames, remembered provenance, or guessed executable revisions.

This improves evidence provenance only. It does not increase the semantic strength of any heuristic candidate and, by itself, does not prove that the byte sequence is an independently authenticated retail revision.

## Evidence grade

- **CONFIRMED** for deterministic SHA-256/length binding to the exact scanner input once CI passes.
- **OPEN** for the independently reproduced authentication contract of `loadAuthenticatedRetailEBOOT` in this audit note.
- **OPEN** for retail scanner results because the CI environment still does not contain the game asset.

## What this excludes

- It excludes accidentally comparing candidate offsets from different EBOOT byte sequences as though they came from the same executable.
- It does not prove that a candidate is the real C5 handler, substitution dispatcher, or renderer.
- It does not use the loader's function name as proof of executable authenticity.

## New question

Recover and document the loader's actual retail-authentication contract separately. Then, when the retail preflight is run, inspect the candidate set under the bound EBOOT SHA-256 and follow either the `$15` path or PAF-record path to a concrete copy/lookup contract before another gameplay test.
