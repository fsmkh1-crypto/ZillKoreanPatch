#!/usr/bin/env python3
"""QA-3 repeated-source Korean consistency analyzer.

Group accepted Korean records by normalized Japanese semantic source. Build-owned
<line-break> layout markers and whitespace are ignored for grouping, while all
other source text and runtime control tokens remain significant. Groups whose
Korean semantic translations differ are reported for human review; no corpus
text is modified.
"""
from __future__ import annotations

import argparse
from collections import defaultdict
import json
from pathlib import Path
import re
import tomllib

ROOT = Path(__file__).resolve().parents[2]
KOREAN_DIR = ROOT / "translations" / "korean" / "messages"
SECTION_FILE_RE = re.compile(r"^msgsec(\d{3})(?:(?:-part\d+)|b)?\.toml$")
LINE_BREAK = "<line-break>"
WS_RE = re.compile(r"[\s\u3000]+")


def normalize_source(text: str) -> str:
    return WS_RE.sub("", text.replace(LINE_BREAK, ""))


def normalize_korean_for_exact_variant(text: str) -> str:
    # Ignore only incidental whitespace. Punctuation, wording, honorifics, and
    # fixed controls remain significant because they can encode real differences.
    return WS_RE.sub(" ", text).strip()


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--json", type=Path)
    ap.add_argument("--max-examples", type=int, default=40)
    args = ap.parse_args()

    groups: dict[str, list[dict[str, object]]] = defaultdict(list)
    seen_ids: set[int] = set()

    for path in sorted(KOREAN_DIR.glob("msgsec*.toml")):
        m = SECTION_FILE_RE.match(path.name)
        if not m:
            continue
        section = int(m.group(1))
        with path.open("rb") as f:
            data = tomllib.load(f)
        for rid, rec in data.items():
            try:
                numeric = int(rid)
            except ValueError:
                continue
            if numeric in seen_ids:
                continue
            seen_ids.add(numeric)
            ja = rec.get("japanese")
            ko = rec.get("korean")
            if not isinstance(ja, str) or not isinstance(ko, str):
                continue
            key = normalize_source(ja)
            groups[key].append({
                "section": section,
                "id": str(rid),
                "path": path.name,
                "japanese": ja,
                "korean": ko,
            })

    repeated = {k: rows for k, rows in groups.items() if len(rows) >= 2}
    inconsistent: list[dict[str, object]] = []
    inconsistent_records = 0
    for key, rows in repeated.items():
        variants: dict[str, list[dict[str, object]]] = defaultdict(list)
        for row in rows:
            variants[normalize_korean_for_exact_variant(str(row["korean"]))].append(row)
        if len(variants) < 2:
            continue
        inconsistent_records += len(rows)
        ranked = sorted(variants.items(), key=lambda kv: (-len(kv[1]), kv[0]))
        inconsistent.append({
            "normalized_japanese": key,
            "occurrences": len(rows),
            "variant_count": len(ranked),
            "variants": [
                {
                    "korean": variant,
                    "count": len(vrows),
                    "records": [
                        {"section": r["section"], "id": r["id"], "path": r["path"]}
                        for r in vrows
                    ],
                }
                for variant, vrows in ranked
            ],
            "japanese_example": rows[0]["japanese"],
        })

    inconsistent.sort(key=lambda g: (-int(g["occurrences"]), -int(g["variant_count"]), str(g["normalized_japanese"])))
    report = {
        "schema": 1,
        "unique_records": len(seen_ids),
        "normalized_source_groups": len(groups),
        "repeated_source_groups": len(repeated),
        "repeated_source_records": sum(len(v) for v in repeated.values()),
        "inconsistent_groups": len(inconsistent),
        "inconsistent_records": inconsistent_records,
        "groups": inconsistent,
    }
    if args.json:
        args.json.parent.mkdir(parents=True, exist_ok=True)
        args.json.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    print("Korean QA-3 repeated-source consistency scan")
    for key in (
        "unique_records", "normalized_source_groups", "repeated_source_groups",
        "repeated_source_records", "inconsistent_groups", "inconsistent_records",
    ):
        print(f"  {key}: {report[key]}")
    for row in inconsistent[: max(args.max_examples, 0)]:
        compact = {
            "occurrences": row["occurrences"],
            "variant_count": row["variant_count"],
            "japanese_example": row["japanese_example"],
            "variants": [
                {"korean": v["korean"], "count": v["count"], "records": v["records"][:5]}
                for v in row["variants"]
            ],
        }
        print("  example: " + json.dumps(compact, ensure_ascii=False, sort_keys=True))


if __name__ == "__main__":
    main()
