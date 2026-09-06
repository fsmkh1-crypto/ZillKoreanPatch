# Beta1 Review Ledger Policy

## Purpose

`docs/audit/beta1-review-ledger.json` is the authoritative progress ledger for language/context review coverage in Beta1. It is generated, not hand-edited.

## Statuses

- `AUTO_ONLY`: accepted Korean record with automated/static QA coverage, but no proven full contextual review in the ledger.
- `CONTEXT_KEEP`: Japanese, pinned English patch, and current Korean were explicitly reviewed and no edit was needed at that review point.
- `CONTEXT_EDIT`: the same contextual review occurred and an approved copyedit was applied.

Runtime verification and special UI/title audits are orthogonal evidence and should remain separate flags/audits rather than replacing the language-review status.

## Counting rule

Discovery is not review. Regex hits, statistical scans, automated lint, layout QA, and candidate generation do not promote an ID to `CONTEXT_KEEP` by themselves.

Historical KEEP coverage is backfilled only where an auditable scope proves that every ID in that scope was actually read in context. This intentionally understates old coverage rather than manufacturing progress.

Approved contextual-copyedit manifests do prove contextual review for their edited IDs and therefore promote those IDs to `CONTEXT_EDIT`.

## English-patch baseline

For semantic/naturalness decisions, review Japanese source meaning first and then inspect how the pinned English patch handled the same construction and why. Korean should adopt the useful structural reasoning, not blindly translate English wording.

Pinned English reference for current Beta1 work:

`a98d9ce29f361d666ec23da0dcfd351f24537ffd`

## Going forward

Each sequential full-review batch should record both:

1. IDs reviewed and kept unchanged.
2. IDs reviewed and edited through the normal approved manifest/queue flow.

The generated coverage summary can then report exact unique whole-corpus contextual-review progress against the accepted Korean ID population.
