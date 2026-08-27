#!/usr/bin/env python3
from __future__ import annotations
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
KOREAN_DIR = ROOT / "translations" / "korean" / "messages"

FIRST = "너, 나를 모르는 건가…? 그렇다면 충고 하나 하지. 힘에 매달리는 건 어리석은 일이다. 힘은 그저 더 강한 힘에 짓눌릴 뿐이야.<end>"
THANKS = "너, 나를 모르는 건가…? 그렇다면 감사 대신 충고 하나 하지. 힘에 매달리는 건 어리석은 일이다. 힘은 그저 더 강한 힘에 짓눌릴 뿐이야.<end>"

RULES = [
    # Elfas, exact repeated source. Nearby dialogue consistently uses 너/네 for 君.
    ("msgsec149-part99.toml", "1490137", "너, 나를 모르는 건가…? 그럼 한 가지 충고하지. 힘에 매달리는 건 어리석은 일이다. 힘은 그저 더 강한 힘에 의해 눌릴 뿐이야.<end>", FIRST),
    ("msgsec149-part99.toml", "1490148", "너, 나를 모르는 건가…? 그럼 한 가지 충고하지. 힘에 매달리는 건 어리석은 일이다. 힘은 그저 더 강한 힘에 의해 눌릴 뿐이야.<end>", FIRST),
    ("msgsec149-part99.toml", "1490194", "너, 나를 모르는 건가…? 그럼 한 가지 충고하지. 힘에 매달리는 건 어리석은 일이다. 힘은 그저 더 강한 힘에 의해 눌릴 뿐이야.<end>", FIRST),
    ("msgsec149-part99.toml", "1490270", "자네, 나를 모르는 건가…? 그렇다면 충고 하나 하지. 힘에 매달리는 건 어리석은 짓이다. 힘은 더 강한 힘에 짓눌릴 뿐이다.<end>", FIRST),
    ("msgsec149-part99.toml", "1490323", "자네, 나를 모르는 건가…? 그렇다면 충고 하나 하지. 힘에 매달리는 건 어리석은 짓이다. 힘은 더 강한 힘에 짓눌릴 뿐이다.<end>", FIRST),
    ("msgsec150-part99.toml", "1500022", "자네, 나를 모르는 건가…? 그럼 충고 하나 하지. 힘에 의지하는 건 어리석은 일이야. 힘은 그저 더 강한 힘에 짓눌릴 뿐이지.<end>", FIRST),

    ("msgsec149-part99.toml", "1490177", "자네, 나를 모르는 건가…? 그렇다면 감사 대신 하나 충고하지. 힘에 매달리는 건 어리석은 일이다. 힘은 그저 더 강한 힘에 짓눌릴 뿐이야.<end>", THANKS),
    ("msgsec149-part99.toml", "1490241", "자네, 나를 모르는 건가…? 그렇다면 감사 대신 하나 충고하지. 힘에 매달리는 건 어리석은 일이다. 힘은 그저 더 강한 힘에 짓눌릴 뿐이야.<end>", THANKS),
    ("msgsec149-part99.toml", "1490255", "자네, 나를 모르는 건가…? 그렇다면 감사 대신 하나 충고하지. 힘에 매달리는 건 어리석은 일이다. 힘은 그저 더 강한 힘에 짓눌릴 뿐이야.<end>", THANKS),
    ("msgsec149-part99.toml", "1490308", "너, 나를 모르는 건가…? 그렇다면 감사 대신 한 가지 충고하지. 힘에 의지하는 건 어리석은 일이다. 힘은 그저 더 강한 힘에게 짓눌릴 뿐이야.<end>", THANKS),
    ("msgsec151-part99.toml", "1510038", "너, 나를 모르는 건가…? 그렇다면 감사 대신 한 가지 충고하지. 힘에 의지하는 건 어리석은 일이다. 힘은 그저 더 강한 힘에게 짓눌릴 뿐이야.<end>", THANKS),
    ("msgsec153-part99.toml", "1530033", "너, 나를 모르는 건가…? 그렇다면 감사 대신 한 가지 충고하지. 힘에 의지하는 건 어리석은 일이다. 힘은 그저 더 강한 힘에게 짓눌릴 뿐이야.<end>", THANKS),

    # Clear contextual/semantic defects in the same Elfas volume.
    ("msgsec149-part99.toml", "1490049", "맑아질 리가 없어!<end>", "후련해질 리가 없어!<end>"),
    ("msgsec149-part99.toml", "1490105", "용왕은 신이 아니다. 그저 신의 이름으로이 세계를 감시할 뿐. 그것을 스스로 신이 된 것처럼어느새 자만하고,타락한 거군.<end>", "용왕은 신이 아니다. 그저 신의 이름으로 이 세계를 감시할 뿐. 그것을 스스로 신이 된 것처럼 어느새 자만하고, 타락한 거군.<end>"),
    ("msgsec149-part99.toml", "1490167", "으윽…. 파, 팔을 당했어.<end>", "으윽…. 파, 팔을 다쳤어.<end>"),
    ("msgsec149-part99.toml", "1490202", "으윽…. 파, 팔을 당했어.<end>", "으윽…. 파, 팔을 다쳤어.<end>"),
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
    print(f"elfas speaker batch5: changed={changed} files={len(files)}")

if __name__ == "__main__":
    main()
