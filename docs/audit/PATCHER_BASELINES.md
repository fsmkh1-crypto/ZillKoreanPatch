# Korean patcher U-series baselines

This file records immutable recovery points for user-tested/release-candidate Korean patchers. U-series labels are milestone names, not claims of root-cause proof or universal runtime safety.

## U0 — first non-freeze baseline

- Git commit: `3c3d5a04d5b677f84342a9c7486981e11feb8f60`
- Recovery branch: `milestone/U0`
- Meaning: first patcher reported by the user to complete the previously freezing path twice without a freeze.
- Evidence interpretation: two successful runs are two non-reproductions only; they do not prove universal safety.
- Static gates at this baseline: full Korean consumer-storage census `canonical=42016 checked=42016`, 141 renderer-aware profile layouts with `contracts=PASS visual=PASS`, A-054 scanner census, standard CI, Android RC and embedded-payload verification.
- APK SHA-256 recorded at generation time: `e5a55a1ae2749a1f1a7046e8156b4ccc01ccd3a495fb9c94abbfecd4e4342980`.

## U1 — system UI expansion baseline

- Git commit: `f55817ba7981ca983fdc38bdb547023676ade684`
- Recovery branch: `milestone/U1`
- Meaning: U0-derived line with environment/settings/title/save-load fixed EBOOT strings expanded under the upstream-English fixed-width authority.
- Static gates: standard CI PASS; A-054 scanner census PASS; English-consumer storage/visual census `canonical=42016 checked=42016`, `contracts=PASS`, `visual=PASS`; Android unit tests/build/embedded payload verification PASS.
- Android RC artifact id: `9740652487`.
- Artifact digest: `sha256:4af4ab9e85ef2e6ec85a9afb8d3bb1f712f68a6f3180bb20afb6868daa0c7a7c`.
- APK SHA-256 recorded at generation time: `4f730d7a56d1bf929f412541c2e4052af08735030037c4e942b562b63d7e9f61`.

## U2 — opening / character-creation / tutorial fixed UI baseline

- Git commit: `f290574c92f1ddc4e0b754b21ac85f44538bd19f`
- Recovery branch: `milestone/U2`
- Meaning: U1 plus upstream-English-authorized fixed EBOOT strings for opening and character-creation surfaces: sex/color labels, start locations, tutorial label, birth-date/element labels, and final-confirmation choices. Input-window command strips at `0x243d50` / `0x2465f8` are intentionally excluded.
- Static gates: standard CI PASS; English-first full parity PASS; fixed-data audit PASS; glyph repertoire PASS; A-054 scanner census PASS; full Korean English-consumer storage/visual census PASS; Android unit tests/build/embedded payload verification PASS.
- Android RC artifact id: `9741993896`.
- Artifact digest: `sha256:1e226d23663c8b77f17d0299bd1f3adcb25dbada1c256368591a2cf1e7099bfd`.
- Device runtime evidence: not yet recorded for U2.

## U3 — input-window command-strip baseline

- Git commit: `d7c25b281b4b3241b6797db021739e82e2522919`
- Recovery branch: `milestone/U3`
- Meaning: U2 plus the two input/name-entry command strips at `0x243d50` and `0x2465f8`, translated strictly as upstream-English-owned guarded EBOOT fixed strings.
- Explicit boundary: U3 does not modify retail keyboard character tables, mode dispatch, cursor/navigation mechanics, or input buffers.
- Static gates: standard CI PASS; `audit-input-window-english-authority.py` PASS; English-first fixed-data PASS; A-054 scanner census PASS; full Korean English-consumer storage/visual census PASS; Android unit tests/build/embedded payload verification PASS.
- Android RC run: `33346491038`.
- Android RC artifact id: `9742179788`.
- Artifact digest: `sha256:278c3958b449e1c7779f45d2a8eb41b39cd64445e73756b05512c26f8a7a8189`.
- APK SHA-256 recorded at generation time: `5cd4fb7177a61b3d6cce699f6ea0cccfb92b9df5601654c6bf3a44e70e57b85a`.
- Device evidence after generation: the ABC input page displayed Korean custom glyphs in some Latin/digit/symbol cells, and selecting those cells inserted the displayed Korean character. This established live keyboard-slot ownership and motivated U4.

## U4 — keyboard slot-protection baseline

- Git commit: `b3ad54403becb8d780723a0da887fdd73884a55e`
- Recovery branch: `milestone/U4`
- Meaning: U3 plus structured reservation of the retail input keyboard's CP932 fullwidth-ASCII page (`U+FF01..U+FF5E`) so Korean custom-glyph allocation cannot reuse those live input slots. Keyboard mechanics, cursor logic and input buffers remain stock.
- Evidence basis: user device screenshots and input behavior confirmed that the affected cells are real input consumers, not merely visual aliases; selecting a Korean-looking corrupted cell inserted that Korean character.
- Guard: `KeyboardInputReservedKeys()` is used by both desktop and mobile Korean planners; tests cover uppercase/lowercase/digits/punctuation and forbid Hangul allocation to protected keyboard keys.
- Static gates: standard CI PASS; English-first audits PASS; `go test ./...` and `go vet ./...` PASS; A-054 scanner census PASS; full Korean English-consumer storage/visual census PASS; Android unit tests/build/embedded payload verification PASS.
- Android RC run: `33349410117`.
- Android RC artifact id: `9743109503`.
- Artifact digest: `sha256:e9e5bfe1933af6405b16344456c26a89d985b03b060370acb63aa9b1c9b99dd8`.
- Device runtime evidence: not yet recorded for U4. The keyboard fix remains a release-candidate hypothesis until the ABC page and prior freeze path are re-tested on device.

## U5 — verified dialogue reflow baseline

- Git commit: `49528fd869e06817e17e636445929721ca6b92c7`
- Recovery branch: `milestone/U5`
- Meaning: U4 plus build-local Korean reflow for upstream-English `narrowText` records (verified dialogue and in-world-guidance ranges). This closes the parity gap where upstream English inserted line breaks before warning on width, while Korean previously emitted only the warning and could leave an unbroken device line.
- Scope boundary: canonical Korean text is unchanged; only build-local layouts receive derived `<line-break>` boundaries. Existing authored layouts are not rewritten. Item-description/special consumers remain under their existing dedicated contracts.
- Safety rules: actual Korean renderer advances are measured; derived layout must preserve semantic/control topology; a word that cannot fit the inherited English advance limit fails closed; runtime value adjacency is not newly split at a derived boundary.
- Static gates: work-branch standard CI PASS; active-branch Android RC PASS; `go test ./...` and `go vet ./...` PASS; A-054 scanner census PASS; full Korean English-consumer storage/visual census PASS after dialogue derivation; Android unit tests/build/embedded payload verification PASS.
- Android RC run: `33350619031`.
- Android RC artifact id: `9743494846`.
- Artifact digest: `sha256:ff0b7bbefde7367874f357e949f0631b0aed620363fa4a68296d12eba47d0b35`.
- Device runtime evidence: not yet recorded for U5. The photographed overflowing dialogue is expected to be covered only if its source ID belongs to the verified upstream-English narrow-dialogue ranges; device confirmation remains pending.

## U-series policy

1. Never move or overwrite `milestone/U0` through `milestone/U5` except to correct a documented baseline-recording error against authenticated generation/runtime evidence.
2. Future work happens on a separate work/active branch and receives a new U-number only after its intended scope is complete and its static/compile gates are green.
3. A device success is recorded as non-reproduction, not proof of safety. A freeze/crash is strong failure evidence.
4. Upstream-English consumer/storage/visual contracts remain the primary authority; Korean-only deviations require explicit renderer/encoding evidence.
5. Structured live-consumer ownership (such as the confirmed keyboard page) may reserve renderer slots; arbitrary whole-blob two-byte occurrences remain diagnostic-only and must not establish slot ownership.
6. Runtime confirmation may be deferred; each green release-candidate stage remains separately recoverable and is explicitly labeled as device-unverified until tested.
