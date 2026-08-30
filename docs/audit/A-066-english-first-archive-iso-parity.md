# A-066 — English-first archive and ISO provenance parity

## Scope

Restart the archive/output boundary audit from the upstream English release contract rather than inheriting prior Korean conclusions.

Compared:

- English `internal/release/build.go`
- Korean desktop `internal/release/korean_build.go`
- Korean mobile `internal/release/korean_mobile.go`
- shared `internal/gamefmt/paa/paa.go`
- shared `internal/release/disc.go`

## Result

### PAA archive rebuild — PASS

English and both Korean production ISO paths rebuild `pa`/`pami` through the same `paa.Pair.Rebuild` implementation.

The shared implementation is fail-closed on the relevant engine/archive contracts:

- replacement indexes must resolve to existing retail members;
- the same archive member may not be selected by more than one replacement;
- source files cannot be overwritten as rebuild destinations;
- offsets are recomputed with the archive's 16-byte alignment contract;
- member sizes and offsets must remain representable as `uint32`;
- the rebuilt pair is reopened and structurally validated before publication;
- every rebuilt member payload is compared byte-for-byte with either the requested replacement or the original retail payload;
- member identity/tree metadata and the archive prefix are preserved.

No Korean-only relaxation was found at this boundary.

### Final ISO authoring — PASS

English and both Korean production ISO paths use the same `authorTranslatedISO` helper.

After ISO construction, the helper reopens the authored image and walks the staged `PSP_GAME` tree. Every staged regular file is opened from the authored ISO and compared byte-for-byte. A mismatch, missing file, or length difference deletes the staged ISO and fails the build.

Therefore the verified chain is:

`compiled payload -> PAA replacement -> reopened rebuilt archive -> staged PSP_GAME -> reopened authored ISO`

This is a static provenance contract only. It does not imply that a successful ISO build proves runtime freeze safety.

## Hardening added

`tools/korean/audit-english-first-full-parity.py` now fails CI if:

- shared PAA duplicate-replacement rejection disappears;
- rebuilt-member exact verification disappears;
- English, Korean desktop, or Korean mobile bypasses `paa.Pair.Rebuild`;
- any production release path bypasses `authorTranslatedISO`;
- final staged-`PSP_GAME` exact verification disappears.

## Classification

- Archive replacement/rebuild parity: **PASS**
- Final ISO staged payload provenance parity: **PASS**
- New engine defect discovered in this audit slice: **none**
