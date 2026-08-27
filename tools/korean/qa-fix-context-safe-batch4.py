#!/usr/bin/env python3
from __future__ import annotations
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
KOREAN_DIR = ROOT / "translations" / "korean" / "messages"

RULES = [
    # Same speaker / same source: Rossima refusal line.
    ("msgsec176-part99.toml", "1760012", "정말인가… 그럼 어쩔 수 없지. 난 여기서 기다릴 테니 네가 한가해지면 와. 모험의 재미를 가르쳐주마.<end>", "정말인가…. 그렇다면 어쩔 수 없지. 나는 여기서 기다릴 테니 네가 한가해지면 와라. 모험의 재미를 알려 주마.<end>"),
    # Visible word-concatenation artifacts in Rossima events.
    ("msgsec175-part99.toml", "1750037", "이게 저주 가스라면어쩔 셈이냐. 파티가 전멸한다고!<end>", "이게 저주 가스라면 어쩔 셈이냐. 파티가 전멸한다고!<end>"),
    ("msgsec176-part99.toml", "1760008", "나는 보다시피, 오른눈은 안 보이고왼손은 의수라서 말이야. 내 몸 지키는 것만으로도 벅차다네. 그러니 좀 봐주게나.<end>", "나는 보다시피, 오른눈은 안 보이고 왼손은 의수라서 말이야. 내 몸 지키는 것만으로도 벅차다네. 그러니 좀 봐주게나.<end>"),
    # Same speaker / same source: Noel thanks the player.
    ("msgsec089-part99.toml", "890696", "<value:$28>님, 오늘 정말 감사합니다! 저는 <value:$28>님이라면 반드시 구하러 와 주실 거라고 믿었어요!<end>", "<value:$28> 님, 오늘은 정말 감사했습니다! 저는 <value:$28> 님이라면 분명 구하러 와 주실 거라고 믿고 있었어요!<end>"),
    # Visible flattening/spacing artifacts in section 089.
    ("msgsec089-part99.toml", "890064", "라드라스가… 난다고?…공중도시라고? 그럴 수가….…설령 네 말이맞다고 해도, 라드라스를 띄우는 게네 소원이냐!?<end>", "라드라스가… 난다고? …공중도시라고? 그럴 수가…. …설령 네 말이 맞다고 해도, 라드라스를 띄우는 게 네 소원이냐!?<end>"),
    ("msgsec089-part99.toml", "890103", "아니, 이 수정이 라드라스를 비춘다면우리는 라드라스로 돌아가야겠지. <value:$28> 님도 꼭라드라스까지 와서 저희에게힘을 빌려 주십시오.<end>", "아니, 이 수정이 라드라스를 비춘다면 우리는 라드라스로 돌아가야겠지. <value:$28> 님도 꼭 라드라스까지 와서 저희에게 힘을 빌려 주십시오.<end>"),
    ("msgsec089-part99.toml", "890265", "오, 자네인가. 이런 때에 나를 찾아오다니자네도 참 운이 없군.<end>", "오, 자네인가. 이런 때에 나를 찾아오다니 자네도 참 운이 없군.<end>"),
    ("msgsec089-part99.toml", "890275", "일부러 나와 주느라수고했지만,당장 사라져 줘야겠어.<end>", "일부러 나와 주느라 수고했지만, 당장 사라져 줘야겠어.<end>"),
    ("msgsec089-part99.toml", "890331", "인류 혁신의 방침은 다르지만,<value:$28>는,짐에게도 무엇과도 바꿀 수 없는 존재다. 이런 곳에서,속절없이 잃고 싶지는 않다. 나쁜 제안은 아니겠지?<end>", "인류 혁신의 방침은 다르지만, <value:$28>는 짐에게도 무엇과도 바꿀 수 없는 존재다. 이런 곳에서 속절없이 잃고 싶지는 않다. 나쁜 제안은 아니겠지?<end>"),
    ("msgsec089-part99.toml", "890661", "…겨우,소울 이터를 쓰러뜨렸는데…. 이럴 리가 없었는데…. 나, 뭘 하고 있는 거지…결국 또 소중한 사람들을 잃고….<end>", "…겨우 소울 이터를 쓰러뜨렸는데…. 이럴 리가 없었는데…. 나, 뭘 하고 있는 거지… 결국 또 소중한 사람들을 잃고….<end>"),
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
    print(f"context-safe batch4: changed={changed} files={len(files)}")

if __name__ == "__main__":
    main()
