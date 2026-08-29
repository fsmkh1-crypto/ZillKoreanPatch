# A-025 — `$15` caller-substitution provenance and ID 10010

## Trigger

ID 10010 remains freeze-adjacent evidence and is a retained C5 consumer. Its payload begins with `<value:$15>`, while the current C5 known-expansion audit has no proven maximum for opcode `$15`.

## Hypothesis

A useful next step was to determine whether `$15` has one repository-wide semantic/storage contract that would justify a global encoded-byte maximum, or whether its value is supplied contextually by callers and therefore requires consumer/call-site-specific provenance.

## Verification

- Rechecked ID 10010 against the retained consumer map and verified category data.
- Compared the Japanese, English and Korean forms of ID 10010.
- Inspected upstream `internal/message/projection.go` and its history.
- Inspected upstream layout rules and storage-contract documentation.
- Rechecked the commit that introduced explicit fixed-buffer byte limits.

## Result

### Confirmed repository facts

- ID 10010 is in the retained C5 consumer set and is categorized as `character-creation-prompt`.
- The source is `<value:$15>よ...`; upstream English renders the same substitution as `<value:$15>, I ask...`; Korean retains it as `<value:$15>여...`.
- Upstream message projection explicitly separates substitution opcodes into two groups:
  - `pureMovable`: `$24`, `$25`, `$28`, `$2B`
  - `callerMovable`: `$15`, `$16`, `$17`, `$1A`, `$1B`
- Commit `63c19ea22e987fa46b71f10161828ca4af7a5251` is specifically titled `Allow caller substitutions to move` and includes `$15` in the caller-substitution path.
- Upstream fixed-buffer rules/documentation provide an independently documented 16-byte bound for player-name substitutions, which supports `$28`; they do not provide an equivalent global maximum for `$15`.
- The explicit fixed-buffer work in commit `36edf1a82e401dd3e34724263b3b20e538d9c9a0` added known expansion maxima for documented consumers but did not establish a global `$15` maximum.

### Interpretation

The `callerMovable` name is source-code provenance, not independent retail-runtime proof of the exact producer. It nevertheless materially weakens any attempt to infer one global `$15` semantic or maximum merely from the opcode number. The safest current model is that `$15` belongs to a family whose concrete value depends on the calling context until a runtime producer/storage contract is recovered.

ID 10010's language strongly indicates that its particular `$15` value is used as a vocative/addressed label or name. This is a semantic observation only. It does **not** prove that `$15` is the player-name substitution, and `$28` remains the substitution with the documented player-name storage bound.

## Evidence grade

- **CONFIRMED**: code-level `callerMovable` classification; ID 10010 C5/category membership; Japanese/English/Korean substitution placement; absence of a `$15` maximum from the inspected upstream fixed-buffer contract.
- **STRONG**: a global `$15` byte bound should not be guessed from opcode identity alone; investigation should resolve the caller/call-site supplying ID 10010's value.
- **OPEN**: exact retail producer of `$15` for ID 10010, its maximum encoded length, intermediate scratch/staging capacity, final expansion destination, and any causal link to the freeze.

## What this excludes

- It excludes treating `$15` as equivalent to the documented `$28` player-name contract without additional evidence.
- It excludes using a guessed global `$15` maximum as a release-safety proof.
- It does not exclude `$15` overflow or staging-buffer corruption in the ID 10010 call path.

## New questions

1. Which CDC/event-script call site supplies `$15` when ID 10010 is displayed?
2. Does that producer point to a fixed-size name/label field, a temporary formatter buffer, or another message-derived string?
3. What is the maximum encoded length at that specific producer?
4. Does substitution expansion write directly into C5's retained 256-byte page destination or into an earlier smaller scratch/staging buffer?
5. Can the ID 10010 call site be reproduced from authenticated retail CDC/EBOOT assets before another device test?

## Related commits

- `b89c9cdb3644f42586410af4e1e0d26dc06d9111` — quantify unbounded inline opcodes within C5
- `f337f34f9ae08e25cbee23592357130b06b39a25` — compare freeze-adjacent opcode with verified category peers
- upstream `63c19ea22e987fa46b71f10161828ca4af7a5251` — allow caller substitutions to move
- upstream `36edf1a82e401dd3e34724263b3b20e538d9c9a0` — enforce explicit message buffer byte limits
