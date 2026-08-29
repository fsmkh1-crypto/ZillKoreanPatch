#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""Inventory runtime <value:$XX> uses in the accepted Korean corpus.

This is an evidence/audit tool, not a safety validator. A <value> token can be
used as an inline rendered value, an <if> predicate operand, or a <select>
selector. Those roles are counted separately so control-flow reads are not
mistaken for text expansion.
"""

from __future__ import annotations

import collections
import glob
import pathlib
import re
import tomllib

ROOT = pathlib.Path(__file__).resolve().parents[2]
VALUE_RE = re.compile(r"<value:\$([0-9A-Fa-f]{2})>")
# Concrete blind spot raised during independent review: a value used as the
# right-hand side of a comparison would not be immediately preceded by <if> and
# the lightweight classifier below would otherwise call it inline.
COMPARISON_VALUE_RE = re.compile(
    r"<(?:equal|less-equal|greater-equal)><value:\$([0-9A-Fa-f]{2})>"
)


def load_toml(path: pathlib.Path):
    with path.open("rb") as f:
        return tomllib.load(f)


def id_set(values):
    return {int(v) for v in values}


def compact(text: str, limit: int = 180) -> str:
    text = text.replace("\r", " ").replace("\n", " ").replace("\t", " ")
    text = " ".join(text.split())
    return text if len(text) <= limit else text[: limit - 1] + "…"


def classify_value_uses(text: str):
    """Return (all, inline, predicate, selector, inline-adjacencies).

    This is deliberately a lightweight corpus classifier, not the message
    grammar parser. Predicate values immediately after <if> and selector values
    immediately after <select> are classified structurally enough for the
    current corpus; explicit counterexample scans below prevent the known
    right-hand-comparison blind spot from silently contaminating headline counts.
    """
    all_tags = []
    inline = []
    predicate = []
    selector = []
    adjacency = []
    for match in VALUE_RE.finditer(text):
        opcode = match.group(1).upper()
        all_tags.append(opcode)
        prefix = text[: match.start()]
        if prefix.endswith("<if>"):
            predicate.append(opcode)
            continue
        if prefix.endswith("<select>"):
            selector.append(opcode)
            continue
        inline.append(opcode)
        if match.end() < len(text):
            ch = text[match.end()]
            if ch != "<" and not ch.isspace():
                adjacency.append((opcode, ch))
    return all_tags, inline, predicate, selector, adjacency


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
    inline_counts = collections.Counter()
    predicate_counts = collections.Counter()
    selector_counts = collections.Counter()
    consumer_counts = collections.Counter()
    inline_consumer_counts = collections.Counter()
    adjacency_counts = collections.Counter()
    classifier_counterexamples = []
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
            mid = int(key)
            rhs_values = [m.group(1).upper() for m in COMPARISON_VALUE_RE.finditer(korean)]
            if rhs_values:
                classifier_counterexamples.append((mid, rhs_values, korean))
            tags, inline, predicate, selector, adjacency = classify_value_uses(korean)
            if not tags:
                continue
            opcode_counts.update(tags)
            inline_counts.update(inline)
            predicate_counts.update(predicate)
            selector_counts.update(selector)
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
            if inline:
                inline_consumer_counts.update(labels)
            adjacency_counts.update(op for op, _ in adjacency)
            rows.append(
                {
                    "id": mid,
                    "tags": tags,
                    "inline": inline,
                    "predicate": predicate,
                    "selector": selector,
                    "labels": labels,
                    "category": category_for(mid),
                    "adjacency": adjacency,
                    "japanese": japanese if isinstance(japanese, str) else "",
                    "korean": korean,
                }
            )

    inline_records = sum(1 for row in rows if row["inline"])
    predicate_records = sum(1 for row in rows if row["predicate"])
    selector_records = sum(1 for row in rows if row["selector"])

    print("DYNAMIC_SUBSTITUTION_AUDIT")
    print(f"accepted_korean_records={total_records}")
    print(f"records_with_value_tags={len(rows)}")
    print(f"value_tag_occurrences={sum(opcode_counts.values())}")
    print(f"distinct_value_opcodes={len(opcode_counts)}")
    print(
        "value_roles="
        f"inline:{sum(inline_counts.values())},predicate:{sum(predicate_counts.values())},"
        f"selector:{sum(selector_counts.values())}"
    )
    print(
        "role_records="
        f"inline:{inline_records},predicate:{predicate_records},selector:{selector_records}"
    )
    print(f"inline_nonspace_adjacencies={sum(adjacency_counts.values())}")
    print(f"classifier_rhs_value_counterexamples={len(classifier_counterexamples)}")
    for mid, opcodes, text in classifier_counterexamples[:20]:
        print(
            f"CLASSIFIER_RHS_VALUE_COUNTEREXAMPLE id={mid} "
            f"opcodes={','.join('$'+op for op in opcodes)} KO={compact(text)!r}"
        )
    print("consumer_overlap=" + ", ".join(f"{k}:{v}" for k, v in sorted(consumer_counts.items())))
    print("inline_consumer_overlap=" + ", ".join(f"{k}:{v}" for k, v in sorted(inline_consumer_counts.items())))
    print("opcodes=" + ", ".join(f"${k}:{v}" for k, v in opcode_counts.most_common()))
    print("inline_opcodes=" + ", ".join(f"${k}:{v}" for k, v in inline_counts.most_common()))
    print("predicate_opcodes=" + ", ".join(f"${k}:{v}" for k, v in predicate_counts.most_common()))
    print("selector_opcodes=" + ", ".join(f"${k}:{v}" for k, v in selector_counts.most_common()))
    if adjacency_counts:
        print("inline_adjacency_opcodes=" + ", ".join(f"${k}:{v}" for k, v in adjacency_counts.most_common()))

    print("OPCODE_BRIEF_BEGIN")
    for opcode in sorted(opcode_counts):
        candidates = [row for row in rows if opcode in row["tags"]]
        row = candidates[0]
        print(
            f"${opcode} all={opcode_counts[opcode]} inline={inline_counts[opcode]} "
            f"predicate={predicate_counts[opcode]} selector={selector_counts[opcode]} "
            f"records={len(candidates)} sample_id={row['id']} "
            f"consumers={'+'.join(row['labels'])} category={row['category']} "
            f"JP={compact(row['japanese'], 100)!r} KO={compact(row['korean'], 100)!r}"
        )
    print("OPCODE_BRIEF_END")

    print("OPCODE_CONTEXT_SAMPLES_BEGIN")
    for opcode in sorted(opcode_counts):
        candidates = [row for row in rows if opcode in row["tags"]]
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
        print(
            f"OPCODE ${opcode} all={opcode_counts[opcode]} inline={inline_counts[opcode]} "
            f"predicate={predicate_counts[opcode]} selector={selector_counts[opcode]} records={len(candidates)}"
        )
        for row in chosen:
            print(
                f"  id={row['id']} consumers={'+'.join(row['labels'])} category={row['category']} "
                f"JP={compact(row['japanese'])!r} KO={compact(row['korean'])!r}"
            )
    print("OPCODE_CONTEXT_SAMPLES_END")

    def score(row):
        bounded_weight = sum(
            1 for label in row["labels"] if label != "unmapped-by-audited-fixed-consumers"
        )
        return (
            bool(row["inline"]),
            bounded_weight > 0,
            len(row["inline"]),
            len(row["adjacency"]),
            -row["id"],
        )

    review = sorted(rows, key=score, reverse=True)
    print("HIGH_VALUE_DYNAMIC_RECORDS_BEGIN")
    for row in review[:120]:
        adj = ",".join(f"${op}->{ch}" for op, ch in row["adjacency"]) or "-"
        print(
            f"id={row['id']} consumers={'+'.join(row['labels'])} category={row['category']} "
            f"inline={','.join('$'+t for t in row['inline']) or '-'} "
            f"predicate={','.join('$'+t for t in row['predicate']) or '-'} "
            f"selector={','.join('$'+t for t in row['selector']) or '-'} adjacency={adj}"
        )
    print("HIGH_VALUE_DYNAMIC_RECORDS_END")

    for row in rows:
        if row["id"] == 10010:
            adj = ",".join(f"${op}->{ch}" for op, ch in row["adjacency"]) or "-"
            print(
                "ID10010_DYNAMIC_CONTEXT "
                f"consumers={'+'.join(row['labels'])} category={row['category']} "
                f"inline={','.join('$'+t for t in row['inline']) or '-'} "
                f"predicate={','.join('$'+t for t in row['predicate']) or '-'} "
                f"selector={','.join('$'+t for t in row['selector']) or '-'} adjacency={adj} "
                f"JP={compact(row['japanese'])!r} KO={compact(row['korean'])!r}"
            )
            break


if __name__ == "__main__":
    main()
