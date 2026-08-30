# A-063 — English-first restart: build-entrypoint parity

Status: fixed and CI-gated
Branch: `audit/english-first-full-parity-restart`

## Premise

The upstream English patch is the primary source of engine-facing contracts. The Korean path may differ only where the Korean renderer requires a different representation, and all production/preflight entry points must execute the same established storage-contract gates before compiling Korean banks.

## Restart finding

The full restart audit found a real entry-point drift in `PreflightKoreanAlphaISOOnly`.

The real mobile ISO path already executed:

1. Korean C5 layout derivation
2. upstream-English consumer layout derivation
3. C22-specific scanner hardening
4. upstream-English consumer contract validation
5. Korean exact C5 storage validation
6. Korean bank compilation

The authenticated mobile preflight executed only C5 derivation/validation before compiling banks. It therefore could report a successful deterministic preflight without exercising the same upstream-English consumer contracts that the APK's real ISO path enforced.

Classification: `MISSING`.

## Fix

`PreflightKoreanAlphaISOOnly` now executes the same English-first contract chain before compilation:

- `DeriveKoreanC5StorageLayouts`
- `DeriveKoreanEnglishConsumerLayouts`
- `DeriveKoreanC22RetailScannerLayouts`
- `ValidateKoreanEnglishConsumerContracts`
- `validateKoreanRuntimeStorage`
- only then `compileKoreanBanksWithPlan`

The project parity audit now checks all three Korean entry points structurally:

- desktop Korean release
- mobile Korean ISO
- authenticated mobile preflight

A missing/reordered English-first gate fails CI.

## Separate audit-gate correction

The first restart CI failure was not an engine-contract defect. The new source-structure audit accidentally classified the helper parameter `bufferCapacityBytes` as a shared engine constant based only on its suffix. The gate was corrected to derive authoritative capacity names from `internal/layout/rules.go` first, then compare only those constants across English/Korean validators.

Classification: audit-tool false positive, fixed without weakening any engine contract.

## Evidence discipline

This closes one demonstrable validation-path drift. It does not prove runtime freeze freedom. Runtime success remains non-reproduction only; any freeze/crash remains strong failure evidence.
