# Current Claude review request — renderer slot audit

Last updated: 2026-08-25

Read these first:

- `docs/CLAUDE_HANDOFF.md`
- `docs/CLAUDE_REVIEW_PROTOCOL.md`
- `docs/CLAUDE_REVIEW_LOG.md`

This review is a required gate before any renderer key is promoted from **candidate** to **safe/reusable**, and before the first real Korean-sentence ISO PoC is produced.

## Review target

Primary implementation commits/files:

- Go slot-audit implementation through `0d3c8fca4397cf415084d1f7c19a26eb901ac6c3`
  - `internal/slotaudit/scan.go`
  - `internal/slotaudit/scan_test.go`
  - `cmd/zill/korean_font.go`
- Android audit extraction through `c2b9ee969ac0bacf8dea63baa09c0aedff090869`
  - `android-patcher/app/src/main/java/com/fsmkh1/zillfontdump/FontExtractor.java`
  - `android-patcher/app/src/main/java/com/fsmkh1/zillfontdump/MainActivity.java`
  - `android-patcher/app/src/test/java/com/fsmkh1/zillfontdump/PoCPatcherTest.java`

Do not spend time redesigning unrelated Gemini translation exchange code.

## Intended safety model

The current audit intentionally does **not** claim that a remaining key is safe.

A key is removed from the candidate set if it is referenced by any of the currently audited sources:

1. current English message corpus;
2. retail Japanese message corpus;
3. canonical fixed EBOOT strings from `release/strings/eboot.toml`;
4. canonical equipment strings from `release/strings/equipment.toml`;
5. CP932-looking NUL-terminated literals recovered from the authenticated retail EBOOT;
6. CP932-looking NUL-terminated literals recovered from authenticated retail `data/bindata.dat`.

The remaining set is reported only as `AUDITED CANDIDATES ONLY`. Other archive/UI/script resources are not yet semantically parsed.

The binary scanner deliberately does not treat every Shift-JIS-looking byte pair in arbitrary binary data as a text reference, because compressed/image/random binary content would create massive false positives. It instead looks for plausible NUL-terminated CP932 string runs.

## Questions Claude must answer

### MUST review

1. **False-negative risk in `ScanCP932Literals`:** Can a plausible runtime-visible Japanese string in EBOOT/bindata be missed because of the current acceptance rules (NUL termination, minimum double-byte count, ASCII mix, control bytes, half-width kana, reset behavior)? Give concrete byte examples where possible.
2. **False-positive risk:** Could current scanner rules eliminate large numbers of actually reusable keys due to binary lookalikes? Is that merely conservative, or can it make the approach unusable?
3. **CP932 byte validity:** Verify lead/trail rules and `GlyphKeyFromBytes` interaction, including 0x7F exclusion and single-byte handling.
4. **Trust boundary:** Verify `korean-slots` only interprets retail EBOOT/bindata after the expected game/version/source authentication checks. Look for a path where modified/wrong assets could silently contribute to the audit.
5. **Bindata discovery:** Review pa/pami search behavior for duplicate/missing `data/bindata.dat`, archive bounds, and source fingerprint handling in both Go and Android implementations.
6. **Android extraction:** Verify it is read-only toward the source ISO, extracts the exact intended members/files, revalidates before export, cleans partial output on failure, and does not accidentally overwrite the source.
7. **Safety language:** Search the changed code/docs for any place that incorrectly promotes a candidate to `safe`, `unused`, or `reusable` without evidence.
8. **Coverage gap:** Identify the minimum additional source/resource classes that must be audited before one or a small handful of keys can reasonably be used for the first real-sentence PoC. Distinguish a hard blocker from desirable completeness.

### Regression awareness

The previous Claude review MUST-FIX about unmatched Korean replacement IDs is already closed. Do not reopen it unless the current code regressed.

The duplicated stock/Korean projection logic remains known non-blocking technical debt and is outside this review unless the new slot-audit work interacts with it.

## Required output format

For each finding:

- severity: MUST-FIX / SHOULD-FIX / NOTE;
- file + function;
- concrete failure mode;
- minimal reproducible test or byte example when possible;
- proposed fix direction;
- whether it blocks using the audit to choose the first PoC slot.

End with an explicit gate decision:

- `BLOCKED` — one or more MUST-FIX findings prevent using the audit results; or
- `PASS FOR POC CANDIDATE SELECTION` — no blocking finding, while still preserving the distinction that candidate != production-safe.

If GitHub write access is unavailable, output the full review in chat for manual transfer to GPT.
