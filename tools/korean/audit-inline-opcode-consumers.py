#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""Summarize inline <value:$XX> uses by audited consumer and category.

This is a corpus triage report, not a runtime contract proof. It intentionally
uses the same lightweight role rule as audit-dynamic-substitutions.py so its
counts can be compared directly with that audit's headline numbers.
"""

from __future__ import annotations

import collections
import glob
import pathlib
import re
import tomllib

ROOT = pathlib.Path(__file__).resolve().parents[2]
VALUE_RE = re.compile(r"<value:\$([0-9A-Fa-f]{2})>")
# Freeze-adjacent records are reporting anchors only. Inclusion here does not
# elevate any record or opcode to a root cause.
FOCUS_IDS = (10010,)


def load_toml(path: pathlib.Path):
    with path.open("rb") as f:
        return tomllib.load(f)


def id_set(values):
    return {int(v) for v in values}


def inline_opcodes(text: str):
    out = []
    for match in VALUE_RE.finditer(text):
        prefix = text[: match.start()]
        if prefix.endswith("<if>") or prefix.endswith("<select>"):
            continue
        out.append(match.group(1).upper())
    return out


def compact(text: str, limit: int = 160) -> str:
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

    occurrence_consumers: dict[str, collections.Counter[str]] = collections.defaultdict(collections.Counter)
    record_consumers: dict[str, collections.Counter[str]] = collections.defaultdict(collections.Counter)
    categories_by_opcode: dict[str, collections.Counter[str]] = collections.defaultdict(collections.Counter)
    record_ids: dict[str, set[int]] = collections.defaultdict(set)
    direct_occurrences: collections.Counter[str] = collections.Counter()
    c5_occurrences: collections.Counter[str] = collections.Counter()
    c5_record_ids: dict[str, set[int]] = collections.defaultdict(set)
    c5_categories: dict[str, collections.Counter[str]] = collections.defaultdict(collections.Counter)
    focus_rows: dict[int, dict[str, object]] = {}
    category_opcode_rows: dict[tuple[str, str], list[dict[str, object]]] = collections.defaultdict(list)

    for filename in sorted(glob.glob(str(ROOT / "translations/korean/messages/msgsec*.toml"))):
        data = load_toml(pathlib.Path(filename))
        for key, record in data.items():
            if not isinstance(record, dict):
                continue
            text = record.get("korean")
            if not isinstance(text, str) or not text:
                continue
            opcodes = inline_opcodes(text)
            if not opcodes:
                continue
            mid = int(key)
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
            category = category_for(mid)
            row = {
                "id": mid,
                "opcodes": opcodes,
                "labels": labels,
                "category": category,
                "japanese": record.get("japanese") if isinstance(record.get("japanese"), str) else "",
                "korean": text,
            }

            if mid in FOCUS_IDS:
                focus_rows[mid] = row

            for opcode in opcodes:
                direct_occurrences[opcode] += 1
                record_ids[opcode].add(mid)
                occurrence_consumers[opcode].update(labels)
                category_opcode_rows[(category, opcode)].append(row)
                if mid in c5:
                    c5_occurrences[opcode] += 1
                    c5_record_ids[opcode].add(mid)
            for opcode in set(opcodes):
                record_consumers[opcode].update(labels)
                categories_by_opcode[opcode][category] += 1
                if mid in c5:
                    c5_categories[opcode][category] += 1

    print("INLINE_OPCODE_CONSUMER_AUDIT")
    for opcode in sorted(record_ids):
        print(
            f"OPCODE ${opcode} inline_occurrences={direct_occurrences[opcode]} records={len(record_ids[opcode])} "
            f"consumer_records=" + ",".join(f"{k}:{v}" for k, v in sorted(record_consumers[opcode].items())) + " "
            f"categories=" + ",".join(f"{k}:{v}" for k, v in categories_by_opcode[opcode].most_common())
        )

    print("C5_INLINE_OPCODE_AUDIT")
    for opcode in sorted(c5_record_ids):
        bound = "16" if opcode == "28" else "unknown"
        print(
            f"C5_OPCODE ${opcode} inline_occurrences={c5_occurrences[opcode]} records={len(c5_record_ids[opcode])} "
            f"known_max_encoded_bytes={bound} categories="
            + ",".join(f"{k}:{v}" for k, v in c5_categories[opcode].most_common())
        )

    print("FOCUS_INLINE_OPCODE_AUDIT")
    for mid in FOCUS_IDS:
        row = focus_rows.get(mid)
        if row is None:
            print(f"FOCUS_ID id={mid} inline_values=none-or-record-missing")
            continue
        opcodes = row["opcodes"]
        labels = row["labels"]
        category = str(row["category"])
        assert isinstance(opcodes, list)
        assert isinstance(labels, list)
        print(
            f"FOCUS_ID id={mid} inline_opcodes=" + ",".join(f"${opcode}" for opcode in opcodes)
            + " consumers=" + ",".join(str(label) for label in labels)
            + f" category={category}"
        )
        print(
            f"FOCUS_TEXT id={mid} JP={compact(str(row['japanese']))!r} "
            f"KO={compact(str(row['korean']))!r}"
        )
        for opcode in sorted(set(str(opcode) for opcode in opcodes)):
            peers = category_opcode_rows[(category, opcode)]
            unique_peers = {int(peer["id"]): peer for peer in peers}
            ordered = [unique_peers[peer_id] for peer_id in sorted(unique_peers)]
            ids = ",".join(str(peer["id"]) for peer in ordered[:40])
            if len(ordered) > 40:
                ids += ",..."
            print(
                f"FOCUS_CATEGORY_OPCODE id={mid} opcode=${opcode} category={category} "
                f"occurrences={len(peers)} records={len(ordered)} ids={ids or '-'}"
            )
            for peer in ordered[:12]:
                print(
                    f"  FOCUS_PEER id={peer['id']} consumers={'+'.join(str(v) for v in peer['labels'])} "
                    f"JP={compact(str(peer['japanese']), 120)!r} KO={compact(str(peer['korean']), 120)!r}"
                )

    print("NOTE consumer labels may overlap; this report classifies corpus use, not runtime substitution source or maximum length")
    print("NOTE C5 membership comes from retained release/layout/consumer-map.toml evidence; it is not independent retail-runtime proof")
    print("NOTE focus-category peers are semantic triage only; category co-membership does not prove a shared runtime source or buffer")


if __name__ == "__main__":
    main()
