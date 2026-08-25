#!/usr/bin/env python3
"""Move legacy Korean line breaks out of semantic text and into layout metadata.

Bulk Korean text is translator-owned wording. ``<line-break>`` is build-owned layout.
Older rows inherited Japanese break positions in ``korean``; this migration preserves
all non-whitespace Korean characters, converts each legacy break boundary to one
semantic space, and preserves the old visual break in ``layout`` until reflow tooling
replaces it.
"""
from __future__ import annotations

import json
from pathlib import Path
import re
import tomllib

ROOT = Path(__file__).resolve().parents[2]
KOREAN = ROOT / "translations" / "korean" / "messages"
LINE_BREAK = "<line-break>"
BREAK_WITH_SPACE_RE = re.compile(r"\s*<line-break>\s*")


def semanticize(text: str) -> str:
    return BREAK_WITH_SPACE_RE.sub(" ", text)


def layoutize(text: str) -> str:
    return BREAK_WITH_SPACE_RE.sub(LINE_BREAK, text)


def q(text: str) -> str:
    return json.dumps(text, ensure_ascii=False)


def migrate_file(path: Path) -> int:
    raw = path.read_text(encoding="utf-8")
    with path.open("rb") as f:
        decoded = tomllib.load(f)

    changed: dict[str, tuple[str, str]] = {}
    for rid, rec in decoded.items():
        if not isinstance(rec, dict):
            continue
        korean = rec.get("korean")
        if not isinstance(korean, str) or LINE_BREAK not in korean:
            continue
        semantic = semanticize(korean)
        existing_layout = rec.get("layout")
        layout = layoutize(str(existing_layout)) if existing_layout else layoutize(korean)
        changed[str(rid)] = (semantic, layout)

    if not changed:
        return 0

    lines = raw.splitlines()
    current_id: str | None = None
    seen_layout: set[str] = set()
    output: list[str] = []
    for line in lines:
        header = re.fullmatch(r'\["([0-9]+)"\]', line)
        if header:
            current_id = header.group(1)
        if current_id in changed and line.startswith("korean = "):
            semantic, layout = changed[current_id]
            output.append(f"korean = {q(semantic)}")
            if not isinstance(decoded[current_id].get("layout"), str) or not decoded[current_id].get("layout"):
                output.append(f"layout = {q(layout)}")
                seen_layout.add(current_id)
            continue
        if current_id in changed and line.startswith("layout = "):
            output.append(f"layout = {q(changed[current_id][1])}")
            seen_layout.add(current_id)
            continue
        output.append(line)

    missing = set(changed) - seen_layout
    if missing:
        raise SystemExit(f"{path}: failed to materialize layout for migrated IDs {sorted(missing)}")
    path.write_text("\n".join(output) + "\n", encoding="utf-8")
    return len(changed)


def main() -> None:
    total = 0
    files = 0
    for path in sorted(KOREAN.glob("msgsec*.toml")):
        count = migrate_file(path)
        if count:
            files += 1
            total += count
            print(f"migrated {count} rows in {path.name}")
    print(f"migrated {total} legacy semantic line-break rows across {files} files")


if __name__ == "__main__":
    main()
