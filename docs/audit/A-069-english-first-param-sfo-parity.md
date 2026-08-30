# A-069 — English-first PARAM.SFO parity

Status: **PASS**

## Scope

Audit the Korean `PARAM.SFO` build path from the English-patch implementation rather than treating the Korean release as an independent transform.

## English baseline

`internal/release/build.go::buildSFO`:

1. reads retail `PARAM.SFO` from the authenticated PSP_GAME tree;
2. reads `patches/system/param-sfo.toml`;
3. parses the shared manifest with `sfo.ParseManifest`;
4. applies it with `sfo.Apply`;
5. supplies only the English localized display title.

The shared manifest pins the supported retail source and the engine-facing mutation:

- source size: 472 bytes;
- PSF magic: `0x46535000`;
- version: `0x00000101`;
- `MEMSIZE` must be absent in retail;
- append `MEMSIZE` as format `0x0404`, value `1`, alignment `16`.

## Korean result

`internal/release/korean_build.go::buildKoreanAlphaSFO` uses the same retail input, the same `patches/system/param-sfo.toml`, the same parser, and the same `sfo.Apply` implementation. The only difference is the localized display title:

`Zill O'll Infinite Plus [Korean Beta %s]`

Classification: **DIFFERENT-BY-DESIGN only for the title string; engine/storage contract PASS.**

## Gate

`tools/korean/audit-english-first-param-sfo.py` now fails closed if either release path stops using the shared manifest/transform or if the authenticated SFO manifest loses its retail fingerprint/shape/MEMSIZE guards. CI runs this audit before Go tests.

## Freeze relevance

This closes a static system-file parity surface. It does **not** prove absence of runtime freezes and should not be interpreted as runtime safety evidence.
