# Gemini translation exchange

This project keeps the repository/TOML corpus under maintainer control and uses Gemini only as an external translation worker. Gemini does not need GitHub access.

## Contract

Source batches use JSON Lines (`.jsonl`) with one object per line. The current schema is `zill-gemini-v1`.

Generate a 100-record batch:

```text
./zill gemini-export --batch 0 --out work/gemini/batch-000.source.jsonl
```

The default batch size is 100. `--count` may be set from 1 through 150. Batch N starts at project offset `N * count`, so keep the same count when advancing sequentially.

One exported row looks like:

```json
{"schema":"zill-gemini-v1","id":1234,"section":0,"record_index":1234,"source_file":"translations/messages/msgsec000.toml","japanese":"...<end>","english_reference":"...<end>","speaker":null,"context":null}
```

`speaker` and `context` are deliberately `null` unless the project has source-backed evidence. Adjacent storage records are not assumed to be dialogue chronology, and an LLM must not invent speaker identity from file order.

Gemini must return exactly one raw JSON object per source row, in the same order, with no Markdown fences or prose:

```json
{"id":1234,"korean":"...<end>","uncertain":false,"note":"","glossary_candidates":[]}
```

`glossary_candidates` entries use:

```json
{"source":"ヴァン","korean":"반","type":"character"}
```

Candidates are suggestions only. They are not automatically promoted into the canonical glossary.

## Translation instructions for Gemini

- Japanese is the authoritative source. English is context/reference only; do not translate from English as the primary text.
- Preserve every record ID and output order exactly.
- Never add or omit records.
- Preserve machine-readable controls exactly, including angle-bracket tokens such as `<end>`, `<if>`, `<select>`, `<value:$28>`, double-brace placeholders, and printf tokens such as `%s`.
- Return natural Korean JRPG localization rather than literal Japanese syntax, but do not invent or delete meaning.
- If interpretation, speaker tone, terminology, or a proper noun is uncertain, set `uncertain` to true and explain briefly in `note`.
- Suggest newly discovered proper nouns/terms in `glossary_candidates`; do not silently choose inconsistent spellings across batches.
- Do not constrain Korean vocabulary to the current temporary font PoC. The final Korean character set is measured from the completed translation and the font is built to match it.
- Output raw JSONL only. No code fences, headings, comments, blank lines, or extra keys.

## Validate returned data

Save Gemini's raw response as a `.jsonl` file, then run:

```text
./zill gemini-check \
  --input work/gemini/batch-000.source.jsonl \
  --result work/gemini/batch-000.gemini.jsonl \
  --out work/gemini/batch-000.accepted.jsonl
```

The checker rejects the batch before producing an accepted output if any of these fail:

- invalid JSON or Markdown/prose outside JSON objects
- missing or unknown response fields
- `glossary_candidates: null` instead of an array
- duplicate IDs
- changed record count
- changed ID order
- empty Korean text
- changed angle-bracket control-token sequence
- changed double-brace placeholder sequence
- changed printf-token sequence
- malformed glossary candidates

The checker emits warnings, without rejecting the batch, for:

- records explicitly marked `uncertain`
- Korean results that still contain hiragana or katakana

An accepted file is normalized JSONL. It is still **staging data**, not canonical Korean TOML. Canonical Korean storage/import is a separate step so external model output can never directly overwrite the contributor corpus.

## Recommended batch workflow

1. GPT/maintainer exports 100 records.
2. The user gives the `.source.jsonl` file and the translation instructions above to Gemini.
3. Gemini returns raw JSONL.
4. The user returns that raw result to GPT/maintainer.
5. `gemini-check` validates and normalizes it.
6. Claude/reviewer performs translation QA on the accepted batch and flags terminology/tone issues.
7. Only reviewed data is imported into the canonical Korean corpus.
8. The next batch is exported.

Start at 100 records. Raise to 150 only after multiple batches complete without truncation, missing rows, or JSON corruption.
