#!/usr/bin/env python3
"""Fix the two exact-source Balor inflection residues left by earlier Korean text."""
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
WS_RE = re.compile(r"[\s\u3000]+")

TARGET = WS_RE.sub("", "ウフフアハハ…！かわいいこと言うね、エルファス。そういう甘ちゃんなところは大好きさ。でも、君はもう舞台を降りちゃったし、ちょっと惜しいけど、代役を立てさせてもらうよ。破壊の化身、魔王バロルをね。<end>")


def norm_source(text: str) -> str:
    return WS_RE.sub("", text.replace("<line-break>", ""))


def main() -> None:
    seen: set[int] = set()
    changed_records = 0
    changed_files = 0
    for path in sorted(KOREAN_DIR.glob("msgsec*.toml")):
        if not SECTION_FILE_RE.match(path.name):
            continue
        with path.open("rb") as f:
            data = tomllib.load(f)
        lines = path.read_text(encoding="utf-8").splitlines()
        current: str | None = None
        dirty = False
        for i, line in enumerate(lines):
            hm = HEADER_RE.match(line)
            if hm:
                current = hm.group("id")
                continue
            km = KO_RE.match(line)
            if current is None or km is None:
                continue
            rid = int(current)
            if rid in seen:
                current = None
                continue
            seen.add(rid)
            rec = data.get(current)
            if not isinstance(rec, dict):
                current = None
                continue
            ja = rec.get("japanese")
            ko = rec.get("korean")
            if isinstance(ja, str) and isinstance(ko, str) and norm_source(ja) == TARGET:
                new = ko.replace("마왕 바로를 말이야", "마왕 발로르를 말이야")
                if new != ko:
                    lines[i] = "korean = " + json.dumps(new, ensure_ascii=False)
                    dirty = True
                    changed_records += 1
            current = None
        if dirty:
            path.write_text("\n".join(lines) + "\n", encoding="utf-8")
            changed_files += 1
    print(f"Exact Balor residue correction: {changed_records} records across {changed_files} files")


if __name__ == "__main__":
    main()
