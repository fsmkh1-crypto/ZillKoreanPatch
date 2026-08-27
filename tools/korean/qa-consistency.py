#!/usr/bin/env python3
"""QA-3 repeated-source Korean consistency analyzer.

Group accepted Korean records by normalized Japanese semantic source. Build-owned
<line-break> layout markers and whitespace are ignored for grouping, while all
other source text and runtime control tokens remain significant. Groups whose
Korean semantic translations differ are reported for human review. Reviewed
context-dependent differences can be registered as exact record-ID sets; those
remain visible as contextual exceptions while actionable inconsistencies are
reported separately.
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
EXCEPTIONS_PATH = ROOT / "tools" / "korean" / "qa-consistency-exceptions.json"
SECTION_FILE_RE = re.compile(r"^msgsec(\d{3})(?:(?:-part\d+)|b)?\.toml$")
LINE_BREAK = "<line-break>"
WS_RE = re.compile(r"[\s\u3000]+")


def normalize_source(text: str) -> str:
    return WS_RE.sub("", text.replace(LINE_BREAK, ""))


def normalize_korean_for_exact_variant(text: str) -> str:
    # Ignore only incidental whitespace. Punctuation, wording, honorifics, and
    # fixed controls remain significant because they can encode real differences.
    return WS_RE.sub(" ", text).strip()


def group_ids(group: dict[str, object]) -> tuple[str, ...]:
    ids: list[str] = []
    for variant in group["variants"]:  # type: ignore[index]
        for rec in variant["records"]:  # type: ignore[index]
            ids.append(str(rec["id"]))
    return tuple(sorted(ids, key=int))


def load_exceptions() -> dict[tuple[str, ...], dict[str, object]]:
    if not EXCEPTIONS_PATH.exists():
        return {}
    raw = json.loads(EXCEPTIONS_PATH.read_text(encoding="utf-8"))
    if not isinstance(raw, list):
        raise SystemExit(f"{EXCEPTIONS_PATH}: expected a JSON list")
    out: dict[tuple[str, ...], dict[str, object]] = {}
    for idx, item in enumerate(raw):
        if not isinstance(item, dict) or not isinstance(item.get("ids"), list):
            raise SystemExit(f"{EXCEPTIONS_PATH}: invalid entry #{idx}")
        ids = tuple(sorted((str(v) for v in item["ids"]), key=int))
        if len(ids) < 2 or len(set(ids)) != len(ids):
            raise SystemExit(f"{EXCEPTIONS_PATH}: invalid ids in entry #{idx}: {ids}")
        if ids in out:
            raise SystemExit(f"{EXCEPTIONS_PATH}: duplicate exception ids: {ids}")
        out[ids] = item
    return out


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

    exceptions = load_exceptions()
    matched_exception_ids: set[tuple[str, ...]] = set()
    contextual: list[dict[str, object]] = []
    actionable: list[dict[str, object]] = []
    for group in inconsistent:
        ids = group_ids(group)
        exception = exceptions.get(ids)
        if exception is None:
            actionable.append(group)
            continue
        matched_exception_ids.add(ids)
        contextual.append({
            **group,
            "exception_category": exception.get("category", "contextual"),
            "exception_reason": exception.get("reason", "reviewed contextual difference"),
        })

    stale = sorted(set(exceptions) - matched_exception_ids)
    if stale:
        rendered = [list(ids) for ids in stale]
        raise SystemExit(f"stale or mismatched QA-3 exception registry entries: {rendered}")

    actionable_records = sum(int(g["occurrences"]) for g in actionable)
    contextual_records = sum(int(g["occurrences"]) for g in contextual)
    report = {
        "schema": 2,
        "unique_records": len(seen_ids),
        "normalized_source_groups": len(groups),
        "repeated_source_groups": len(repeated),
        "repeated_source_records": sum(len(v) for v in repeated.values()),
        "inconsistent_groups": len(inconsistent),
        "inconsistent_records": inconsistent_records,
        "actionable_inconsistent_groups": len(actionable),
        "actionable_inconsistent_records": actionable_records,
        "contextual_exception_groups": len(contextual),
        "contextual_exception_records": contextual_records,
        # Preserve the historical raw list for consumers that inspect all variants.
        "groups": inconsistent,
        "actionable_groups": actionable,
        "contextual_exceptions": contextual,
    }
    if args.json:
        args.json.parent.mkdir(parents=True, exist_ok=True)
        args.json.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    print("Korean QA-3 repeated-source consistency scan")
    for key in (
        "unique_records", "normalized_source_groups", "repeated_source_groups",
        "repeated_source_records", "inconsistent_groups", "inconsistent_records",
        "actionable_inconsistent_groups", "actionable_inconsistent_records",
        "contextual_exception_groups", "contextual_exception_records",
    ):
        print(f"  {key}: {report[key]}")
    for row in actionable[: max(args.max_examples, 0)]:
        compact = {
            "occurrences": row["occurrences"],
            "variant_count": row["variant_count"],
            "japanese_example": row["japanese_example"],
            "variants": [
                {"korean": v["korean"], "count": v["count"], "records": v["records"][:5]}
                for v in row["variants"]
            ],
        }
        print("  actionable_example: " + json.dumps(compact, ensure_ascii=False, sort_keys=True))
    for row in contextual[: max(args.max_examples, 0)]:
        compact = {
            "ids": list(group_ids(row)),
            "category": row["exception_category"],
            "reason": row["exception_reason"],
            "japanese_example": row["japanese_example"],
        }
        print("  contextual_exception: " + json.dumps(compact, ensure_ascii=False, sort_keys=True))


if __name__ == "__main__":
    main()
