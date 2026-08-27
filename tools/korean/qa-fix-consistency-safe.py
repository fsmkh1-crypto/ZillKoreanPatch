#!/usr/bin/env python3
"""Conservative QA-3 exact-source and typography consistency fixer.

Reviewed lexical/spacing fixes remain source-gated. Corpus-wide typography
repairs are limited to punctuation/terminal whitespace and cannot alter words,
register, control tokens, or punctuation. Internal spacing repairs are allowed
only for explicitly reviewed Japanese sources or exact current-record rewrites.
Proper-name repairs are source-gated by the Japanese canonical surface so a
Korean substring can never be rewritten outside the matching source context.
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

SOURCE_TERM_VARIANTS: dict[str, dict[str, str]] = {
    "フゴー": {"후고": "휴고", "푸고": "휴고"},
    "オルナット": {"올나트": "오르나트", "오르낫": "오르나트"},
    "アンギルダン": {"안길단": "앙길단", "앙기르단": "앙길단"},
    "ダルケニス": {"달케니스": "다르케니스"},
    "ジュサプブロス": {"주삽브로스": "주사프브로스", "주사브로스": "주사프브로스"},
    "レムオン": {"레무온": "레몬", "레므온": "레몬"},
    "ウルグ": {"우르그": "울그"},
    "エア": {"에어": "에아"},
    "ジリオン": {"질리온": "지리온"},
    "ゼリグ": {"제리그": "젤리그"},
    "ヴァイライラ": {"바이라이라": "바일라이라"},
    "エリス": {"엘리스": "에리스"},
    "シェムハザ": {"셰므하자": "셈하자", "셰무하자": "셈하자"},
    "イズキヤル": {"이즈키알": "이즈키얄"},
}

SOURCE_SUBSTITUTIONS: dict[str, tuple[tuple[str, str], ...]] = {
    norm_source("君、僕を知らないのか…？それならひとつ忠告しよう。力にすがるのは愚かなことだ。力はただ、より強い力によって叩き伏せられるのみだ。<end>"): (
        ("의해눌릴", "의해 눌릴"),
    ),
    norm_source("究極生物を作ろうとして、その製法を神器、禁断の聖杯から得ようと、聖杯をねらっています。円卓騎士には他にうちのネモなどがいますね。<end>"): (
        ("만들려고,그", "만들려고, 그"),
        ("그 밖에도저희", "그 밖에도 저희"),
    ),
}

EXACT_RECORD_REWRITES: dict[int, tuple[str, str]] = {
    270054: (
        "이걸로 <value:$28> 님은,결승 대회에 출전할 권리를 얻었습니다.　결승 대회에 참가할 생각이 있다면 8월에여기로 와 주세요. 기다리고 있겠습니다.<end>",
        "이걸로 <value:$28> 님은, 결승 대회에 출전할 권리를 얻었습니다. 결승 대회에 참가할 생각이 있다면 8월에 여기로 와 주세요. 기다리고 있겠습니다.<end>",
    ),
    560189: (
        "…죽음의 날갯소리 레이븐이이 근처에서 갈 만한 곳이라….　　　…나는 역시 운이 없어. 저주가 걸려 있어서그걸 말하면 나는 더….<end>",
        "…죽음의 날갯소리 레이븐이 이 근처에서 갈 만한 곳이라…. …나는 역시 운이 없어. 저주가 걸려 있어서 그걸 말하면 나는 더….<end>",
    ),
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

            exact = EXACT_RECORD_REWRITES.get(numeric)
            if exact is not None and new == exact[0]:
                new = exact[1]
                counts["exact-current-record-rewrite"] = counts.get("exact-current-record-rewrite", 0) + 1

            canonical = SPACE_CANONICAL.get(src)
            if canonical is not None and new != canonical and no_space(new) == no_space(canonical):
                new = canonical
            elif src in EXACT_VARIANTS and new in EXACT_VARIANTS[src]:
                new = EXACT_VARIANTS[src][new]

            for ja_term, variants in SOURCE_TERM_VARIANTS.items():
                if ja_term not in ja:
                    continue
                for old, replacement in variants.items():
                    n = new.count(old)
                    if n:
                        new = new.replace(old, replacement)
                        key = f"proper-name:{ja_term}:{old}->{replacement}"
                        counts[key] = counts.get(key, 0) + n

            for old, replacement in SOURCE_SUBSTITUTIONS.get(src, ()):
                n = new.count(old)
                if n:
                    new = new.replace(old, replacement)
                    key = f"source-gated:{old}->{replacement}"
                    counts[key] = counts.get(key, 0) + n

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
