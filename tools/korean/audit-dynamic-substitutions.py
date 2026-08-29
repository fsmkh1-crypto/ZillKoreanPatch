#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""Inventory runtime <value:$XX> substitutions in the accepted Korean corpus.

This is an evidence/audit tool, not a safety validator. It deliberately reports
where static storage validation stops being a complete proof because the game
supplies bytes at runtime.
"""

from __future__ import annotations

import collections
import glob
import pathlib
import re
import tomllib

ROOT = pathlib.Path(__file__).resolve().parents[2]
VALUE_RE = re.compile(r"<value:\$([0-9A-Fa-f]{2})>")
ADJ_RE = re.compile(r"<value:\$([0-9A-Fa-f]{2})>([^<\s])")


def load_toml(path: pathlib.Path):
    with path.open("rb") as f:
        return tomllib.load(f)


def id_set(values):
    return {int(v) for v in values}


def compact(text: str, limit: int = 180) -> str:
    text = text.replace("\r", " ").replace("\n", " ").replace("\t", " ")
    text = " ".join(text.split())
    return text if len(text) <= limit else text[: limit - 1] + "…"


def main() -> None:
    consumers = load_toml(ROOT / "release/layout/consumer-map.toml")
    c5 = id_set(consumers.get("c5_ids", []))
    single_c5 = id_set(consumers.get("single_page_c5_ids", []))
    c22 = id_set(consumers.get("c22_ids", []))
    bounded = id_set(consumers.get("bounded_label_ids", []))
    guild_client = id_set(consumers.get("guild_client_ids", []))
    guild_region = id_set(consumers.get("guild_region_ids", []))
    c20 = set()
    for group in consumers.get("c20_group", []):
        c20.update(int(v) for v in group.get("ids", []))

    categories = load_toml(ROOT / "release/layout/categories.toml")
    ranges = [
        (int(r["first"]), int(r["last"]), str(r["category"]))
        for r in categories.get("range", [])
    ]

    def category_for(mid: int) -> str:
        names = [name for first, last, name in ranges if first <= mid <= last]
        return ",".join(names) if names else "-"

    rows = []
    opcode_counts = collections.Counter()
    consumer_counts = collections.Counter()
    adjacency_counts = collections.Counter()
    total_records = 0

    for filename in sorted(glob.glob(str(ROOT / "translations/korean/messages/msgsec*.toml"))):
        data = load_toml(pathlib.Path(filename))
        for key, record in data.items():
            if not isinstance(record, dict):
                continue
            korean = record.get("korean")
            japanese = record.get("japanese")
            if not isinstance(korean, str) or not korean:
                continue
            total_records += 1
            tags = VALUE_RE.findall(korean)
            if not tags:
                continue
            mid = int(key)
            tags = [t.upper() for t in tags]
            opcode_counts.update(tags)
            labels = []
            if mid in c5:
                labels.append("C5")
            if mid in single_c5:
                labels.append("C5-single")
            if mid in c22:
                labels.append("C22")
            if mid in c20:
                labels.append("C20")
            if mid in bounded:
                labels.append("bounded-label")
            if mid in guild_client:
                labels.append("guild-client")
            if mid in guild_region:
                labels.append("guild-region")
            if not labels:
                labels.append("unmapped-by-audited-fixed-consumers")
            consumer_counts.update(labels)
            adjacency = [(m.group(1).upper(), m.group(2)) for m in ADJ_RE.finditer(korean)]
            adjacency_counts.update(op for op, _ in adjacency)
            rows.append(
                {
                    "id": mid,
                    "tags": tags,
                    "labels": labels,
                    "category": category_for(mid),
                    "adjacency": adjacency,
                    "japanese": japanese if isinstance(japanese, str) else "",
                    "korean": korean,
                }
            )

    print("DYNAMIC_SUBSTITUTION_AUDIT")
    print(f"accepted_korean_records={total_records}")
    print(f"records_with_value_tags={len(rows)}")
    print(f"value_tag_occurrences={sum(opcode_counts.values())}")
    print(f"distinct_value_opcodes={len(opcode_counts)}")
    print(f"immediate_nonspace_adjacencies={sum(adjacency_counts.values())}")
    print("consumer_overlap=" + ", ".join(f"{k}:{v}" for k, v in sorted(consumer_counts.items())))
    print("opcodes=" + ", ".join(f"${k}:{v}" for k, v in opcode_counts.most_common()))
    if adjacency_counts:
        print("adjacency_opcodes=" + ", ".join(f"${k}:{v}" for k, v in adjacency_counts.most_common()))

    # Representative real-corpus contexts per opcode. These are evidence for
    # semantic reverse engineering, not labels for the opcode contract.
    print("OPCODE_CONTEXT_SAMPLES_BEGIN")
    for opcode in sorted(opcode_counts):
        candidates = [row for row in rows if opcode in row["tags"]]
        # Keep samples deterministic while preferring category/consumer diversity.
        chosen = []
        seen_shapes = set()
        for row in candidates:
            shape = (tuple(row["labels"]), row["category"])
            if shape in seen_shapes and len(chosen) >= 3:
                continue
            seen_shapes.add(shape)
            chosen.append(row)
            if len(chosen) == 6:
                break
        print(f"OPCODE ${opcode} occurrences={opcode_counts[opcode]} records={len(candidates)}")
        for row in chosen:
            print(
                f"  id={row['id']} consumers={'+'.join(row['labels'])} category={row['category']} "
                f"JP={compact(row['japanese'])!r} KO={compact(row['korean'])!r}"
            )
    print("OPCODE_CONTEXT_SAMPLES_END")

    # Highest-value review set: dynamic substitutions at known bounded/display
    # consumers, then records with multiple substitutions, then immediate
    # substitution/text adjacency. This is a triage list, not a verdict.
    def score(row):
        bounded_weight = sum(
            1 for label in row["labels"] if label != "unmapped-by-audited-fixed-consumers"
        )
        return (bounded_weight > 0, len(row["tags"]), len(row["adjacency"]), -row["id"])

    review = sorted(rows, key=score, reverse=True)
    print("HIGH_VALUE_DYNAMIC_RECORDS_BEGIN")
    for row in review[:120]:
        adj = ",".join(f"${op}->{ch}" for op, ch in row["adjacency"]) or "-"
        print(
            f"id={row['id']} consumers={'+'.join(row['labels'])} "
            f"category={row['category']} tags={','.join('$'+t for t in row['tags'])} "
            f"adjacency={adj}"
        )
    print("HIGH_VALUE_DYNAMIC_RECORDS_END")

    # Make the historically interesting record explicit without assigning
    # causality, so CI logs can be compared across future corpus revisions.
    for row in rows:
        if row["id"] == 10010:
            adj = ",".join(f"${op}->{ch}" for op, ch in row["adjacency"]) or "-"
            print(
                "ID10010_DYNAMIC_CONTEXT "
                f"consumers={'+'.join(row['labels'])} category={row['category']} "
                f"tags={','.join('$'+t for t in row['tags'])} adjacency={adj} "
                f"JP={compact(row['japanese'])!r} KO={compact(row['korean'])!r}"
            )
            break


if __name__ == "__main__":
    main()
