#!/usr/bin/env python3
"""Normalize full-width layout-space residue in Korean prose.

The Korean overlay intentionally removes semantic <line-break> markers. During
first-pass translation some Japanese full-width padding survived in the Korean
text, creating visual gaps after line flattening. Convert those runs to one
ASCII space, except for a tiny source-gated set of UI strings where spacing is
part of the display layout.
"""
from __future__ import annotations

import json
from pathlib import Path
import re
import tomllib

ROOT = Path(__file__).resolve().parents[2]
KOREAN_DIR = ROOT / "translations" / "korean" / "messages"
SECTION_FILE_RE = re.compile(r"^msgsec(\d{3})(?:(?:-part\d+)|b)?\.toml$")
HEADER_RE = re.compile(r'^\["(?P<id>\d+)"\]$')
KO_RE = re.compile(r'^korean = (?P<value>".*")$')
HANGUL_RE = re.compile(r"[가-힣]")
FULLWIDTH_RUN_RE = re.compile(r"\u3000+")

# These are genuine compact UI layouts where a wide separator remains useful.
PRESERVE_IDS = {
    190373, 190378, 190383,  # date display rows
    200077,                 # ×닫기　○닫기 button legend
}


def main() -> None:
    seen_ids: set[int] = set()
    changed_records = 0
    changed_files = 0
    replacements = 0

    for path in sorted(KOREAN_DIR.glob("msgsec*.toml")):
        if not SECTION_FILE_RE.match(path.name):
            continue
        with path.open("rb") as f:
            data = tomllib.load(f)
        lines = path.read_text(encoding="utf-8").splitlines()
        dirty = False
        current: str | None = None

        for i, line in enumerate(lines):
            hm = HEADER_RE.match(line)
            if hm:
                current = hm.group("id")
                continue
            km = KO_RE.match(line)
            if current is None or km is None:
                continue
            numeric = int(current)
            if numeric in seen_ids:
                current = None
                continue
            seen_ids.add(numeric)
            if numeric in PRESERVE_IDS:
                current = None
                continue
            rec = data.get(current)
            if not isinstance(rec, dict):
                current = None
                continue
            ko = rec.get("korean")
            if not isinstance(ko, str) or "\u3000" not in ko or not HANGUL_RE.search(ko):
                current = None
                continue

            new, n = FULLWIDTH_RUN_RE.subn(" ", ko)
            # Full-width padding often sat next to ordinary whitespace as well.
            new = re.sub(r" {2,}", " ", new)
            if new != ko:
                lines[i] = "korean = " + json.dumps(new, ensure_ascii=False)
                dirty = True
                changed_records += 1
                replacements += n
            current = None

        if dirty:
            path.write_text("\n".join(lines) + "\n", encoding="utf-8")
            changed_files += 1

    print(f"Full-width space cleanup: {changed_records} records across {changed_files} files; {replacements} runs normalized")


if __name__ == "__main__":
    main()
