#!/usr/bin/env python3
"""Quarantine legacy Korean overlay records that no longer validate against canonical.

Historical overlays may contain stale duplicated Japanese text or Korean strings whose
fixed runtime control-token sequence no longer matches canonical. Korean ``<line-break>``
is build-owned layout and is deliberately excluded from this destructive validation.
"""
from __future__ import annotations

from pathlib import Path
import re
import tomllib

from control_tags import fixed_tokens

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


def record_problem(expected: str, japanese: object, korean: object) -> str | None:
    if japanese is None or korean is None or str(korean) == "":
        return "missing/empty required field"
    if str(japanese) != expected:
        return "stale Japanese reference"
    if fixed_tokens(expected) != fixed_tokens(str(korean)):
        return "fixed control-token mismatch"
    return None


def remove_record(text: str, rid: str, path: Path) -> str:
    header = f'["{rid}"]'
    start = text.find(header)
    if start < 0:
        raise SystemExit(f"{path}: cannot locate record {rid}")
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

        bad: list[tuple[str, str]] = []
        for rid, rec in data.items():
            if not isinstance(rec, dict):
                continue
            rid = str(rid)
            expected = canon.get(rid)
            if expected is None:
                raise SystemExit(f"{path}: overlay ID {rid} does not exist in canonical source")
            reason = record_problem(expected, rec.get("japanese"), rec.get("korean"))
            if reason is not None:
                bad.append((rid, reason))

        if not bad:
            continue

        text = path.read_text(encoding="utf-8")
        for rid, reason in bad:
            text = remove_record(text, rid, path)
            removed += 1
            details.append(f"{path.name}:{rid} ({reason})")
        path.write_text(text, encoding="utf-8")

    print(f"removed {removed} invalid legacy Korean overlay records for retranslation")
    for detail in details:
        print(f"  {detail}")


if __name__ == "__main__":
    main()
