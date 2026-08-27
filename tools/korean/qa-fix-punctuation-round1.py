#!/usr/bin/env python3
"""Source-gated QA-3 punctuation normalization for unambiguous repeated lines."""
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


def norm(text: str) -> str:
    return WS_RE.sub("", text.replace("<line-break>", ""))


CANONICAL = {
    norm("受け入れられぬ運命であれば、私はこれからも戦い続けるだろう。<end>"):
        "받아들일 수 없는 운명이라면 나는 앞으로도 계속 싸울 것이다.<end>",
    norm("それでは、失礼いたします。<end>"):
        "그럼, 실례하겠습니다.<end>",
    norm("今は忙しい<end>"):
        "지금은 바빠<end>",
    norm("今、なんて…？<end>"):
        "방금, 뭐라고…?<end>",
    norm("今日こそ、その大剣、天動地鳴が抜ける日だ！<end>"):
        "오늘이야말로 그 대검, 천동지명이 뽑히는 날이다!<end>",
    norm("何よ、あんただって…<end>"):
        "뭐야, 너도…<end>",
    norm("きっ、貴様は解放軍の<value:$28>警備隊を呼べっ！早く！<end>"):
        "네, 네놈은 해방군의 <value:$28>! 경비대를 불러! 빨리!<end>",
}


def main() -> None:
    seen: set[int] = set()
    changed_records = 0
    changed_files = 0
    by_source: dict[str, int] = {}
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
            if current is None or not KO_RE.match(line):
                continue
            numeric = int(current)
            if numeric in seen:
                current = None
                continue
            seen.add(numeric)
            rec = data.get(current)
            if not isinstance(rec, dict):
                current = None
                continue
            ja = rec.get("japanese")
            ko = rec.get("korean")
            if not isinstance(ja, str) or not isinstance(ko, str):
                current = None
                continue
            key = norm(ja)
            canonical = CANONICAL.get(key)
            if canonical is not None and ko != canonical:
                # This pass is intentionally limited to groups pre-classified as
                # punctuation-only; refuse to alter wording beyond punctuation/
                # whitespace by comparing alphanumeric/Hangul/token skeletons.
                def skeleton(s: str) -> str:
                    return re.sub(r"[\s\u3000,.!?…~。！？、]", "", s)
                if skeleton(ko) != skeleton(canonical):
                    raise SystemExit(f"refusing lexical punctuation rewrite for id {numeric}: {ko!r} -> {canonical!r}")
                lines[i] = "korean = " + json.dumps(canonical, ensure_ascii=False)
                dirty = True
                changed_records += 1
                by_source[key] = by_source.get(key, 0) + 1
            current = None
        if dirty:
            path.write_text("\n".join(lines) + "\n", encoding="utf-8")
            changed_files += 1
    print(f"Punctuation round 1: {changed_records} records across {changed_files} files")
    for key, count in sorted(by_source.items(), key=lambda kv: -kv[1]):
        print(f"  {count:3d}  {key[:80]}")


if __name__ == "__main__":
    main()
