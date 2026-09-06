#!/usr/bin/env python3
import argparse
import difflib
import glob
import re
import tomllib
from pathlib import Path

LINE_BREAK = "<line-break>"
SPACE_RE = re.compile(r"\s+")
SECTION_RE = re.compile(r'^\["([0-9]+)"\]$')
CONTROL_RE = re.compile(r"<[^>]+>")

# Characters that must not be stranded at the beginning of a display line.
# This is intentionally broader than sentence-closing punctuation because Korean
# layouts can contain Japanese-style punctuation/quotes inherited from source
# structure or deliberate typography.
KINSOKU_LINE_START = set(",，、。.!?！？…」』”’)]）］】〉》・~〜─;；:：")


def content_only(text: str) -> str:
    """Return semantic/control characters excluding layout boundaries/whitespace."""
    return "".join(ch for ch in text.replace(LINE_BREAK, "") if not ch.isspace())


def semantic_spacing_key(text: str) -> str:
    """Compare semantic text while treating a display break as one whitespace run.

    Unlike content_only(), this deliberately preserves whether lexical tokens are
    separated. It therefore catches stale persisted layouts after edits such as
    `,다음` -> `, 다음` while still accepting a semantic space represented by a
    <line-break> in display layout.
    """
    return SPACE_RE.sub(" ", text.replace(LINE_BREAK, " ")).strip()


def changes(korean: str, layout: str):
    a = content_only(korean)
    b = content_only(layout)
    sm = difflib.SequenceMatcher(a=a, b=b, autojunk=False)
    out = []
    for tag, i1, i2, j1, j2 in sm.get_opcodes():
        if tag != "equal":
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


def _visible_line(line: str) -> str:
    line = line.replace("<end>", "")
    # Control tokens do not themselves make punctuation legal at line start.
    return CONTROL_RE.sub("", line)


def assert_layout_postconditions(original: str, result: str) -> None:
    """Cheap structural guards for the exceptional boundary-preserving sync path."""
    before = original.split(LINE_BREAK)
    after = result.split(LINE_BREAK)
    if len(before) != len(after):
        raise ValueError("layout sync changed display-line count")

    before_empty = sum(1 for line in before if not _visible_line(line).strip())
    after_empty = sum(1 for line in after if not _visible_line(line).strip())
    if after_empty > before_empty:
        raise ValueError("layout sync introduced an empty display line")

    for index, line in enumerate(after):
        visible = _visible_line(line)
        if visible and (visible[0].isspace() or visible[-1].isspace()):
            raise ValueError(f"layout sync introduced edge whitespace on line {index + 1}")
        stripped = visible.lstrip()
        if stripped and stripped[0] in KINSOKU_LINE_START:
            raise ValueError(
                f"layout sync stranded prohibited line-start character {stripped[0]!r} on line {index + 1}"
            )


def sync_layout_content(korean: str, layout: str) -> str:
    """Exceptional lexical sync for intentionally fixed display boundaries.

    Normal Beta1 semantic edits should invalidate persisted layout and let the
    established reflow engine derive a new layout. This helper remains only for
    intentionally fixed boundaries and therefore fails closed on ambiguous edits.
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
                if i1 > 0:
                    previous_end = positions[i1 - 1] + 1
                    gap = layout[previous_end:start]
                    boundary = gap.rfind(LINE_BREAK)
                    if boundary >= 0 and replacement:
                        prefix_len = 0
                        while prefix_len < len(replacement) and replacement[prefix_len] in KINSOKU_LINE_START:
                            prefix_len += 1
                        if prefix_len:
                            if prefix_len != len(replacement):
                                raise ValueError(
                                    "mixed prohibited-punctuation/text insertion at layout boundary requires reflow"
                                )
                            start = previous_end + boundary
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
    if semantic_spacing_key(result) != semantic_spacing_key(korean):
        raise ValueError("content sync did not preserve semantic whitespace; invalidate/reflow instead")
    assert_layout_postconditions(layout, result)
    return result


def remove_layout_lines(path: Path, ids: set[str]) -> int:
    """Atomically remove persisted layout fields for stale semantic rows."""
    if not ids:
        return 0
    lines = path.read_text(encoding="utf-8").splitlines(keepends=True)
    current = None
    removed = set()
    rendered_lines = []
    for line in lines:
        stripped = line.rstrip("\r\n")
        match = SECTION_RE.match(stripped)
        if match:
            current = match.group(1)
        if current in ids and stripped.startswith('layout = "'):
            removed.add(current)
            continue
        rendered_lines.append(line)
    missing = ids - removed
    if missing:
        raise ValueError(f"failed to locate layout lines for IDs: {sorted(missing)}")
    rendered = "".join(rendered_lines)
    tomllib.loads(rendered)
    path.write_text(rendered, encoding="utf-8")
    return len(removed)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--invalidate-drift",
        action="store_true",
        help="remove persisted layout when Korean semantic/spacing no longer matches so build-owned reflow rederives it",
    )
    parser.add_argument(
        "--fix",
        action="store_true",
        help="deprecated alias for --invalidate-drift; Beta1 no longer lexical-syncs ordinary semantic edits",
    )
    args = parser.parse_args()
    invalidate = args.invalidate_drift or args.fix

    root = Path(__file__).resolve().parents[2]
    drift_count = 0
    invalidated = 0
    persisted = 0
    for filename in sorted(glob.glob(str(root / "translations/korean/messages/msgsec*.toml"))):
        path = Path(filename)
        with path.open("rb") as f:
            data = tomllib.load(f)
        stale_ids = set()
        for key, row in data.items():
            korean = row.get("korean", "")
            layout = row.get("layout", "")
            if not layout:
                continue
            persisted += 1
            if semantic_spacing_key(korean) == semantic_spacing_key(layout):
                continue
            drift_count += 1
            stale_ids.add(str(key))
            rel = path.relative_to(root)
            ops = changes(korean, layout)
            rendered = "; ".join(f"{tag} korean={a!r} layout={b!r}" for tag, a, b in ops)
            if not rendered:
                rendered = "whitespace-only semantic drift"
            print(f"{key}@{rel}: {rendered}")
        if invalidate and stale_ids:
            invalidated += remove_layout_lines(path, stale_ids)

    print(f"persisted_layout_count={persisted}")
    print(f"layout_semantic_drift_count={drift_count}")
    if invalidate:
        print(f"layout_invalidated_for_rederive={invalidated}")
        return 0
    return 1 if drift_count else 0


if __name__ == "__main__":
    raise SystemExit(main())
