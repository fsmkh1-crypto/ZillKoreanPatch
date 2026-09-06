#!/usr/bin/env python3
import argparse
import json
import tomllib
from collections import defaultdict
from pathlib import Path


def load_english(root: Path) -> dict[str, dict]:
    out = {}
    base = root / "translations/messages"
    for path in sorted(base.glob("msgsec*.toml")):
        data = tomllib.loads(path.read_text(encoding="utf-8"))
        for rid, row in data.items():
            if not isinstance(row, dict):
                continue
            rid = str(rid)
            if rid in out:
                raise ValueError(f"duplicate English id {rid}: {path}")
            out[rid] = {
                "english": row.get("english", ""),
                "english_japanese": row.get("japanese", ""),
                "english_file": path.relative_to(root).as_posix(),
            }
    return out


def load_korean(root: Path) -> dict[str, list[dict]]:
    out = defaultdict(list)
    base = root / "translations/korean/messages"
    for path in sorted(base.glob("*.toml")):
        data = tomllib.loads(path.read_text(encoding="utf-8"))
        rel = path.relative_to(root).as_posix()
        for rid, row in data.items():
            if isinstance(row, dict) and isinstance(row.get("korean"), str):
                out[str(rid)].append({"path": rel, **row})
    return out


def canonicalize(rid: str, rows: list[dict], english: dict[str, dict] | None) -> dict:
    first = rows[0]
    for row in rows[1:]:
        for key in ("japanese", "korean", "layout"):
            if row.get(key, "") != first.get(key, ""):
                raise ValueError(f"non-identical Korean alias {rid} key={key}")
    result = {
        "id": rid,
        "paths": [row["path"] for row in rows],
        "japanese": first.get("japanese", ""),
        "english": "" if english is None else english.get("english", ""),
        "korean": first.get("korean", ""),
        "layout": first.get("layout", ""),
        "consumer": first.get("consumer", ""),
        "english_file": "" if english is None else english.get("english_file", ""),
    }
    if english is not None:
        ej = english.get("english_japanese", "")
        if ej and result["japanese"] and ej != result["japanese"]:
            raise ValueError(f"Japanese mismatch with English patch for id {rid}")
    return result


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--english-root", required=True, type=Path)
    parser.add_argument("--jsonl", required=True, type=Path)
    parser.add_argument("--summary", required=True, type=Path)
    args = parser.parse_args()
    root = Path(__file__).resolve().parents[2]
    english = load_english(args.english_root.resolve())
    korean = load_korean(root)
    records = []
    missing_english = []
    for rid in sorted(korean, key=lambda x: int(x)):
        erow = english.get(rid)
        if erow is None:
            missing_english.append(rid)
        records.append(canonicalize(rid, korean[rid], erow))
    with args.jsonl.open("w", encoding="utf-8") as fh:
        for record in records:
            fh.write(json.dumps(record, ensure_ascii=False) + "\n")
    summary = {
        "status": "READ_ONLY_EXPORT",
        "accepted_korean_ids": len(records),
        "physical_korean_rows": sum(len(v) for v in korean.values()),
        "alias_extra_rows": sum(len(v) - 1 for v in korean.values()),
        "english_ids": len(english),
        "missing_english_ids": missing_english,
        "records_with_layout": sum(bool(r["layout"]) for r in records),
        "records_with_consumer": sum(bool(r["consumer"]) for r in records),
    }
    args.summary.write_text(json.dumps(summary, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(summary, ensure_ascii=False, indent=2))
    if len(records) != 42016:
        raise SystemExit(f"expected 42016 accepted Korean IDs, found {len(records)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
