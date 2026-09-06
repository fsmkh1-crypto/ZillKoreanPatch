# Beta1 section 001 copyedit layout pre-flight

Date: 2026-09-06

Status: **PRE-FLIGHT PASS — semantic copyedit is still unapplied**.

This note continues `docs/BETA1_WORK_HANDOFF.md` and the reviewed proposal manifest `docs/audit/beta1-copyedit-batch001-proposals.json`.

## Scope

- Korean file: `translations/korean/messages/msgsec001.toml`
- Reviewed IDs: `10000..10175` (176 records)
- Proposed semantic modifications: 29
- Proposed changes already reviewed against Japanese + repository English: punctuation 27, translation 2, naturalness 2 (overlapping categories)
- Persisted layout present among proposed records: 27
- Proposed records without persisted layout: `10022`, `10035`
- New Hangul required by the proposal manifest: 0
- No proposal changes runtime control tokens or substitution topology.

Proposed-change IDs:

`10002, 10006, 10007, 10008, 10009, 10010, 10012, 10015, 10018, 10021, 10022, 10024, 10027, 10030, 10033, 10035, 10036, 10039, 10042, 10045, 10048, 10051, 10054, 10057, 10060, 10063, 10066, 10069, 10072`

## English-patch / consumer principle

The existing Beta1 implementation mirrors the upstream English separation between semantic text and display layout: fixed/source controls retain their semantics, ordinary display line boundaries are layout-owned, and consumer/storage contracts are validated after projection.

For Korean, `KoreanProject.WithKorean` explicitly invalidates a previous layout when semantic Korean changes because generated layout is build-owned. The release contract chain subsequently derives/validates dialogue, English-consumer storage and visual layouts before compilation.

However, the checked-in section-001 overlay already contains persisted `layout =` values for 27 of these 29 proposals. `effectiveKoreanText` treats a persisted row layout as authoritative unless a newer derived overlay replaces it. The generic Korean English-consumer derivation repairs C22 storage; the hard visual auto-reflow path specifically derives character-profile layouts. Character-creation prompt width outside hard categories remains represented in the upstream-style warning census. Therefore changing only `korean =` while leaving stale section-001 `layout =` lexical content is not acceptable.

## Repository-native safe synchronization path

`tools/korean/qa-layout-drift.py` is the existing repository mechanism for this exact situation.

Its `--fix` path:

1. compares canonical `korean` against persisted `layout` after removing whitespace and `<line-break>`;
2. updates only stale lexical content while preserving existing whitespace / line-break boundaries;
3. refuses an edit that would cross a whitespace or `<line-break>` boundary rather than guessing a structural layout;
4. verifies the synchronized layout converges to the semantic Korean content;
5. reparses the output as TOML before writing.

The 27 layout-bearing proposals in batch001 are compatible with that model: sentence-ending periods are inserted at existing line boundaries; the three dependent-clause commas (`10012`, `10066`, `10069`) are removed at existing boundaries; `10021` changes `지평선` to source-correct `수평선` within the same first-line segment and restores the sentence boundary. `10007` and `10010` likewise restore punctuation at their already-existing line boundaries. No proposal requires inventing a new semantic `<line-break>`.

`10022` and `10035` have no persisted layout, so only their semantic Korean changes. Their proposal forms introduce no new Hangul outside the current custom catalog according to the saved proposal census.

## Acceptance procedure for this batch

Do not count the 29 proposals as applied until all steps below complete.

1. Verify branch/HEAD is a descendant of the current Beta1 checkpoint; never reset legitimate newer work.
2. Apply only the 29 manifest `PROPOSED_CHANGE` semantic values, requiring the current Korean to equal each manifest `before_korean` first.
3. Preserve Japanese, IDs, controls and substitutions byte-for-textually.
4. Run `python3 tools/korean/qa-layout-drift.py --fix` so the 27 existing layouts receive only the repository-approved lexical synchronization described above.
5. Run `python3 tools/korean/qa-layout-drift.py` and require `layout_lexical_drift_count=0`.
6. Run Korean glyph audit and require no missing/new-unhandled Hangul.
7. Run integrity/terminology/consistency/voice/text-sanity checks.
8. Run Go/consumer gates through a provisioned environment or CI: `go test ./...`, `go vet ./...`, `./zill korean-check`, `./zill korean-font-check`, full English-consumer/storage/visual census and full static reflow census.
9. Require the established full-corpus assertions to remain green, including 42,016 accepted rows unless a deliberate source-accounting change was made, and `residual_overflow=0` for the verified static dialogue population.
10. Keep the 47 runtime-unbounded rows as PENDING unless independent runtime evidence closes them.
11. Only after all gates pass, mark batch001 as APPLIED/ACCEPTED and update its audit statistics.

## Important non-actions

- No semantic Korean was changed by this pre-flight.
- No persisted layout was changed by this pre-flight.
- No renderer/font code changed.
- No generated glyph catalog changed.
- No Android RC was requested.
- No U-series branch was deleted.

Next action: apply and validate the 29 section-001 proposals using the procedure above, then checkpoint the accepted batch before beginning section 000 or section 002.
