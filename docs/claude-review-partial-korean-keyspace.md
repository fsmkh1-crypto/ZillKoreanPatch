# Claude review request: partial-Korean alpha CP932 key-space failure

## Current device failure

Android Korean Alpha v0.7.3 reaches the mobile full-font slot planner and fails with:

```
mobile alpha full-font plan allocation: need 1076 custom glyphs but only 257 reusable two-byte slots are available
```

Previous experiments were:

- retail-cell reuse: 192 reusable slots
- metric-rewrite reuse: 202 reusable slots
- full 2637-glyph atlas/PAF repack: 257 reusable slots

The full repack removed physical atlas-cell geometry as the limiting factor, but allocation still cannot provide 1076 renderer keys.

## GPT diagnosis

The likely limiting resource is no longer atlas area; it is the two-byte CP932 renderer-key namespace that can safely be repurposed while preserving untranslated Japanese.

Relevant current code:

### `internal/corpus/korean_runtime.go`

`KoreanProject.RuntimeTexts(source)` returns Korean for translated rows but deliberately falls back to `item.Translation.Japanese` for every untranslated source record.

Therefore the runtime text set used for slot planning includes a large amount of original Japanese.

### `cmd/zill/build_korean_mobile_plan.go`

The mobile full-font planner now calls:

```go
plan, err := koreanslots.BuildPlan(
    texts,
    font.DoubleByteKeys(),
    rendererKeySetSlice(reserved),
)
```

So all installed two-byte PAF keys are initially eligible, but `BuildPlan()` blocks:

1. `RequiredStockKeys(texts)` — stock CP932 keys still required by final runtime text, including untranslated Japanese;
2. fixed/binary reservations from BOOT/BINDATA/etc.

Only the remainder can be mapped to custom Hangul runes.

### `internal/koreanslots/plan.go`

`BuildPlan()` explicitly blocks stock keys required by final runtime text before calling `Allocate(custom, candidates)`.

## Why this may differ from the English patch

The upstream English patch installs a frozen complete font transform (`font/zillfont.par` + `2d/font/jillbtn.par`) and a complete 2637-glyph metrics table. But an almost-fully-English runtime corpus no longer needs most Japanese double-byte glyph keys, so those keys can be repurposed much more freely.

Our alpha is intentionally sparse: translated rows are Korean, untranslated rows remain Japanese. That creates a simultaneous requirement for:

- ~1076 custom Hangul glyphs, and
- a broad stock Japanese repertoire.

A full atlas repack alone cannot create additional renderer keys if the game text encoding still uses only the existing CP932 two-byte key namespace.

## Questions for Claude

Please review the current branch `translation/section001-batch2`, especially:

- `internal/corpus/korean_runtime.go`
- `internal/koreanslots/plan.go`
- `internal/koreanslots/slots.go`
- `cmd/zill/build_korean_mobile_plan.go`
- `internal/zillfont/full_repack.go`
- `internal/release/korean_font_mobile.go`
- upstream English implementation in `HK47196/zill`, particularly `release/font/*` and `internal/release/build.go`

Please answer:

1. Is the diagnosis correct that the 257-slot ceiling is now a **CP932 renderer-key namespace / stock-Japanese coexistence problem**, not an atlas capacity problem?
2. Does the upstream English patch avoid this mostly because its translated corpus frees Japanese double-byte keys, rather than because it expands the renderer-key namespace?
3. For a **partial Korean alpha that must keep untranslated Japanese readable**, what is the safest practical architecture?
   - temporarily translate/replace all remaining Japanese with a reduced placeholder repertoire for device testing;
   - add a second font bank/page selector or custom escape encoding via EBOOT renderer patch;
   - use dynamic paging/banked mappings by message section;
   - extend the renderer to accept another encoding/key range;
   - another approach visible in the upstream implementation.
4. Which option is the smallest-risk route for an immediate PPSSPP integration test, and which is the right long-term production architecture?
5. Please call out any incorrect assumptions in GPT's analysis and any safety issues with the current full-font repack implementation.

Do not assume prior chat context; review the repository code and upstream English patch directly.
