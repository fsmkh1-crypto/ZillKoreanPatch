#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path
import tomllib

ROOT = Path(__file__).resolve().parents[2]
KOREAN_DIR = ROOT / "translations" / "korean" / "messages"

RULES = [
    ("msgsec198-part99.toml", "1980083", "좋아, 오늘이야말로 밤을 새서라도 별똥별을 찾아야지. 힘내자.<end>", "좋아, 오늘이야말로 밤을 새워서라도 별똥별을 찾아야지. 힘내자.<end>"),
    ("msgsec149-part99.toml", "1490254", "그래…. 그럼 나를 저 녀석들에게서 구했다고 생각한 건가? 말해 두지만 저 정도 인간에게 당할 만큼 나는 약하지 않아.<end>", "그래…. 그럼 날 저 녀석들에게서 구했다고 생각한 건가? 말해 두지만 저 정도 인간에게 당할 만큼 난 약하지 않아.<end>"),
    ("msgsec183-part99.toml", "1830045", "그럼 서류는 보내 둘 테니 아큐류스 신전으로 가 줘. 거기서 자세한 이야기를 들을 수 있을 거야. 앞으로는 아큐류스 병사라는 걸로 잘 부탁한다.<end>", "그럼 서류는 보내 둘 테니 아큐류스 신전으로 가 줘. 거기서 자세한 이야기가 있을 거야. 앞으로는 아큐류스 병사라는 걸로 잘 부탁한다.<end>"),
    ("msgsec139-part99.toml", "1390015", "귀환 중이던 제네테스도 억울한 죄로 처형되었다.<end>", "귀환하던 제네테스도 억울한 죄로 처형되었다.<end>"),
    ("msgsec139-part99.toml", "1390007", "정적을 제거하고 로스토올을 손에 넣은 레몬이었지만, 그 레몬마저 수수께끼의 실종을 맞았다.<end>", "정적을 제거하고 로스토올을 손에 넣은 레몬이었지만, 그 레몬마저 의문의 실종을 맞았다.<end>"),
    ("msgsec139-part99.toml", "1390016", "정적을 제거하고 로스토올을 손에 넣은 레몬이었지만, 그 레몬마저 수수께끼처럼 실종됐다.<end>", "정적을 제거하고 로스토올을 손에 넣은 레몬이었지만, 그 레몬마저 의문의 실종을 맞았다.<end>"),
    ("msgsec163-part99.toml", "1630131", "살아남은 딩갈 병사들은 드워프 왕국을 지나 딩갈 영내로 철수했다.<end>", "살아남은 딩갈 병사들은 드워프 왕국을 빠져나가 딩갈 영내로 철수했다.<end>"),
    ("msgsec056-part99.toml", "560485", "…만지지 마. 저항은 하지 않을게.<end>", "…만지지 마. 저항은 안 할게.<end>"),
    ("msgsec055-part99.toml", "550184", "반…, 이럴 때 그런 농담은 그만해. 그것보다 <value:$28>, 저 검은 옷을 입은 자도….<end>", "반…, 이런 때 그런 농담은 그만하자. 그것보다, <value:$28>, 저 검은 옷의 사람도….<end>"),
    ("msgsec106-part99.toml", "1060009", "앞으로의 활약을 기대하고 있겠다. 앞으로도 잘 부탁한다. 우리의 동지, 로센의 종의 영웅이여.<end>", "앞으로의 활약을 기대하겠다. 앞으로도 잘 부탁한다. 우리의 동지, 로센의 종의 영웅이여.<end>"),
    ("msgsec142-part99.toml", "1420394", "세상은 다시 당신에게 다양한 모습을 보여 줄 겁니다. 자유로운 여행을!<end>", "세계는 다시 당신에게 여러 모습을 보여 줄 겁니다. 자유로운 여행을!<end>"),
    ("msgsec135-part99.toml", "1350103", "남의 안색을 살핀다는 건 그만큼 배려할 줄 아는 사람이라는 뜻이지.<end>", "남의 눈치를 살핀다는 건 배려할 줄 아는 사람이라는 뜻이야.<end>"),
    ("msgsec134-part99.toml", "1340522", "물으면 안 된다 고브! 딱히 휴고의 보물이 탐나서 부흥을 돕는 게 아니다 고브!<end>", "들으면 안 된다 고브! 휴고의 보물이 탐나서 부흥을 돕는 게 아니다 고브!<end>"),
    ("msgsec059-part99.toml", "590115", "물러나라! 흩어져서 지하왕국을 통과해라! 모두… 살아남는 거다…!<end>", "물러나라! 흩어져 지하 왕국을 통과해라! 모두… 살아남는… 거다…!<end>"),
    ("msgsec125-part99.toml", "1250164", "꺅! 반란이라고! 이길 수 있을까?<end>", "꺅! 반란이라고! 승산은 있는 거야?<end>"),
    ("msgsec089-part99.toml", "890148", "<value:$28> 님. 의뢰드릴 일이 있습니다. 급히 엔샨트 정청으로 와 주십시오. 자기브 딘갈<end>", "<value:$28> 공께. 의뢰드리고 싶은 일이 있습니다. 지체 없이 엔샨트 정청으로 와 주십시오. 자기브 딘갈<end>"),
    ("msgsec089-part99.toml", "890150", "<value:$28> 공께. 의뢰드리고 싶은 일이 있습니다. 지체 없이 엔샨트 정청으로 와 주십시오. 자기브 딩갈.<end>", "<value:$28> 공께. 의뢰드리고 싶은 일이 있습니다. 지체 없이 엔샨트 정청으로 와 주십시오. 자기브 딘갈<end>"),
    ("msgsec020-part96.toml", "200028", "HP 회복 계열 마법<end>", "HP 회복 계열 마법<end>"),
    ("msgsec165-part99.toml", "1650074", "HP 회복계 마법<end>", "HP 회복 계열 마법<end>"),
]


def replace_record(path: Path, rid: str, old: str, new: str) -> bool:
    text = path.read_text(encoding="utf-8")
    marker = f'["{rid}"]'
    start = text.find(marker)
    if start < 0:
        raise SystemExit(f"missing id {rid} in {path.name}")
    nxt = text.find('\n["', start + len(marker))
    end = len(text) if nxt < 0 else nxt
    block = text[start:end]
    old_line = f'korean = "{old.replace(chr(92), chr(92)+chr(92)).replace(chr(34), chr(92)+chr(34))}"'
    new_line = f'korean = "{new.replace(chr(92), chr(92)+chr(92)).replace(chr(34), chr(92)+chr(34))}"'
    if old == new:
        if old_line not in block:
            raise SystemExit(f"expected canonical record missing {path.name}:{rid}")
        return False
    if old_line not in block:
        if new_line in block:
            return False
        raise SystemExit(f"unexpected Korean text {path.name}:{rid}")
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
    print(f"context-safe batch1: changed={changed} files={len(files)}")

if __name__ == "__main__":
    main()
