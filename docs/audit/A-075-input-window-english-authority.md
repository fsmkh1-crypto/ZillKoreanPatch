# A-075 — Input-window English authority

## Scope

This audit isolates the two stock input/name-entry command strips from broader Korean input-method work.

Upstream English owns exactly these visible command-strip fields through the same authenticated fixed-width EBOOT mechanism used for the other 557 executable strings:

- `0x243d50`: `カナ　　かな　　英数　　ゆだねる　　空白　　取消　　入力完了`
- `0x2465f8`: `  カナ   かな   記号   ゆだねる   空白   取消   中止   完了`

The upstream executable patch manifest contains no keyboard/input/name-entry feature. `internal/fixeddata/eboot.go` applies these fields generically with source-byte authentication, NUL guard, non-overlap, CP932 encoding, and replacement-capacity checks.

## Korean U3 treatment

U3 follows that proven English contract only:

- `0x243d50` -> `가타  히라  영문  자동  공백  취소  완료`
- `0x2465f8` -> `  가타   히라   기호   자동   공백   취소   중지   완료`

Command order and count are preserved. These are label substitutions only.

## Contract status

| Surface | Status | Evidence |
| --- | --- | --- |
| `0x243d50` fixed field | PASS when gates are green | upstream-English fixed EBOOT authority + Korean source/capacity/glyph gate |
| `0x2465f8` fixed field | PASS when gates are green | upstream-English fixed EBOOT authority + Korean source/capacity/glyph gate |
| command order/count | PASS when `audit-input-window-english-authority.py` is green | exact expected strings + 7/8 label census |
| keyboard executable patch | ABSENT upstream | no keyboard/input feature in upstream executable manifest |
| retail character-table contents | UNKNOWN | not established by upstream fixed-string authority |
| mode-dispatch indices | UNKNOWN | not established by upstream fixed-string authority |
| cursor/navigation mechanics | UNKNOWN | not established by upstream fixed-string authority |
| Korean character entry | NOT IMPLEMENTED by U3 | U3 does not alter stock input mechanics or character tables |

## Interpretation

A green U3 proves only that the two visible command strips follow the upstream-English fixed-string contract and remain inside the authenticated executable fields. It does not prove that the stock input method can enter Hangul, nor does it justify changing keyboard tables or mode dispatch without separate evidence.

If later work adds real Korean character entry, it must be a new U-series scope with its own table/index/buffer/renderer contract and must preserve U0/U1/U2/U3 as recovery points.
