# A-061 — Dynamic substitution contract parity

Status: **audited; no new Korean-only bound introduced**

Project premise: `AGENTS.md` — established English-patch engine contracts are the primary baseline. Korean-specific behavior may only diverge where representation requires it and must preserve the underlying contract.

## Scope

Compared the retained English layout/storage validator (`internal/layout/validate.go`) with the Korean materialized-byte path (`internal/layout/validate_korean.go`, `internal/layout/c5_dynamic_audit.go`, and `internal/layout/validate_korean_english_contract.go`) for runtime `<value:$XX>` substitutions.

## Findings

| Contract / consumer | English baseline | Korean path | Classification |
| --- | --- | --- | --- |
| C5 static branch/page storage | Materialize exact stored bytes, omit runtime substitution expansion from static page payload, track that the leaf is dynamic | Same branch/page walker and same 256-byte page contract using authenticated Korean renderer bytes | `PASS` |
| C5 dynamic expansion | English static validator does not invent a runtime expansion length | Korean additionally records known expansion maxima only where independently established; currently `<value:$28>` player name = 16 encoded bytes. Other substitutions remain explicitly unknown | `DIFFERENT-BY-DESIGN` (strictly additive evidence; does not weaken English contract) |
| Chronicle `<value:$28>` | Adds `playerNameMaxEncodedBytes` for each occurrence before enforcing 764-byte payload maximum | Same maximum and same contract, with Korean literal bytes measured through renderer slots | `PASS` |
| Trap runtime values | Adds `trapValueMaxBytes` per `<value>` occurrence before enforcing the 104-byte buffer | Same runtime allowance and same capacity, with Korean literal bytes measured through renderer slots | `PASS` |
| Guild posting bindings | Adds integer or candidate-role maxima before enforcing the 316-byte posting buffer | Same role/binding model and same limits, with Korean literal bytes measured through renderer slots | `PASS` |
| Other inline `<value:$XX>` expansion maxima | No universal English bound established by the retained validator | No guessed Korean bound. Corpus tooling reports uses/consumer overlap and leaves unknown expansions as runtime-QA/evidence work | `PASS` relative to English baseline; runtime expansion size remains `UNKNOWN` |

## Important consequence

Do **not** assign guessed maximum lengths to `$15` or other inline substitutions merely because they appear near a freeze. The English path does not provide a universal expansion bound for them. Under the project-wide premise, such a Korean-only bound would be an unjustified divergence unless independent runtime/engine evidence establishes it.

Likewise, C5 static validation success is not proof that a dynamic page is runtime-safe. The English validator itself separates the exact stored/static payload from runtime expansion. The Korean path must retain that distinction.

## Existing safeguards confirmed

- Korean C5 lowering uses the same source-record projection/control flow and the same page-count/page-capacity constants as the English path, changing only literal-byte encoding to the authenticated Korean renderer mapping.
- Known player-name expansion uses the same 16-byte engine bound used by the English validator.
- Trap, chronicle, and guild-posting dynamic allowances reuse the same shared engine constants rather than Korean-specific guesses.
- Unknown substitutions remain explicitly visible in the C5 known-expansion audit instead of being silently assigned an arbitrary size.

## Remaining boundary

The semantic/runtime source and maximum expansion sizes for inline opcodes other than the already-established cases remain an evidence gap. Their presence is not itself a defect and must not be promoted to root cause. Future work should derive those bounds from the corresponding English/runtime consumer before adding any release gate.
