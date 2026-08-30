# A-064 — English-first font source parity

Status: implemented on `audit/english-first-full-parity-restart`.

## English contract

`release/font/manifest.toml` authenticates both retail font members before applying the English static transform:

- `pa[13611] font/zillfont.par` source SHA-256 `0d3d6d2648870e87a01636cdfc7cc7af8100ea40b71e5ed05f82ac197606584a`
- `pa[13612] 2d/font/jillbtn.par` source SHA-256 `95b48379092db4db72f890d5a221ba8c4094dd438cb4c4eba98eb5520c7b17aa`

This is an engine-asset identity contract, not an English-language policy.

## Restart finding

The Korean font paths checked member index/name/size and PAF structure, but did not independently pin the atlas payload to the exact source SHA used by the English manifest. A same-sized unexpected atlas could therefore pass the Korean precondition farther than the English path allows.

Classification: `MISSING`.

## Fix

Both desktop and mobile Korean font transforms now call `verifyKoreanFontRetailSources` before any Korean raster/PAF transform. The helper pins the exact two source SHA-256 values already used by the English manifest.

The English-first CI audit also reads the English font manifest and fails if those source fingerprints are no longer pinned by the Korean font path.

## Scope

This closes a static provenance/parity gap. It is not evidence that font handling was the runtime freeze root cause, and it does not turn a successful runtime test into a safety proof.
