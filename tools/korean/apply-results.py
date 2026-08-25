#!/usr/bin/env python3
"""Validate and apply one or more Korean result JSONL files to sparse overlays.

Each input row must contain section/id/korean. Canonical Japanese is loaded locally.
New translations are merged into one deterministic msgsecNNN-auto.toml per section.
Re-applying the same translation is a no-op; conflicting existing translations fail.
"""
from __future__ import annotations

import argparse
import json
from pathlib import Path
import re
import tomllib

ROOT = Path(__file__).resolve().parents[2]
CANON = ROOT / "translations" / "messages"
KOREAN = ROOT / "translations" / "korean" / "messages"
TOKEN_RE = re.compile(r"<(?:end|line-break|value:\$[0-9A-Fa-f]+|if|select|less-equal|equal)>")
SECTION_RE = re.compile(r"^msgsec(\d{3})")


def q(s: str) -> str:
    return json.dumps(s, ensure_ascii=False)


def tokens(s: str) -> list[str]:
    return TOKEN_RE.findall(s)


def section_from_path(path: Path) -> int:
    match = SECTION_RE.match(path.stem)
    if not match:
        raise SystemExit(f"cannot infer section from Korean overlay filename: {path}")
    return int(match.group(1))


def canonical_for(section: int) -> dict[str, dict[str, object]]:
    path = CANON / f"msgsec{section:03d}.toml"
    if not path.exists():
        raise SystemExit(f"missing canonical section {section}: {path}")
    with path.open("rb") as f:
        raw = tomllib.load(f)
    return {str(k): v for k, v in raw.items() if isinstance(v, dict)}


def load_existing() -> tuple[dict[tuple[int, str], str], dict[int, dict[str, dict[str, str]]]]:
    owners: dict[tuple[int, str], Path] = {}
    existing: dict[tuple[int, str], str] = {}
    auto: dict[int, dict[str, dict[str, str]]] = {}
    for path in sorted(KOREAN.glob("msgsec*.toml")):
        section = section_from_path(path)
        with path.open("rb") as f:
            data = tomllib.load(f)
        for rid, rec in data.items():
            if not isinstance(rec, dict) or not rec.get("korean"):
                continue
            key = (section, str(rid))
            ko = str(rec["korean"])
            if key in existing:
                if existing[key] != ko:
                    raise SystemExit(f"conflicting Korean id {section}/{rid}: {owners[key]} and {path}")
                continue
            existing[key] = ko
            owners[key] = path
            if path.name == f"msgsec{section:03d}-auto.toml":
                auto.setdefault(section, {})[str(rid)] = {
                    "japanese": str(rec.get("japanese", "")),
                    "korean": ko,
                }
    return existing, auto


def render(records: dict[str, dict[str, str]]) -> str:
    out = ["# SPDX-License-Identifier: CC-BY-SA-4.0", ""]
    for rid in sorted(records, key=lambda x: int(x) if x.isdigit() else x):
        rec = records[rid]
        out += [f"[{q(rid)}]", f"japanese = {q(rec['japanese'])}", f"korean = {q(rec['korean'])}", ""]
    return "\n".join(out)


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("inputs", nargs="+", type=Path)
    args = ap.parse_args()

    existing, auto = load_existing()
    canonical_cache: dict[int, dict[str, dict[str, object]]] = {}
    seen_input: dict[tuple[int, str], str] = {}
    added = 0
    unchanged = 0

    for input_path in args.inputs:
        with input_path.open(encoding="utf-8") as f:
            for lineno, line in enumerate(f, 1):
                if not line.strip():
                    continue
                obj = json.loads(line)
                section, rid, ko = int(obj["section"]), str(obj["id"]), str(obj["korean"])
                key = (section, rid)
                if key in seen_input:
                    if seen_input[key] != ko:
                        raise SystemExit(f"{input_path}:{lineno}: conflicting duplicate input id {section}/{rid}")
                    continue
                seen_input[key] = ko
                if not ko:
                    raise SystemExit(f"{input_path}:{lineno}: empty Korean translation {section}/{rid}")
                canon = canonical_cache.setdefault(section, canonical_for(section))
                rec = canon.get(rid)
                if rec is None or "japanese" not in rec:
                    raise SystemExit(f"{input_path}:{lineno}: unknown canonical id {section}/{rid}")
                ja = str(rec["japanese"])
                if tokens(ja) != tokens(ko):
                    raise SystemExit(
                        f"{input_path}:{lineno}: control-token mismatch {section}/{rid}: "
                        f"{tokens(ja)} != {tokens(ko)}"
                    )
                if key in existing:
                    if existing[key] != ko:
                        raise SystemExit(f"{input_path}:{lineno}: conflicting existing translation {section}/{rid}")
                    unchanged += 1
                    continue
                auto.setdefault(section, {})[rid] = {"japanese": ja, "korean": ko}
                existing[key] = ko
                added += 1

    KOREAN.mkdir(parents=True, exist_ok=True)
    for section, records in sorted(auto.items()):
        path = KOREAN / f"msgsec{section:03d}-auto.toml"
        path.write_text(render(records), encoding="utf-8")
    print(f"applied {added} new translations; {unchanged} already identical")


if __name__ == "__main__":
    main()
