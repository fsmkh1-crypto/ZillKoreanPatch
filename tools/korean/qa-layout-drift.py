#!/usr/bin/env python3
import argparse
import difflib
import glob
import json
import re
import tomllib
from pathlib import Path

LINE_BREAK = "<line-break>"
SPACE_RE = re.compile(r"\s+")
SECTION_RE = re.compile(r'^\["([0-9]+)"\]$')


def content_only(text: str) -> str:
    """Return semantic/control characters excluding layout boundaries/whitespace."""
    return "".join(ch for ch in text.replace(LINE_BREAK, "") if not ch.isspace())


def changes(korean: str, layout: str):
    a = content_only(korean)
    b = content_only(layout)
    sm = difflib.SequenceMatcher(a=a, b=b, autojunk=False)
    out = []
    for tag, i1, i2, j1, j2 in sm.get_opcodes():
        if tag == "equal":
            continue
        out.append((tag, a[i1:i2], b[j1:j2]))
    return out


def layout_content_with_positions(layout: str):
    chars = []
    positions = []
    i = 0
    while i < len(layout):
        if layout.startswith(LINE_BREAK, i):
            i += len(LINE_BREAK)
            continue
        ch = layout[i]
        if not ch.isspace():
            chars.append(ch)
            positions.append(i)
        i += 1
    return chars, positions


def sync_layout_content(korean: str, layout: str) -> str:
    """Update stale lexical content while preserving existing whitespace/breaks.

    Refuse edits that would cross a whitespace or <line-break> boundary; those are
    structural layout cases and must remain human-reviewed rather than guessed.
    """
    old_chars, positions = layout_content_with_positions(layout)
    target_chars = [ch for ch in korean if not ch.isspace()]
    sm = difflib.SequenceMatcher(a=old_chars, b=target_chars, autojunk=False)
    edits = []
    for tag, i1, i2, j1, j2 in sm.get_opcodes():
        if tag == "equal":
            continue
        replacement = "".join(target_chars[j1:j2])
        if i1 == i2:
            if i1 < len(positions):
                start = positions[i1]
            else:
                start = len(layout)
                end_tag = layout.rfind("<end>")
                if end_tag >= 0:
                    start = end_tag
            end = start
        else:
            start = positions[i1]
            end = positions[i2 - 1] + 1
            crossed = layout[start:end]
            if LINE_BREAK in crossed or any(ch.isspace() for ch in crossed):
                raise ValueError(f"lexical edit crosses layout boundary: {crossed!r}")
        edits.append((start, end, replacement))
    result = layout
    for start, end, replacement in reversed(edits):
        result = result[:start] + replacement + result[end:]
    if content_only(result) != content_only(korean):
        raise ValueError("content sync did not converge to Korean semantic text")
    return result


def toml_quote(value: str) -> str:
    # JSON double-quoted strings are compatible with TOML basic strings for the
    # escapes used by this corpus; keep Unicode readable.
    return json.dumps(value, ensure_ascii=False)


def rewrite_file(path: Path, fixes: dict[str, str]) -> None:
    lines = path.read_text(encoding="utf-8").splitlines(keepends=True)
    current = None
    changed = set()
    for index, line in enumerate(lines):
        stripped = line.rstrip("\r\n")
        match = SECTION_RE.match(stripped)
        if match:
            current = match.group(1)
            continue
        if current in fixes and stripped.startswith('layout = "'):
            newline = "\n" if line.endswith("\n") else ""
            lines[index] = f"layout = {toml_quote(fixes[current])}{newline}"
            changed.add(current)
    missing = set(fixes) - changed
    if missing:
        raise ValueError(f"failed to locate layout lines for IDs: {sorted(missing)}")
    rendered = "".join(lines)
    # Fail before writing if our output is not valid TOML.
    tomllib.loads(rendered)
    path.write_text(rendered, encoding="utf-8")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--fix", action="store_true", help="sync lexical layout content to canonical Korean while preserving boundaries")
    args = parser.parse_args()

    root = Path(__file__).resolve().parents[2]
    count = 0
    fixed = 0
    for filename in sorted(glob.glob(str(root / "translations/korean/messages/msgsec*.toml"))):
        path = Path(filename)
        with path.open("rb") as f:
            data = tomllib.load(f)
        file_fixes = {}
        for key, row in data.items():
            korean = row.get("korean", "")
            layout = row.get("layout", "")
            if not layout or content_only(korean) == content_only(layout):
                continue
            count += 1
            rel = path.relative_to(root)
            ops = changes(korean, layout)
            rendered = "; ".join(f"{tag} korean={a!r} layout={b!r}" for tag, a, b in ops)
            print(f"{key}@{rel}: {rendered}")
            if args.fix:
                file_fixes[key] = sync_layout_content(korean, layout)
        if args.fix and file_fixes:
            rewrite_file(path, file_fixes)
            fixed += len(file_fixes)
    print(f"layout_lexical_drift_count={count}")
    if args.fix:
        print(f"layout_lexical_fixed={fixed}")
        return 0
    return 1 if count else 0


if __name__ == "__main__":
    raise SystemExit(main())
