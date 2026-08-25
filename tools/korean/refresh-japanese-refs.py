#!/usr/bin/env python3
"""Refresh stale Japanese reference fields in Korean overlays from canonical source.

This never changes Korean translations. It exists so historical overlays cannot block
bulk result application solely because their duplicated Japanese reference text drifted.
Canonical TOML remains the single source of truth.
"""
from __future__ import annotations

import json
from pathlib import Path
import re
import tomllib

ROOT = Path(__file__).resolve().parents[2]
CANON = ROOT / "translations" / "messages"
KOREAN = ROOT / "translations" / "korean" / "messages"
FILE_RE = re.compile(r"^msgsec([0-9]{3})")


def q(value: str) -> str:
    return json.dumps(value, ensure_ascii=False)


def load_canonical(section: int) -> dict[str, str]:
    path = CANON / f"msgsec{section:03d}.toml"
    with path.open("rb") as f:
        raw = tomllib.load(f)
    out: dict[str, str] = {}
    for rid, rec in raw.items():
        if isinstance(rec, dict) and rec.get("japanese") is not None:
            out[str(rid)] = str(rec["japanese"])
    return out


def main() -> None:
    cache: dict[int, dict[str, str]] = {}
    repaired = 0
    for path in sorted(KOREAN.glob("msgsec*.toml")):
        match = FILE_RE.match(path.name)
        if not match:
            continue
        section = int(match.group(1))
        canon = cache.setdefault(section, load_canonical(section))
        with path.open("rb") as f:
            data = tomllib.load(f)
        text = path.read_text(encoding="utf-8")
        changed = False
        for rid, rec in data.items():
            if not isinstance(rec, dict) or rec.get("japanese") is None:
                continue
            rid = str(rid)
            expected = canon.get(rid)
            if expected is None or str(rec["japanese"]) == expected:
                continue
            # Overlay records are emitted as a table header followed by one-line fields.
            # Replace only the Japanese field inside this record; preserve comments,
            # Korean text, file splitting, and all unrelated formatting.
            header = f'["{rid}"]'
            start = text.find(header)
            if start < 0:
                raise SystemExit(f"{path}: cannot locate record {rid}")
            next_header = text.find('\n["', start + len(header))
            end = len(text) if next_header < 0 else next_header
            block = text[start:end]
            new_block, count = re.subn(
                r"(?m)^japanese\s*=\s*.*$",
                "japanese = " + q(expected),
                block,
                count=1,
            )
            if count != 1:
                raise SystemExit(f"{path}: cannot locate japanese field for {rid}")
            text = text[:start] + new_block + text[end:]
            repaired += 1
            changed = True
        if changed:
            path.write_text(text, encoding="utf-8")
    print(f"refreshed {repaired} stale Japanese reference fields")


if __name__ == "__main__":
    main()
