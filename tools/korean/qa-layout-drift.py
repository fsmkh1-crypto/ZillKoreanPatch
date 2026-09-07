#!/usr/bin/env python3
import argparse
import difflib
import glob
import re
import tomllib
from pathlib import Path

LINE_BREAK = "<line-break>"
SECTION_RE = re.compile(r'^\["([0-9]+)"\]$')
CONTROL_RE = re.compile(r"<[^>]+>")

# Characters that should not be newly stranded at the beginning of a display
# line by lexical synchronization. Pre-existing intentional line starts are not
# retroactively rejected; the guard is about newly-created break artifacts.
KINSOKU_LINE_START = set(",，、。.!?！？…」』”’)]）］】〉》・~〜─;；:：")


def content_only(text: str) -> str:
    """Return semantic/control characters excluding layout boundaries/whitespace."""
    return "".join(ch for ch in text.replace(LINE_BREAK, "") if not ch.isspace())


def _gap_signature(text: str) -> tuple[str, set[int], set[int]]:
    """Return lexical text plus regular-space and line-break gap positions.

    A gap position N means a separator occurred after lexical character N-1 and
    before lexical character N. Display line breaks are allowed inside Korean
    eojeols, while semantic whitespace must remain represented either by a real
    layout space or by a display line break.
    """
    chars = []
    regular_space_gaps = set()
    line_break_gaps = set()
    pending_regular = False
    pending_break = False
    i = 0
    while i < len(text):
        if text.startswith(LINE_BREAK, i):
            pending_break = True
            i += len(LINE_BREAK)
            continue
        ch = text[i]
        if ch.isspace():
            pending_regular = True
            i += 1
            continue
        if chars:
            gap = len(chars)
            if pending_regular:
                regular_space_gaps.add(gap)
            if pending_break:
                line_break_gaps.add(gap)
        chars.append(ch)
        pending_regular = False
        pending_break = False
        i += 1
    return "".join(chars), regular_space_gaps, line_break_gaps


def spacing_equivalent(korean: str, layout: str) -> bool:
    """Legacy gap-level equivalence used by the exceptional sync helper.

    This is intentionally not the authoritative stale-layout predicate because a
    set of gap positions cannot retain repeated-boundary cardinality. The main
    audit uses preserves_layout_semantics(), which mirrors the Go compiler.
    """
    k_chars, k_spaces, k_breaks = _gap_signature(korean)
    l_chars, l_spaces, l_breaks = _gap_signature(layout)
    if k_chars != l_chars:
        return False
    required = k_spaces | k_breaks
    represented = l_spaces | l_breaks
    if not required.issubset(represented):
        return False
    if not l_spaces.issubset(required):
        return False
    return True


def semantic_units(text: str) -> list[tuple[str, str]]:
    """Tokenize exactly like internal/message semanticUnits.

    Annotated controls stay atomic, <line-break> is a layout boundary, Unicode
    whitespace is semantic whitespace, and every other Unicode code point is one
    literal unit. Python strings already iterate by Unicode code point, matching
    the Go rune-level contract for valid UTF-8 source text.
    """
    units: list[tuple[str, str]] = []
    i = 0
    while i < len(text):
        match = CONTROL_RE.match(text, i)
        if match:
            value = match.group(0)
            units.append(("boundary" if value == LINE_BREAK else "control", value))
            i = match.end()
            continue
        ch = text[i]
        units.append(("whitespace" if ch.isspace() else "literal", ch))
        i += 1
    return units


def preserves_layout_semantics(semantic: str, layout: str) -> bool:
    """Mirror internal/message.preservesSemantics for preflight invalidation.

    Keeping the Python drift audit and the Go compiler on the same contract is
    critical: in particular, multiple consecutive layout boundaries require at
    least the same number of semantic whitespace runes, while one generated
    boundary may split adjacent ordinary text runes at zero width.
    """
    want = semantic_units(semantic)
    got = semantic_units(layout)
    want_index = 0
    got_index = 0

    while got_index < len(got):
        if want_index < len(want) and want[want_index] == got[got_index]:
            want_index += 1
            got_index += 1
            continue

        if (
            want_index < len(want)
            and want[want_index][0] == "whitespace"
            and got[got_index][0] == "whitespace"
        ):
            while want_index < len(want) and want[want_index][0] == "whitespace":
                want_index += 1
            while got_index < len(got) and got[got_index][0] == "whitespace":
                got_index += 1
            continue

        if got[got_index][0] != "boundary":
            return False

        boundary_end = got_index
        while boundary_end < len(got) and got[boundary_end][0] == "boundary":
            boundary_end += 1
        boundary_count = boundary_end - got_index

        if want_index < len(want) and want[want_index][0] == "whitespace":
            whitespace_end = want_index
            while whitespace_end < len(want) and want[whitespace_end][0] == "whitespace":
                whitespace_end += 1
            whitespace_count = whitespace_end - want_index
            if boundary_count > 1 and boundary_count > whitespace_count:
                return False
            want_index = whitespace_end
            got_index = boundary_end
            continue

        if (
            boundary_count == 1
            and want_index > 0
            and want_index < len(want)
            and want[want_index - 1][0] == "literal"
            and want[want_index][0] == "literal"
        ):
            got_index += 1
            continue
        return False

    return want_index == len(want)


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
    return CONTROL_RE.sub("", line)


def _line_start_char(line: str) -> str:
    visible = _visible_line(line).lstrip()
    return visible[0] if visible else ""


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
        before_visible = _visible_line(before[index])
        before_edge_space = bool(before_visible and (before_visible[0].isspace() or before_visible[-1].isspace()))
        after_edge_space = bool(visible and (visible[0].isspace() or visible[-1].isspace()))
        if after_edge_space and not before_edge_space:
            raise ValueError(f"layout sync introduced edge whitespace on line {index + 1}")

        before_start = _line_start_char(before[index])
        after_start = _line_start_char(line)
        if after_start in KINSOKU_LINE_START and before_start != after_start:
            raise ValueError(
                f"layout sync stranded prohibited line-start character {after_start!r} on line {index + 1}"
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
    if not preserves_layout_semantics(korean, result):
        raise ValueError("content sync did not preserve compiler semantic contract; invalidate/reflow instead")
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
            if preserves_layout_semantics(korean, layout):
                continue
            drift_count += 1
            stale_ids.add(str(key))
            rel = path.relative_to(root)
            ops = changes(korean, layout)
            rendered = "; ".join(f"{tag} korean={a!r} layout={b!r}" for tag, a, b in ops)
            if not rendered:
                rendered = "whitespace/boundary semantic drift"
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
