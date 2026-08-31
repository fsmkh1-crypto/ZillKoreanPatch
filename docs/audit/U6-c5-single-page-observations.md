# U6 C5 single-page unresolved observations

## Scope

This record keeps the current C5 single-page validation findings inside the ongoing U6 audit trail. It is documentation-only: no Korean translation, derived layout, glyph mapping, storage validator, runtime patch payload, or consumer behavior is changed by this record.

The findings are intentionally **not repaired** at this stage. They are preserved as unresolved audit evidence rather than converted into speculative Korean-only edits.

## Observed validation failures

The asset-backed C5 validation reached 15,679 records and reported **9 message IDs / 15 branches** where the derived result occupied `2 pages` while the authenticated consumer contract allows `maximum 1` page.

| Message ID | Branch(es) | Observation |
| ---: | --- | --- |
| 1280007 | b7 | 2 pages > maximum 1 |
| 1280008 | b5, b7 | 2 pages > maximum 1 |
| 1280012 | b1, b2, b5, b6 | 2 pages > maximum 1 |
| 1280017 | b5 | 2 pages > maximum 1 |
| 1280020 | b2, b7 | 2 pages > maximum 1 |
| 1280021 | b3, b5 | 2 pages > maximum 1 |
| 1280043 | b4 | 2 pages > maximum 1 |
| 1280050 | b4 | 2 pages > maximum 1 |
| 1280051 | b4 | 2 pages > maximum 1 |

Total: **9 messages / 15 branches**.

## Interpretation boundary

- The upstream-English C5 consumer/storage contract is real; the `maximum 1` result is not being dismissed as a generic authoring warning.
- The U5 general verified-dialogue reflow path intentionally excludes C5, so these findings must not be generalized into a U5 dialogue-reflow defect.
- The current evidence does not justify guessing a Korean-only break pattern, shortening translation text, or relaxing the validator merely to make the count disappear.
- These findings do not establish a historical freeze root cause and must not be conflated with `$15`, record `10010`, or other prior hypotheses.

## U6 handling decision

Keep all 15 branch findings **unchanged and unresolved** under U6.

Do not:

1. edit the affected Korean strings solely to satisfy this gate,
2. insert speculative branch/page breaks,
3. weaken or bypass the authenticated C5 `max_pages=1` contract,
4. promote the population to a proven runtime failure without device evidence.

Future mutation requires stronger same-consumer evidence: an upstream-English transformation that applies to these exact branches, authenticated asset/runtime evidence establishing the correct Korean projection, or practical device evidence demonstrating an actual failure mode.

Until then, these 9 message IDs / 15 branches remain U6 audit observations and regression targets, not a new U-stage feature/fix.
