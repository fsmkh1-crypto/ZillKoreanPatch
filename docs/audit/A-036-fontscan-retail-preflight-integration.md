# A-036 — Font scanner retail-preflight integration

## Trigger

A-035 strengthened `font-renderer-scan.go`, but the scanner still existed only as a standalone forensic tool. The real authenticated ISO preflight automatically ran C5 and `$15` substitution scans, while font-renderer candidate collection required a separate manual invocation.

## Hypothesis

Moving the linked stride-32 scanner into a shared internal package and invoking it from the same authenticated-retail EBOOT audit path will make font evidence reproducible in every real ISO build without increasing its evidence grade.

## Verification

- Moved the reusable scanner logic to `internal/forensics/fontscan`.
- Moved the synthetic positive/negative tests with it.
- Updated `tools/forensics/font-renderer-scan.go` to call the shared package instead of maintaining a second implementation.
- Added `fontscan.Scan(retailEBOOT, 12)` to `auditC5RuntimeCandidates()`.
- The real ISO preflight now emits:
  - `FORENSIC FONT_RUNTIME_SCAN`
  - ranked `FORENSIC FONT_RUNTIME_CANDIDATE` records
  - compact `FORENSIC FONT_RUNTIME_WINDOW` disassembly for the strongest candidates.
- Zero candidates remain non-fatal and explicitly do not disprove a renderer that computes PAF record addresses differently.

## Result

Once an authenticated retail EBOOT reaches the real ISO build path, C5, `$15` substitution and linked PAF-stride font candidates are captured in the same reproducible build log.

This closes an observability/integration gap only. It does not establish that any returned candidate is the retail renderer, does not prove BST traversal, and does not prove that all glyph consumers obey patched Page/X/Y metadata.

## Evidence grade

- **CONFIRMED** for shared scanner integration and synthetic contract once CI passes.
- **OPEN** for authenticated-retail candidate results and renderer semantics.

## What this excludes

- It excludes future analysis accidentally running a stale standalone font-scanner implementation different from the ISO preflight implementation.
- It does not exclude renderers using multiply instructions, pointer tables, precomputed record pointers, alternate stride construction, or direct atlas references.

## New question

When the authenticated retail EBOOT is scanned, can a surviving candidate be connected from glyph key lookup through the PAF record fields to Page/X/Y consumers and then to atlas sampling without relying on remembered addresses?
