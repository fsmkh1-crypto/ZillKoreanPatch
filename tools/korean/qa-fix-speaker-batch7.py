#!/usr/bin/env python3
from pathlib import Path
import re

ROOT = Path("translations/korean/messages")
FIXES = {
    "990109": (
        "…이제 당신은 나 없이 가야 해.<end>",
        "…이제 너는 나 없이 가야 해.<end>",
    ),
    "990114": (
        "…당신도 참 질리지 않는 사람이네. 같은 말 몇 번씩 하게 하지 마. 우리는 반대 길을 걷고 있어. 이제 여기 오지 마. 다음엔 봐주지 않을 테니까….<end>",
        "…너도 참 질리지 않는 사람이네. 같은 말 몇 번씩 하게 하지 마. 우리는 반대 길을 걷고 있어. 이제 여기 오지 마. 다음엔 봐주지 않을 테니까….<end>",
    ),
    "1010031": (
        "…빨리 가세요.<end>",
        "…빨리 가.<end>",
    ),
    "1010033": (
        "이제 저는 당신을 지킬 수 없어요. 이제 저는….<end>",
        "이제 난 너를 지켜 줄 수 없어. 이제 난….<end>",
    ),
    "1010036": (
        "<value:$28>…. …모처럼이지만 지금은 바빠. 시간이 있다면 대련 정도는 괜찮지만…<end>",
        "<value:$28>…. …모처럼이지만 지금은 바빠. 시간이 있으면 대련 정도는 괜찮지만….<end>",
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
print(f"speaker batch7: changed={changed} files={len(files_changed)}")
