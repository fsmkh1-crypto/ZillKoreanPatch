# PSP Koreanization Reconnaissance Method — Oreshika-Derived

## Purpose

This document turns the Oreshika Korean-input research into a reusable PSP localization investigation method.

The goal is **not** to assume every PSP game can be localized the same way as Oreshika. The goal is to test for the cheapest existing text path first, and only fall back to game-specific encoding/font-slot hacks when the game actually requires them.

## Core rule

Before building a custom Hangul renderer or remapping Korean into unused Japanese code points, determine whether the target game already has a 16-bit/Unicode-capable text path that can be reused or extended.

The preferred order is:

1. inspect existing string storage and runtime buffers;
2. inspect existing input/OSK path;
3. inspect the renderer/font consumer;
4. verify serialization/save-load behavior;
5. only then choose an implementation strategy.

Do **not** infer general Unicode support merely because Korean input works in one screen. Input, storage, rendering, and serialization are separate contracts.

---

## What Oreshika currently proves

The recovered final Oreshika patch package proves, with high confidence:

- a custom Korean manual surname/given-name entry subsystem exists;
- EBOOT-side Korean OSK initialization was modified;
- a dedicated name UI asset region exists;
- Korean names survive save/load and cold-load validation;
- PGF/font, name UI, OSK, and name serialization are treated as connected protected surfaces by the patch QA lineage.

The exact Hangul composition algorithm is still unresolved. We do **not** yet claim that Oreshika necessarily uses standard UTF-16 Hangul arithmetic, the PSP system OSK directly, or a universal game-wide Unicode renderer. Those remain implementation questions to verify from the final EBOOT/tables.

Therefore the reusable lesson is methodological:

> **Look for an existing 16-bit/Unicode-capable consumer path before inventing a CP932 slot-remap architecture.**

---

## Phase 0 — Preserve evidence before modifying anything

Record the exact target before reverse engineering:

- game ID and revision/version;
- clean ISO hash;
- patched ISO/patch hash if an existing fan patch is being studied;
- EBOOT hash where available;
- known text/input/font asset locations;
- emulator/runtime version used for validation.

Keep original binaries and generated artifacts separate. Do not treat filenames, screenshots, or a single working input field as proof of architecture.

---

## Phase 1 — Determine the game’s text representation

### A. Static inspection

Look for evidence of:

- UTF-16LE or other 16-bit strings;
- CP932/Shift-JIS byte strings;
- custom 16-bit glyph IDs;
- fixed-width text records;
- pointer tables and length fields;
- byte-oriented versus 16-bit-oriented string functions.

Useful questions:

- Are Japanese strings stored as CP932 bytes or 16-bit code units?
- Does the game copy text one byte at a time or two bytes at a time?
- Are terminators `00` or `00 00`?
- Do text buffers have byte capacities or character/code-unit capacities?

### B. Runtime inspection

For one known string, follow:

`asset/script -> decode/copy -> runtime buffer -> renderer`

If possible, alter one character in memory and observe what the renderer consumes.

Do not translate the whole game yet. Prove one string end-to-end first.

---

## Phase 2 — Inspect the input/OSK path separately

If the game has name entry, chat, naming, or other user input, determine:

- what the input UI returns;
- buffer element width;
- whether composition occurs in the game or an external/system utility;
- whether the returned value is Unicode, game-specific IDs, or encoded bytes;
- whether the same buffer later reaches the normal text renderer.

A successful Korean name entry proves only that **that input path** can represent Korean. It does not automatically prove that dialogue/script storage or all UI consumers can.

For Oreshika specifically, the custom Korean OSK/name path is verified; the exact composition and stored representation are still under investigation.

---

## Phase 3 — Inspect rendering independently from storage

Determine how a code unit becomes pixels.

Possible models include:

### Model 1 — Unicode/16-bit text + font lookup capable of Hangul

Best case.

If the renderer already accepts Unicode-like code points and the font path can supply Hangul, Korean localization may avoid per-character fake code allocation.

Potential work:

- provide/extend Hangul glyph coverage;
- patch missing font tables or fallback rules;
- adjust metrics/layout;
- preserve original string/control-code logic.

### Model 2 — 16-bit/Unicode storage but font lacks Hangul

Still favorable.

Keep Unicode/16-bit strings and extend only the glyph/font consumer.

This is usually simpler and more scalable than mapping every Korean syllable into unused CP932 slots.

### Model 3 — CP932 text decoder feeding a renderer that can be extended

Intermediate case.

Possible options:

- extend the decoder for a new Korean encoding range;
- add a separate Korean escape/opcode path;
- convert to an internal wider representation before rendering.

This requires engine work but may still avoid a fixed fake-slot repertoire.

### Model 4 — CP932 byte stream where code values directly select a custom game font

Zill O’ll is the current reference example.

Typical path:

`CP932 bytes -> renderer glyph key -> game PAF/custom font -> pixels`

In this case, slot remapping/reuse may be the safest solution because the renderer architecture itself is code-page keyed.

---

## Phase 4 — Test whether Korean can travel through the whole pipeline

Before mass translation, construct a minimal proof string containing:

- simple syllables: `가나다`;
- complex final consonants: `값`, `앉`, `없`;
- punctuation and spaces;
- if input is relevant, composed user-entered names such as `김`, `태호`.

Verify separately:

1. source/static storage;
2. runtime buffer;
3. rendering;
4. line measurement/wrapping;
5. save serialization if applicable;
6. reload/cold-load rendering.

A path is not considered proven until the final consumer displays the same Korean text after the relevant persistence boundary.

---

## Phase 5 — Choose the least invasive architecture

Use this decision order.

### Path A — Reuse a proven Unicode/16-bit game-wide text path

Choose this when:

- static text can be stored in that representation;
- the normal renderer accepts it;
- Hangul glyphs can be supplied;
- control codes and layout remain stable.

Advantages:

- no fake CP932 code allocation per Korean character;
- easier expansion to newly translated Hangul;
- easier user-entered Korean names;
- fewer repertoire-management problems;
- potentially simpler tooling.

### Path B — Reuse Unicode/16-bit internally and patch only font coverage

Choose this when string representation is already suitable but glyph coverage is incomplete.

This is usually preferable to replacing the text engine.

### Path C — Add a bounded Korean decoder/bridge

Choose this if storage is legacy but the renderer can consume a wider internal representation.

Keep the patch narrow and auditable.

### Path D — Legacy code-page slot reuse

Choose this when the engine is fundamentally keyed to CP932/custom-font slots and replacing that architecture would create disproportionate risk.

Zill’s current Korean implementation belongs here.

---

## Why an Oreshika-like path can be much simpler

If a target game really has a reusable 16-bit/Unicode-capable text path, then Korean syllables can potentially remain their own logical characters instead of being disguised as unused Japanese codes.

That can eliminate or reduce:

- per-rune CP932 slot allocation;
- collision/reservation management;
- fake nominal Japanese code mappings;
- fixed custom-glyph repertoire maintenance;
- some text compiler complexity.

However, this simplification applies only after the renderer and storage path are proven. A Unicode-capable input widget alone is insufficient.

---

## What still remains game-specific even in the best case

A reusable Unicode/16-bit path does **not** remove all localization work.

Still inspect:

- font/glyph coverage;
- glyph metrics and line height;
- UI box widths and wrapping;
- control codes;
- fixed-size buffers;
- save-data field sizes;
- endian/layout rules;
- executable literals;
- texture-rendered text;
- sorting/search/comparison behavior;
- checksum or integrity rules.

The main savings are in character representation and repertoire management, not in translation QA or UI fitting.

---

## Fast reconnaissance checklist for a new PSP game

Before writing a Korean patcher, answer these questions in order:

1. What encoding/representation are ordinary dialogue strings stored in?
2. What representation does the runtime text buffer use?
3. Does the game already contain any 16-bit/Unicode text path?
4. Does name entry/OSK use the same representation as ordinary text?
5. What renderer consumes the final string?
6. Can that renderer address Hangul glyphs directly?
7. Where do the glyphs come from: game font, PGF/system font, texture atlas, or custom table?
8. Can one Korean test string survive render + save/load where applicable?
9. Do control codes and wrapping still behave correctly?
10. Only after those answers: choose Unicode reuse, decoder bridge, or legacy slot remap.

---

## Comparison with Zill O’ll

### Oreshika research lesson

The Korean name-input subsystem demonstrates that a PSP title can have a Korean-capable input/serialization path worth reusing or studying before attempting a code-page slot hack.

Exact Oreshika representation/rendering details are still being reverse engineered, so portability is **plausible, not yet proven**.

### Zill O’ll current architecture

Zill’s verified text/font path is fundamentally CP932-keyed and custom-font based. The English patch also preserves that engine contract and transforms data/font assets around it rather than converting the game to Unicode.

Therefore Zill should continue using the English-patch-first CP932/custom-font architecture unless new runtime evidence proves a cheaper compatible path.

The two projects should not be mechanically merged. Oreshika provides a **reconnaissance strategy**; Zill provides the fallback model for a legacy CP932 renderer.

---

## Standing rule for future PSP Koreanization work

> **Unicode/16-bit reuse first if proven; custom decoder second if bounded and justified; CP932 slot remap only when the existing renderer architecture makes it the safer choice.**

For every new game, prove the consumer path before choosing the patch architecture.
