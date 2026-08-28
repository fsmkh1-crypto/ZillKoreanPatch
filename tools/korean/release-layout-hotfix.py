#!/usr/bin/env python3
from pathlib import Path
import tomllib

ROOT = Path(__file__).resolve().parents[2]


def patch_record(path: Path, record_id: str, replacements: list[tuple[str, str]]) -> bool:
    text = path.read_text(encoding="utf-8")
    marker = f'["{record_id}"]\n'
    start = text.find(marker)
    if start < 0:
        raise SystemExit(f"missing record {record_id} in {path}")
    next_start = text.find('\n["', start + len(marker))
    if next_start < 0:
        next_start = len(text)
    block = text[start:next_start]
    original = block
    for old, new in replacements:
        if old not in block:
            if new in block:
                continue
            raise SystemExit(f"{record_id}: expected neither old nor already-new text: {old!r} -> {new!r}")
        block = block.replace(old, new)
    if block == original:
        return False
    rendered = text[:start] + block + text[next_start:]
    tomllib.loads(rendered)
    path.write_text(rendered, encoding="utf-8")
    return True


def main() -> int:
    changed = 0
    section1 = ROOT / "translations/korean/messages/msgsec001b.toml"
    # Restore canonical terminology after the temporary 0.9.4 experiment
    # incorrectly made semantic Korean follow stale generated layouts.
    changed += patch_record(section1, "10191", [("로스톨", "로스토올")])
    changed += patch_record(section1, "10193", [("로스톨", "로스토올")])
    changed += patch_record(section1, "10195", [("로시마", "롯시마")])
    changed += patch_record(section1, "10197", [("에어", "에아")])

    section3 = ROOT / "translations/korean/messages/msgsec003-part02.toml"
    # The generated layout carried one stale full-width trailing space that is
    # absent from translator-owned semantic Korean.
    changed += patch_record(section3, "30011", [("알고 있다면 알려 줘.　<end>", "알고 있다면 알려 줘.<end>")])

    print(f"release_layout_hotfix_changed_records={changed}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
