#!/usr/bin/env python3
"""Source-gated correction pass for proper names with official English anchors.

This deliberately covers only names whose official/legacy English surface is
present in translations/terminology/names.toml and whose Korean variant is an
obvious transliteration inconsistency. No dialogue/register text is touched.
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

# Japanese surface -> noncanonical Korean surface -> canonical Korean surface.
# Anchors from translations/terminology/names.toml:
#   ダルケニス = Darkenith
#   フゴー = Hugo
#   ジリオン = Zillion
#   アンギルダン = Angeerdan
#   ヴァイライラ = Vailaila
#   ゼリグ = Zelig
#   シェムハザ = Shemhaza
#   エリス = Eris
VARIANTS: dict[str, dict[str, str]] = {
    "ダルケニス": {"달케니스": "다르케니스"},
    "フゴー": {"후고": "휴고", "푸고": "휴고"},
    "ジリオン": {"지리온": "질리온"},
    "アンギルダン": {"안길단": "앙길단", "앙기르단": "앙길단"},
    "ヴァイライラ": {"바이라이라": "바일라이라", "바이아이라": "바일라이라"},
    "ゼリグ": {"제리그": "젤리그"},
    "シェムハザ": {"셰므하자": "셈하자", "셰무하자": "셈하자"},
    "エリス": {"엘리스": "에리스"},
}


def main() -> None:
    seen_ids: set[int] = set()
    changed_records = 0
    changed_files = 0
    counts: dict[str, int] = {}

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
            rec = data.get(current)
            if not isinstance(rec, dict):
                current = None
                continue
            ja = rec.get("japanese")
            ko = rec.get("korean")
            if not isinstance(ja, str) or not isinstance(ko, str):
                current = None
                continue

            new = ko
            for ja_name, variants in VARIANTS.items():
                if ja_name not in ja:
                    continue
                for old, canonical in variants.items():
                    n = new.count(old)
                    if n:
                        new = new.replace(old, canonical)
                        key = f"{ja_name}:{old}->{canonical}"
                        counts[key] = counts.get(key, 0) + n

            if new != ko:
                lines[i] = "korean = " + json.dumps(new, ensure_ascii=False)
                dirty = True
                changed_records += 1
            current = None

        if dirty:
            path.write_text("\n".join(lines) + "\n", encoding="utf-8")
            changed_files += 1

    print(f"Proper-name correction pass: {changed_records} records across {changed_files} files")
    for key, count in sorted(counts.items(), key=lambda kv: (-kv[1], kv[0])):
        print(f"  {count:4d}  {key}")


if __name__ == "__main__":
    main()
