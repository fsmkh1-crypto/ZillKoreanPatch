# Current Claude review status — Review 5 remediation complete

Last updated: 2026-08-26
Branch: `translation/section001-batch2`

Claude Review 5 evaluated whether the repository could safely continue accumulating large-scale Korean translations without risking later loss or forced retranslation of translator-owned wording.

The review decision returned:

`BLOCKED`

That historical decision is preserved exactly. The review then identified two MUST-FIX issues, both of which have now been remediated and regression-tested. There has **not** been a second Claude review, so this document does not relabel Claude's original decision as a PASS.

Full review/remediation record:

- `docs/CLAUDE_REVIEW_5.md`

## MUST-FIX 1 — CLOSED

`tools/korean/refresh-japanese-refs.py` previously used a stale local control-token regex and treated `<line-break>` as fixed during destructive quarantine.

Current state:

- refresh imports shared `fixed_tokens` from `tools/korean/control_tags.py`;
- Korean line-break placement no longer participates in destructive fixed-control comparison;
- `next-packet.py` uses the same shared runtime-control grammar;
- regression tests prove a legitimately reflowed Korean row survives refresh;
- historical auto-commit audit found eight removed rows;
- two true false-positive removals (`30000`, `60011`) were safely recovered;
- six rows with changed canonical Japanese were deliberately left for retranslation rather than restoring stale wording.

## MUST-FIX 2 — CLOSED

Semantic Korean and build-owned wrapping are now structurally separated.

Current invariant:

- translator-owned `korean` **must not contain `<line-break>`**;
- build-owned `layout` may contain line breaks;
- Python import/apply paths reject semantic line breaks;
- Go corpus loading and `WithKorean` enforce the same rule;
- legacy Japanese-derived semantic breaks were deterministically migrated to spaces while their former visual placement was retained in `layout`;
- `CompileBankKorean` compiles semantic text independently from optional layout.

## Claude SHOULD-FIX items

- Python runtime-control grammar duplication: **CLOSED** by shared `control_tags.py` imports.
- Hangul-boundary reflow: **CLOSED**. `preservesSemantics` can now insert a generated line break between adjacent Hangul syllables while still rejecting wording/control changes; existing whitespace normalization remains supported.
- theoretical `<color:c>` / `<discard:c:$XX>` case where the embedded raw byte is literal `<` or `>`: **OPEN, NON-BLOCKING FOR BULK TRANSLATION**. No matching canonical emitted form has been found; keep for the final runtime-control/ISO audit.

## Data integrity result

The destructive-history audit is complete rather than assumed clean. The two Korean rows that were actually lost because of the old line-break-sensitive quarantine have been restored from history only after exact Japanese-source equality and fixed-control verification. Rows whose source changed are not restored against stale Japanese.

Current accepted corpus after recovery:

- 3,304 unique Korean records;
- 43,116 total source records;
- 932 custom renderer glyphs currently required.

The raster catalog has been regenerated and validated for the 932-rune requirement (`167a9fab788b8a36149983ee95df097297f57715`).

## Next action

Once the post-remediation CI head is green, bulk translation may resume under the repository's enforced invariant:

1. translate canonical Japanese meaning naturally into Korean;
2. preserve fixed runtime controls in exact order;
3. **do not emit `<line-break>` in semantic Korean**;
4. leave wrapping to build-owned `layout` generation.

No additional Claude review should be spent immediately unless a new failure changes one of these invariants. The next high-value adversarial review is the final integrated Korean ISO/build gate.
