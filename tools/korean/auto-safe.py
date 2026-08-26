#!/usr/bin/env python3
"""Emit only high-confidence automatic Korean translations.

Sources:
1) exact translation-memory propagation when every accepted Korean rendering for
   the same canonical Japanese source agrees;
2) a deliberately tiny whitelist of full-string numeric/date patterns.

All emitted rows are still expected to pass tools/korean/apply-results.py, which
remains the authoritative fail-closed gate for duplicate IDs and fixed controls.
"""
from __future__ import annotations

import argparse
import json
from collections import defaultdict
from pathlib import Path
import re
import tomllib
from typing import Callable, Match

from control_tags import fixed_tokens

ROOT = Path(__file__).resolve().parents[2]
KOREAN = ROOT / "translations" / "korean" / "messages"
SECTION_RE = re.compile(r"^msgsec(\d{3})")
LINE_BREAK = "<line-break>"

PatternRenderer = Callable[[Match[str]], str]

# Only context-independent, full-string patterns belong here. Proper names and
# place names are intentionally excluded unless separately whitelisted.
SAFE_PATTERNS: list[tuple[re.Pattern[str], PatternRenderer]] = [
    (re.compile(r"^(\d+)月(\d+)日<end>$"), lambda m: f"{m.group(1)}월 {m.group(2)}일<end>"),
    (re.compile(r"^(\d+)月<end>$"), lambda m: f"{m.group(1)}월<end>"),
    (re.compile(r"^第(\d+)章<end>$"), lambda m: f"제{m.group(1)}장<end>"),
    (re.compile(r"^(\d+)個<end>$"), lambda m: f"{m.group(1)}개<end>"),
]


def section_from_path(path: Path) -> int:
    match = SECTION_RE.match(path.stem)
    if not match:
        raise SystemExit(f"cannot infer section from Korean overlay filename: {path}")
    return int(match.group(1))


def load_translation_memory() -> dict[str, str]:
    candidates: dict[str, set[str]] = defaultdict(set)
    for path in sorted(KOREAN.glob("msgsec*.toml")):
        section_from_path(path)
        with path.open("rb") as f:
            data = tomllib.load(f)
        for rec in data.values():
            if not isinstance(rec, dict):
                continue
            ja = rec.get("japanese")
            ko = rec.get("korean")
            if ja is None or not ko:
                continue
            ja_s, ko_s = str(ja), str(ko)
            if LINE_BREAK in ko_s:
                continue
            if fixed_tokens(ja_s) != fixed_tokens(ko_s):
                continue
            candidates[ja_s].add(ko_s)
    return {ja: next(iter(values)) for ja, values in candidates.items() if len(values) == 1}


def safe_pattern(japanese: str) -> str | None:
    for regex, render in SAFE_PATTERNS:
        match = regex.fullmatch(japanese)
        if not match:
            continue
        korean = render(match)
        if fixed_tokens(japanese) == fixed_tokens(korean):
            return korean
    return None


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("packet", type=Path)
    ap.add_argument("--out", type=Path)
    args = ap.parse_args()

    tm = load_translation_memory()
    rows: list[str] = []
    seen: set[tuple[int, str]] = set()
    with args.packet.open(encoding="utf-8") as f:
        for lineno, line in enumerate(f, 1):
            if not line.strip():
                continue
            obj = json.loads(line)
            section = int(obj["section"])
            rid = str(obj["id"])
            japanese = str(obj["japanese"])
            key = (section, rid)
            if key in seen:
                raise SystemExit(f"{args.packet}:{lineno}: duplicate packet id {section}/{rid}")
            seen.add(key)

            korean = tm.get(japanese)
            if korean is None:
                korean = safe_pattern(japanese)
            if korean is None:
                continue
            if fixed_tokens(japanese) != fixed_tokens(korean):
                raise SystemExit(f"{args.packet}:{lineno}: internal fixed-control mismatch {section}/{rid}")
            rows.append(json.dumps({"section": section, "id": rid, "korean": korean}, ensure_ascii=False))

    payload = "\n".join(rows) + ("\n" if rows else "")
    if args.out:
        args.out.parent.mkdir(parents=True, exist_ok=True)
        args.out.write_text(payload, encoding="utf-8")
    else:
        print(payload, end="")


if __name__ == "__main__":
    main()
