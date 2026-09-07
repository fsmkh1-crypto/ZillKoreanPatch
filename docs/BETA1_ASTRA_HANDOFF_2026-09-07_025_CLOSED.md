# Zill O’ll Infinite Plus PSP Korean Patch — Astra Handoff after Beta1 025/026 closure

Date: 2026-09-07
Repository: `fsmkh1-crypto/ZillKoreanPatch`
Working branch: `milestone/Beta1`
Reference HEAD immediately before this handoff document: `b4f3f06dbc7fc29c54cf51b6eecafc14410ef921`

**Always re-fetch the actual remote `milestone/Beta1` HEAD before every mutation.** Other legitimate work may advance the branch. Never reset a legitimate newer commit backward. Never force-push.

**Beta1 is incomplete. Do not start Beta2.**

Pinned English implementation/reference authority:
`HK47196/zill@a98d9ce29f361d666ec23da0dcfd351f24537ffd`

Japanese remains semantic authority. For a correction, read **JP -> pinned EN -> KO**. The English patch is not semantic authority over Japanese, but it is the implementation/reference baseline for how the existing patch solved wording, dynamic-substitution, storage/runtime, naming, and UI constraints. Before inventing a Korean-specific workaround, inspect what English did and why.

---

## 1. Read these first

1. `docs/BETA1_ASTRA_HANDOFF_2026-09-07_025_CLOSED.md` — this file, newest operational handoff.
2. `docs/BETA1_ASTRA_RESUME_NOW.md`
3. `docs/BETA1_REVIEW_PIPELINE_V3_CHECKPOINT.md`
4. `docs/BETA1_REVIEW_LEDGER_POLICY.md`
5. `docs/audit/beta1-review-coverage.md`
6. `docs/audit/beta1-review-distribution.md`
7. `AGENTS.md`
8. `CONTRIBUTING.md` and Korean terminology/style docs as needed.

Do not treat older checkpoint counts as current if this handoff or generated coverage is newer.

---

## 2. Current authoritative state

After dense Batch 025 and the exact-JP whole-corpus consistency closure registered as Batch 026:

- Accepted Korean IDs: **42,016**
- Valid contextual review: **2,714 / 42,016 = 6.459%**
- `CONTEXT_STALE`: **0**
- `LAYOUT_RECHECK`: **0**
- contextual `UNREVIEWED`: **39,302**
- pending batches lacking registered review basis: **none**
- propagated coverage: **0**

Generated authoritative source: `docs/audit/beta1-review-coverage.md`.

### Dense Batch 025

- range/population: **700 directly read accepted IDs**
- first selected ID: `200405`
- last selected ID: `250014`
- KEEP: **623**
- EDIT: **77**
- exclusions: **0**
- propagation: **0**
- semantic commit: `113b0245cda3f17a33a496d3aa729d4030003ac0`
- scope: `docs/audit/review/scope-025.json`
- deterministic packet SHA256: `486634fa06d1dea13c72abfdc13621032e604ed32e9820008d37c009257ad239`

025 was directly reviewed source-order using JP -> pinned EN -> KO. Because the first-pass edit density was high, repeated patterns and KEEP samples were voluntarily rechecked before scope registration. **Do not confuse that with the formal >5% rule:** the policy’s mandatory expansion threshold is the **second-pass KEEP-audit correction rate**, where **more than 5%** triggers expansion. First-pass EDIT rate itself is not the formal 5% trigger.

### Batch 026 — post-025 consistency closure

Whole-corpus QA found seven exact-JP consistency groups after the 025 semantic work. They were individually rechecked against Japanese and pinned English and then corrected through the normal heavy queue.

Canonical evidence:

- manifest: `docs/audit/beta1-contextual-copyedit-026-reviewed.json`
- semantic commit: `2fa647d18d09ee1a19a73ad5d67302fec75227bb`
- records: **7**
- no dense scope / no propagation

IDs:

- `190086` `封士 / Sealer` -> `봉인사`
- `60031` `王者` / English champion handling -> `챔피언`
- `60028` Dragon King defeat wording -> natural Korean direct-object structure
- `230018` dynamic `<value:$15>` request -> remove fixed Korean particle, matching English’s placeholder-safe structure
- `240018` dynamic delivery request -> placeholder-safe no-fixed-particle structure and correct living-target wording
- `1260226` `ゴブゴブ団員` -> `고브고브단 단원`
- `1670290` `謎の冒険者` -> `수수께끼의 모험자`

Net valid coverage became 2,714 because some of these seven already had contextual evidence and some were newly credited direct edits. The ledger correctly deduplicates by ID.

### Final validation state

After 026 and after restoring the materializer described below:

- whole-corpus QA-1 integrity: PASS
- QA-2 terminology: PASS
- QA-3 repeated-source consistency: **actionable groups 0 / actionable records 0**
- QA-4 voice hazard scan: PASS
- text sanity: PASS
- layout lexical drift: 0
- glyph repertoire/font checks: PASS
- dynamic-substitution/runtime/inline-opcode audits in general CI: PASS
- Go tests/vet: PASS

Do not call a future batch complete merely because the edit queue committed successfully. In particular, explicitly confirm whole-corpus QA-3 has `actionable_inconsistent_groups=0` and `actionable_inconsistent_records=0`. The queue invokes `qa-consistency.py`, but the dedicated Korean data CI contains the explicit zero-actionable assertion.

---

## 3. The speed problem we were solving

The user explicitly objected that too much time was being spent designing scanners, evidence machinery, and policies instead of **reading and fixing actual game text**.

The operating rule is now:

> **If work does not directly help read/fix the current sentences or safely apply those reviewed fixes, defer it.**

Do not restart pipeline design. Do not invent a new scanner before every batch. Do not turn candidate detection into review credit. Use the existing system and spend most effort on direct contextual reading.

### What the measurements showed

Batch 022 was the useful small-batch timing baseline:

- 134 contextual IDs
- wall clock: **737 sec**
- direct reading: **138.8 sec** = ~1.036 sec/ID
- edit drafting: **120 sec**
- evidence/manifest/queue/scope construction: **207.2 sec**
- successful CI: **69 sec**
- additional tooling/API/wait remainder: **93 sec**

This showed CI itself was **not** the dominant bottleneck. Boundary/evidence/manual manifest work was a major avoidable cost.

Batch 023 increased direct scope to **430 IDs**:

- direct reading: **469.05 sec** = ~1.091 sec/ID
- EDIT 39 / KEEP 391
- coverage moved from 880 to 1,310

Batch 024 tested **700 IDs**:

- EDIT 34 / KEEP 666
- coverage moved 1,310 -> 2,010
- candidate-prepared to completed direct review/manifest work was about **20m50s wall/workflow time**, including connector/chunk I/O; it was **not** a clean pure-reading stopwatch
- heavy apply/QA: about **55 sec**
- final dense scope verification: about **38 sec**
- final general CI around **45 sec**

Conclusion: **700 direct IDs is workable and greatly reduces per-batch boundary overhead.**

Batch 025 confirmed 700 is still workable even when defect density is much higher:

- EDIT 77 / KEEP 623
- high-density patterns were still identifiable late in the packet
- the important failure was not “700 is too large”; it was that manual manifest/evidence handling and final consistency closure must be disciplined.

### Default throughput strategy

Use **~700 accepted IDs per dense batch** unless the content itself provides a strong reason to shrink the batch (e.g. a highly stateful runtime-sensitive cluster). Do not shrink merely because 700 feels large.

The next dense batch should normally be **027**, beginning from the first still-unreviewed accepted row after `250014`. A candidate request can use `start_id: "250015"`; the existing selector will handle skipped/non-accepted/already-reviewed IDs.

Do not create parallel review lanes or propagation lanes just to chase a percentage. Direct reading remains the default.

---

## 4. Why the reviewed-manifest materializer exists

Files:

- `tools/korean/materialize-beta1-reviewed-manifest.py`
- `tools/korean/test_materialize_beta1_reviewed_manifest.py`

These files were first created during 025, temporarily removed because the 025 closure had not yet been fully proven, and then **intentionally restored after 025/026 reached stale=0 and whole-corpus QA green**. Do not remove them again simply because they were once temporarily reverted.

### The bottleneck it removes

With a 700-row batch, especially 025 with 77 edits, manually building the final reviewed manifest requires repeatedly copying:

- exact current `before_korean`
- one or more overlay paths for the ID
- long strings with JSON escaping
- control/substitution tags such as `<value:$15>`, `<end>`, `<select>`, `<if>`

That is mechanical work, not linguistic review. It is slow and creates transcription risk. A single copied-before typo can cause a queue failure; a wrong path or accidentally changed control token is worse.

The materializer lets the reviewer spend time on **`ID + final proposed Korean`** and derives the mechanical fields from the exact reviewed repository commit.

### What it DOES

Given a proposal JSON and an explicit reviewed `--base-sha`, it:

1. resolves the exact base commit and requires that base to be an ancestor of current HEAD;
2. indexes Korean overlay rows at that base commit;
3. finds every physical overlay path for each ID;
4. requires aliases for an ID to have identical JP/KO at the base;
5. pins English reference SHA to `a98d9ce29f361d666ec23da0dcfd351f24537ffd`;
6. rejects duplicate proposal IDs;
7. optionally verifies `expected_japanese` and `expected_before_korean` if provided;
8. rejects `before_korean == proposed_korean`;
9. rejects any change to runtime control/substitution token topology;
10. emits the normal multifile reviewed manifest consumed by the existing queue;
11. reports record count, target files, source head, and manifest SHA256.

Its unit tests cover:

- correct materialization of before/path/pinned-English fields;
- control-topology rejection;
- duplicate-ID rejection;
- stale expected-before rejection.

Both general CI and Korean data CI passed after final restoration.

### What it MUST NOT do

The materializer:

- **does not translate**;
- **does not decide KEEP vs EDIT**;
- **does not decide terminology**;
- **does not propagate to unread siblings**;
- **does not create contextual review coverage**;
- **does not apply the manifest automatically**;
- **does not replace JP -> pinned EN -> KO review**.

Never expand it into an auto-translator or scanner-driven bulk fixer. Its value is precisely that it automates only boring, deterministic manifest assembly while leaving judgment with the contextual reviewer and safety with the existing queue/QA.

### Recommended use for every substantial dense batch

During direct review, maintain a lightweight proposal list, for example in `/tmp/beta1-027-proposals.json`:

```json
{
  "purpose": "Beta1 dense source-order contextual review 027",
  "source_head": "<EXACT REVIEWED BASE SHA>",
  "english_reference_sha": "a98d9ce29f361d666ec23da0dcfd351f24537ffd",
  "records": [
    {
      "id": "250123",
      "proposed_korean": "...<end>"
    }
  ]
}
```

Optional fail-closed checks can be carried in each record:

```json
{
  "id": "250123",
  "expected_japanese": "...<end>",
  "expected_before_korean": "...<end>",
  "proposed_korean": "...<end>"
}
```

Then run:

```bash
python3 tools/korean/materialize-beta1-reviewed-manifest.py \
  --proposals /tmp/beta1-027-proposals.json \
  --output docs/audit/beta1-contextual-copyedit-027-reviewed.json \
  --base-sha <EXACT REVIEWED BASE SHA>
```

Use `--base-sha` explicitly. The base must be the commit whose Korean text was actually reviewed. Do not silently substitute a later HEAD if translations changed in between.

After materialization, inspect the reported record count/target files and the manifest diff. Then use the **existing** reviewed queue. Do not bypass its exact-before and heavy QA gates.

Current queue before 027:

- file: `docs/audit/beta1-copyedit-queue.json`
- revision: **28**
- next normal revision: **29**

For 027, point the queue to `docs/audit/beta1-contextual-copyedit-027-reviewed.json` and use an appropriate commit message such as `copyedit: apply source-order contextual review 027`.

---

## 5. English-patch handling rule — important examples

The user’s project instruction is explicit: **when deciding how to fix something, inspect what the English patch did and why, and use that as the implementation/reference baseline.**

Examples already established:

- `封士` -> pinned EN `Sealer`; Korean `봉사` is ambiguous/wrong, so `봉인사` is preferred.
- `東方諸国` -> EN `Eastern Nations`; Korean `동방 제국들` incorrectly implies empires, so use a country/nations rendering.
- `流れ矢` -> EN `Stray arrows`; Korean `유탄` is semantically wrong for arrows.
- dynamic `<value:$15>` request/delivery text: English places the placeholder in a grammar that needs no gender/particle allomorph. Korean should likewise prefer a **particle-free/restructured sentence** instead of hard-coding `을/를` after an unknown runtime value.
- `王者` in the relevant arena line is handled as `champion` by pinned English, supporting `챔피언`, not the accidental Korean `왕자`.
- `冒険者` project terminology is `모험자`; `勇者` is `용사` where established.

English is a useful implementation disambiguator, not permission to ignore Japanese. If JP and EN disagree, expose the mismatch and resolve from Japanese/context; do not silently copy English.

---

## 6. Exact next dense-review procedure

The next normal dense batch is **027**.

### A. Before mutation

1. Fetch actual remote `milestone/Beta1` HEAD.
2. If newer legitimate commits exist, continue from them; do not reset to a SHA in this document.
3. Confirm generated coverage still has stale/layout zero before starting another translation batch.

### B. Candidate selection

Use the existing candidate pipeline. Suggested request:

```json
{
  "batch_id": "027",
  "start_id": "250015",
  "candidate_count": 700
}
```

Do not create a new candidate scanner. The existing selector already knows accepted IDs and current coverage.

### C. Actual review

For **every selected ID**, read:

1. Japanese
2. pinned English
3. current Korean

Decide KEEP or EDIT contextually.

Use search/grep only to investigate a pattern discovered during reading (terminology consistency, repeated source, name spelling, etc.). Search is not a substitute for direct reading and creates no KEEP coverage.

Prefer high-confidence corrections. Do not rewrite merely to make every sentence stylistically identical. Preserve character voice and legitimate contextual variants.

### D. Build EDIT manifest with the materializer

Keep `ID + proposed_korean` as you review. After the 700 are read, run the restored materializer against the exact reviewed base SHA.

Do not manually re-copy 50–100 long before-values if the materializer can derive them safely.

### E. Apply through existing queue

Use `docs/audit/beta1-copyedit-queue.json` revision 29 for normal 027 work.

The existing queue must remain responsible for:

- exact before-value verification
- reviewed values apply
- persisted-layout invalidation/re-derivation
- glyph/font checks
- integrity/terminology/consistency/text sanity
- pinned-English consumer/storage/effective-layout contract
- semantic commit registration for the normal numbered manifest

Do not hand-edit translations around the queue unless diagnosing a real workflow defect.

### F. Explicit whole-corpus closure before declaring the batch complete

After the semantic result, require:

- QA-1 critical 0
- QA-2 mismatch 0
- QA-3 **actionable groups 0 / actionable records 0**
- text sanity findings 0
- layout drift 0
- glyph/font/data gates pass

If QA-3 finds an exact-JP consistency group, inspect **JP + pinned EN + all Korean contexts** and fix only genuinely inconsistent members. Generic reactions/voice lines may be legitimate contextual exceptions; do not flatten speaker voice merely to satisfy equality.

If a post-batch correction changes a previously reviewed ID, register that direct correction as reconstructable manifest evidence (as 026 did). Never leave generated coverage stale and never hand-edit the generated ledger.

### G. Register dense scope

Create `docs/audit/review/scope-027.json` following the exact current schema and the 025 example.

The scope must contain the exact 700 IDs, edit IDs, exclusions, reviewer, pinned EN SHA, source files, reviewed commit, and deterministic packet SHA.

Avoid the old deliberate placeholder-hash failure trick when possible. The packet builder can print the real hash directly:

```bash
python3 tools/korean/build-beta1-review-packet.py \
  --english-root <PINNED_ENGLISH_CHECKOUT> \
  --reviewed-commit <REVIEWED_COMMIT> \
  --source-file <SOURCE_FILE> \
  --ids <COMMA_SEPARATED_SCOPE_IDS>
```

or create the scope then use the same script without `--verify` to obtain the digest before committing the final scope. The final scope workflow must regenerate and verify the exact hash.

Do not award aggregate KEEP counts without the exact reconstructable ID population.

### H. Final coverage check

Only after scope/ledger regeneration is successful, record the new authoritative coverage from `docs/audit/beta1-review-coverage.md`.

Do not claim completion from a candidate count or a local KEEP count.

---

## 7. Accuracy audit state

The independent 200-row KEEP audit after 023 found **10 corrections / 200 = exactly 5.00%**.

Policy consequence:

- exactly 5.00% = formal PASS
- **more than 5%** = expansion trigger

The audit exposed terminology/naturalness consistency risks, which is one reason whole-corpus QA and English-reference comparison remain important. Do not invent a new formal rule that first-pass EDIT rate above 5% automatically triggers expansion.

Do not run a full second-pass audit after every dense batch unless the established checkpoint/cadence requires it. Direct contextual coverage is currently the dominant unfinished task.

---

## 8. Things Astra must NOT do

- Do not start Beta2.
- Do not force-push.
- Do not reset to any SHA in this handoff if remote HEAD is newer.
- Do not redesign the review ledger or pipeline before continuing text review.
- Do not create a new scanner to replace reading.
- Do not credit scanner-only KEEP.
- Do not bulk-propagate unread duplicates.
- Do not treat English as semantic authority over Japanese.
- Do not ignore English implementation handling when it resolves a Korean patching/grammar/runtime question.
- Do not remove or rewrite the materializer unless an actual reproducible bug is found.
- Do not expand the materializer into translation judgment.
- Do not call a dense batch done until semantic apply, whole-corpus actionable QA closure, scope/packet verification, and ledger stale/layout closure are all confirmed.

Historical accidental branch pointers `do-not-use`, `no-op`, and `discard-me` were created during earlier connector probing. They do not affect `milestone/Beta1`. Do not use them as working branches and do not create more cleanup branches.

---

## 9. Practical priority

The project has **39,302 contextually unreviewed accepted IDs** remaining as of this handoff. The highest-value next action is therefore not more infrastructure work; it is:

> **generate Batch 027 (~700 IDs) -> directly read JP/EN/KO -> record only real EDITs -> materialize safely -> existing queue -> explicit whole-corpus QA closure -> dense scope -> ledger -> repeat.**

The materializer exists specifically so the larger 700-ID review strategy does not lose its speed advantage to manual JSON/before/path transcription.

---

## 10. Copy-paste mission for Astra

When taking over, state-check first, then execute. Do not merely propose a plan and stop.

1. Re-fetch current remote `milestone/Beta1` HEAD.
2. Read this handoff and the mandatory policy/checkpoint files.
3. Verify authoritative coverage and zero stale/layout state.
4. Continue with dense Batch 027, default 700 accepted IDs from the first unreviewed row after 250014.
5. Directly review JP -> pinned EN -> KO for every selected ID.
6. Use `tools/korean/materialize-beta1-reviewed-manifest.py` for reviewed EDIT manifest assembly; do not automate translation judgment.
7. Apply through the existing queue and explicitly close whole-corpus actionable QA-3 before declaring the batch done.
8. Register deterministic dense scope and refresh ledger.
9. Keep the user informed with actual reviewed counts, EDIT/KEEP counts, coverage gain, and measured wall time; distinguish clean reading time from workflow/API wall time when not separately instrumented.
10. Continue dense review rather than returning to tool/scanner design unless a real blocker is demonstrated.
