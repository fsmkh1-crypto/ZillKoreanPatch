# Claude review request — continued bulk Korean translation gate

Review the current head of branch `translation/section001-batch2` directly from the repository. Do not rely on prior-chat assumptions; use this document only as scope/context and verify claims against code, tests, workflows, and representative Korean overlay files.

## Gate being decided

The first real Korean sentence renderer/message PoC already received:

`PASS FOR FIRST-SENTENCE POC TEST`

and subsequently passed empirically in PPSSPP. The observed first screen rendered Korean correctly while leaving the following Japanese lines intact.

This review is **not** the final ISO-release gate. It is the gate for whether we can safely continue translating tens of thousands of records without creating canonical Korean data that will later need to be discarded or substantially rewritten because of pipeline/schema/control/layout mistakes.

The active Korean corpus currently contains roughly 3.3k accepted records out of 43,116 source records. Translation automation works in large packets and the corpus will grow quickly after this gate.

## Explicitly out of scope for this gate

`internal/release/build.go::Build` is not yet fully switched from the existing stock/English release path to the integrated Korean bank/font plan. Final Korean ISO wiring and final source-asset authentication are still pending production work.

Do **not** fail this gate solely because that final `release.Build` Korean ISO integration is incomplete. Do fail it if something in the current canonical corpus, import/validation pipeline, slot/font planning, or layout ownership would make continued translation unsafe or force large-scale retranslation/data migration later.

## Recent changes that require adversarial review

1. Korean overlays may be sharded as `msgsecNNN-partNN.toml`; legacy `msgsecNNNb.toml` is temporarily accepted. Both Korean loaders now accept those forms and reject conflicting duplicate IDs.
2. `KoreanEntry` now separates translator-owned semantic `Korean` text from optional machine/build-owned `Layout`. Editing semantic Korean invalidates an old layout. Runtime text prefers layout when present.
3. Japanese `<line-break>` locations are no longer intended to be mandatory in future Korean translation imports. Fixed runtime bytecode controls remain mandatory and ordered.
4. A precise runtime-control recognizer was added because angle-bracketed natural text such as `<未使用>` is actual translatable game text, not bytecode. Both Go loaders use the shared recognizer. Python import tools use a matching recognizer and now allow line-break reflow.
5. `apply-results.py` preserves an existing `layout` field rather than silently dropping it when rewriting auto-part files.
6. Python Korean tooling is syntax-checked and its runtime-control grammar is regression-tested in CI.
7. Korean raster catalog generation is deterministic and now self-refreshes on translation branches; slot/font planning is corpus-driven rather than tied to the five PoC keys.

## Known unresolved issue — inspect carefully

The current Korean compiler still has legacy line-break semantics that may not match the intended canonical ownership model:

- many already-imported Korean rows inherited Japanese `<line-break>` positions because older import validation required the same token sequence;
- `internal/message/compile_korean.go` currently treats any `<line-break>` in `KoreanRecord.Text` as explicitly authored and rejects a different generated layout;
- `preservesSemantics` currently allows whitespace spans to become line breaks, but Korean UI wrapping may need breaks at Hangul syllable boundaries, not only at existing whitespace;
- source-fixed runtime line breaks must still remain fixed wherever the retail projection proves they are structurally locked.

Determine whether this requires a migration/schema/compiler fix **before more bulk translation**, and propose the smallest safe rule. In particular, distinguish movable/layout line breaks from structurally fixed source controls rather than treating every Japanese line break identically.

## MUST inspect

### A. Canonical corpus durability

Check `internal/corpus/korean.go`, `internal/koreancorpus/*`, representative `translations/korean/messages/*`, and the import/apply scripts.

Determine whether new translations can be accumulated now without later losing semantic text, layout information, Japanese source binding, or ID ownership. Check shard handling, duplicate handling, deterministic ordering, stale Japanese rejection, and update/render behavior.

### B. Runtime-control correctness

Cross-check the runtime-control regex/grammar against `internal/corpus/bank.go::displayText` and the actual projection/compiler rules.

Look specifically for both:

- false negatives that would allow a translator/import script to alter real runtime bytecode;
- false positives that would freeze ordinary translatable angle-bracketed game text such as `<未使用>`.

Check Go and Python implementations for drift.

### C. Semantic Korean versus layout

Check whether `Korean`, optional `Layout`, `Texts()`, `Layouts()`, `RuntimeTexts()`, `compileKoreanBanks`, and `CompileBankKorean` agree on ownership and precedence.

Determine whether legacy Japanese-derived line breaks in existing Korean rows can be normalized safely without changing translated wording, and whether future Korean semantic translations should contain zero layout-only line breaks by default.

### D. Korean wrapping model

Assess whether the current whitespace-only `preservesSemantics` rule is sufficient for Korean. If not, specify a safe build-owned reflow model that can insert line breaks between Hangul syllables/words while preserving runtime substitutions and fixed controls.

The PPSSPP PoC visibly demonstrated why padding/fake width is not acceptable: correct Korean rendered, but the short padded sentence appeared horizontally offset. Production output must rebuild record lengths/offsets and produce real Korean layout.

### E. Translation import safety

Review `tools/korean/control_tags.py`, `apply-results.py`, `import-translations.py`, `auto-trivial.py`, `next-packet.py`, and relevant workflows. Look for silent data loss, duplicate/conflict bypass, layout dropping, control drift, stale source references, or packet races.

### F. Renderer-slot/font scalability

Review `internal/koreanslots`, `internal/koreanfont`, Korean raster workflow, and release Korean font helpers. Confirm deterministic allocation/catalog behavior and fail-closed handling of missing rasters, capacity exhaustion, unsuitable cells, duplicate keys, or missing authenticated inputs.

Do not call currently audited candidates globally production-safe if the evidence does not support that statement.

### G. Existing translated-data risk

Most important: state explicitly whether any current issue would require retranslation of already translated Korean **wording**, versus merely a mechanical layout/control metadata migration. If a migration is needed, describe how to make it deterministic and lossless before translation volume grows further.

## Output format

Organize findings by severity and include file/function references plus a concrete failure mode or reproducible test for every blocker.

Use these headings where applicable:

- `🔴 MUST-FIX BEFORE MORE BULK TRANSLATION`
- `🟡 SHOULD-FIX BEFORE FINAL ISO BUILD`
- `🟢 VERIFIED / SAFE TO CONTINUE`

For every MUST-FIX, say whether it threatens existing Korean wording, only layout/metadata, or runtime correctness.

End the complete review with **exactly one** of these lines:

`BLOCKED`

or

`PASS FOR CONTINUED BULK TRANSLATION`
