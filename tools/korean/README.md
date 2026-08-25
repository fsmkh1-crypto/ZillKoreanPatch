# Bulk Korean translation workflow

This path avoids repeatedly copying/fetching large canonical TOML records while translating.

## 1. Export compact source

```sh
python3 tools/korean/export-corpus.py --section 3 > /tmp/sec003.jsonl
# or export the complete corpus
python3 tools/korean/export-corpus.py > /tmp/all.jsonl
```

Each row contains only `section`, `id`, and canonical `japanese`.

## 2. Translate in large batches

Produce JSONL containing only the identity and result:

```json
{"section":3,"id":"30000","korean":"...control tokens preserved..."}
```

Do not translate, remove, reorder, or invent runtime control tokens.

## 3. Import and fail closed

```sh
python3 tools/korean/import-translations.py /tmp/batch.jsonl --out /tmp/korean-import
```

The importer looks canonical Japanese up locally and rejects unknown IDs, duplicate IDs,
empty Korean strings, and ordered control-token mismatches before emitting TOML. This means
the translation payload no longer needs to carry a duplicate copy of Japanese text.

The generated TOML retains `japanese` because the current project loader uses it for canonical
source validation. Once imported into the repository, run the normal Korean checks before merge.
