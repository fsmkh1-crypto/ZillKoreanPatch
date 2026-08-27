#!/usr/bin/env python3
"""Conservative QA-3 exact-source and typography consistency fixer.

Reviewed lexical/spacing fixes remain source-gated. Two corpus-wide typography
repairs are also allowed because they do not change words, register, control
tokens, or punctuation:
  * insert one space after ASCII sentence-final .!? when Hangul follows directly
  * remove trailing ASCII/full-width whitespace immediately before <end>
Full-width punctuation and internal full-width layout spacing are untouched.
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
GLUED_ASCII_SENTENCE_RE = re.compile(r"(?<=[.!?])(?=[가-힣])")
SPACE_BEFORE_END_RE = re.compile(r"[ \u3000]+(?=<end>)")


def norm_source(text: str) -> str:
    return WS_RE.sub("", text.replace(LINE_BREAK, ""))


def no_space(text: str) -> str:
    return WS_RE.sub("", text)


SPACE_CANONICAL: dict[str, str] = {
    norm_source("<value:$28>様！<end>"): "<value:$28>님!<end>",
    norm_source("ノエル！　<end>"): "노엘!<end>",
    norm_source("…いつも言ってるはずだぞ。一番大切なのは仲間だってな。<end>"): "…늘 말했을 텐데. 가장 중요한 건 동료라고.<end>",
    norm_source(" 装   備<end>"): "장비<end>",
    norm_source("魔法<end>"): "마법<end>",
    norm_source("<value:$28>様がお通りになるのですね。<end>"): "<value:$28>님이 지나가시는군요.<end>",
    norm_source("<value:$28>さん！<end>"): "<value:$28>씨!<end>",
    norm_source("イクスキュア<end>"): "익스 큐어<end>",
    norm_source("サプキュア<end>"): "서브 큐어<end>",
    norm_source("ゴブゴブ団員<end>"): "고브고브 단원<end>",
    norm_source("気がついたか？放っておいてくれて大丈夫だったのに、…大きなお世話だ。<end>"): "정신이 들었나? 그냥 내버려 둬도 괜찮았는데, …쓸데없는 참견이군.<end>",
}

EXACT_VARIANTS: dict[str, dict[str, str]] = {
    norm_source("冒険者<end>"): {
        "모험자<end>": "모험가<end>",
    },
}


def apply_typography(text: str) -> tuple[str, int, int]:
    text, glued = GLUED_ASCII_SENTENCE_RE.subn(" ", text)
    text, trailing = SPACE_BEFORE_END_RE.subn("", text)
    return text, glued, trailing


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

            src = norm_source(ja)
            new = ko
            canonical = SPACE_CANONICAL.get(src)
            if canonical is not None and ko != canonical and no_space(ko) == no_space(canonical):
                new = canonical
            elif src in EXACT_VARIANTS and ko in EXACT_VARIANTS[src]:
                new = EXACT_VARIANTS[src][ko]

            new, glued, trailing = apply_typography(new)
            if glued:
                counts["insert-space-after-ascii-sentence-punctuation"] = counts.get("insert-space-after-ascii-sentence-punctuation", 0) + glued
            if trailing:
                counts["remove-space-before-end"] = counts.get("remove-space-before-end", 0) + trailing

            if new != ko:
                lines[i] = "korean = " + json.dumps(new, ensure_ascii=False)
                dirty = True
                changed_records += 1
            current = None

        if dirty:
            path.write_text("\n".join(lines) + "\n", encoding="utf-8")
            changed_files += 1

    print(f"QA-3 safe consistency fixer: {changed_records} records across {changed_files} files")
    for key, count in sorted(counts.items(), key=lambda kv: (-kv[1], kv[0])):
        print(f"  {count:4d}  {key}")


if __name__ == "__main__":
    main()
