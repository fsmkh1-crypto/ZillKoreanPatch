#!/usr/bin/env python3
"""Conservative QA-3 exact-source consistency fixer.

Only explicitly reviewed Japanese sources are eligible. Within those sources,
known-safe spacing variants may be matched after normalizing ASCII/full-width
whitespace. Dialogue/register, punctuation, and unresolved names are untouched.
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
WS_RE = re.compile(r"[\s\u3000]+")
LINE_BREAK = "<line-break>"


def norm_source(text: str) -> str:
    return WS_RE.sub("", text.replace(LINE_BREAK, ""))


def no_space(text: str) -> str:
    return WS_RE.sub("", text)


# normalized Japanese source -> reviewed canonical Korean.
# A row is rewritten only when its Korean differs from canonical by whitespace
# alone, so this cannot collapse lexical or punctuation differences.
SPACE_CANONICAL: dict[str, str] = {
    norm_source("<value:$28>様！<end>"): "<value:$28>님!<end>",
    norm_source("ノエル！　<end>"): "노엘!<end>",
    norm_source("…いつも言ってるはずだぞ。一番大切なのは仲間だってな。<end>"): "…늘 말했을 텐데. 가장 중요한 건 동료라고.<end>",
    norm_source(" 装   備<end>"): "장비<end>",
    norm_source("魔法<end>"): "마법<end>",
    norm_source("<value:$28>様がお通りになるのですね。<end>"): "<value:$28>님이 지나가시는군요.<end>",
    norm_source("<value:$28>さん！<end>"): "<value:$28>씨!<end>",
}


def main() -> None:
    changed_records = 0
    changed_files = 0
    counts: dict[str, int] = {}
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

            canonical = SPACE_CANONICAL.get(norm_source(ja))
            if canonical is not None and ko != canonical and no_space(ko) == no_space(canonical):
                lines[i] = "korean = " + json.dumps(canonical, ensure_ascii=False)
                dirty = True
                changed_records += 1
                key = f"{ko}->{canonical}"
                counts[key] = counts.get(key, 0) + 1
            current = None

        if dirty:
            path.write_text("\n".join(lines) + "\n", encoding="utf-8")
            changed_files += 1

    print(f"QA-3 safe consistency fixer: {changed_records} records across {changed_files} files")
    for key, count in sorted(counts.items(), key=lambda kv: (-kv[1], kv[0])):
        print(f"  {count:4d}  {key}")


if __name__ == "__main__":
    main()
