#!/usr/bin/env python3
"""Conservative QA-2 canonical terminology normalizer.

Only source-aware, spelling-only variants are rewritten. A replacement is
eligible only when the Japanese record contains the matching canonical source
term. Runtime control tokens are untouched and the main corpus validators run
before any commit in the companion workflow.
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

# Japanese source term -> (canonical Korean, known-safe spelling variants)
RULES: dict[str, tuple[str, tuple[str, ...]]] = {
    "アーギルシャイア": ("아르길샤이어", ("아길샤이어",)),
    "ロストール": ("로스토올", ("로스톨", "로스토르", "로스트올")),
    "リベルダム": ("리벨덤", ("리베르담", "리벨담")),
    "ギア": ("기어", ("기아",)),
    "ソウルイーター": ("소울 이터", ("소울이터",)),
    "エンシャント": ("엔샨트", ("엔션트",)),
    "オッシ": ("오시", ("옷시",)),
    "フルーヴ": ("프루브", ("플루브",)),
    "クリュセイス": ("크류세이스", ("크리세이스",)),
    "ソリアス": ("솔리아스", ("소리아스",)),
    "ロッシマ": ("롯시마", ("로시마",)),
    "ラミリー山": ("라밀리 산", ("라미리 산",)),
    "アルノートゥン": ("알노툰", ("아르노툰",)),
    "リルビー": ("릴비", ("릴루비",)),
    "ソウルリープ": ("소울 리프", ("소울리프",)),
    "タルテュバ": ("타르튜바", ("탈테튜바", "타르테튜바")),
    "ルブルグ": ("루브르그", ("루브루그",)),
}


def main() -> None:
    changed_records = 0
    changed_files = 0
    replacement_counts: dict[str, int] = {}
    seen_ids: set[int] = set()

    for path in sorted(KOREAN_DIR.glob("msgsec*.toml")):
        if not SECTION_FILE_RE.match(path.name):
            continue
        with path.open("rb") as f:
            data = tomllib.load(f)
        lines = path.read_text(encoding="utf-8").splitlines()
        dirty = False
        current: str | None = None
        for i, line in enumerate(lines):
            m = HEADER_RE.match(line)
            if m:
                current = m.group("id")
                continue
            km = KO_RE.match(line)
            if current is None or km is None:
                continue
            numeric = int(current)
            # Identical legacy duplicate overlays are not edited twice; the main
            # loader already treats them as one accepted record.
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
            touched: list[str] = []
            for ja_term, (canonical, variants) in RULES.items():
                if ja_term not in ja:
                    continue
                for variant in variants:
                    n = new.count(variant)
                    if n:
                        new = new.replace(variant, canonical)
                        replacement_counts[f"{ja_term}:{variant}->{canonical}"] = replacement_counts.get(
                            f"{ja_term}:{variant}->{canonical}", 0
                        ) + n
                        touched.append(ja_term)
            if new != ko:
                lines[i] = "korean = " + json.dumps(new, ensure_ascii=False)
                dirty = True
                changed_records += 1
            current = None

        if dirty:
            path.write_text("\n".join(lines) + "\n", encoding="utf-8")
            changed_files += 1

    print(f"QA-2 safe terminology fixer: {changed_records} records across {changed_files} files")
    for key, count in sorted(replacement_counts.items(), key=lambda kv: (-kv[1], kv[0])):
        print(f"  {count:4d}  {key}")


if __name__ == "__main__":
    main()
