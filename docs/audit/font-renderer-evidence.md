# Korean font / renderer audit evidence

This document records what the Korean patch has actually demonstrated, and separates it from assumptions that still require runtime reverse engineering.

## Confirmed from authenticated retail resources

- `2d/font/jillbtn.par` contains a PAF table with 2,637 records.
- The authenticated retail PAF size is `0x149d0`.
- Records begin at `0x30` and use a `0x20`-byte stride.
- The header reports BST root index 1318.
- Renderer keys are strictly ascending in the authenticated retail table.
- Each complete record contains key, width/height, atlas x/y, bearing x/y, advance, left/right links, page, and a zero reserved tail.

The earliest parser temporarily modeled a truncated final record; that interpretation was corrected after checking the authenticated retail member. Historical hypotheses must therefore not be treated as equivalent to current authenticated facts.

## Confirmed by the earliest Korean font PoC

The one-glyph Korean smoke test replaced only atlas bytes for the retail renderer key used by CP932 `の` (`82 CC` on disk / `GlyphKey 0xCC82` in the little-endian PAF record representation). It deliberately left PAF key, metrics, tree links, record count, and all other font-member bytes unchanged.

That experiment demonstrated that a message encoded with an installed retail renderer key can display a different bitmap when only the atlas cell for that key is replaced. It is evidence that the relevant observed rendering path resolves the key to that retail glyph cell. It does **not** independently prove the implementation of the runtime lookup routine or that all keys/consumers follow the same path.

## Confirmed by current Korean full-repack code

- Full repack preserves glyph index, renderer key, and left/right links.
- Stock glyph width/height, bearings, advance and raster are intended to remain semantically identical while Page/X/Y may be relocated.
- Custom Korean slots retain the selected renderer key/tree position but receive Korean raster dimensions/metrics and a new atlas position.
- `VerifyFullRepackSemantics` now reparses the produced PAF and re-extracts atlas rasters to verify these postconditions on asset-backed builds.

## Runtime claims that remain unproven

The repository currently does not contain a reproducible disassembly/test proving the exact retail runtime glyph lookup algorithm. In particular, these remain hypotheses until reconstructed from the authenticated executable:

1. The PAF `Left`/`Right` fields are traversed as a BST rooted at 1318 for every ordinary glyph lookup.
2. Every glyph consumer reads current PAF Page/X/Y rather than relying on hard-coded stock atlas coordinates.
3. There is no pre-lookup special handling for particular renderer-key ranges such as `0x87xx`.
4. There are no consumers that bypass the PAF lookup and address icon/font atlas cells directly.

Earlier reverse-engineering notes referred to a runtime `lhu` key read and `0x20` record stride near a renderer lookup path. That evidence is not presently preserved in a repository artifact that can be independently reproduced, so it must not be promoted to confirmed status during this audit.

## 0x87 status

Observed device behavior showed at least two mappings using lead byte `0x87` rendering game icons rather than the expected Korean glyphs. This is strong evidence that reusing those particular keys is unsafe for the current patch. It does **not** by itself prove whether the cause is executable special-key dispatch, an alternate font/icon resource, a direct atlas reference, or another rendering path.

`Minimal87` should therefore be treated as a diagnostic avoidance policy, not as proof that the whole `0x87xx` range is a formally identified renderer-private namespace.

## Next reverse-engineering gate

Recover the authenticated retail executable path that consumes a two-byte renderer key and reaches PAF metadata. Document the exact instruction range and data-flow that establishes:

- key formation / byte order;
- PAF base address;
- record stride;
- root selection;
- left/right traversal or alternate lookup mechanism;
- any special-key branches before lookup;
- Page/X/Y consumption after lookup.

Only after that evidence exists should the audit classify full PAF relocation as runtime-contract-safe rather than merely internally self-consistent.
