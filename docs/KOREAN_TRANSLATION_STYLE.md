# Korean translation style

This file records the default dialogue/register policy for bulk Korean translation.

## Register

- Japanese plain/casual forms should normally become Korean banmal. Do not defer a row solely because the speaker identity is unknown when the Japanese morphology already makes the register clear.
- Examples: `なんだ`, `待て`, `任せて`, `誰だ` are casual/plain. Translate them as banmal by default.
- Explicit polite Japanese such as `です` / `ます` should normally become Korean polite speech unless scene context clearly requires otherwise.
- Honorifics, titles, archaic speech, military commands, and strongly character-specific verbal tics should preserve their marked tone.

## Context versus register

Register and lexical meaning are separate questions. A plain form can have obvious banmal register while the best Korean verb still depends on context. For example, `待て` is casual either way, but a scene may favor `기다려` or `멈춰`. Resolve that from nearby context rather than postponing the row merely to decide politeness.

## Bulk workflow

- Preserve runtime control tokens exactly.
- Semantic Korean text must not contain build-owned `<line-break>` tokens; layout owns wrapping.
- Keep terminology and proper-name spellings consistent with the project terminology tables and previously accepted Korean corpus.
- If context is genuinely insufficient to choose meaning or a proper-name spelling, leave only that row for later review; do not stop the whole packet.
