#!/usr/bin/env python3
import argparse
import json
import subprocess
import tomllib
from collections import defaultdict
from pathlib import Path


def git_sha(root: Path) -> str:
    return subprocess.check_output(["git", "-C", str(root), "rev-parse", "HEAD"], text=True).strip()


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


def canonicalize(
    rid: str,
    rows: list[dict],
    english: dict[str, dict] | None,
    korean_sha: str,
    english_sha: str,
) -> tuple[dict, bool]:
    first = rows[0]
    for row in rows[1:]:
        for key in ("japanese", "korean", "layout", "consumer"):
            if row.get(key, "") != first.get(key, ""):
                raise ValueError(f"non-identical Korean alias {rid} key={key}")

    japanese_mismatch = False
    if english is not None:
        ej = english.get("english_japanese", "")
        japanese_mismatch = bool(ej and first.get("japanese", "") and ej != first.get("japanese", ""))

    result = {
        "id": rid,
        "paths": [row["path"] for row in rows],
        "japanese": first.get("japanese", ""),
        "english": "" if english is None else english.get("english", ""),
        "korean": first.get("korean", ""),
        "layout": first.get("layout", ""),
        "consumer": first.get("consumer", ""),
        "english_file": "" if english is None else english.get("english_file", ""),
        "japanese_mismatch": japanese_mismatch,
        "korean_sha": korean_sha,
        "english_sha": english_sha,
    }
    return result, japanese_mismatch


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--english-root", required=True, type=Path)
    parser.add_argument("--jsonl", required=True, type=Path)
    parser.add_argument("--summary", required=True, type=Path)
    args = parser.parse_args()

    root = Path(__file__).resolve().parents[2]
    english_root = args.english_root.resolve()
    korean_sha = git_sha(root)
    english_sha = git_sha(english_root)
    english = load_english(english_root)
    korean = load_korean(root)

    records = []
    missing_english = []
    japanese_mismatch_ids = []
    for rid in sorted(korean, key=lambda x: int(x)):
        erow = english.get(rid)
        if erow is None:
            missing_english.append(rid)
        record, mismatch = canonicalize(rid, korean[rid], erow, korean_sha, english_sha)
        records.append(record)
        if mismatch:
            japanese_mismatch_ids.append(rid)

    with args.jsonl.open("w", encoding="utf-8") as fh:
        for record in records:
            fh.write(json.dumps(record, ensure_ascii=False) + "\n")

    summary = {
        "status": "READ_ONLY_EXPORT",
        "korean_sha": korean_sha,
        "english_sha": english_sha,
        "accepted_korean_ids": len(records),
        "physical_korean_rows": sum(len(v) for v in korean.values()),
        "alias_extra_rows": sum(len(v) - 1 for v in korean.values()),
        "english_ids": len(english),
        "missing_english_ids": missing_english,
        "japanese_mismatch_ids": japanese_mismatch_ids,
        "records_with_layout": sum(bool(r["layout"]) for r in records),
        "records_with_consumer": sum(bool(r["consumer"]) for r in records),
    }
    args.summary.write_text(json.dumps(summary, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(summary, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
