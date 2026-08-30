# A-023 — C5 256-byte contract is upstream-retained evidence, not yet independently reproducible proof

## Trigger

The runtime-message branch now depends heavily on a precise statement: C5 dialogue pages use a buffer with 256-byte capacity, while ordinary static validation excludes runtime substitution expansion bytes. The latter behavior is directly testable in current source. The provenance of the former runtime capacity needed a separate audit.

## Upstream history audit

### `release/layout/consumer-map.toml`

GitHub path history shows that `consumer-map.toml` already exists in upstream's root/initial public Go-project commit:

- `932655c531ceebab0471b64b4d02a9cac771f6e4` — `Initial Go translation project`.

Later path-history changes are ordinary maintained snapshot changes such as portrait-dialogue and guild-mission reflow membership updates. There is no earlier public commit containing a generator or derivation step for the initial consumer map.

The current upstream repository tree contains the maintained snapshot and its consumers, but no checked-in generator that reconstructs the C5/C20/C22 membership and buffer contracts from authenticated game assets.

### Explicit byte-limit commit

Upstream commit:

- `36edf1a82e401dd3e34724263b3b20e538d9c9a0` — `Enforce explicit message buffer byte limits`

renamed the layout constants to explicit buffer-capacity names and expanded `release/layout/README.md` to document the runtime contracts, including:

- C5 starts a new page after every third line break;
- at most nine pages;
- each statically known page payload must use fewer than 256 encoded bytes;
- unbounded runtime substitutions remain a runtime-QA boundary.

That commit strengthens the evidence that the upstream maintainer intentionally treated 256 bytes as a verified runtime buffer contract. However, it does not include the retail executable address, disassembly, signature, or reverse-engineering script from which that value was derived.

## Evidence classification

### CONFIRMED

- Current upstream/Korean validator logic uses a 256-byte C5 page-buffer contract.
- Upstream documentation explicitly describes it as a verified storage contract for the supported game revision.
- The contract was already part of the public Go project's initial maintained data/code and was later made more explicit in commit `36edf1a...`.
- Current source does not preserve a public generator or executable-address provenance for independently reconstructing that C5 buffer contract.

### STRONG, not independently CONFIRMED

- `256` is very likely a real runtime property: it is deliberate, consistently documented, enforced in release validation, and paired with other reverse-engineered consumer contracts.

### OPEN

- The exact retail function/address implementing C5 page splitting.
- Whether runtime substitution expansion occurs directly into the same 256-byte destination or first into another scratch/staging buffer.
- The order of substitution expansion versus page splitting/copying.
- The exact overflow/check semantics in executable code.
- Whether `$15`, `$28`, and other inline value opcodes share one formatter/copy path.

## Meaning for the freeze investigation

The current C5 audit proves a real **static-proof gap**: ordinary validation assigns zero payload bytes to substitution tokens. It does **not yet prove a runtime overflow**, because the executable expansion/copy topology is not independently reconstructed.

Therefore:

- do not demote the dynamic-expansion branch merely because upstream validation exists;
- do not promote a page-overflow warning to root-cause proof merely because `static + known substitution max >= 256`;
- reconstruct the retail runtime path before turning forensic warnings into production-blocking limits.

## Highest-value next gate

Add an authenticated-retail C5 contract scanner/analyzer that searches the executable for candidate code regions consistent with the retained contract, then manually/structurally verifies candidates before assigning semantics.

The scanner must be heuristic-only at first and must not hardcode a remembered function address as truth. Useful candidate features include local code windows containing combinations of:

- byte-capacity immediates `0x100` / boundary values `0xFF`;
- line/page-count values `3` and `9`;
- byte load/store or copy-loop behavior;
- comparisons/branches around those values.

A candidate becomes evidence only after data/control flow shows that it consumes message bytes and performs the relevant page/expansion operation.

Once found, record reproducible file offsets/runtime addresses/instruction signatures in this audit rather than relying on private recollection.
