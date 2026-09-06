# Claude Independent Review Brief — Full Korean Dialogue QA / Reflow / Terminology Plan

Repository: `fsmkh1-crypto/ZillKoreanPatch`
Branch to review: `fix/u6-korean-reflow-parity`
Authoritative base: `milestone/U6`

## Purpose

Independently audit the **entire planned Korean dialogue cleanup and QA process before mass editing begins**.

This review is deliberately broader than copyediting. The planned work includes:

1. full-corpus Korean proofreading and minimal polishing;
2. spacing, punctuation, particles, awkward syntax, repetition, register and dialogue-naturalness review;
3. canonicalization of proper nouns, place names, people, factions, items and other terminology;
4. explicit Korean transliteration rules for Japanese long vowels, sokuon, epenthetic vowels/consonant clusters and foreign-name reconstruction;
5. English-patch-first terminology and engine-contract parity;
6. detection of missing or incorrect dialogue reflow/line-break generation;
7. source-aware Korean reflow parity with the English implementation;
8. runtime-control / substitution / literal-occupancy invariants;
9. stale-layout invalidation and regeneration after semantic edits;
10. consumer-specific fixed-layout / bounded-buffer / select-control handling;
11. static validation, tests, asset-backed build checks and runtime QA;
12. evidence logging, coverage accounting, regression cases and handoff discipline.

The goal is to determine whether our **rules, workflow, validation gates and planned coverage are actually sufficient to edit tens of thousands of Korean dialogue records without silently breaking the patch**.

This is not merely a prose/style review. Check the rules against the repository's actual implementation and the English patch's behavior.

Project premise: **ENGLISH PATCH FIRST.** Where the English patch already implements an engine-facing contract, that implementation is the primary reference for Korean. A Korean-specific rule is allowed only when Korean/custom-renderer requirements make a divergence necessary and the underlying engine contract remains preserved with evidence.

---

## Mandatory files / code to inspect

At minimum inspect all of the following on `fix/u6-korean-reflow-parity`:

### Project rules / documentation
- `AGENTS.md`
- `CONTRIBUTING.md`
- `README.md`
- `docs/KOREAN_TRANSLATION_STYLE.md`
- `docs/KOREAN_DIALOGUE_QA_PROTOCOL.md`
- this review brief

### Korean corpus / control contracts
- `internal/corpus/korean.go`
- `internal/koreancorpus/corpus.go`
- `internal/corpus/control_tags.go`
- any relevant semantic/control validators and tests

### English and Korean layout/reflow
- `internal/layout/engine.go`
- `internal/layout/korean_dialogue_reflow.go`
- `internal/layout/korean_source_aware_reflow.go`
- `internal/layout/korean_repository_source_aware.go`
- related tests on this branch, including C5/dialogue/source-aware/select-control cases
- consumer/category/configuration data used by the layout engine

### Translation / terminology data
- representative records under `translations/korean/messages/`
- corresponding English message records and Japanese source fields
- `translations/terminology/` files, especially canonical names/place names
- any existing Korean terminology or style tables

### Known regression / investigation cases
Inspect at least these known areas rather than reasoning abstractly:

- early-tutorial Ferme/`フェルム` dialogue in `msgsec198`, including ID `1980005` and its `<value:$28>` substitution;
- the previously fixed `280181` select-count/fixed-control case;
- representative C5 / C5Portrait dialogue;
- records with movable runtime value tags;
- records with fixed runtime controls;
- records that already contain generated `layout`;
- records with authored/fixed visual breaks where applicable.

If paths/names differ in the current branch, locate the actual implementation instead of assuming this list is exhaustive.

---

# A. Review the project-wide English-first rule

Check whether the current rule in `AGENTS.md` is technically sound and appropriately scoped.

Questions:

1. Does the English patch actually provide the relevant engine-facing contract for each of message compilation, reflow, page boundaries, consumer limits, controls and substitutions?
2. Are there areas where "English patch first" is too broad because Korean/custom-renderer behavior necessarily differs?
3. Are those differences documented and testable rather than heuristic?
4. Is any Korean-specific implementation currently overriding an English-established contract without sufficient evidence?
5. Are there remaining English/Korean parity gaps that would make a mass dialogue pass unsafe?

Classify material comparisons as:

- `PASS`
- `DIFFERENT-BY-DESIGN` with evidence
- `MISSING`
- `UNKNOWN`

---

# B. Review immutable/protected dialogue data rules

Proposed invariants:

1. Record IDs are immutable during copyediting.
2. `japanese` source text is immutable.
3. TOML record identity/structure and unrelated metadata must not drift.
4. Runtime control topology and semantics must remain compatible with the source/consumer contract.
5. Runtime substitutions such as `%s`, `%u`, `<value:$XX>` and other tagged/value controls must not be accidentally changed, removed, reordered or invented.
6. Fixed runtime controls must remain compatible in required byte/order semantics.
7. Interior literal occupancy around fixed controls must not be accidentally destroyed where the existing validator requires visible literal slots.
8. If proof of runtime/control equivalence is unavailable, fail closed / hold for review.

Verify these rules against actual validators. Identify anything that is overstated, understated, or technically inaccurate.

Pay particular attention to the distinction between **movable value tags** and **fixed controls**.

---

# C. Review semantic Korean vs generated layout separation

Proposed rule:

- `korean` is canonical semantic authored text.
- ordinary dialogue copyediting does not manually insert `<line-break>` merely to fit the screen;
- generated `layout` belongs to the build/reflow path;
- changing `korean` makes an existing generated `layout` stale;
- touched stale layout should be invalidated and rederived through the current reflow path where applicable;
- verified fixed visual-layout consumers may be exceptions according to the established English/project contract.

Questions:

1. Is this correct for every relevant consumer?
2. Which consumers are exceptions?
3. Could removal of an existing `layout` ever lose intentionally authored/verified layout rather than generated layout?
4. Is there sufficient metadata or code distinction to know which is which?
5. Is `WithKorean` or another existing path the safest way to mutate semantic Korean?
6. If direct TOML edits are used, what exact checks are necessary to prevent stale layout from surviving?

---

# D. Review the full Korean copyedit scope

The planned pass is **full proofreading + minimal polishing**, not wholesale retranslation.

We intend to actively review/fix where context supports it:

- spacing;
- unnecessary/excessive commas;
- malformed punctuation;
- particles and endings;
- awkward word order;
- duplicated/redundant wording;
- stiff Japanese/English literal-translation structure;
- unnatural Korean dialogue;
- obvious grammar errors;
- inconsistent honorific/register usage;
- inconsistent terminology/proper-noun spelling.

We intend to preserve unless clearly wrong:

- character catchphrases;
- deliberate speech quirks;
- archaic/stylized diction;
- intentional stutters, hesitations and emotional repetition;
- meaningful punctuation/cadence;
- concise system/choice/UI wording constrained by its consumer;
- ambiguous lines until context resolves them.

Review whether this scope is appropriate for this project. Identify any classes that need a stricter gate or separate workflow.

Also answer whether any spacing/punctuation checks are safe as **hard validators**, versus candidate-only linting.

---

# E. Review proper-noun / place-name / foreign-name normalization rules

The planned principle is:

> Use the English patch's canonical spelling to establish the identity/original form of a name, use the Japanese source to recover the intended game pronunciation, and then choose a natural Korean transliteration. Do not mechanically transliterate every Japanese mora.

Review the following proposed subrules:

1. **Priority order**
   - English-patch canonical name/spelling for identity;
   - Japanese source for intended pronunciation;
   - natural Korean transliteration for final display;
   - existing Korean usage as consistency evidence;
   - raw kana-to-Hangul mapping only as a last clue.

2. **Japanese long vowels**
   - do not mechanically render `ー`, `オウ`, `オオ`, etc. as doubled Korean vowels;
   - re-evaluate against English canonical spelling and intended pronunciation;
   - known candidate: `ロストール / Rostorl`, currently proposed normalization `로스토올 -> 로스톨`.

3. **Japanese epenthetic vowels / consonant clusters**
   - do not automatically preserve Japanese support vowels inserted because Japanese cannot represent certain consonant clusters/codas;
   - reconstruct the canonical name first, then choose natural Korean syllabification.

4. **Sokuon `ッ`**
   - do not mechanically map every `ッ` to a Korean tense consonant;
   - use canonical spelling and intended pronunciation.

5. **Japanese ラ行 / final ル**
   - do not blindly map every `ル` to `루` where the canonical foreign name ends in a consonant or Korean coda is more natural;
   - R/L identity should be informed by canonical spelling, not kana alone.

6. **V/B/F and similar distinctions**
   - recover distinctions from canonical spelling when Japanese kana has collapsed them.

7. **シ/ジ/チ/ツ and contracted sounds**
   - do not assume kana alone uniquely determines the original consonant/phoneme;
   - use English spelling + Japanese pronunciation together.

8. **No English-spelling-only pronunciation**
   - English is a canonical identity reference, not automatically an instruction to pronounce the name as contemporary English.

9. **No blind global normalization**
   - collect a table such as `Japanese | English canonical | current Korean variants | proposed canonical Korean | evidence | status`;
   - classify as `KEEP`, `NORMALIZE`, or `REVIEW`;
   - only confirmed items may be globally normalized.

10. **Whole-project consistency**
    - once a canonical Korean name is confirmed, check dialogue, UI, maps, chronicle, quests, descriptions and other applicable consumers rather than changing only one occurrence.

Please judge these as a localization/engineering policy, not as a generic Korean-language opinion. Flag names for which the English patch itself may be invented/ambiguous and therefore insufficient to establish pronunciation.

Also inspect the actual Ferme/`フェルム` Korean variants (e.g. `페름`/`펠름` if present) and explain what evidence would be required before standardizing them. Do not guess a canonical Korean form without evidence.

---

# F. Review candidate discovery and editing workflow

Proposed workflow:

1. Freeze branch/base/record identity before editing.
2. Scan the entire Korean corpus for candidates.
3. Automated scanning is **candidate discovery only**, not blind rewriting.
4. For each real edit, inspect as needed:
   - Korean;
   - Japanese source;
   - English patch;
   - adjacent lines / scene context;
   - consumer/layout constraints.
5. Classify candidate as:
   - simple copyedit;
   - contextual rewrite;
   - terminology normalization;
   - layout/reflow defect;
   - hold/uncertain.
6. Work in reviewable section-sized batches rather than one giant rewrite.
7. Validate each batch before progressing.

Check whether this is sufficient and whether any categories require a different transaction/edit API.

---

# G. Review full reflow / line-break coverage strategy

A key motivation is that an early tutorial Ferme dialogue was observed at runtime without expected line breaking despite earlier statements that reflow had been broadly audited.

We therefore plan to distinguish:

- static full-corpus audit;
- contextual full-corpus copyedit;
- reflow-eligibility/coverage audit;
- sampled runtime QA;
- true full-path runtime QA.

Do not allow "full audit" to mean a few representative samples.

Review the proposed coverage principle:

1. Derive the **expected English reflow consumer/population set**.
2. Derive the corresponding **Korean eligible/reflow population set**.
3. Compare those sets explicitly.
4. Explain every set difference as `PASS`, `DIFFERENT-BY-DESIGN`, `MISSING`, or `UNKNOWN`.
5. Separately enumerate records excluded due to runtime-width-unknown substitutions or other special cases.
6. Confirm that excluded records still have a proven path to safe display/layout.
7. Inspect records whose semantic Korean changed but whose layout was not rederived.
8. Treat known failures/anomalies as permanent regression tests where feasible.

Specifically investigate ID `1980005` with `<value:$28>`:

- determine its English consumer/reflow behavior;
- determine its Korean consumer/reflow behavior;
- determine whether its observed no-wrap state is caused by reflow eligibility, movable substitution handling, missing consumer mapping, stale layout, build-materialization behavior, or something else;
- do **not** recommend manual Korean `<line-break>` insertion unless the English/consumer contract proves that authored breaks are the correct mechanism.

Also verify the branch's existing `280181` fixed select-count regression and whether its fix suggests other untested classes.

---

# H. Review reflow algorithm parity, not just population parity

For eligible dialogue, verify whether Korean uses the same relevant English concepts:

- source-derived layout hints;
- preferred vs greedy break selection;
- consumer-specific width limits;
- portrait/narrow-text handling;
- preservation of semantic/control topology;
- final width/materialization validation;
- behavior when retail projection/token data is available;
- repository/asset-free fallback behavior;
- fail-closed behavior when final projection cannot be proven.

Identify any algorithmic divergence that could produce a Korean-only overflow/freeze or silently different pagination.

---

# I. Review validation gates after semantic edits

Planned minimum gates:

- record ID set unchanged;
- Japanese text unchanged byte-for-byte;
- fixed runtime control sequence preserved;
- substitutions preserved;
- TOML parses;
- no improper authored `<line-break>` added to semantic Korean;
- Korean semantic/control-contract validation passes;
- stale touched layouts are invalidated/regenerated where applicable;
- Korean reflow/source-aware tests pass;
- consumer-specific limits are checked with final encoded/materialized bytes where applicable;
- `go test ./...`;
- `go vet ./...`;
- `./zill check`;
- asset-backed build validation for release confidence;
- runtime QA for representative/high-risk paths.

Review:

1. Which of these gates genuinely exist today?
2. Which are only documentation promises?
3. Which important invariants have no automated gate yet?
4. Which gates should be added **before** mass editing?
5. Does `./zill check` omit any critical retail/reflow/fixed-buffer validation, as currently documented?
6. Which tests are redundant versus complementary?

---

# J. Review evidence / audit trail requirements

We want a mass-edit process that can prove what it did.

Proposed per-edit/batch evidence may include:

`ID | old Korean | new Korean | category | reason | JP checked | EN checked | controls safe | layout invalidated/rederived | validation result`

For reflow coverage, require counts/sets rather than prose claims.

Review whether this is sufficient. Recommend a smaller or better evidence format if appropriate.

The process must never claim:

- "runtime full QA" when only static checks ran;
- "reflow complete" when only representative records were sampled;
- "all dialogue reviewed" when only candidate-linted records were manually read.

Suggest precise vocabulary for each coverage level if the current protocol is ambiguous.

---

# K. Review runtime QA strategy

The final patch cannot rely solely on static Unicode/source checks.

Review what runtime QA should minimally cover after the mass pass, including:

- early tutorial / initial routes;
- portrait dialogue;
- narrow dialogue;
- runtime name/value substitutions;
- select/fixed-control cases;
- long multi-page dialogue;
- known previously failing IDs;
- boundary-width cases;
- any consumer where final encoded byte size matters.

Distinguish what can be proven statically from what genuinely requires asset-backed/materialized or emulator/device runtime testing.

---

# L. Review workflow / branch / handoff discipline

Current working constraints:

- work only on `fix/u6-korean-reflow-parity`;
- do not modify `milestone/U6` directly;
- do not modify `translation/section001-batch2`;
- do not touch unrelated U7/debug/forensic/experiment/improve branches;
- no force-push;
- no merge into `milestone/U6` as part of this task.

Review whether the QA documentation should additionally require:

- baseline commit recording before each batch;
- changed-record manifest;
- validation result recording;
- explicit rollback point;
- handoff notes if work stops mid-batch;
- prohibition on leaving partially edited semantic/layout state committed.

---

# M. Find anything we missed

Do not limit yourself to the questions above.

Search for **any additional failure mode** relevant to this planned mass pass, including:

- encoding/glyph coverage;
- TOML escaping;
- duplicate IDs;
- stale source hashes/authenticated Japanese matching;
- storage/bank limits;
- page/record byte limits;
- UI or non-dialogue consumers accidentally included in dialogue normalization;
- choice-text constraints;
- untested control tokens;
- build provenance / using wrong branch or wrong generated asset;
- cases where a visually acceptable line could still violate final encoded storage constraints.

If such issues already have sufficient gates, say so. If not, recommend the minimal precondition needed before editing.

---

# Required questions Claude must answer

1. **Overall verdict:** Is the complete planned workflow safe enough to begin? `PASS`, `PASS WITH CHANGES`, or `FAIL`.
2. Is the **English-patch-first** premise correctly scoped and implemented?
3. Are the **copyedit rules** safe and appropriate?
4. Are the **proper-noun/transliteration rules** technically and linguistically defensible for this project?
5. Are the **semantic/layout separation rules** correct for all relevant consumers?
6. Does current Korean reflow have **population parity** with the relevant English reflow path?
7. Does it have sufficient **algorithm/consumer-contract parity**?
8. What is the most likely code-grounded explanation for the known `1980005`/Ferme no-wrap observation, and what evidence is still needed?
9. Are dynamic/movable value-tag cases sufficiently handled and tested?
10. Are fixed-control/select cases sufficiently handled and tested?
11. What **hard invariants/gates are missing** before mass editing?
12. What is safe as automatic validation, and what must remain candidate/manual review only?
13. Is our claimed notion of "full audit" measurable and honest?
14. Is the runtime QA plan sufficient?
15. Are there branch/workflow/handoff risks that should be added to the protocol?
16. What minimal rule/code/test changes are required **before GO**?

---

# Required response format

Return the review in this structure so it can be relayed back to ChatGPT verbatim:

```text
VERDICT: PASS | PASS WITH CHANGES | FAIL

MASS-PASS GO/NO-GO
- GO | GO AFTER FIXES | NO-GO
- Preconditions: ...

CRITICAL ISSUES
- ...

HIGH-RISK GAPS
- ...

RULES TO CHANGE
- Rule / wording:
  Problem:
  Evidence:
  Recommended replacement:

MISSING SAFETY RULES / GATES
- ...

RULES CONFIRMED CORRECT
- ...

ENGLISH-PATCH PARITY FINDINGS
- Population parity:
- Algorithm/consumer parity:
- Known divergences:
- UNKNOWN items:

REFLOW / LINE-BREAK FINDINGS
- 1980005 / Ferme:
- Dynamic substitution cases:
- Fixed/select control cases:
- Other uncovered classes:

TERMINOLOGY / TRANSLITERATION FINDINGS
- Rules confirmed:
- Rules to revise:
- Names requiring REVIEW rather than normalization:

COPYEDIT QA FINDINGS
- Safe candidate linting:
- Safe hard validators:
- Manual-only decisions:

VALIDATION COVERAGE
- Existing automated gates:
- Missing automated gates:
- Asset-backed requirements:
- Runtime-only requirements:

WORKFLOW / EVIDENCE FINDINGS
- ...

MINIMUM CHANGES BEFORE START
1. ...
2. ...

FINAL RECOMMENDATION
- ...
```

For every material finding, cite repository file/function/test evidence where possible.

Distinguish clearly between:

- `PROVEN` — supported by code/test/data evidence;
- `LIKELY` — strong evidence but incomplete proof;
- `UNKNOWN` — repository evidence insufficient.

Do not invent project behavior. If a claim cannot be verified from the repository, mark it `UNKNOWN` and state what evidence/test would resolve it.
