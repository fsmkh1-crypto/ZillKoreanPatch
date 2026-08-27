#!/usr/bin/env python3
"""Source-gated high-confidence Korean consistency pass, round 3.

This pass only changes identical Japanese sources whose scene/register is stable,
plus independently anchored proper-name surfaces. Proper-name replacements are
performed only when the matching Japanese name is present in the same record.
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


CANONICAL: dict[str, str] = {
    norm_source("究極生物を作ろうとして、<line-break>その製法を神器、禁断の聖杯から<line-break>得ようと、聖杯をねらっています。<line-break>円卓騎士には他に<line-break>うちのネモなどがいますね。<line-break>　<end>"): "궁극 생물을 만들려고, 그 제조법을 신기, 금단의 성배에서 얻기 위해 성배를 노리고 있습니다. 원탁기사에는 그 밖에도 저희 네모 같은 이들이 있죠.<end>",
    norm_source("　　　ネメアよ、生きテいたか…。<line-break>だが、うぬハもはや…闇に落ちシ者…。<line-break>　　　我が軍門ニ降るがよい。<end>"): "네메아여, 살아 있었는가…. 하지만 너 또한 이미… 어둠에 떨어진 자…. 내 휘하로 들어오라.<end>",
    norm_source("　　　　クックックッ。ネメアよ。<line-break>　　　　　　闇ニ落ちテなお、<line-break>　　　我ト刺し違えようと…するとは…。<line-break>　　アクマデ運命神ノことのはニ従うカ。<line-break>　　　ここでうぬヲ滅ぼスノミでは<line-break>　　　我ガ憎しみハ…満足できぬ…。<line-break>　　　うぬガ我ヲ滅ぼす運命ナラバ<line-break>　我ハうぬヲ砕キ、運命ヲ飲み喰ラオウ。<line-break>　　　謁見ノ間に…来るガよイ。<line-break>　　皇帝の玉座…コタビハうぬの血で<line-break>　　朱に染メ、世界ノ運命ソノモノヲ<line-break>　　ココニ終わラセテくレヨウぞ…！<end>"): "크크크. 네메아여. 어둠에 떨어지고도 나와 함께 죽으려 하다니…. 끝까지 운명신의 말씀을 따르는가. 여기서 네놈을 없애는 것만으로는 내 증오가 만족하지 않는다…. 네놈이 나를 멸할 운명이라면 내가 네놈을 부수고 운명까지 삼켜 버리겠다. 알현의 방으로 와라. 황제의 옥좌를 이번에는 네 피로 붉게 물들이고 세계의 운명 그 자체를 여기서 끝내 주마…!<end>",
    norm_source("聞きたいことがあれば答えますよ。<end>"): "궁금한 게 있으면 대답해 드리죠.<end>",
    norm_source("なんとか、助かりました。<line-break>どうもありがとうございました。<line-break>　<line-break>お礼金です、受け取ってください。<line-break>ほんとうにありがとうございました。<end>"): "어떻게든 살았습니다. 정말 감사합니다. 사례금입니다, 받아 주세요. 정말 감사합니다.<end>",
    norm_source("　　　　　　　　　　よし。　　　　　　　　　<line-break>　　　　　　　では我は、汝を<line-break>　　　　　　 精霊の座へと導かん。<line-break>　<line-break>　　　　 　汝、気高き地の光を越え<line-break>　　　　 　　彼の巨人に挑むがよい。<line-break>　　　　  　挑み、打ち破るがよい…。<end>"): "좋다. 그럼 내가 그대를 정령의 자리로 인도하리라. 그대, 고귀한 땅의 빛을 넘어 저 거인에게 도전하라. 도전하여 쓰러뜨려라….<end>",
    norm_source("いや～、はやいもんだよねぇ。<line-break>もう、バイアシオン大陸なんて<line-break>ちっちゃくて見えないわ。<line-break>あんな、ちっぽけな陸の上で<line-break>あたしたちは戦ってたんだよね。<line-break>ここからこっちは自分の国だって。<end>"): "이야~ 빠르네. 벌써 바이아시온 대륙은 작아서 안 보일 정도야. 저 조그만 땅 위에서 우리는 싸우고 있었던 거지. 여기서부터는 자기 나라라고 말하면서.<end>",
    norm_source("おう、昨日のケガは大丈夫か？　　　　　　<line-break>　<line-break>　<end>"): "어이, 어제 다친 곳은 괜찮아?<end>",
    norm_source("こ、こんなはずは…。<line-break>ええい、まだだ！<line-break>俺様の無敵伝説はこれから始まるんだ！<end>"): "이, 이럴 리가…. 젠장, 아직이다! 이 몸의 무적 전설은 이제부터 시작이야!<end>",
    norm_source("この戦いで、ゼネテスの声望は<line-break>一気に上がり、奇跡の名将と呼ばれたが<line-break>ロストールの被害は甚大なものがあった。<end>"): "이 전투로 제네테스의 명망은 단숨에 높아져 기적의 명장이라 불렸지만, 로스토올이 입은 피해도 막대했다.<end>",
    norm_source("さあ、そなたも行くがよい。<line-break>無限のソウルを持つ者よ。<line-break>　<line-break>最後に、礼を言うぞ。<line-break>そなたのおかげで、わらわも<line-break>未来に希望を見出せそうじゃ。<end>"): "자, 그대도 가거라. 무한의 소울을 지닌 자여. 마지막으로 감사하마. 그대 덕분에 나도 미래에서 희망을 찾을 수 있을 것 같구나.<end>",
    norm_source("ジンガ<end>"): "징가<end>",
    norm_source("ゾフォル<end>"): "조포르<end>",
    norm_source("ダイダロ<end>"): "다이다로<end>",
    norm_source("バロル<end>"): "발로르<end>",
}

SOURCE_TERM_VARIANTS: dict[str, dict[str, str]] = {
    "ジンガ": {"진가": "징가"},
    "ゾフォル": {"조폴": "조포르"},
    "ダイダロ": {"다이달로": "다이다로"},
    "バロル": {"바롤": "발로르", "바로르": "발로르"},
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

            new = CANONICAL.get(norm_source(ja), ko)
            for ja_term, variants in SOURCE_TERM_VARIANTS.items():
                if ja_term not in ja:
                    continue
                for old, replacement in variants.items():
                    n = new.count(old)
                    if n:
                        new = new.replace(old, replacement)
                        counts[f"proper-name:{ja_term}:{old}->{replacement}"] = counts.get(f"proper-name:{ja_term}:{old}->{replacement}", 0) + n

            if new != ko:
                lines[i] = "korean = " + json.dumps(new, ensure_ascii=False)
                dirty = True
                changed_records += 1
            current = None

        if dirty:
            path.write_text("\n".join(lines) + "\n", encoding="utf-8")
            changed_files += 1

    print(f"Repeated high-confidence round 3: {changed_records} records across {changed_files} files")
    for key, count in sorted(counts.items(), key=lambda kv: (-kv[1], kv[0])):
        print(f"  {count:4d}  {key}")


if __name__ == "__main__":
    main()
