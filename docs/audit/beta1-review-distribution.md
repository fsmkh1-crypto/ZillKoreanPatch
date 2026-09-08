# Beta1 Review Workload Distribution

> Generated workload measurement. These classes do not grant review coverage.

- Accepted IDs: **42,016**
- Rows with non-`<end>` controls: **5,021**
- Rows with sentence punctuation: **31,114**
- Conservative rapid-scan candidates: **10,409**

Rapid-scan candidate means visible KO <= 20 and visible JP <= 20, no non-`<end>` controls, and no sentence punctuation. `FIXED_BUFFER` is intentionally not used as a shortcut.

## Visible length distribution

| Length | KO IDs | JP IDs | FIXED_BUFFER within KO bin |
| --- | ---: | ---: | ---: |
| 1-10 | 9,786 | 10,136 | 3,877 |
| 11-20 | 8,870 | 9,006 | 2,935 |
| 21-40 | 11,204 | 11,388 | 5,135 |
| 41-80 | 8,568 | 8,532 | 5,392 |
| 81-160 | 2,989 | 2,443 | 1,797 |
| 161+ | 599 | 511 | 244 |

## Language duplicate-group population

| Group size | IDs belonging to groups of this size |
| ---: | ---: |
| 1 | 33,065 |
| 2 | 4,058 |
| 3 | 972 |
| 4 | 540 |
| 5 | 280 |
| 6 | 168 |
| 7 | 126 |
| 8 | 104 |
| 9 | 81 |
| 10 | 110 |
| 11 | 44 |
| 12 | 96 |
| 13 | 65 |
| 14 | 28 |
| 16 | 48 |
| 17 | 34 |
| 18 | 54 |
| 19 | 19 |
| 21 | 21 |
| 22 | 22 |
| 23 | 46 |
| 33 | 33 |
| 60 | 60 |
| 62 | 62 |
| 94 | 94 |
| 96 | 96 |
| 99 | 99 |
| 101 | 101 |
| 148 | 148 |
| 170 | 170 |
| 1172 | 1,172 |

## Top consumers

| Consumer | IDs |
| --- | ---: |
| `unproven` | 17,169 |
| `c5-portrait` | 15,225 |
| `verified-narrow-dialogue` | 6,458 |
| `bounded-label` | 1,516 |
| `c22` | 697 |
| `item-description` | 420 |
| `c5` | 395 |
| `c5-single-page` | 54 |
| `guild-region` | 42 |
| `guild-commentary` | 40 |

## Top categories

| Category | IDs |
| --- | ---: |
| `unknown` | 26,646 |
| `dialogue` | 7,477 |
| `uncategorized` | 1,959 |
| `objective-advice` | 1,597 |
| `location-route-label` | 913 |
| `character-profile` | 705 |
| `choice` | 347 |
| `combat-action-name` | 269 |
| `equipment-description` | 204 |
| `general-ui-menu` | 160 |
| `equipment-feedback` | 148 |
| `ability-description` | 135 |
| `equipment-name` | 120 |
| `quest-item-description` | 112 |
| `item-effect-description` | 104 |
| `system-help` | 103 |
| `player-name-option` | 100 |
| `cinematic-text` | 99 |
| `inventory-and-status-help` | 87 |
| `guild-ui-voiced-copy` | 86 |
| `chronicle-entry` | 80 |
| `audio-cue-label` | 74 |
| `ability-requirement` | 68 |
| `epithet-component` | 65 |
| `quest-topic-label` | 54 |
| `notification` | 53 |
| `ability-command-name` | 51 |
| `character-creation-choice` | 42 |
| `character-creation-prompt` | 34 |
| `battle-notification` | 33 |

## Planning rule

Do not set IDs/hour from `FIXED_BUFFER` population or from this heuristic alone. Run the first dense-review pilot and replace estimates with measured seconds/ID for rapid-scan, ordinary dialogue, and high-risk rows.
