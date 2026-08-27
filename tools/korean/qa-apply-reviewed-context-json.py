#!/usr/bin/env python3
from pathlib import Path
import json
import re

ROOT = Path("translations/korean/messages")
FIX_PATH = Path("tools/korean/reviewed-context-fixes.json")

fixes = json.loads(FIX_PATH.read_text(encoding="utf-8"))
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
        if current in fixes and line.startswith('korean = "'):
            item = fixes[current]
            old = item["old"]
            new = item["new"]
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

missing = set(fixes) - found
if missing:
    raise SystemExit(f"missing ids: {sorted(missing)}")
print(f"reviewed-context-json: changed={changed} already={already} files={len(files_changed)}")
