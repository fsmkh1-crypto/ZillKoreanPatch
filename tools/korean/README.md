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

## 3. Apply results and fail closed

```sh
python3 tools/korean/apply-results.py /tmp/batch.jsonl
```

`apply-results.py` looks canonical Japanese up locally and rejects unknown IDs,
empty Korean strings, semantic `<line-break>` markers, and fixed control-token mismatches.
Ordinary conflicting translations remain fail-closed. When multiple ordinary conflicts
exist in one batch, all conflicting section/IDs are reported together and **no overlay
files are written**.

## 4. Recover a discarded historical result packet

Historical recovery is only for translations that were already produced but whose
result packet was later discarded because another process filled some of the same IDs.
Create a manifest row that points at the old result packet in Git history:

```json
{"recover_from":{"commit":"0123456789abcdef0123456789abcdef01234567","path":"work/korean-results/old.jsonl"}}
```

Recovery is intentionally narrower than ordinary apply:

- `commit` must be a full 40-character hexadecimal commit SHA.
- `path` must be a `.jsonl` file below `work/korean-results/`.
- Nested recovery manifests are rejected.
- Canonical Japanese and fixed control tokens are revalidated against the **current** corpus.
- An existing current Korean translation is never overwritten. A differing historical
  translation is skipped and logged with section/id, reason, and historical source.
- Missing IDs are restored normally.

This makes recovery append-only with respect to the current Korean corpus while keeping
ordinary translation application strict.

## 5. CI / maintenance behavior

The Korean apply workflow may run when translation tooling itself changes. Maintenance-only
runs still validate the corpus, refresh the packet, and keep the raster catalog synchronized,
but they do **not** run automatic TM/trivial translation mutation unless an actual
`work/korean-results/*.jsonl` packet is being processed.

The generated TOML retains `japanese` because the current project loader uses it for canonical
source validation. Run the normal Korean checks before merge.
