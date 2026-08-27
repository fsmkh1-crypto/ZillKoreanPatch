#!/usr/bin/env python3
from pathlib import Path
import re

ROOT = Path("translations/korean/messages")
FIXES = {
    "330001": (
        "예. 여기서 장사도 끝났으니,로스토올로 향할 생각입니다.<end>",
        "예. 여기서 장사도 끝났으니 로스토올로 향할 생각입니다.<end>",
    ),
    "330007": (
        "지난번처럼,엄청 강한 몬스터에게 습격당하면어쩌려고 그러니.<end>",
        "지난번처럼 엄청 강한 몬스터에게 습격당하면 어쩌려고 그러니.<end>",
    ),
    "330010": (
        "루루안타, 나는 그 아이는 그 아이대로좋아하는 길을 가면 된다고 생각해.<end>",
        "루루안타, 나는 그 아이는 그 아이대로 좋아하는 길을 가면 된다고 생각해.<end>",
    ),
    "330015": (
        "…그렇구나. …릴비는 아빠도 엄마도 없어서 잘 모르겠어.<end>",
        "…그렇구나. …루루안타는 아빠도 엄마도 없어서 잘 모르겠어.<end>",
    ),
    "330033": (
        "<if><value:$29><equal>%0네가 플린트의 아들,<value:$28>지?<end>네가 플린트의 딸,<value:$28>지?<end>",
        "<if><value:$29><equal>%0네가 플린트의 아들, <value:$28>로군?<end>네가 플린트의 딸, <value:$28>로군?<end>",
    ),
    "330044": (
        "아니야! 정반대야! 게다가 아저씨도 아니야!22살! 한창 청춘이라고!!<end>",
        "아니야! 정반대야! 게다가 아저씨도 아니야! 22살! 한창 청춘이라고!!<end>",
    ),
    "330227": (
        "…고마워. 상냥하네, <value:$28>은. 그럼 가자! 빨리 돌아가지 않으면 플린트 씨가 걱정할 거야!<end>",
        "…고마워. 다정하네, <value:$28>. 자, 가자! 빨리 돌아가지 않으면 플린트 씨가 걱정할 거야!<end>",
    ),
    "1860020": (
        "릴비가 우리와 함께 다니게 된 지도 10년인가. …순식간이었군. 그 아이의 미소는 어머니를 잃은 어린 <value:$28>에게 큰 힘이 되어 줬어.<end>",
        "루루안타가 우리와 함께 다니게 된 지도 10년인가. …순식간이었군. 그 아이의 미소는 어머니를 잃은 어린 <value:$28>에게 큰 힘이 되어 줬어.<end>",
    ),
    "1860023": (
        "그래. 아니, 방금 널 찾으러 갔었어. 엇갈린 모양이네. 이제 로스토올로 출발할 때야. 미안하지만 <value:$28>, 릴비를 찾아와 주지 않을래?<end>",
        "그래. 아니, 방금 널 찾으러 갔었어. 엇갈린 모양이네. 이제 로스토올로 출발할 때야. 미안하지만 <value:$28>, 루루안타를 찾아와 주지 않을래?<end>",
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
print(f"context batch8: changed={changed} files={len(files_changed)}")
