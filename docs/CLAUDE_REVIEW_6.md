# Claude Review 6 — unattended bulk translation gate

Reviewed branch/head: `translation/section001-batch2` @ `6d545615ef74d2ace3a6537b0635d6d34f95a594`

Decision returned by Claude:

`PASS FOR UNATTENDED BULK TRANSLATION`

## Findings

### SHOULD-FIX — packet publication coordination

`korean-apply-results.yml` and `korean-next-packet.yml` used separate concurrency groups while both could publish `work/korean-packets/current.jsonl` to `korean-source-chunks`. A slower workflow could therefore publish a stale, redundant packet after a newer one. Existing result application is idempotent, so this was assessed as wasted work rather than corruption or data loss.

Resolution after review:

- both workflows now use the shared concurrency group `korean-translation-pipeline`;
- both use `cancel-in-progress: false`, so pipeline runs serialize instead of cancelling one another;
- each workflow still checks out the current translation branch and rebases its writes before push.

### NOTE — explicit font content hash

The raster workflows already fetched UnDotum from a commit-SHA-addressed GitHub URL, but did not independently verify the downloaded bytes.

Resolution after review:

- both translation-ingestion/raster paths now verify `/tmp/UnDotum.ttf` against SHA-256 `5b8373e126bb61f59105cf7f54a47eb1b089c2b0aacb70c6cd688bd8ea76cdc9` before raster generation.

## Verified by Claude

- result packets are serialized safely within `korean-apply-results.yml` and are not silently lost when another result packet arrives mid-run;
- raster synchronization is owned by the automated ingestion path;
- `apply-results.py`, `refresh-japanese-refs.py`, and `next-packet.py` fail closed on stale Japanese/control-token violations;
- semantic `korean` text cannot contain `<line-break>` and layout ownership is enforced in both Python and Go;
- the `korean-check` accepted-record count and `next-packet.py` `records_translated` metric intentionally count different sets.

## Remaining production ISO blockers

This PASS covers unattended bulk translation only. It does not establish production ISO readiness. Remaining high-value gates include full Korean integration in `internal/release/build.go::Build`, whole-game renderer-slot capacity/safety at the completed corpus size, complete authentication inventory for all modified retail assets, and broad PPSSPP visual verification.
