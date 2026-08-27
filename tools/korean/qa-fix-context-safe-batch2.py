#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
KOREAN_DIR = ROOT / "translations" / "korean" / "messages"

RULES = [
    ("msgsec101-part99.toml", "1010064", "<value:$28>…. 당신도 참 질리지 않는 사람이네. 언제까지고 옛정이 통할 거라 생각하지 않았으면 좋겠어.<end>", "<value:$28>…. 당신도 참 끈질기네. 언제까지나 옛정이 통할 거라 생각하지 않았으면 해.<end>"),
    ("msgsec137-part99.toml", "1370117", "<value:$28>, …힘내라.<end>", "<value:$28>, …힘내.<end>"),
    ("msgsec256-part99.toml", "2560014", "…어, 어라? 이상하네, 안 돌아오잖아….<end>", "…어, 어라? 이상하네, 돌아오지 않아….<end>"),
    ("msgsec063-part99.toml", "630278", "호오, 웨딩드레스 가봉이라도 시작하는 건가?<end>", "호오, 웨딩드레스 가봉이라도 시작하나?<end>"),
    ("msgsec092-part99.toml", "920039", "정말 싫다니까, 자기 기준으로밖에 사물을 못 보다니. 역시 하등생물이란 그 정도인가 봐.<end>", "싫다~ 자기 기준으로밖에 사물을 못 보다니. 결국 하등생물이란 그 정도인가 봐.<end>"),
    ("msgsec137-part99.toml", "1370475", "좋아. 그럼 부탁한다. 네 혼자서도 충분하겠지만… 혹시 모르니 네 소꿉친구도 뒤따라 보내겠다. 사이좋게 지내.<end>", "좋아. 그럼 부탁한다. 뭐, 너 혼자서도 충분하겠지만… 혹시 모르니 네 소꿉친구도 뒤따라 보내겠다. 사이좋게 지내라.<end>"),
    ("msgsec140-part99.toml", "1400431", " 좋아 그렇다면 그대를 정령의 자리로 이끌겠다 고귀한 땅의 빛을 넘어 저 거인에게 도전하라 도전하고, 그리고 쓰러뜨려라<end>", "좋다. 그렇다면 너를 정령의 자리로 이끌겠다. 고결한 대지의 빛을 넘어 저 거인에게 도전하라. 도전하고, 쓰러뜨려라.<end>"),
    ("msgsec051-part99.toml", "510108", "후후…. 운명의 여신은 잔혹하구나! 여기까지 와서 날 버리다니…. 제네테스여…. 너 같은 사내가… 파로스의… 후계자였다니… 크윽….<end>", "후후… 운명의 여신은 잔혹하구나! 여기까지 와서 나를 버리다니…. 제네테스여…. 자네 같은 사내가… 파로스의… 후계자였다니… 윽…<end>"),
    ("msgsec101-part99.toml", "1010028", "로스토올의 <value:$28>! 이런 곳에서 뭘 하고 있지! 당장 나가! 이런 곳에 와도 되는 신분이라고 생각해?<end>", "로스토올의 <value:$28>! 이런 곳에서 뭘 하고 있지! 어서 나가! 이런 곳에 와도 되는 신분이라고 생각해?<end>"),
    ("msgsec101-part99.toml", "1010067", "로스토올의 <value:$28>! 이런 곳에서 뭘 하고 있지! 어서 나가! 이런 곳에 와도 되는 신분이라고 생각해?<end>", "로스토올의 <value:$28>! 이런 곳에서 뭘 하고 있지! 어서 나가! 이런 곳에 와도 되는 신분이라고 생각해?<end>"),
    ("msgsec106-part99.toml", "1060111", "그, 그럴 수가…. <value:$28>님이… 그럴 수가….<end>", "그, 그럴 수가…. <value:$28> 님이… 그런….<end>"),
    ("msgsec142-part99.toml", "1420628", "뭐, 뭐야 당신! 태도가 크네! 이미 고귀한 나와 같은 선에서 있잖아!<end>", "뭐, 뭐야 너! 태도가 건방지네! 이미 고귀한 나와 같은 선에 서 있잖아!<end>"),
    ("msgsec137-part99.toml", "1370020", "뭐야. 여기 귀족들 이야기인가?<end>", "뭐야. 여기 귀족들 이야기야?<end>"),
    ("msgsec137-part99.toml", "1370064", "뭐, 그렇지. 이런 시대에 일일이 그런 걸 생각할 여유도 없겠지. 하지만 모든 일에는 반드시 하나쯤 배울 것이 있다. 그리고 누구나 거기서 같은 걸 보는 건 아니야.<end>", "뭐, 그렇지. 이런 시대에 그런 것까지 일일이 생각할 순 없겠지. 하지만 말이야, 어떤 일이든 반드시 하나쯤 배울 게 있는 법이다. 그리고 누구나 거기서 같은 것을 발견하는 건 아니지.<end>"),
    ("msgsec149-part99.toml", "1490054", "놈을 떠올릴 때마다,온몸이 분함으로 꿰뚫리는 것 같아! 놈의 힘을 두려워한 건,죽음의 날갯소리뿐만이 아니야. 나도 그랬다고!<end>", "그 녀석을 떠올릴 때마다 온몸이 분함에 휩싸여! 그 힘을 두려워한 건 죽음의 날갯소리만이 아니야. 나도 그랬어!<end>"),
    ("msgsec215-part99.toml", "2150016", "…그래도 그녀 말도 당연하지. 그녀와 나는 사는 세계뿐 아니라 수명도 달라…. 완전히 다른 종족이니까….<end>", "…하지만 그녀가 하는 말도 당연하긴 하지. 확실히 그녀와 나는 사는 세계뿐 아니라 수명도 달라…. 완전히 다른 종족이니까….<end>"),
    ("msgsec196-part99.toml", "1960508", "박격! 트리플 고블린!<end>", "맹공! 트리플 고블린!<end>"),
    ("msgsec194-part99.toml", "1940007", "Burnin'! 부모와 자식 매<end>", "Burnin'! 부전자전<end>"),
    ("msgsec196-part99.toml", "1960512", "Burnin'! 부모와 자식 매<end>", "Burnin'! 부전자전<end>"),
    ("msgsec001-part01.toml", "10182", "Ｂｕｒｎｉｎ’！부전자전<end>", "Burnin'! 부전자전<end>"),
    ("msgsec002-part02.toml", "20112", "도망<end>", "도망친다<end>"),
]


def esc(s: str) -> str:
    return s.replace('\\', '\\\\').replace('"', '\\"')


def replace_record(path: Path, rid: str, old: str, new: str) -> bool:
    text = path.read_text(encoding="utf-8")
    marker = f'["{rid}"]'
    start = text.find(marker)
    if start < 0:
        raise SystemExit(f"missing id {rid} in {path.name}")
    nxt = text.find('\n["', start + len(marker))
    end = len(text) if nxt < 0 else nxt
    block = text[start:end]
    old_line = f'korean = "{esc(old)}"'
    new_line = f'korean = "{esc(new)}"'
    if old_line not in block:
        if new_line in block:
            return False
        raise SystemExit(f"unexpected Korean text {path.name}:{rid}")
    if old == new:
        return False
    block2 = block.replace(old_line, new_line, 1)
    path.write_text(text[:start] + block2 + text[end:], encoding="utf-8")
    return True


def main() -> None:
    changed = 0
    files: set[str] = set()
    for filename, rid, old, new in RULES:
        path = KOREAN_DIR / filename
        if replace_record(path, rid, old, new):
            changed += 1
            files.add(filename)
    print(f"context-safe batch2: changed={changed} files={len(files)}")

if __name__ == "__main__":
    main()
