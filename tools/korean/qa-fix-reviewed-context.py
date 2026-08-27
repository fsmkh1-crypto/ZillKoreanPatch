#!/usr/bin/env python3
from pathlib import Path
import re

ROOT = Path("translations/korean/messages")

# Human-reviewed context fixes. Keep this map idempotent: an already-correct
# record is accepted, while an unexpected third state aborts the run.
FIXES = {
    "1050007": (
        "…아저씨. 얼마 전 저는 누군가에게 목숨을 위협받았습니다. 다행히 <value:$28>이 달려와 줘서 무사했지만….<end>",
        "…아저씨. 며칠 전 저는 누군가에게 목숨을 위협받았습니다. 다행히 <value:$28>이 달려와 줘서 무사했지만….<end>",
    ),
    "1910109": (
        "…네 말투는 듣고 있을 수가 없군. 살인 기계였던 나를 구해 준 건 감사하고 있어. 하지만 노엘을 슬프게 하는 건 용서 못 해.<end>",
        "…네 말투는 도저히 듣고 있을 수가 없군. 살인 기계였던 나를 구해 준 건 고맙게 생각한다. 하지만 노엘을 슬프게 하는 건 용서 못 해.<end>",
    ),
    "300130": (
        "…아, 그러고 보니 부탁 하나. 미안하지만 가는 길에 술집에 들러서 이걸 페름에게 전해 줘.<end>",
        "…아, 내친김에 하나 더. 미안하지만 가는 길에 술집에 들러 이걸 페름에게 전해 줘.<end>",
    ),
    "1370479": (
        "…아, 내친김에 하나 더. 미안하지만 가는 길에 술집에 들러 이걸 펠름에게 전해 줘.<end>",
        "…아, 내친김에 하나 더. 미안하지만 가는 길에 술집에 들러 이걸 페름에게 전해 줘.<end>",
    ),
    "2560006": (
        "…잠깐만요! 나를 무시하고 실험을진행하지 말아 줘요!<end>",
        "…잠깐만요! 나를 무시하고 실험을 진행하지 말아 줘요!<end>",
    ),
    # Same Luruanta conditional dialogue reused in sections 036 and 159.
    "1590176": (
        "<if><value:$29><equal>%0후후, 그럼 루루안타가 오빠랑 함께 있어 줄게. 오빠, 이름이 뭐야?<end>후후, 그럼 루루안타가 언니랑 함께 있어 줄게. 언니, 이름이 뭐야?<end>",
        "<if><value:$29><equal>%0후후, 그럼 루루안타가 오빠랑 같이 있어 줄게. 오빠, 이름이 뭐야?<end>후후, 그럼 루루안타가 언니랑 같이 있어 줄게. 언니, 이름이 뭐야?<end>",
    ),
    # Same mentor scene and surrounding lines; section 032 diverged only here.
    "320064": (
        "…강해져야 한다.<end>",
        "…강해져라.<end>",
    ),
    # Same speaker/event: surrounding lines in sections 099 and 101 are also reused.
    "1010037": (
        "…이제 당신은 나 없이 가야 해.<end>",
        "…이제 너는 나 없이 가야 해.<end>",
    ),
    # Exact repeated hero-song lyric; use the standard noun already used elsewhere.
    "460059": (
        "♪뜨겁게 타오르는 세계 제일의 혼~ ♪끓어오르는 악당을 향한 분노~<end>",
        "♪뜨겁게 타오르는 세계 제일의 영혼~ ♪끓어오르는 악당을 향한 분노~<end>",
    ),
    # Exact repeated guild-registration line. The same guild-master context uses
    # casual 너 in the surrounding dialogue, so normalize the older 자네 variants.
    "420135": (
        "환술계 수치도 없고 범죄 등록에도 해당하지 않음. 축하해. 자네도 모험자로 등록됐어.<end>",
        "환술계 수치도 없고 범죄 등록에도 해당하지 않는군. 축하해. 너도 모험자로 등록됐어.<end>",
    ),
    "420143": (
        "환술계 수치도 없고 범죄 등록에도 해당하지 않음. 축하해. 자네도 모험자로 등록됐어.<end>",
        "환술계 수치도 없고 범죄 등록에도 해당하지 않는군. 축하해. 너도 모험자로 등록됐어.<end>",
    ),
    "460109": (
        "환술계 수치도 없고 범죄 등록에도 해당하지 않음. 축하해. 자네도 모험자로 등록됐어.<end>",
        "환술계 수치도 없고 범죄 등록에도 해당하지 않는군. 축하해. 너도 모험자로 등록됐어.<end>",
    ),
    "900002": (
        "환술계 수치도 없고 범죄 등록에도 해당하지 않음. 축하해. 자네도 모험자로 등록됐어.<end>",
        "환술계 수치도 없고 범죄 등록에도 해당하지 않는군. 축하해. 너도 모험자로 등록됐어.<end>",
    ),
}

record_re = re.compile(r'^\["(\d+)"\]$')
changed = 0
already = 0
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
            if actual == new:
                already += 1
                found.add(current)
                current = None
                continue
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
print(f"reviewed-context: changed={changed} already={already} files={len(files_changed)}")
