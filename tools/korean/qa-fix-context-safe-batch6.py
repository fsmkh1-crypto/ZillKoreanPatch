#!/usr/bin/env python3
from pathlib import Path
import re

ROOT = Path("translations/korean/messages")

FIXES = {
    "1360009": (
        "떠올리려 했다는 사실조차기억하지 못하겠지만나는 알고 있어.<end>",
        "떠올리려 했다는 사실조차 기억하지 못하겠지만 나는 알고 있어.<end>",
    ),
    "1360021": (
        "있잖아? 발로르을 쓰러뜨린 네메아는 왜 세계를 구한 용사라고 칭송받는 걸 싫어하는 거야? 응?<end>",
        "있잖아? 발로르를 쓰러뜨린 네메아는 왜 세계를 구한 용사라고 칭송받는 걸 싫어하는 거야? 응?<end>",
    ),
    "1360294": (
        "모든 것을 무로 돌릴 필요는 없어…. 내 안에서 잠들어. 그 마음과 함께….<end>",
        "모든 것이 무로 돌아가지는 않아…. 내 안에서 잠들렴. 그 마음과 함께….<end>",
    ),
    "1530007": (
        "제멋대로인 말이지만,나 때문에 언제까지나 슬퍼하지 말고빨리 기운을 차려 줬으면 좋겠어. 그럼…마지막으로 만나서 다행이야. 마타….<end>",
        "제멋대로인 말이지만, 나 때문에 언제까지나 슬퍼하지 말고 빨리 기운을 차려 줬으면 좋겠어. 그럼… 마지막으로 만나서 다행이야. 마타….<end>",
    ),
    "1530025": (
        "자, 덤벼!<end>",
        "자, 따라와!<end>",
    ),
    "2560027": (
        "…잠깐만요! 나를 무시하고 실험을진행하지 말아 줘요!<end>",
        "…잠깐만요! 나를 무시하고 실험을 진행하지 말아 줘요!<end>",
    ),
}

record_re = re.compile(r'^\["(\d+)"\]$')
changed = 0
files_changed = set()
found = set()

for path in ROOT.glob("msgsec*-part99.toml"):
    lines = path.read_text(encoding="utf-8").splitlines()
    current = None
    dirty = False
    for i, line in enumerate(lines):
        m = record_re.match(line)
        if m:
            current = m.group(1)
            continue
        if current in FIXES and line.startswith('korean = "'):
            old, new = FIXES[current]
            actual = line[len('korean = "'):-1]
            if actual != old:
                raise SystemExit(f"guard mismatch id={current} path={path}: {actual!r}")
            lines[i] = f'korean = "{new}"'
            changed += 1
            found.add(current)
            files_changed.add(str(path))
            dirty = True
            current = None
    if dirty:
        path.write_text("\n".join(lines) + "\n", encoding="utf-8")

missing = set(FIXES) - found
if missing:
    raise SystemExit(f"missing ids: {sorted(missing)}")

print(f"context-safe batch6: changed={changed} files={len(files_changed)}")
