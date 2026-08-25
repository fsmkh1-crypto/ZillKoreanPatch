#!/usr/bin/env python3
"""Export canonical message TOML files into compact JSONL for bulk translation.

Usage:
  python3 tools/korean/export-corpus.py > /tmp/zill-korean-corpus.jsonl
  python3 tools/korean/export-corpus.py --section 3 > /tmp/sec003.jsonl

The compact stream avoids repeatedly fetching huge TOML ranges during translation.
Control tokens remain embedded in `japanese` so translation/import validation can
compare them against the Korean result.
"""
from __future__ import annotations

import argparse
import json
from pathlib import Path
import tomllib

ROOT = Path(__file__).resolve().parents[2]
MESSAGES = ROOT / "translations" / "messages"


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--section", type=int)
    ap.add_argument("--start-id", type=int)
    ap.add_argument("--end-id", type=int)
    args = ap.parse_args()

    paths = sorted(MESSAGES.glob("msgsec*.toml"))
    if args.section is not None:
        paths = [MESSAGES / f"msgsec{args.section:03d}.toml"]

    for path in paths:
        if not path.exists():
            raise SystemExit(f"missing canonical section: {path}")
        section = int(path.stem.removeprefix("msgsec"))
        with path.open("rb") as f:
            data = tomllib.load(f)
        for rid, rec in data.items():
            try:
                numeric = int(rid)
            except ValueError:
                continue
            if args.start_id is not None and numeric < args.start_id:
                continue
            if args.end_id is not None and numeric > args.end_id:
                continue
            japanese = rec.get("japanese")
            if japanese is None:
                continue
            print(json.dumps({"section": section, "id": rid, "japanese": japanese}, ensure_ascii=False, separators=(",", ":")))


if __name__ == "__main__":
    main()
