#!/usr/bin/env python3
from pathlib import Path
import re

ROOT = Path("translations/korean/messages")
FIXES = {
    "1360021": (
        "있잖아? 발로르을 쓰러뜨린 네메아는 왜 세계를 구한 용사라고 칭송받는 걸 싫어하는 거야? 응?<end>",
        "있잖아? 발로르를 쓰러뜨린 네메아는 왜 세계를 구한 용사라고 칭송받는 걸 싫어하는 거야? 응?<end>",
    ),
    "1530007": (
        "제멋대로인 말이지만,나 때문에 언제까지나 슬퍼하지 말고빨리 기운을 차려 줬으면 좋겠어. 그럼…마지막으로 만나서 다행이야. 마타….<end>",
        "제멋대로인 말이지만, 나 때문에 언제까지나 슬퍼하지 말고 빨리 기운을 차려 줬으면 좋겠어. 그럼… 마지막으로 만나서 다행이야. 마타….<end>",
    ),
    "1910017": (
        "아니요, <value:$28> 님이곁에 있어 주시면 제가 불안에휩싸일 일도 없어질 테니까요. 그러니, <value:$28> 님. 또 저를 만나러 와 주세요.<end>",
        "아니요, <value:$28> 님이 곁에 있어 주시면 제가 불안에 휩싸일 일도 없어질 테니까요. 그러니, <value:$28> 님. 또 저를 만나러 와 주세요.<end>",
    ),
    "1910019": (
        "제 불안을 없앨 수 있는 건<value:$28> 님이 곁에 계실 때느끼는 안도감인 것 같아요. 그러니, <value:$28> 님. 또 저를 만나러 와 주세요.<end>",
        "제 불안을 없앨 수 있는 건 <value:$28> 님이 곁에 계실 때 느끼는 안도감인 것 같아요. 그러니, <value:$28> 님. 또 저를 만나러 와 주세요.<end>",
    ),
    "1910020": (
        "…티아나님은 아직 돌아오지 않으셨다고 들었습니다. 성 사람들은 모두 걱정하고 있어요. 하지만 만약 제가 사라진다면 이렇게까지 걱정해 줄까요? 문득 그런 생각을 하게 됩니다.<end>",
        "…티아나 님은 아직 돌아오지 않으셨다고 들었습니다. 성 사람들은 모두 걱정하고 있어요. 하지만 만약 제가 사라진다면 이렇게까지 걱정해 줄까요? 문득 그런 생각을 하게 됩니다.<end>",
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
print(f"surface batch9: changed={changed} files={len(files_changed)}")
