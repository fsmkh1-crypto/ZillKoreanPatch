# A-067 — English-first font container result boundary

## Scope

Restart audit of the renderer font boundary using the upstream English patch as the primary reference.

Upstream English `release/font/manifest.toml` authenticates both the retail source and the complete transformed result for `font/zillfont.par` and `2d/font/jillbtn.par` with SHA-256 fingerprints. Korean mobile full repack cannot use one frozen result fingerprint because renderer-key allocation is derived from the accepted Korean corpus, but the underlying engine/container contract must still be at least as strict.

## Finding

Before this audit, `VerifyFullRepackSemantics` reparsed the produced PAF and verified renderer keys, BST links, stock metrics/rasters, custom metrics/rasters and glyph count. That was strong semantic verification, but it did not independently prove that bytes outside modeled atlas/PAF mutation fields remained untouched.

This was weaker than the English transform's complete `result_sha256` boundary: an accidental change to a GIM/container header, an inter-page byte, a PAF wrapper/header byte, or another unmodeled PAF field could theoretically survive the semantic checks.

Status before fix: **MISSING** English-equivalent result boundary for dynamic Korean font output.

## Fix

`VerifyFullRepackSemantics` now first enforces an exact byte mutation surface:

- output atlas member size must equal authenticated retail size;
- only the four 512x512 4bpp GIM image payload ranges may differ in `font/zillfont.par`;
- all atlas/container headers and inter-page bytes must remain retail-exact;
- output PAF member size must equal authenticated retail size;
- only modeled per-glyph width/height, x/y, bearing, advance and page fields may differ in `2d/font/jillbtn.par`;
- renderer keys, BST links, reserved tails, PAF wrapper/header and every other byte must remain retail-exact;
- existing semantic verification still reparses the result and proves stock/custom raster and metric behavior.

The English-first parity CI now requires this postcondition to remain wired.

## Classification

- English frozen full-result SHA: **ENGINE-INVARIANT intent / English implementation policy**.
- Korean static result SHA: **DIFFERENT-BY-DESIGN** because the mapping is corpus-derived.
- Exact permitted mutation surface plus semantic reparse: **PASS**, Korean equivalent of the English complete-result boundary.

## Evidence discipline

This removes one class of static font/container corruption. It does not prove runtime freeze elimination and does not promote any prior runtime hypothesis to root cause.
