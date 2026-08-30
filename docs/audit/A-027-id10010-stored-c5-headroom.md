# A-027 — ID10010 stored C5 headroom without runtime assumptions

## Trigger

ID 10010 is a retained C5 consumer and contains inline `<value:$15>`. The dynamic audit previously tracked literal/runtime-known bytes but did not separately report the two compiled substitution bytes `02 15`.

## Hypothesis

Quantifying the exact compiled page payload can show whether ID10010 is already close to the retained 256-byte C5 boundary before runtime substitution. This must remain separate from any claim about the runtime expansion destination or replacement semantics.

## Verification

Current canonical layout:

`<value:$15>여<line-break>나는 그대가 나아갈 길을 묻노라<line-break>그 길을 보이고 운명의 문을 열어라<end>`

Before the block terminator, the rendered literal portion contains:

- 28 Hangul syllables, each mapped to a two-byte custom renderer key = 56 bytes
- 11 single-byte ASCII/space/line-break bytes = 11 bytes
- compiled inline substitution token `02 15` = 2 stored bytes

Exact compiled C5 page payload before `<end>`: **69 bytes**.

Distance from that stored payload to the 256-byte boundary: **187 bytes**. Equivalently, 186 additional bytes could be added to the stored page while remaining below 256.

The C5 dynamic audit now separates:

- `StoredBytes`: exact compiled bytes, including `02 XX` substitution tokens
- `StaticBytes`: literal runtime payload without assigning substitution semantics
- `KnownMaxBytes`: literal bytes plus only independently proven runtime substitution maxima

`StoredHeadroomBytes()` is explicitly documented as a bank-storage diagnostic, not a runtime expansion threshold.

## Result

ID10010 is **not near the 256-byte boundary in its compiled form**. The stored payload is 69 bytes.

This materially weakens a simple theory in which the Korean translation by itself nearly fills the retained C5 page before `$15` is processed.

It does **not** establish that `$15` is safe. A runtime problem could still arise if:

- `$15` is expanded into a different fixed-size scratch/staging buffer;
- the runtime appends rather than replaces data;
- the source supplied by the caller is malformed or unexpectedly long;
- another consumer copies the expanded text into a smaller buffer;
- the freeze is in the renderer/font path instead.

## Evidence grade

- **CONFIRMED**: current ID10010 canonical layout and exact compiled byte count under the installed two-byte Korean renderer mapping contract.
- **OPEN**: runtime `$15` source, maximum encoded length, substitution write semantics, destination buffer, and causal role in the freeze.

## What this excludes

The current ID10010 Korean page is not statically close to overflowing the retained 256-byte C5 page. Any `$15`-related memory explanation now requires a runtime contract different from the naive `69B + ordinary short name` picture.

## New question

What exact CDC caller supplies `$15` for ID10010, and what destination/storage path consumes the expanded result?

## Commit

Implementation commits immediately preceding this note separate stored substitution-token bytes from runtime-known expansion accounting and add regression coverage. This note must not be cited as proof of the runtime substitution algorithm.
