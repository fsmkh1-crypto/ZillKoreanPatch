#!/usr/bin/env python3
"""Import translated JSONL into sparse Korean TOML overlays.

Input JSONL rows:
  {"section":3,"id":"30000","korean":"..."}

Canonical Japanese text is looked up locally, so bulk translation payloads do not
need to duplicate it. Output remains compatible with the current Korean loader:
entries contain both canonical `japanese` and translated `korean`.

Example:
  python3 tools/korean/import-translations.py batch.jsonl --out /tmp/korean-import
"""
from __future__ import annotations

import argparse
import json
from pathlib import Path
import re
import tomllib

ROOT = Path(__file__).resolve().parents[2]
CANON = ROOT / "translations" / "messages"
TOKEN_RE = re.compile(r"<(?:end|line-break|value:\$[0-9A-Fa-f]+|if|select|less-equal|equal)>")


def q(s: str) -> str:
    # JSON strings are valid TOML basic strings for the characters used here.
    return json.dumps(s, ensure_ascii=False)


def tokens(s: str) -> list[str]:
    return TOKEN_RE.findall(s)


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
            ja = rec["japanese"]
            if not ko:
                raise SystemExit(f"empty Korean translation {section}/{rid}")
            if tokens(ja) != tokens(ko):
                raise SystemExit(f"control-token mismatch {section}/{rid}: {tokens(ja)} != {tokens(ko)}")
            out += [f"[{q(rid)}]", f"japanese = {q(ja)}", f"korean = {q(ko)}", ""]
            total += 1
        opath = args.out / f"msgsec{section:03d}.toml"
        opath.write_text("\n".join(out), encoding="utf-8")
        print(f"wrote {opath}: {len(items)} records")
    print(f"imported {total} records")


if __name__ == "__main__":
    main()
