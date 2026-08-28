#!/usr/bin/env python3
import difflib
import glob
import re
import tomllib
from pathlib import Path

LINE_BREAK = "<line-break>"
SPACE_RE = re.compile(r"\s+")


def comparable(text: str) -> str:
    return SPACE_RE.sub(" ", text.replace(LINE_BREAK, " ")).strip()


def changes(korean: str, layout: str):
    a = comparable(korean)
    b = comparable(layout)
    sm = difflib.SequenceMatcher(a=a, b=b, autojunk=False)
    out = []
    for tag, i1, i2, j1, j2 in sm.get_opcodes():
        if tag == "equal":
            continue
        out.append((tag, a[i1:i2], b[j1:j2]))
    return out


def main() -> int:
    root = Path(__file__).resolve().parents[2]
    count = 0
    for filename in sorted(glob.glob(str(root / "translations/korean/messages/msgsec*.toml"))):
        with open(filename, "rb") as f:
            data = tomllib.load(f)
        for key, row in data.items():
            korean = row.get("korean", "")
            layout = row.get("layout", "")
            if not layout:
                continue
            if comparable(korean) == comparable(layout):
                continue
            count += 1
            rel = Path(filename).relative_to(root)
            ops = changes(korean, layout)
            rendered = "; ".join(f"{tag} korean={a!r} layout={b!r}" for tag, a, b in ops)
            print(f"{key}@{rel}: {rendered}")
    print(f"layout_drift_count={count}")
    return 1 if count else 0


if __name__ == "__main__":
    raise SystemExit(main())
