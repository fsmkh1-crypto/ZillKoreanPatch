# A-072 — Dynamic substitution consumer closure

## Scope

This audit closes the static classification gap around runtime `<value:$XX>` tokens without promoting them to a universal freeze cause.

The accepted Korean corpus currently contains:

- 42,028 Korean records
- 4,980 records containing `<value:$XX>`
- 10,123 value-tag occurrences
- 18 distinct value opcodes

The role-aware audit classifies:

- 6,566 inline/rendered occurrences
- 3,272 predicate occurrences
- 285 selector occurrences
- 5,742 inline non-space adjacencies

The comparison-RHS blind spot is now explicitly handled. `<value>` immediately following `<equal>`, `<less-equal>`, or `<greater-equal>` is classified as a predicate operand rather than rendered inline text. CI fails if such a comparison RHS is still misclassified. Current counterexamples: **0**.

## Consumer mapping result

Known English-patcher bounded/storage consumers remain covered by the release validator and are now named consistently in both dynamic-substitution audit tools:

- C5
- C20
- C22
- bounded labels
- guild client/region
- trap
- character-creation choices
- chronicle entries
- category-backed equipment-feedback / guild-posting labels when present

The latest corpus report still contains 1,830 records with inline `<value>` usage that are labelled `unmapped-by-audited-fixed-consumers`. This label does **not** mean a missing engine contract. It means only that those records are not members of the currently evidenced fixed/bounded consumer sets. Samples are overwhelmingly ordinary dialogue, notifications, choices, prompts, and quest text.

Therefore this audit does **not** invent a synthetic maximum for every inline value token. Doing so would repeat the earlier mistake of applying one consumer-specific NUL-scanner invariant universally.

## Freeze-adjacent record 10010

ID 10010 remains a useful reporting anchor:

- category: `character-creation-prompt`
- inline opcode: `$15`
- it is not a member of the English patcher's `character-creation-choice` fixed-width consumer contract

This means the English-first static audit does not justify assigning the 31-byte character-choice buffer to 10010. A new bound for this prompt would require independent engine-consumer evidence.

The same rule applies to other ordinary inline `$15/$16/$17/$1A/$1B/$28` uses: their presence or adjacency alone is not evidence of a bounded consumer.

## Closure statement

For the known English-patcher bounded/fixed consumer surface, the dynamic-substitution census found no newly exposed missing consumer contract.

This is a static classification result only. Runtime-provided substitution values can still affect concrete rendered length, and a later freeze remains strong failure evidence.

## CI evidence

Commit `7bdb4c66b9da6bc3bd3b16de40058e6586f0911a` passed:

- project-premise enforcement
- full English-first parity audit
- entrypoint closure audit
- Go tests and vet
- Python compile/unit tests
- runtime-text audit
- role-aware dynamic-substitution audit
- aligned inline-opcode consumer audit
- `zill check`
- `zill korean-check`
- `zill korean-font-check`
- whole-corpus Korean QA workflow
