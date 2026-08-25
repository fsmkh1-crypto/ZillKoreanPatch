#!/usr/bin/env python3
"""Import translated JSONL into sparse Korean TOML overlays.

Input JSONL rows:
  {"section":3,"id":"30000","korean":"..."}

Canonical Japanese text is looked up locally, so bulk translation payloads do not
need to duplicate it. Fixed runtime controls are validated. ``<line-break>`` is
build-owned layout and must not appear in translator-owned ``korean`` text.

Example:
  python3 tools/korean/import-translations.py batch.jsonl --out /tmp/korean-import
"""
from __future__ import annotations

import argparse
import json
from pathlib import Path
import tomllib

from control_tags import fixed_tokens

ROOT = Path(__file__).resolve().parents[2]
CANON = ROOT / "translations" / "messages"
LINE_BREAK = "<line-break>"


def q(s: str) -> str:
    # JSON strings are valid TOML basic strings for the characters used here.
    return json.dumps(s, ensure_ascii=False)


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("input", type=Path)
    ap.add_argument("--out", type=Path, default=ROOT / "translations" / "korean" / "imported")
    args = ap.parse_args()

    rows: dict[int, list[tuple[str, str]]] = {}
    seen: set[tuple[int, str]] = set()
    with args.input.open(encoding="utf-8") as f:
        for lineno, line in enumerate(f, 1):
            if not line.strip():
                continue
            obj = json.loads(line)
            section, rid, ko = int(obj["section"]), str(obj["id"]), str(obj["korean"])
            key = (section, rid)
            if key in seen:
                raise SystemExit(f"line {lineno}: duplicate {section}/{rid}")
            seen.add(key)
            rows.setdefault(section, []).append((rid, ko))

    args.out.mkdir(parents=True, exist_ok=True)
    total = 0
    for section, items in sorted(rows.items()):
        cpath = CANON / f"msgsec{section:03d}.toml"
        if not cpath.exists():
            raise SystemExit(f"missing canonical section {section}: {cpath}")
        with cpath.open("rb") as f:
            canonical = tomllib.load(f)
        out = ["# SPDX-License-Identifier: CC-BY-SA-4.0", ""]
        for rid, ko in items:
            rec = canonical.get(rid)
            if rec is None or "japanese" not in rec:
                raise SystemExit(f"unknown canonical id {section}/{rid}")
            ja = str(rec["japanese"])
            if not ko:
                raise SystemExit(f"empty Korean translation {section}/{rid}")
            if LINE_BREAK in ko:
                raise SystemExit(
                    f"semantic Korean {section}/{rid} contains {LINE_BREAK}; "
                    "omit layout breaks from translation output"
                )
            if fixed_tokens(ja) != fixed_tokens(ko):
                raise SystemExit(
                    f"fixed-control mismatch {section}/{rid}: "
                    f"{fixed_tokens(ja)} != {fixed_tokens(ko)}"
                )
            out += [f"[{q(rid)}]", f"japanese = {q(ja)}", f"korean = {q(ko)}", ""]
            total += 1
        opath = args.out / f"msgsec{section:03d}.toml"
        opath.write_text("\n".join(out), encoding="utf-8")
        print(f"wrote {opath}: {len(items)} records")
    print(f"imported {total} records")


if __name__ == "__main__":
    main()
