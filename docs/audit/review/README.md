# Beta1 Dense Review Scope Evidence

Files named `scope-*.json` are authoritative reviewer assertions for sequential full-read coverage. They are source evidence; `translations/korean/review-ledger.jsonl` is generated from them and must not be hand-edited.

## Required schema

```json
{
  "schema_version": 1,
  "scope_id": "S-018",
  "batch_id": "018",
  "kind": "sequential_full_read",
  "reviewed_commit": "<40-hex Korean repository commit read by reviewer>",
  "english_sha": "a98d9ce29f361d666ec23da0dcfd351f24537ffd",
  "source_files": ["translations/korean/messages/msgsecNNN.toml"],
  "ids": ["..."],
  "edit_ids": ["..."],
  "exclusions": ["..."],
  "packet_sha256": "<SHA-256 of deterministic review packet bytes>",
  "reviewer": "<reviewer/model identity>",
  "reviewed_at": "YYYY-MM-DD"
}
```

`ids` is the exact ordered population presented for review. `edit_ids` and `exclusions` must be subsets of `ids` and must not overlap. The contextual KEEP population is reconstructed as `ids - edit_ids - exclusions`; storing only aggregate counts is forbidden.

## Evidence rules

1. A scope is valid only if `build-beta1-review-packet.py` can regenerate the exact packet from `(reviewed_commit, english_sha, ids, source_files)` and reproduce `packet_sha256`.
2. Stale evaluation is **ID-granular, never scope-granular**. The ledger builder reconstructs each ID from `reviewed_commit` and compares its language basis against current JP/EN/KO. One changed ID cannot invalidate its unchanged siblings.
3. `language_basis = SHA256(JP + NUL + pinned_EN + NUL + KO)`. Only a mismatch here produces `CONTEXT_STALE`.
4. Layout/consumer/runtime changes are structural evidence. They may set `LAYOUT_RECHECK` or other flags but do not erase unchanged language review.
5. `edit_ids` are not promoted to KEEP by the scope. Their later reviewed edit manifest/semantic commit provides `CONTEXT_EDIT` evidence.
6. Scanner hits, mechanical migrations, and AI verdicts alone never create scope coverage. Every non-excluded scope ID must actually pass through the review packet in source order.
7. Propagation is separate evidence. Exact JP + exact pinned EN + exact KO is required; any English mismatch forbids propagation and forces individual review.
8. A propagated KEEP group must be reviewed with the surrounding context of every group member visible. Propagated KEEP is deliberately over-sampled in second-pass accuracy QA.

## Review packet contract

The packet format is deterministic UTF-8 with LF newlines and fixed field ordering. The packet SHA is therefore not merely a reviewer claim: lightweight CI regenerates the packet from repository history and the pinned English checkout and rejects any mismatch.
