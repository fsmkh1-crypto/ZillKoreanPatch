# A-025 — `$15` projection classification and ID 10010

## Trigger

ID 10010 remains freeze-adjacent evidence and is a retained C5 consumer. Its payload begins with `<value:$15>`, while the current C5 known-expansion audit has no proven maximum for opcode `$15`.

## Hypothesis

A useful next step was to determine whether `$15` has one repository-wide semantic/storage contract that would justify a global encoded-byte maximum, and whether upstream's `callerMovable` naming actually identifies the runtime producer of `$15`.

## Verification

- Rechecked ID 10010 against the retained consumer map and verified category data.
- Compared the Japanese, English and Korean forms of ID 10010.
- Inspected upstream `internal/message/projection.go` and its history.
- Inspected upstream layout rules and storage-contract documentation.
- Rechecked the commit that introduced explicit fixed-buffer byte limits.
- Inspected the retained CDC context decoder's exact C5 argument contract.

## Result

### Confirmed repository facts

- ID 10010 is in the retained C5 consumer set and is categorized as `character-creation-prompt`.
- The source is `<value:$15>よ...`; upstream English renders the same substitution as `<value:$15>, I ask...`; Korean retains it as `<value:$15>여...`.
- Upstream message projection explicitly separates substitution opcodes into two editor/projection groups:
  - `pureMovable`: `$24`, `$25`, `$28`, `$2B`
  - `callerMovable`: `$15`, `$16`, `$17`, `$1A`, `$1B`
- Commit `63c19ea22e987fa46b71f10161828ca4af7a5251` is titled `Allow caller substitutions to move` and includes `$15` in that projection path.
- Upstream fixed-buffer rules/documentation provide an independently documented 16-byte bound for player-name substitutions, which supports `$28`; they do not provide an equivalent global maximum for `$15`.
- The retained CDC decoder establishes C5's command shape as three to seven integer arguments: `mode`, `association handle`, followed by one or more message IDs. C5 has no explicit substitution-string argument.

### Corrected interpretation

`callerMovable` is a **translation projection/editor classification**, not proof that the CDC C5 caller directly supplies the runtime string for `$15`. The retained C5 command grammar now directly falsifies the stronger interpretation that a C5 call passes `$15` as one of its explicit arguments.

The classification still supports one narrow conclusion: `$15` must not be assigned the documented `$28` player-name maximum merely because both are movable substitutions. But the exact runtime producer of `$15` remains unresolved and may live in shared formatter/global state rather than in the C5 command itself.

ID 10010's language strongly indicates that its particular `$15` value is used as a vocative/addressed label or name. This is a semantic observation only. It does **not** prove that `$15` is the player-name substitution, and `$28` remains the substitution with the documented player-name storage bound.

## Evidence grade

- **CONFIRMED**: code-level projection classification; ID 10010 C5/category membership; Japanese/English/Korean substitution placement; absence of a `$15` maximum from the inspected upstream fixed-buffer contract; C5's explicit argument shape does not carry a substitution string.
- **STRONG**: a global `$15` byte bound should not be guessed from opcode identity alone.
- **SUPERSEDED**: the earlier wording that treated `callerMovable` as evidence that the C5/event-script call site itself supplies `$15`.
- **OPEN**: exact retail producer of `$15` for ID 10010, its maximum encoded length, intermediate scratch/staging capacity, final expansion destination, and any causal link to the freeze.

## What this excludes

- It excludes treating `$15` as equivalent to the documented `$28` player-name contract without additional evidence.
- It excludes using a guessed global `$15` maximum as a release-safety proof.
- It excludes the specific model in which C5's explicit command arguments directly carry the `$15` replacement string.
- It does not exclude `$15` expansion or formatter-buffer corruption after C5 resolves the message record.

## New questions

1. Which shared formatter/state source resolves opcode `$15` after C5 selects ID 10010?
2. Is that source a fixed-size name/label field, a global formatter slot, another message-derived string, or something else?
3. What is the maximum encoded length at that source?
4. Does substitution expansion write directly into C5's retained 256-byte page destination or into an earlier smaller scratch/staging buffer?
5. Can authenticated retail EBOOT analysis identify the `0x02 <opcode>` substitution dispatcher and the `$15` case before another device test?

## Related commits

- `b89c9cdb3644f42586410af4e1e0d26dc06d9111` — quantify unbounded inline opcodes within C5
- `f337f34f9ae08e25cbee23592357130b06b39a25` — compare freeze-adjacent opcode with verified category peers
- upstream `63c19ea22e987fa46b71f10161828ca4af7a5251` — allow caller substitutions to move
- upstream `36edf1a82e401dd3e34724263b3b20e538d9c9a0` — enforce explicit message buffer byte limits
