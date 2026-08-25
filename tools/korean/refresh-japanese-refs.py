#!/usr/bin/env python3
"""Remove stale Korean overlay records whose duplicated Japanese no longer matches canonical.

A Korean translation was produced for the Japanese text stored beside it. If that Japanese
reference differs from the current canonical source, silently replacing only the Japanese
field can mislabel an unrelated Korean translation as valid. For bulk work, the safe repair
is to drop the stale overlay record entirely so the deterministic next-packet flow emits the
canonical record again for retranslation.
"""
from __future__ import annotations

from pathlib import Path
import re
import tomllib

ROOT = Path(__file__).resolve().parents[2]
CANON = ROOT / "translations" / "messages"
KOREAN = ROOT / "translations" / "korean" / "messages"
FILE_RE = re.compile(r"^msgsec([0-9]{3})")


def load_canonical(section: int) -> dict[str, str]:
    path = CANON / f"msgsec{section:03d}.toml"
    with path.open("rb") as f:
        raw = tomllib.load(f)
    return {
        str(rid): str(rec["japanese"])
        for rid, rec in raw.items()
        if isinstance(rec, dict) and rec.get("japanese") is not None
    }


def remove_record(text: str, rid: str, path: Path) -> str:
    header = f'["{rid}"]'
    start = text.find(header)
    if start < 0:
        raise SystemExit(f"{path}: cannot locate record {rid}")
    # Include the blank line immediately before a non-first record when present so repeated
    # cleanup does not accumulate whitespace. Preserve the SPDX/comment preamble.
    cut_start = start
    if cut_start >= 2 and text[cut_start - 2 : cut_start] == "\n\n":
        cut_start -= 1
    next_header = text.find('\n["', start + len(header))
    end = len(text) if next_header < 0 else next_header + 1
    return text[:cut_start] + text[end:]


def main() -> None:
    cache: dict[int, dict[str, str]] = {}
    removed = 0
    details: list[str] = []

    for path in sorted(KOREAN.glob("msgsec*.toml")):
        match = FILE_RE.match(path.name)
        if not match:
            continue
        section = int(match.group(1))
        canon = cache.setdefault(section, load_canonical(section))
        with path.open("rb") as f:
            data = tomllib.load(f)

        stale: list[str] = []
        for rid, rec in data.items():
            if not isinstance(rec, dict) or rec.get("japanese") is None:
                continue
            rid = str(rid)
            expected = canon.get(rid)
            if expected is None:
                raise SystemExit(f"{path}: overlay ID {rid} does not exist in canonical source")
            if str(rec["japanese"]) != expected:
                stale.append(rid)

        if not stale:
            continue

        text = path.read_text(encoding="utf-8")
        for rid in stale:
            text = remove_record(text, rid, path)
            removed += 1
            details.append(f"{path.name}:{rid}")
        path.write_text(text, encoding="utf-8")

    print(f"removed {removed} stale Korean overlay records for retranslation")
    for detail in details:
        print(f"  {detail}")


if __name__ == "__main__":
    main()
