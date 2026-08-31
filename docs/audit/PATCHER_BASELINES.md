# Korean patcher U-series baselines

This file records immutable recovery points for user-tested/release-candidate Korean patchers. U-series labels are milestone names, not claims of root-cause proof or universal runtime safety.

## U0 — first non-freeze baseline

- Git commit: `a9a693eb66b5d05ba4e5eca859815f4b97e25b22`
- Recovery branch: `milestone/U0`
- Descriptive recovery branch: `milestone/U0-first-nonfreeze`
- Meaning: first patcher reported by the user to complete the previously freezing path twice without a freeze.
- Evidence interpretation: two successful runs are two non-reproductions only; they do not prove universal safety.
- Static gates at this baseline: full Korean consumer-storage census populated at `canonical=42016 checked=42016`, plus A-054 scanner census and Android payload verification.
- APK SHA-256 recorded at generation time: `bdfcd0588aecfded22947737ef3d1e7983e58574272bf9597d3871415e938bd0`.

## U1 — system UI expansion baseline

- Git commit: `f55817ba7981ca983fdc38bdb547023676ade684`
- Recovery branch: `milestone/U1`
- Descriptive recovery branch: `milestone/U1-system-ui`
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

## U-series policy

1. Never move or overwrite `milestone/U0`, `milestone/U1`, or `milestone/U2`.
2. Future work happens on a separate work/active branch and receives a new U-number only after its intended scope is complete and its static/compile gates are green.
3. A device success is recorded as non-reproduction, not proof of safety. A freeze/crash is strong failure evidence.
4. Upstream-English consumer/storage contracts remain the primary authority; Korean-only deviations require explicit renderer/encoding evidence.
5. Input-window work is isolated from ordinary fixed-string/UI localization until its renderer/input consumer contract is established.
