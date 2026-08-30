# A-074 — Practical success baseline and next localization scope

Date: 2026-08-31 (KST)

## Evidence status

A practical successful run has been reported against the last green Android/CI build at commit `98c74d3715e1575f9d25f466101048067a960452`.

Per project evidence policy, this is recorded strictly as **one non-reproduction of the prior freeze in that run**. It is not proof that the build is globally safe and it does not establish any particular hypothesis as root cause.

The APK associated with that baseline was built from the same commit and previously verified to contain payload version `98c74d3715e1575f9d25f466101048067a960452` with its embedded manifest passing `sha256sum -c`.

Subsequent commits on `audit/english-first-full-parity-restart` strengthen static/parity auditing and must not overwrite the interpretation of the practical evidence above. If a later build regresses, `98c74d...` remains the practical comparison baseline.

## Baseline interpretation

- SUCCESS once = freeze not reproduced in that execution only.
- Any later freeze/crash = strong failure evidence for the tested later build.
- Same visible screen does not by itself establish the same machine-state cause.
- Do not fold future localization changes into the historical success claim.

## Next localization scope

The next work intentionally moves from freeze containment into completeness/usability. The requested areas are:

1. Early-game/environment settings and other untranslated UI/system text.
2. `흑황적` tutorial/help material and other tutorial/help regions that still remain untranslated.
3. The incomplete English-oriented text/input window path, including character entry/input rendering and any fixed UI labels around it.

These areas must be treated as separate consumer families until their actual storage/render/input contracts are established from the upstream English patch and authenticated retail assets.

## Change isolation policy

Do not change all three scopes in one experiment. Use independent checkpoints:

- **U1 — settings/system untranslated text**
- **U2 — tutorial/help untranslated text**
- **U3 — input-window/text-entry path**

For each scope:

1. Identify exact source records/fixed-data fields and their runtime consumer.
2. Compare the upstream English implementation first.
3. Reuse the English storage/visual/input contract where applicable.
4. Add static or asset-backed validation before modifying Korean data.
5. Build and run the full English-first contract chain.
6. Test the affected screen separately.
7. Record a successful run only as non-reproduction; record any freeze/crash as strong failure evidence.

## Regression risk assessment

Completing untranslated text can reintroduce failures if a newly translated field belongs to a fixed-size or special-purpose consumer that was previously left in retail Japanese/English bytes. The main risks are:

- fixed-width/fixed-buffer UI fields,
- non-message fixed strings in EBOOT/BINDATA/other assets,
- special tutorial/help layouts with narrower visual boxes,
- input-method or text-entry tables whose byte semantics differ from ordinary message rendering,
- new Korean glyph requirements changing the slot plan,
- dynamic substitutions or controls in newly covered records.

Therefore untranslated completion is not considered a pure text-only change until consumer classification says it is.

The input window is the highest-risk of the three because it may combine rendering, key tables, cursor/indexing, allowable-character tables and fixed executable UI text. It should be audited as an input subsystem rather than treated as another ordinary message bank.

## Rollback/reference point

Practical comparison baseline: `98c74d3715e1575f9d25f466101048067a960452`

Current audit work continues on `audit/english-first-full-parity-restart`; any localization expansion should remain separable from this baseline by commit history and, preferably, by one scope per commit/branch checkpoint.
