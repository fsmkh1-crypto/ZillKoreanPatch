# A-021 — Common static conditions in the two observed freeze builds

## Trigger

After downgrading historical single-run PASS results to non-reproduction only (A-020), the most reliable runtime evidence in the PR #14 matrix is the actual freeze observed in the Combined full-Han+EBOOT build and the Stable-minimal+EBOOT build. This note asks what those two failing builds genuinely share without treating H0/A/B non-reproduction as proof of safety.

## Current-branch distinction

The current audit branch production file `release/korean/strings/eboot.toml` contains only the two H0-era fixed fields. It is **not** the expanded PR #14 EBOOT set used by the two observed freeze builds. The historical expanded set contains 46 fixed fields and is now preserved separately at `docs/audit/fixtures/pr14-eboot-full.toml`, copied from commit `4c04e52329290995d3ed4f6a2d742547a9d2ff7a` for forensic replay only.

Therefore a fingerprint emitted by the current production EBOOT path must not be described as the fingerprint of the historical Combined/Stable-minimal EBOOT. PR #14 policy replay must explicitly load the historical fixture. Earlier shorthand such as “51-line EBOOT set” referred to diff-added source lines, not the number of translation fields, and is superseded by the exact count of 46 fields.

## Confirmed common conditions

### 1. The expanded Korean EBOOT fixed-string table is the same

The Combined commit `4c04e52329290995d3ed4f6a2d742547a9d2ff7a` and Stable-minimal commit `52b9a8aab3ecf3a9910ed0010d5a234fd0ecc9cb` contain the same expanded `release/korean/strings/eboot.toml` content: 46 fixed fields.

### 2. EBOOT overlay cannot exceed the original fixed-field byte capacities

`internal/fixeddata/korean_eboot.go::ApplyKoreanEBOOT`:

- authenticates the already-manifest-patched ELF fingerprint;
- CP932-encodes the exact Japanese source field;
- encodes Korean through the active renderer mapping;
- rejects any Korean replacement whose encoded byte length exceeds the original source-field byte length;
- requires the exact source bytes followed by NUL at the target offset;
- rejects overlapping replacement fields;
- clears and rewrites only the original field capacity.

`internal/release/korean_fixed.go::applyKoreanFixedEBOOT` then re-verifies the executable runtime manifest after the sparse Korean overlay.

Therefore an on-disk "Korean EBOOT string grew into adjacent ELF data/instructions" explanation is strongly reduced. Likewise, for a normal downstream C-string copy whose original Japanese string was safe, the Korean replacement cannot be longer in encoded bytes than that source string. This does **not** prove renderer/key semantics are safe.

### 3. Both failing builds remove the known visible 0x87 assignments, but by very different mapping policies

- Combined uses the historical full 1,308-rune Han-only round-trip candidate policy.
- Stable-minimal preserves the H0 mapping except mappings whose encoded lead byte is `0x87`, which are relocated to unused round-trip Han keys.

Thus the known `게`/`깃` icon corruption can be absent while the freeze still occurs. This is evidence that the visible 0x87 symptom is not sufficient by itself to explain the freeze.

### 4. Both use the dynamic full atlas+PAF repack

`FullRepackAuthenticatedRetailFont` rebuilds atlas positions for all 2,637 PAF glyph records. Mapping identity determines which PAF keys receive the Korean 10x10 raster/metrics. Different mapping policies can therefore lead to different final PAF/atlas geometry even though both use the same repacker.

Consequently, "both use full repack" is a genuine common condition, but **final font outputs cannot be assumed equal** between Combined and Stable-minimal.

## Important non-conclusions

- The two observed failures do not prove that the EBOOT table is causal.
- The two observed failures do not prove that Han-key relocation is causal.
- The two observed failures do not prove an EBOOT×mapping interaction, because the historical H0/A/B PASS arms were only single-run non-reproductions.
- Full repack is also present in other historical configurations, so its presence alone is not a discriminating cause.
- EBOOT replacement byte-length safety does not prove that every custom renderer key is semantically safe, nor that every fixed-string UI consumer treats those keys like ordinary text.

## Strongest static implication

The EBOOT overlay is byte-capacity bounded strongly enough that a simple fixed-field overwrite is no longer a leading explanation. If EBOOT participation matters, the more plausible remaining mechanisms are:

1. renderer-key semantic behavior while rendering fixed UI strings;
2. state established by a particular fixed-string/UI rendering path;
3. a mapping/font geometry effect that changes lookup/rendering independently of string byte length;
4. an unrelated intermittent bug that happened to reproduce in those two runs.

## Highest-value next asset-backed gate

From one authenticated retail asset set and the current canonical Korean corpus, replay the historical planner policies without device execution:

- H0 = original candidate policy + two-field H0 EBOOT table;
- B = original candidate policy + historical 46-field expanded EBOOT table;
- A = Han-only round-trip policy + two-field H0 EBOOT table;
- Combined = Han-only round-trip policy + historical 46-field expanded EBOOT table;
- Stable-minimal = H0 mapping preserved, only 0x87-lead custom assignments relocated, then historical 46-field EBOOT table.

The replay is explicitly labeled `current_corpus_replay=true`; it is not an exact historical ISO reconstruction because the canonical corpus has changed since PR #14.

For each policy, record:

- custom-rune set and mapping SHA-256;
- number of rune→key differences versus H0;
- mapped-key exact-byte ownership hits in BOOT/EBOOT/BINDATA;
- encoded EBOOT-field digest;
- then, where practical, final PAF SHA-256 and atlas SHA-256 plus PAF geometry deltas.

This comparison is static and deterministic, so it does not depend on interpreting a single runtime PASS. Only after these output deltas are known should another device A/B be selected.
