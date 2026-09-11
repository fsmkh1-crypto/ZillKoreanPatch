# Beta1 archive and runtime QA checkpoint

Date: 2026-09-11

Repository: `fsmkh1-crypto/ZillKoreanPatch`  
Branch: `milestone/Beta1`

This checkpoint records the first post-language-closeout runtime observation and the corresponding external archive layout.

## 1. Language and validation baseline

- Residual-quality frozen baseline: `2c29f53c8c949e3b10f27ec2ea3044abb2278bec`
- Batch 060 semantic commit: `f34341dc7d4675066b4d495ee6b84af41e06350d`
- Batch 060 evidence commit: `a57cf323ab7bddc7f0f11ddc4273ae8ccdf5a9db`
- Beta1 language-closeout checkpoint: `e9d1e2269017b21cf1c411db0c6cd2adec873e73`
- Large-scale validation reuse guide commit: `36b28cdb73ad3558a5382660246ed87d50aee259`
- Pinned English reference: `HK47196/zill@a98d9ce29f361d666ec23da0dcfd351f24537ffd`

Accepted Korean IDs remain 42,016 with valid contextual review 42,016/42,016, `UNREVIEWED=0`, `CONTEXT_STALE=0`, and `LAYOUT_RECHECK=0` at language closeout.

## 2. B-font readability profile carried by the tested RC

The tested post-Batch-060 RC includes the Beta1 B-font profile:

- source raster: 10x10
- BearingX: 1
- Advance: 12
- alpha gamma: 0.60
- rounded 4bpp quantization
- render rule: `opentype-10px-72dpi-hinting-none-origin-0,-2-alpha-gamma-0.60-round-4bpp-v2`

The readability adjustment changes alpha intensity before 4bpp quantization while preserving the established geometry/advance contract.

## 3. Fresh post-Batch-060 build evidence

General CI:

- run `34315400302`
- result: SUCCESS

Android full-corpus RC:

- run `34315400270`
- result: SUCCESS
- artifact ID: `10089928239`
- artifact name: `zill-korean-a054-full-corpus-patcher`
- artifact ZIP SHA-256: `98049c94446200effe01cc87f23671c8a3df63ce9138cfe450d6352b2888c8ff`
- raw APK SHA-256: `a757b51ae19b245776bf80526c8cd38cb5da6be2640327c6ef2b1a9d73b96983`

The ZIP digest and raw APK digest are intentionally recorded separately.

## 4. First runtime observation after language closeout

User-observed test result on 2026-09-11:

- the previously problematic first freezing section was passed successfully;
- the conspicuous line-break problems previously visible in early play were not observed in this pass.

Disposition: positive runtime non-reproduction evidence. This does not prove universal runtime safety across all routes or substitutions.

## 5. Runtime direction from this point

Beta1 language/static QA remains closed. Further work should be driven by concrete runtime evidence rather than another broad polishing pass.

When a new runtime defect is found:

1. record screenshot/scene/dialogue context and, when possible, record ID;
2. inspect the Japanese source;
3. inspect how the pinned English patch handles the same consumer/structure and why;
4. change only the proven affected scope;
5. rerun the relevant regression/QA and produce a fresh RC when source changes.

Minor wording, taste-level punctuation, or non-critical polishing discovered after the declared cutoff should move to Beta2 unless it contributes to a demonstrated runtime/semantic defect.

## 6. Google Drive archive layout

The human-readable and binary archive is stored under the user's Google Drive logical path:

`GPT/질올 한글패치/`

Layout:

- `00_현재상태/`
  - `BETA1_CURRENT_STATE_2026-09-11.md`
- `01_릴리즈_APK/`
  - `Zill_Oll_Infinite_Plus_Korean_Beta1_RC.apk`
  - `zill-korean-a054-full-corpus-patcher_post060.zip`
- `02_검증기록/`
  - `BETA1_LARGE_SCALE_VALIDATION_REUSE_GUIDE.md`
  - `BETA1_RUNTIME_QA_LOG.md`
- `03_참고자료/`
  - `GITHUB_REFERENCE_INDEX.md`

GitHub remains the source of truth for code, repository-owned evidence, and formal project history. Google Drive is the companion archive for distributable binaries and fast human handoff.

## 7. Reusable large-scale validation reference

For a future large three-party audit or full-corpus residual-quality pass, read:

`docs/audit/BETA1_LARGE_SCALE_VALIDATION_REUSE_GUIDE.md`

It records the successful bounded strategy, including reviewer separation, blind frozen sampling, Pattern Registry full-corpus census, one Late Addendum only, terminology normalization rules, apply/finalize protocol, and failure modes to avoid such as infinite resampling, same-reviewer full rereads, blind mechanical replacements, excessive CI, and unnecessary new framework development.
