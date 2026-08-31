# U5 full-corpus visual census

U5 remains provisional until the repository-only 42,016-row visual census executes the same asset-independent derivation order as the production Korean contract chain.

The first U5 RC (`49528fd869e06817e17e636445929721ca6b92c7`) added verified dialogue/in-world-guidance reflow to production, but `TestCurrentKoreanCorpusEnglishConsumerStorageContracts` still skipped that derivation. Its warning totals therefore described the pre-U5 layout path and must not be used as the final post-U5 overflow census.

Closure requirements:

- canonical Korean rows: 42,016 checked;
- repository census includes fixed-consumer -> verified-dialogue -> hard-visual derivation in production order;
- known verified dialogue/in-world-guidance rows with no authored layout must be renderer-width safe after derivation;
- remaining `line_exceeds_authoring_ceiling` warnings are classified rather than silently equated with known dialogue-window overflow;
- standard CI and Android asset-backed RC gates remain green;
- only then is `milestone/U5` moved from the provisional RC to the final U5 functional commit.

U0-U4 remain immutable.
