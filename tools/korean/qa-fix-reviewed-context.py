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
    "1590176": (
        "<if><value:$29><equal>%0후후, 그럼 루루안타가 오빠랑 함께 있어 줄게. 오빠, 이름이 뭐야?<end>후후, 그럼 루루안타가 언니랑 함께 있어 줄게. 언니, 이름이 뭐야?<end>",
        "<if><value:$29><equal>%0후후, 그럼 루루안타가 오빠랑 같이 있어 줄게. 오빠, 이름이 뭐야?<end>후후, 그럼 루루안타가 언니랑 같이 있어 줄게. 언니, 이름이 뭐야?<end>",
    ),
    "320064": (
        "…강해져야 한다.<end>",
        "…강해져라.<end>",
    ),
    "1010037": (
        "…이제 당신은 나 없이 가야 해.<end>",
        "…이제 너는 나 없이 가야 해.<end>",
    ),
    "460059": (
        "♪뜨겁게 타오르는 세계 제일의 혼~ ♪끓어오르는 악당을 향한 분노~<end>",
        "♪뜨겁게 타오르는 세계 제일의 영혼~ ♪끓어오르는 악당을 향한 분노~<end>",
    ),
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
    "1050016": (
        "그래, 네 아버지의 친우였지. 그 덕분에 꽤 쓰라린 일을 겪었어. 겉치레가 좋고 사람 마음을 사는 데 능했지. 같은 일을 해도 그 녀석은 칭찬받고 나는 그 몫까지 비난받았어.<end>",
        "그래, 네 아버지의 절친한 친구였지. 덕분에 아주 쓰라린 경험을 많이 했다. 겉모습이 좋고 인기를 끄는 데 능했지. 같은 일을 해도 녀석은 칭송받고 나는 그 녀석 몫까지 비난받았다.<end>",
    ),
    "1370244": (
        "앗, <value:$28> 님이시군.<end>",
        "앗, <value:$28> 님이구나.<end>",
    ),
    "1340378": (
        "『힘·예지·박애심, 모든 것을 갖춘 자에게 나의 유품 스트라스에지를 주어라』<end>",
        "『힘·지혜·박애심, 모든 것을 갖춘 자에게 나의 유품 스트라스 엣지를 주어라』<end>",
    ),
    "1340381": (
        "저희는 당신을이 유언에 걸맞은 인물이라고판단한 것입니다.<end>",
        "저희는 당신을 이 유언에 걸맞은 인물이라고 판단한 것입니다.<end>",
    ),
    "1340384": (
        "그 타르튜바를 구하려고 한 거야아.<end>",
        "그 타르튜바를 구하려고 한 거야.<end>",
    ),
    # Duplicate section-083 version of the same inheritance-event surface defects.
    "830007": (
        "저희는 당신을이 유언을 받을 만한 인물이라고판단한 것입니다.<end>",
        "저희는 당신을 이 유언을 받을 만한 인물이라고 판단한 것입니다.<end>",
    ),
    "830011": (
        "그 타르튜바를 구하려고 한 거야아.<end>",
        "그 타르튜바를 구하려고 한 거야.<end>",
    ),
    # Same Zagiv event and speaker. Use one polished rendering for both copies.
    "890163": (
        "당신이 길드에서 받은 몬스터 퇴치 의뢰 중에는 어둠의 괴물도 포함되어 있었어. 나도 계속 쫓고 있었지. 한 마리라도 더 이 세상에서 없애기 위해서.<end>",
        "당신이 길드에서 받은 몬스터 퇴치 의뢰 중에는 어둠의 괴물도 포함되어 있었어. 나도 계속 쫓고 있었지. 한 마리라도 더 이 세상에서 없애려고.<end>",
    ),
    "1140003": (
        "당신이 길드에서 받은 몬스터 토벌 의뢰 중에는 어둠의 괴물도 섞여 있었어. 나도 계속 쫓고 있었지. 한 마리라도 더 이 세상에서 없애려고.<end>",
        "당신이 길드에서 받은 몬스터 퇴치 의뢰 중에는 어둠의 괴물도 포함되어 있었어. 나도 계속 쫓고 있었지. 한 마리라도 더 이 세상에서 없애려고.<end>",
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
