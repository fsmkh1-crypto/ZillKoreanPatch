# Current Claude review request — renderer slot audit follow-up

Last updated: 2026-08-25

Read these first:

- `docs/CLAUDE_HANDOFF.md`
- `docs/CLAUDE_REVIEW_PROTOCOL.md`
- `docs/CLAUDE_REVIEW_LOG.md`

This is the follow-up gate after Claude's first renderer-slot audit returned `BLOCKED`.
Do not promote any renderer key from **candidate** to **safe/reusable** based on this review alone. The requested decision is only whether the implementation is sound enough to select a very small set of **PoC candidates** for empirical PPSSPP testing.

## First-review findings and changes

Claude's blocking finding was valid:

- `cmd/zill/korean_font.go::loadRetailBindata` opened both `pa` and `pami` unconditionally.
- The retail audit extraction subsequently confirmed `data/bindata.dat` is in `pa` (index 0), so requiring `pami` after it had already been found was unnecessary and could fail on an otherwise valid input tree.

The following changes are now on `main`:

- `f23b5a4cdddff59272a93e8c4a129ad2b85cd59e`
  - `pa` is searched first and `pami` is opened only as a fallback when `data/bindata.dat` is absent from `pa`.
  - Go now pins retail `data/bindata.dat` SHA-256 to `3241fc000f3d52fe8522baaa985fd866e29d64d3a0f23ac4e28b66dee957de3e`.
  - `fixeddata.ApplyEquipment` remains a structural/source-layout consistency check rather than the sole authentication mechanism.
- `aaa2db17abdc0919a3f12676a944973595c98af8`
  - regression test proves a `pa`-resident `data/bindata.dat` succeeds when no `pami.bin`/`pami.arc` exists;
  - regression test proves a wrong bindata fingerprint is rejected.

## Additional blocker found during GPT integration review

While validating the real audit ZIP, GPT found an issue that was not called out in the first review:

- retail `SYSDIR/EBOOT.BIN` is a PSP `~PSP` executable container, not the decrypted ELF text image;
- the canonical release pipeline authenticates `EBOOT.BIN`, but applies executable patches and fixed strings to `SYSDIR/BOOT.BIN`;
- `patches/executable/manifest.toml` explicitly targets `SYSDIR/BOOT.BIN` and pins its source SHA-256 to `5e294dc84a7f0d50719ecd26cb24ffb3792f2d9445803690845a8f1fa1cb85a3`;
- therefore treating raw `EBOOT.BIN` byte runs as executable CP932 literals was semantically wrong and could both create binary false positives and miss actual plaintext references.

This has been corrected:

- `3c63d3674fa5874f85f0c1057939ab28de2c1a97`
  - `korean-slots` still authenticates retail `EBOOT.BIN` to bind the supported game/version;
  - it separately authenticates retail decrypted `BOOT.BIN` with SHA-256 `5e294dc84a7f0d50719ecd26cb24ffb3792f2d9445803690845a8f1fa1cb85a3`;
  - `ScanCP932Literals` now scans authenticated `BOOT.BIN`, not raw `EBOOT.BIN`;
  - candidate output wording now says message/fixed/BOOT/bindata audit.
- `08d78c3af26feeab5eded6b944234579173387dd`
  - Android audit extraction authenticates and exports `SYSDIR/BOOT.BIN` in addition to `EBOOT.BIN`;
  - audit manifest format is now v4 / extractor 0.5.0.
- `b9ad631484f8adfaac91b03036f22d2a7f32e133`
  - Android app version bumped to 0.5.0.

The old v0.4 audit ZIP is therefore insufficient for the final candidate calculation because it does not contain `BOOT.BIN`. It did, however, authenticate the expected font members, EBOOT and bindata and confirmed bindata's archive location.

## Current review target

Review current `main`, with special attention to:

- `cmd/zill/korean_font.go`
  - `loadAuthenticatedRetailBOOT`
  - `loadAuthenticatedRetailEBOOT`
  - `loadBindataFromArchive`
  - `loadRetailBindataWithSHA`
  - `runKoreanSlots`
- `cmd/zill/korean_font_test.go`
- `internal/slotaudit/scan.go`
- `internal/slotaudit/scan_test.go`
- `android-patcher/app/src/main/java/com/fsmkh1/zillfontdump/FontExtractor.java`
- `android-patcher/app/src/main/java/com/fsmkh1/zillfontdump/MainActivity.java`
- `patches/executable/manifest.toml`
- `internal/release/build.go::buildExecutable`

Do not spend time redesigning unrelated Gemini translation exchange code.

## Intended safety model

A key is removed from the current candidate set if referenced by any of these audited sources:

1. current English message corpus;
2. retail Japanese message corpus;
3. canonical fixed EBOOT strings from `release/strings/eboot.toml`;
4. canonical equipment strings from `release/strings/equipment.toml`;
5. CP932-looking NUL-terminated literals recovered from the authenticated decrypted retail `BOOT.BIN`;
6. CP932-looking NUL-terminated literals recovered from authenticated retail `data/bindata.dat`.

The remaining set is still reported only as `AUDITED CANDIDATES ONLY`. Other archive/UI/script resources are not yet semantically parsed.

`ScanCP932Literals` deliberately does not accept every random Shift-JIS-looking byte pair. The known tradeoff from the first review remains: a single two-byte glyph with no sufficient ASCII context can be missed. That limitation must be compensated for before a chosen PoC key is used, for example by exact byte-occurrence inspection of the selected key in authenticated plaintext/bindata sources and by reviewing remaining resource classes.

## Follow-up questions Claude must answer

1. Is the original MUST-FIX actually closed, including the no-`pami` regression test?
2. Is the explicit bindata SHA pin correctly enforced before scanner results affect candidate selection?
3. Is `BOOT.BIN`, rather than the encrypted/raw `EBOOT.BIN`, now the correct executable plaintext source to scan according to the canonical build pipeline and executable manifest?
4. Does the implementation still authenticate `EBOOT.BIN` sufficiently to bind the input to supported ULJM05410 v1.03 while scanning the authenticated `BOOT.BIN` plaintext?
5. Can either the pa→pami fallback or BOOT/EBOOT trust boundary silently accept the wrong file/source?
6. Re-evaluate `ScanCP932Literals` false-negative/false-positive behavior specifically on decrypted `BOOT.BIN` and bindata, not encrypted `EBOOT.BIN`.
7. For a first real-sentence PoC using only a handful of candidate keys, what **minimum additional audit** is a hard blocker? In particular, distinguish:
   - exact raw-byte occurrence checks for the selected key(s);
   - other known plaintext/script/UI archive members that can reference the font;
   - desirable whole-game completeness that can wait until after PoC.
8. Search current output/docs for any accidental `safe`, `unused`, or `reusable` claim that exceeds the evidence.

## Required output format

For each finding:

- severity: MUST-FIX / SHOULD-FIX / NOTE;
- file + function;
- concrete failure mode;
- minimal reproducible test or byte example when possible;
- proposed fix direction;
- whether it blocks selecting first PoC candidate keys.

End with exactly one gate decision:

- `BLOCKED`; or
- `PASS FOR POC CANDIDATE SELECTION`.

A pass means only that candidate selection for empirical PoC is justified; it does **not** mean any slot is production-safe.

If GitHub write access is unavailable, output the full review in chat for manual transfer to GPT.
