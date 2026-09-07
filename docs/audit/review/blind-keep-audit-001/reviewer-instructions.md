# Blind KEEP audit reviewer instructions

Audit ID: `BA-KEEP-001`

You are performing an independent second-pass accuracy audit of previously accepted Korean KEEP rows. Do not search the repository for original IDs, batch membership, prior decisions, or suspected weak areas. Use only the supplied packet.

For every anonymous item:

1. Review only the row marked **TARGET**. Nearby rows are context only.
2. Treat Japanese as semantic authority.
3. Use pinned English as evidence for meaning, structure, and the established patch's solution, but do not translate English mechanically.
4. Return `KEEP` only if the current Korean target needs no correction.
5. Return `SHOULD_EDIT` if you would change the target for any material reason.
6. If `SHOULD_EDIT`, assign one or more error classes from: `semantic`, `speaker_register`, `naturalness`, `punctuation_spacing`.
7. If `SHOULD_EDIT`, provide the full corrected Korean target string, preserving required control tags exactly.
8. Give a short reason tied to Japanese meaning/context and, where useful, the pinned English handling.

Error-class definitions:

- `semantic`: mistranslation, wrong subject/object, polarity, fact, nuance, omission/addition, or other meaning error.
- `speaker_register`: speaker voice, politeness level, relationship-dependent address, honorific/register, or character-consistency error.
- `naturalness`: meaning is essentially correct but the Korean is materially awkward, calqued, or unnatural enough to warrant an edit.
- `punctuation_spacing`: punctuation, spacing, or typography-only correction.

Do not apply any pass/fail threshold yourself. Judge each item independently. Fill the provided CSV template without reordering or deleting rows.
