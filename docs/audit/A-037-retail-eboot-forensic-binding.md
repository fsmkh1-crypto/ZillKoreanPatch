# A-037 — Bind runtime forensic output to the retail EBOOT byte sequence

## Trigger

The C5, `$15` substitution, and font-renderer candidate scanners are invoked automatically from the `build-korean-iso` preflight. Their output records file offsets and virtual addresses, so the exact executable byte sequence producing those candidates must be preserved and authenticated.

## Hypothesis

Recording a cryptographic digest of the EBOOT byte slice used by the preflight, and tying it to the loader's independently reproduced fingerprint contract, makes candidate reports reproducible and prevents results from different executable inputs from being conflated.

## Verification

- `auditC5RuntimeCandidates()` receives the exact byte slice returned by `loadAuthenticatedRetailEBOOT(gameDir)`.
- `cmd/zill/korean_font.go` defines the supported retail EBOOT fingerprint as:
  - SHA-256 `2a52012be00c07512dcde932ff6e9eb9b96912c59dd5a25c7c26ef821c124d68`.
- `loadAuthenticatedRetailEBOOT()` reads `SYSDIR/EBOOT.BIN`, computes SHA-256 over the complete file, compares it to that exact constant, and fails with `unsupported retail EBOOT.BIN fingerprint` on any mismatch.
- Only after that check returns successfully does the forensic preflight hash the same in-memory byte slice and run C5, `$15`, and font scanners against it.
- The preflight now emits:
  - `FORENSIC RETAIL_EBOOT_BINDING`
  - observed SHA-256
  - exact byte length
  - `retail_preflight_input=true`
  - `authenticated_by_sha256_pin=true`
  - the expected SHA-256 pin.

## Evidence-discipline history

The first version of this note briefly emitted `authenticated=true` solely because the loader function was named `loadAuthenticatedRetailEBOOT`. That wording was correctly withdrawn because a function name is not evidence.

A later source-tree audit recovered the actual implementation in `cmd/zill/korean_font.go`. The authentication claim is now restored on a different basis: direct source inspection proves whole-file SHA-256 pinning to one supported EBOOT byte sequence. This correction is retained explicitly rather than erasing the earlier evidence downgrade.

## Result

Every future C5, `$15`, and font-renderer candidate produced by the integrated ISO preflight is bound to the exact EBOOT byte sequence accepted by the revision-specific SHA-256 loader contract.

This establishes executable input identity. It does **not** increase the semantic strength of any scanner candidate and does not identify the C5 handler, substitution dispatcher, glyph renderer, `$15` source, or freeze mechanism.

## Evidence grade

- **CONFIRMED** — whole-file retail EBOOT SHA-256 authentication contract.
- **CONFIRMED** — deterministic SHA-256/length binding of the scanner input.
- **OPEN** — actual scanner results because CI does not contain the retail game asset.
- **OPEN** — runtime semantics and freeze causality of any future candidate.

## What this excludes

- Scanner output from an unsupported/different EBOOT silently being treated as the supported revision.
- Comparing candidate offsets from different EBOOT byte sequences as though they came from one executable.
- Using the loader's function name alone as proof of authenticity.

It does not exclude false-positive/false-negative heuristic scanner behavior or a runtime mechanism outside the scanners' modeled instruction patterns.

## New question

Run the now SHA-256-authenticated retail EBOOT through the integrated or standalone forensic scanners, then follow surviving `$15` or PAF-record candidates to concrete source/destination or lookup contracts before selecting another gameplay experiment.
