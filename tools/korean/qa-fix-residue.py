#!/usr/bin/env python3
"""One-shot QA-1 cleanup for the 17 confirmed Japanese-residue translations.

Only the exact record IDs listed below are touched, and only their `korean =`
line is replaced.  The script fails unless every target is found exactly once,
so it cannot silently drift onto unrelated text.
"""
from __future__ import annotations

import json
from pathlib import Path
import re

ROOT = Path(__file__).resolve().parents[2]
KOREAN_DIR = ROOT / "translations" / "korean" / "messages"

FIXES: dict[str, str] = {
    "20215": "제비 되받아치기<end>",
    "200436": "천공신 노툰이 힘을 나눠 담았다는 액막이<end>",
    "310214": "…가자. 왕도 로스토올로.<end>",
    "1030054": "강도가 이름을 대는 것도 우습지만 난 체라셸이라고 해. 네가 가진 나이트메어의 물방울을 받으러 왔다. 미안하군.<end>",
    "1180089": "케류네이아 V5 종료<end>",
    "1190052": "케류네이아 V2 종료<end>",
    "1200059": "케류네이아 V3 종료<end>",
    "1240147": "………………. 크류세이스에게 전해라.<end>",
    "1340431": "타르튜바 님입니다. 쓰러져 계신 것을 이곳으로 모셔왔습니다.<end>",
    "1340439": "<value:$28>은 타르튜바에게 생명의 조각을 주었다.<end>",
    "1340539": "<value:$28>은 타르튜바에게 생명의 조각을 주었다.<end>",
    "1950124": "#1: 산적의 소문·로스토올·왕도의 문<end>",
    "1960057": "#12: 어두운 방·로스토올·성문 앞<end>",
    "1960623": "쿠데타 「◆로스토올의 문」 부분<end>",
    "1970317": "타르튜바와 부딪힘 종료<end>",
    "1970822": "산 넘은 뒤 타르튜바 발생<end>",
    "1971161": "타르튜바 밀담 발생<end>",
}

HEADER_RE = re.compile(r'^\["(?P<id>\d+)"\]$')
KO_RE = re.compile(r'^korean = ".*"$')


def main() -> None:
    found: dict[str, int] = {rid: 0 for rid in FIXES}
    changed_files = 0

    for path in sorted(KOREAN_DIR.glob("msgsec*.toml")):
        lines = path.read_text(encoding="utf-8").splitlines()
        dirty = False
        current: str | None = None
        for i, line in enumerate(lines):
            m = HEADER_RE.match(line)
            if m:
                current = m.group("id")
                continue
            if current in FIXES and KO_RE.match(line):
                found[current] += 1
                if found[current] > 1:
                    raise SystemExit(f"duplicate target {current} found in {path}")
                replacement = "korean = " + json.dumps(FIXES[current], ensure_ascii=False)
                if line != replacement:
                    lines[i] = replacement
                    dirty = True
                current = None
        if dirty:
            path.write_text("\n".join(lines) + "\n", encoding="utf-8")
            changed_files += 1

    missing = [rid for rid, count in found.items() if count != 1]
    if missing:
        raise SystemExit(f"expected each QA target exactly once; bad targets: {missing}")
    print(f"QA-1 residue fixer: {len(FIXES)} records across {changed_files} files")


if __name__ == "__main__":
    main()
