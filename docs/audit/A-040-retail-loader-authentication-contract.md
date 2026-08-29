# A-040 — Retail executable/archive loader authentication contract

## Trigger

A-037 intentionally downgraded its EBOOT provenance claim when the implementation behind `loadAuthenticatedRetailEBOOT()` could not be reproduced through the earlier code-search path. The forensic scanners are only useful as revision-specific evidence if the bytes reaching them are actually pinned to the supported retail assets.

## Hypothesis

The loader names may either represent real cryptographic authentication or merely convention. The implementation must be inspected directly before any `authenticated` evidence label is used.

## Verification

Recovered `cmd/zill/korean_font.go` from the current audit branch and traced the relevant loaders.

The file defines three SHA-256 pins:

- `retailBOOTSHA256 = 5e294dc84a7f0d50719ecd26cb24ffb3792f2d9445803690845a8f1fa1cb85a3`
- `retailEBOOTSHA256 = 2a52012be00c07512dcde932ff6e9eb9b96912c59dd5a25c7c26ef821c124d68`
- `retailBindataSHA256 = 3241fc000f3d52fe8522baaa985fd866e29d64d3a0f23ac4e28b66dee957de3e`

`loadAuthenticatedRetailBOOT()` and `loadAuthenticatedRetailEBOOT()`:

1. read the complete file from `PSP_GAME/SYSDIR`;
2. compute SHA-256 over the complete byte slice;
3. compare the result against the corresponding fixed constant;
4. fail on any mismatch;
5. return only the exact matched bytes.

`loadRetailBindata()` locates `data/bindata.dat` in the supported PAA archive pair, then calls `loadRetailBindataWithSHA()` which likewise computes SHA-256 over the complete extracted payload and rejects any mismatch against the fixed bindata pin.

The same source also structurally validates the retail PAF location/member count/name/size before parsing it, but this note does not upgrade that structural PAF check into a whole-file SHA-256 claim.

## Result

The EBOOT byte slice supplied to C5, `$15`, and font runtime scanners by `auditC5RuntimeCandidates()` is not merely named “authenticated”: it is accepted only after whole-file SHA-256 equality with the repository's supported retail EBOOT pin.

BOOT and bindata have equivalent explicit SHA-256 pinning in the same source file.

A-037's temporary `authentication_contract_unverified_here=true` state is therefore superseded. The forensic binding log now records `authenticated_by_sha256_pin=true` and the expected EBOOT digest.

## Evidence grade

- **CONFIRMED** — BOOT whole-file SHA-256 authentication contract.
- **CONFIRMED** — EBOOT whole-file SHA-256 authentication contract.
- **CONFIRMED** — extracted bindata payload SHA-256 authentication contract.
- **CONFIRMED** — integrated forensic scanners receive the EBOOT bytes returned after that exact pin check.
- **OPEN** — semantic identity of any future heuristic scanner candidate and freeze causality.

## What this excludes

- Treating an arbitrary same-shaped MIPS EBOOT as the supported executable in the integrated retail forensic path.
- Treating an arbitrary BOOT or bindata payload as equivalent merely because parsing succeeds.
- The earlier concern that `loadAuthenticatedRetailEBOOT` might only be a naming convention.

It does not authenticate the user's ISO as one monolithic file by an ISO-level digest, nor does it prove every other extracted asset is SHA-256 pinned. Claims should stay scoped to the verified loader contracts above.

## New question

With EBOOT provenance now confirmed, the next meaningful evidence requires the actual pinned retail executable bytes: run the integrated/standalone scanners and validate surviving candidate dataflow, or move to a narrowly designed runtime/debugger experiment if those asset-backed static paths are exhausted.
