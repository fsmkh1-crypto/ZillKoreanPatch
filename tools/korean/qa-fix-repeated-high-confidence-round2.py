#!/usr/bin/env python3
"""Normalize reviewed high-confidence repeated Japanese sources to one Korean surface.

Exact-source rewrites are used for repeated UI labels and repeated explanatory
lines. A tiny source-term rewrite table is also allowed for independently
anchored global names so compounds such as `盗人村：ロッシマ邸内` inherit the
same Korean place name without touching unrelated Korean substrings.
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
    norm_source("引き受ける<end>"): "수락한다<end>",
    norm_source("転送機について<end>"): "전송기에 대해<end>",
    norm_source("質問はない<end>"): "질문은 없다<end>",
    norm_source("盗人村<end>"): "도둑 마을<end>",
    norm_source("ディンガル士官<end>"): "딘갈 장교<end>",
    # Menu/choice labels: 町 is consistently localized as 마을 here, and the
    # 출입 pair is kept grammatically symmetrical.
    norm_source("町から出る<end>"): "마을에서 나간다<end>",
    norm_source("町から出ない<end>"): "마을에서 나가지 않는다<end>",
    # Speaker/UI labels whose semantics do not depend on scene context.
    norm_source("スラムの青年<end>"): "빈민가의 청년<end>",
    norm_source("て、敵襲！<end>"): "저, 적습이다!<end>",
    norm_source("へ、へい！<end>"): "예, 예!<end>",
    # Repeated narration / same-register dialogue. These Japanese sources are
    # identical, so divergent Korean register or paraphrase is not contextual.
    norm_source("お主がおるとわかっておれば、<line-break>敗残兵の殲滅を優先させたものを！<line-break>　<end>"): "네놈이 있는 줄 알았다면 패잔병 섬멸을 우선했을 것을!<end>",
    norm_source("ゼネテス…、<line-break>そうか、お主がおったのか！　　　<line-break>　<line-break>お主ならば、あの敗軍を<line-break>まとめる器量がある！<line-break>くっ、ぬかったわ！<end>"): "제네테스…. 그렇군, 네놈이 있었나! 네놈이라면 그 패잔군을 수습할 역량이 있다! 큭, 방심했군!<end>",
    norm_source("フッフッフ…。<line-break><line-break><line-break>ぐふっ。ムダだ…。<line-break>運命は…、変わらぬ…。<end>"): "후후후…. 크흑. 소용없다…. 운명은… 바뀌지 않는다….<end>",
    norm_source("　　　　　　　　　そうか。　　　　　　　　　<line-break>　　　　　ならば無理強いはすまい。<line-break>　<line-break>　<line-break>　　　　　うつろいやすき人の子よ。<line-break>　　　　　　気が変わったならば<line-break>　　　　　　また我を訪ねるがよい。<end>"): "그런가. 그렇다면 억지로 강요하진 않겠다. 변덕스러운 인간의 아이여. 마음이 바뀌면 다시 나를 찾아오거라.<end>",
    norm_source("　　…我ハ再び降臨セリ。<line-break>生あるモノすべてに滅びヲ…。<line-break>　一切ノ事象ヲ無に帰さん…。<end>"): "…나는 다시 강림했노라. 살아 있는 모든 것에 멸망을…. 모든 현상을 무로 돌리리라….<end>",
    norm_source("ついに復活したぜ！<line-break>これで本格的に活動できる！<line-break>ありがとよ。ギーヒッヒッ！<line-break>まず手始めに、お前を殺す！<line-break>マゴス様の恐怖のフルコース、<line-break>とくと味わえ！　ギーヒッ！<end>"): "드디어 부활했다! 이제 본격적으로 활동할 수 있어! 고맙다. 기히히힛! 우선 시작으로 네놈을 죽여 주마! 마고스 님의 공포 풀코스, 실컷 맛봐라! 기힛!<end>",
    norm_source("ネメアにより西方攻略を命じられた　<line-break>白虎将軍ジラークは、　　　　<line-break>難航不落といわれたアキュリュースを<line-break>ほぼ無傷のまま支配下に収めた。<end>"): "네메아에게 서방 공략을 명령받은 백호장군 지라크는 난공불락이라 불리던 아큐류스를 거의 피해 없이 지배하에 두었다.<end>",
    norm_source("王妃様にお会いになるのですね。<line-break>かしこまりました。<line-break>どうぞお通りください。<line-break>しかし、伯爵様、よろしいのですか？<line-break>お兄君のエリエナイ公は…<line-break>い、いえ、出過ぎだ発言でした。<end>"): "왕비님을 만나러 가시는군요. 알겠습니다. 들어가십시오. 하지만 백작님, 괜찮으십니까? 형님이신 에리에나이 공은…. 아, 아닙니다. 제가 주제넘은 말을 했습니다.<end>",
    norm_source("　　　　　　指導者を失った両国は<line-break>　　　一時的な停戦状態に陥るのだった。<end>"): "지도자를 잃은 두 나라는 일시적인 휴전 상태에 들어갔다.<end>",
    norm_source("円卓騎士には他に<line-break>アーギルシャイアなどがいますね。　　<line-break>　<end>"): "원탁기사에는 그 밖에도 아르길샤이어 등이 있습니다.<end>",
    norm_source("将来、旅の途中で出会った仲間を<line-break>呼び出すための装置です。<line-break>　<line-break>一緒に旅する仲間を<line-break>変更したくなったら<line-break>この猫屋敷に来てください。<end>"): "앞으로 여행 중 만난 동료를 불러내기 위한 장치입니다. 함께 여행할 동료를 바꾸고 싶어지면 이 고양이 저택으로 와 주세요.<end>",
    norm_source("破壊神ウルグの復活に<line-break>用いられた強力な魔道器です。　<line-break>　<line-break>１２種類あるといいます。<line-break>禁断の聖杯は<line-break>その中のひとつですね。<end>"): "파괴신 울그의 부활에 쓰인 강력한 마도기입니다. 12종류가 있다고 하지요. 금단의 성배도 그중 하나입니다.<end>",
    norm_source("そいつはお前が持っていてくれ。　　　<line-break>よくわからんが、依頼主が<line-break>そうしてくれって言ってたんでな。<end>"): "그건 네가 가지고 있어. 잘 모르겠지만 의뢰인이 그렇게 해 달라고 했거든.<end>",
    norm_source("他人のために働くことで、<line-break>魂を成長させていけば<line-break>いずれ、わかってきますよ。<end>"): "다른 사람을 위해 일하며 영혼을 성장시켜 가면 언젠가 알게 될 겁니다.<end>",
    norm_source("皇帝ネメアの世界統一構想のもと<line-break>南方攻略軍司令、朱雀将軍アンギルダンが　<line-break>ついにロストール攻略に動き出した。<end>"): "황제 네메아의 세계 통일 구상 아래 남방 공략군 사령관, 주작 장군 앙길단이 마침내 로스토올 공략에 나섰다.<end>",
}

SOURCE_TERM_VARIANTS: dict[str, dict[str, str]] = {
    "盗人村": {"도적 마을": "도둑 마을"},
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
            key = norm_source(ja)
            new = ko
            canonical = CANONICAL.get(key)
            if canonical is not None:
                new = canonical
            for ja_term, variants in SOURCE_TERM_VARIANTS.items():
                if ja_term not in ja:
                    continue
                for old, replacement in variants.items():
                    if old in new:
                        n = new.count(old)
                        new = new.replace(old, replacement)
                        counts[f"term:{ja_term}:{old}->{replacement}"] = counts.get(f"term:{ja_term}:{old}->{replacement}", 0) + n
            if new != ko:
                lines[i] = "korean = " + json.dumps(new, ensure_ascii=False)
                dirty = True
                changed_records += 1
                counts[key] = counts.get(key, 0) + 1
            current = None

        if dirty:
            path.write_text("\n".join(lines) + "\n", encoding="utf-8")
            changed_files += 1

    print(f"Repeated high-confidence round 2: {changed_records} records across {changed_files} files")
    for key, count in sorted(counts.items(), key=lambda kv: (-kv[1], kv[0])):
        print(f"  {count:4d}  {key}")


if __name__ == "__main__":
    main()
