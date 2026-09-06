# Claude Review Brief — Korean Dialogue Copyedit Safety Rules

Repository: `fsmkh1-crypto/ZillKoreanPatch`
Branch to review: `fix/u6-korean-reflow-parity`
Authoritative base: `milestone/U6`

## Purpose

Independently review the proposed rules for a full Korean dialogue copyedit pass before any mass editing begins.

This is **not** a translation-quality review yet. The question is whether the editing rules themselves are safe, correctly scoped, and faithful to the English patch / engine contract.

Project principle: **English patch first.** When the English patch already implements an engine-facing contract, Korean should follow that method unless Korean/custom-renderer requirements provide concrete evidence for a necessary divergence.

## Files / code that must be inspected

At minimum, inspect:

- `AGENTS.md`
- `CONTRIBUTING.md`
- `README.md`
- `internal/corpus/korean.go`
- `internal/koreancorpus/corpus.go`
- `internal/corpus/control_tags.go`
- `internal/layout/engine.go`
- `internal/layout/korean_dialogue_reflow.go`
- related Korean source-aware reflow implementation/tests on this branch
- representative files under `translations/korean/messages/`
- corresponding English patch message data / paths where relevant

Do not review the rules only as prose. Check them against actual code behavior and the English implementation.

## Proposed copyedit rules to review

### A. Immutable / protected data

1. Record IDs are immutable. Do not add/delete/reassign records as part of copyediting.
2. `japanese` source text is immutable.
3. Existing runtime control structure must remain compatible with the source/consumer contract.
4. Runtime substitutions such as `%s`, `%u`, and tagged/value controls must not be accidentally changed, deleted, reordered, or invented.
5. TOML structure, section keys, headers, and unrelated metadata must not be changed by a dialogue copyedit.

### B. Meaning vs layout separation

1. `korean` is semantic authored text.
2. Ordinary dialogue copyediting must not insert authored `<line-break>` tags into `korean` merely to make lines fit.
3. Layout/reflow belongs to the build/layout path unless a consumer is a verified fixed visual layout that explicitly requires authored breaks under the existing project rules.
4. If a touched record already has generated `layout`, that old layout must be treated as stale after changing `korean`; do not manually copyedit the old `layout` in parallel. Invalidate/remove it and let the current U6 Korean reflow path derive a new one where applicable.
5. Do not invent a Korean-only reflow rule when the English patch already establishes the consumer contract.

### C. Control-contract safety

1. Fixed runtime controls must remain byte-for-byte compatible and in the required order.
2. `<line-break>` is layout-authorable and therefore is not treated the same as fixed runtime controls.
3. Literal occupancy around fixed controls must not be destroyed in a way that changes runtime semantics; interior literal slots required by the existing validator must remain occupied.
4. Leading/trailing translation omissions are allowed only where the existing contract/validator allows them.
5. Any case whose runtime/control meaning cannot be proven should fail closed / be held for manual review rather than normalized automatically.

### D. Copyedit scope

The pass is **full proofreading + minimal polishing**, not a wholesale retranslation.

Actively fix, when context supports it:

- Korean spacing errors
- unnecessary or excessive commas, especially punctuation that feels mechanically carried over from Japanese/English pause structure
- awkward particles/endings
- unnatural word order
- redundant wording
- stiff literal-translation phrasing
- obviously malformed or unnatural sentences
- punctuation abnormalities
- clear register/honorific inconsistencies

Do not mechanically normalize:

- character catchphrases or deliberate speech quirks
- archaic/stylized diction
- intentional stutter, repetition, hesitation, or emotional punctuation
- proper nouns, item/place/skill terminology without checking established usage
- choice/system/UI strings whose brevity or wording may be consumer-sensitive
- ambiguous lines whose intended meaning is not clear from context

### E. Candidate detection vs editing

1. Automated scanning may flag candidates, but must not globally rewrite text by regex/search-replace.
2. Comma count alone is not grounds for deletion; retain commas needed for lists, vocatives/interjections, ambiguity prevention, deliberate cadence, or character voice.
3. Each actual edit should be context-reviewed using, as needed:
   - current Korean
   - Japanese source
   - corresponding English patch/reference
   - adjacent dialogue / scene context
4. Classify candidates roughly as:
   - simple copyedit
   - contextual rewrite
   - hold / uncertain

### F. Required validation after edits

At minimum verify:

- ID set unchanged
- Japanese text unchanged byte-for-byte
- fixed runtime control sequence preserved
- runtime substitutions preserved
- TOML parses successfully
- no improper authored `<line-break>` added to semantic Korean
- Korean control-contract validation passes
- touched stale layouts are regenerated through the current U6 path where applicable
- Korean reflow/source-aware tests pass
- `go test ./...`
- `go vet ./...`
- `./zill check`

Also account for the documented limitation that contributor `check` alone does not prove retail/materialized runtime safety; asset-backed build/runtime QA remains relevant for release confidence.

## Questions Claude must answer

Please give a code-grounded review, not a generic style opinion.

1. **Overall verdict:** Are these rules safe and internally consistent for this repository? `PASS`, `PASS WITH CHANGES`, or `FAIL`.
2. Which proposed rules are directly supported by existing English-patch behavior or current code? Cite file/function/test evidence.
3. Is any rule **too strict** and likely to reject valid Korean edits?
4. Is any rule **too loose** and capable of breaking runtime behavior, reflow, consumer limits, control semantics, or patch reproducibility?
5. Is the rule “edit `korean`, invalidate stale generated `layout`, then rederive layout” correct for every relevant consumer? Identify exceptions if any.
6. Are there controls/substitutions for which the proposed wording is misleading? Check movable value tags versus fixed controls carefully.
7. Does the current U6 Korean reflow branch truly mirror the relevant English reflow population/consumer logic, or are there remaining parity gaps that make a mass copyedit unsafe?
8. Do fixed-layout/select/control cases need additional explicit exclusions or review gates? Inspect representative tests and known special cases rather than guessing.
9. Should punctuation/spacing QA be implemented only as candidate linting, or are any parts safe enough to enforce as hard validation? Explain why.
10. Are there missing invariants we should add before editing tens of thousands of dialogue records?
11. Are any proposed rules redundant with current validators/tests and therefore better left as operational review policy rather than new code?
12. What minimal changes, if any, should be made to this rule set before the Korean dialogue copyedit starts?

## Required response format

Use this structure so the result can be relayed back to ChatGPT verbatim:

```text
VERDICT: PASS | PASS WITH CHANGES | FAIL

CRITICAL ISSUES
- ...

RULES TO CHANGE
- Rule / wording:
  Problem:
  Evidence:
  Recommended replacement:

MISSING SAFETY RULES
- ...

RULES CONFIRMED CORRECT
- ...

ENGLISH-PATCH PARITY FINDINGS
- ...

MASS-COPYEDIT GO/NO-GO
- GO | GO AFTER FIXES | NO-GO
- Preconditions: ...
```

Please distinguish proven findings from uncertainty. If repository evidence is insufficient, say `UNKNOWN` rather than inventing a rule.