#!/usr/bin/env python3
from __future__ import annotations
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
KOREAN_DIR = ROOT / "translations" / "korean" / "messages"

RULES = [
    ("msgsec187-part99.toml", "1870001", "오늘은 노엘의 심부름으로너를 만나러 왔다. 실바를 습격한 소울 이터가용왕의 섬에 나타났다. 노엘은 먼저용왕의 섬으로 향했다.<end>", "오늘은 노엘의 심부름으로 너를 만나러 왔다. 실바를 습격한 소울 이터가 용왕의 섬에 나타났다. 노엘은 먼저 용왕의 섬으로 향했다.<end>"),
    ("msgsec187-part99.toml", "1870029", "카핀과 레이븐, 두 사람의 소울을 부활시켜라!<end>", "카핀과 레이븐, 두 사람의 소울을 부활시켜 주세요!<end>"),
    ("msgsec187-part99.toml", "1870052", "<value:$28> 님,오늘은 정말로감사했습니다! 저는 <value:$28> 님이라면분명 구하러 와 주실 거라고믿고 있었어요!<end>", "<value:$28> 님, 오늘은 정말 감사했습니다! 저는 <value:$28> 님이라면 분명 구하러 와 주실 거라고 믿고 있었어요!<end>"),
    ("msgsec187-part99.toml", "1870053", "저, 정말 다행이에요. <value:$28>님이라는 분을 만날 수 있어서요!<end>", "저, 정말 다행이에요. <value:$28> 님이라는 분을 만날 수 있어서요!<end>"),
    ("msgsec187-part99.toml", "1870057", "<value:$28>,우리를 구하러 와 줘서,정말 고마워. 나는 지금까지,용왕님의 가호라는 걸그렇게 믿진 않았지만,<end>", "<value:$28>, 우리를 구하러 와 줘서 정말 고마워. 나는 지금까지 용왕님의 가호라는 걸 그렇게 믿진 않았지만,<end>"),
    ("msgsec187-part99.toml", "1870058", "당신과 노엘을 만나게 해 준 것에는 감사하고 싶어.<end>", "당신과 노엘을 만나게 해 준 것만큼은 감사하고 싶어.<end>"),
    ("msgsec187-part99.toml", "1870059", "그럼, <value:$28>님, 실례하겠습니다!<end>", "그럼, <value:$28> 님, 실례하겠습니다!<end>"),
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
    block2 = block.replace(old_line, new_line, 1)
    path.write_text(text[:start] + block2 + text[end:], encoding="utf-8")
    return old != new

def main() -> None:
    changed = 0
    files: set[str] = set()
    for filename, rid, old, new in RULES:
        path = KOREAN_DIR / filename
        if replace_record(path, rid, old, new):
            changed += 1
            files.add(filename)
    print(f"context-safe batch3: changed={changed} files={len(files)}")

if __name__ == "__main__":
    main()
