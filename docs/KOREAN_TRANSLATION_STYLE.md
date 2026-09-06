# Korean translation style

This file records the default Korean dialogue, register, terminology, and proper-name policy.

It is a language/style companion to `AGENTS.md`, `docs/KOREAN_DIALOGUE_QA_PROTOCOL.md`, and the accepted review decisions in `docs/CLAUDE_FULL_REVIEW_DECISION_2026-09-06.md`. Engine-facing behavior always follows the English-patch-first contract in `AGENTS.md`.

## 1. Register

- Japanese plain/casual forms should normally become Korean banmal. Do not defer a row solely because the speaker identity is unknown when the Japanese morphology already makes the register clear.
- Examples: `なんだ`, `待て`, `任せて`, `誰だ` are casual/plain. Translate them as banmal by default.
- Explicit polite Japanese such as `です` / `ます` should normally become Korean polite speech unless scene context clearly requires otherwise.
- Honorifics, titles, archaic speech, military commands, and strongly character-specific verbal tics should preserve their marked tone.

## 2. Context versus register

Register and lexical meaning are separate questions. A plain form can have obvious banmal register while the best Korean verb still depends on context. For example, `待て` is casual either way, but a scene may favor `기다려` or `멈춰`. Resolve that from nearby context rather than postponing the row merely to decide politeness.

Do not smooth every speaker into the same neutral modern Korean. Preserve deliberate stutter, hesitation, repetition, catchphrases, roughness, formality, childishness, archaic diction, military tone, and other characterization when the source supports it.

## 3. Korean copy-editing defaults

The default review target is full proofreading plus minimal polishing, not wholesale retranslation.

Prefer to fix:

- incorrect spacing and orthography
- awkward particles or connective endings
- unnecessary commas copied from Japanese/English pause structure
- literal source-language word order
- redundant subjects/objects obvious from context
- excessive nominalization
- accidental repeated wording
- grammatical but clearly unnatural Korean dialogue

Do not globally delete commas or punctuation. Keep punctuation when it prevents ambiguity, marks a list/vocative/interjection, or belongs to deliberate character rhythm.

## 4. Proper-name and terminology source hierarchy

For character names, place names, countries, organizations, races, titles, items, skills, and other named terms, determine the Korean form in this order:

1. **English patch canonical spelling** — primary reference for the identity/spelling of the proper noun used by this project.
2. **Japanese source form** — primary pronunciation clue for how the creators rendered the name in Japanese.
3. **Natural Korean transliteration** — final Korean spelling should be readable and natural rather than a mechanical kana-to-Hangul conversion.
4. **Existing accepted Korean terminology/corpus** — preserve a well-established project spelling when it is consistent and not demonstrably wrong.
5. **Mechanical kana correspondence** — only a last-resort clue, never the deciding rule by itself.

`English patch first` does **not** mean `pronounce the Roman letters as English`. The English spelling identifies the name; the Japanese source helps recover intended pronunciation. Use both before deciding the Korean form.

Do not normalize a proper name on intuition alone.

## 5. Long vowels

Japanese long-vowel notation (`ー`, `オウ`, `オオ`, `ウウ`, and similar forms) is pronunciation information, not an instruction to duplicate Korean vowels mechanically.

Avoid mechanical forms such as:

- `オー -> 오오`
- `ウー -> 우우`
- `エー -> 에에`
- `アー -> 아아`

Reconstruct the name from the English spelling plus Japanese pronunciation and choose the natural Korean form.

Approved project target pending repository-wide migration:

- `ロストール / Rostorl -> 로스톨`
- `로스토올` is the legacy corpus/canonical-table form and is not the intended final target.

This normalization is **not complete** until `translations/terminology/korean-canonical.toml`, every proven same-entity user-visible occurrence, and the residual terminology audit all agree. Until that migration batch is complete, do not report Rostorl terminology consistency as finished merely because this style document names the target.

A long vowel may be reflected when it is genuinely part of the intended name/pronunciation, but it must be justified from the name rather than copied from the kana mark.

## 6. Japanese epenthetic vowels

Japanese often inserts vowels to represent consonant clusters or word-final consonants that Japanese phonotactics cannot represent directly. Do not automatically carry those vowels into Korean.

Pay special attention to kana such as:

- `ス / ズ`
- `ト / ド`
- `ク / グ`
- `ル`

These do not automatically require Korean `스 / 즈 / 토 / 도 / 쿠 / 구 / 루` when the English spelling shows that the vowel is only Japanese phonological support.

Use Korean final consonants and natural loanword structure where appropriate rather than making names unnecessarily longer.

## 7. Sokuon (`ッ`)

Do not map Japanese `ッ` mechanically to a Korean tense consonant.

Determine:

1. the English spelling/consonant structure,
2. the Japanese pronunciation clue,
3. the natural Korean realization.

`ッ` is evidence of consonant closure/gemination, not a one-to-one instruction for `ㄲ/ㄸ/ㅃ/ㅆ/ㅉ`.

## 8. Japanese r-row and word-final consonants

Do not automatically render `ラ/リ/ル/レ/ロ` as `라/리/루/레/로` without checking the canonical spelling.

In particular, word-final `ル` may be Japanese support for a final consonant rather than an intended Korean `루`. Check whether the canonical name ends in `r`, `l`, or another consonant structure and choose the natural Korean form.

The same principle applies to other Japanese word-final support vowels: remove them when the underlying name and natural Korean pronunciation support doing so; do not force an unnatural Korean batchim merely to shorten the name.

## 9. B/V/F, J/G/Z and other collapsed distinctions

Japanese spelling may collapse distinctions that are visible in the English canonical spelling.

For kana such as `ヴ`, `バ`, `ファ`, `ジ`, `シ`, `チ`, `ツ`, and related forms:

- use the English spelling to recover the underlying consonant identity,
- use the Japanese form as a pronunciation clue,
- choose the natural Korean result.

Do not assume a single fixed Hangul output for each kana.

## 10. Yoon combinations and unusual fantasy spellings

For combinations such as `キャ/キュ/キョ`, `リャ/リュ/リョ`, and unusual fantasy-name spellings:

- do not transliterate every kana mora mechanically,
- compare the English canonical spelling,
- compare repeated occurrences and surrounding terminology,
- place the name in `REVIEW` when pronunciation cannot be established confidently.

Do not invent a pronunciation from Roman letters alone when the Japanese source does not support it clearly.

## 11. Proper-name decision classes

Every proper-name audit candidate should be classified as:

- `KEEP` — current Korean spelling is natural, consistent, and supported.
- `NORMALIZE` — current spelling has a clear long-vowel/epenthetic-vowel/inconsistency problem and the replacement is well supported.
- `REVIEW` — pronunciation or canonical Korean form remains uncertain.

Only `NORMALIZE` items may proceed to corpus-wide replacement after the mapping is explicitly approved/reviewed.

Before any terminology-wide replacement, build a mapping containing at least:

| Japanese | English | Current Korean variant(s) | Proposed Korean | Decision | Evidence |
| --- | --- | --- | --- | --- | --- |

Do not use blind text replacement. First enumerate all affected records and confirm that every occurrence refers to the same entity/term and that no control or unrelated substring can be altered.

`translations/terminology/korean-canonical.toml` is a canonical-only table, not a review queue. Do not insert unresolved `REVIEW` candidates into that file unless its schema/tooling is deliberately extended to represent review state.

Current required review candidate:

- `フェルム / Ferme`: Korean corpus contains both `페름` and `펠름`. Keep this `REVIEW`; do not globally normalize either form until same-entity surface enumeration and stronger pronunciation/provenance evidence establish the canonical Korean spelling. `Pelm` is not an English canonical spelling and must not be used as if it were English-patch evidence.

## 12. Scope of terminology consistency

Once a canonical Korean term is accepted, audit all applicable surfaces, not dialogue alone:

- dialogue
- choices/prompts
- map/place labels
- character labels
- chronicle/objective text
- items/equipment/skills
- system/help text
- terminology tables and other user-visible strings

Do not change immutable Japanese source text.

## 13. Runtime substitution / control classification

Preserve all runtime substitutions and controls exactly, but do not treat every `<value:$XX>` occurrence as the same visual/layout class.

For layout/reflow review distinguish at least:

- `BOUNDED_INLINE` — rendered inline and a worst-case advance is proven by the engine/game contract; `$28` player name is the current confirmed example.
- `UNBOUNDED_INLINE` — rendered inline but no trustworthy worst-case advance is proven in the current path.
- `CONTROL_FLOW_OPERAND` — part of expression/select/predicate grammar rather than ordinary rendered inline text.
- `FIXED_CONTROL` — source-owned fixed control topology that must remain fixed and ordered.

Use authenticated token/projection grammar where available. Regex presence alone is not enough to decide inline-vs-control-flow semantics.

Do not promote `$15/$16/$17/$1A/$1B/$24/$25/$2B` to bounded merely because they are movable. Each needs its own proven measurement/binding contract.

## 14. Bulk workflow safety

- Preserve runtime control tokens exactly.
- Preserve runtime substitutions such as `<value:$XX>` and printf-style substitutions exactly.
- Semantic Korean text must not contain build-owned `<line-break>` tokens; layout owns wrapping.
- Keep terminology and proper-name spellings consistent with approved project mappings.
- If semantic Korean changes, any previous generated `layout` is stale and must be invalidated and regenerated through the build-owned path.
- Do not manually copy-edit generated layout as though it were semantic text.
- If context is genuinely insufficient to choose meaning or a proper-name spelling, leave only that row as `HOLD/REVIEW`; do not stop the whole packet and do not guess.

All bulk work must additionally follow `docs/KOREAN_DIALOGUE_QA_PROTOCOL.md` and the accepted review decisions in `docs/CLAUDE_FULL_REVIEW_DECISION_2026-09-06.md`.
