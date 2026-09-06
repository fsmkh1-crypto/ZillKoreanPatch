#!/usr/bin/env python3
import argparse
import importlib.util
import json
import re
import tomllib
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
MODULE_PATH = Path(__file__).with_name("qa-layout-drift.py")
SPEC = importlib.util.spec_from_file_location("qa_layout_drift", MODULE_PATH)
qa = importlib.util.module_from_spec(SPEC)
assert SPEC.loader is not None
SPEC.loader.exec_module(qa)

COMMA_NO_SPACE_RE = re.compile(r",(?=[^\s<])")
DIGIT_COMMA_SPACE_DIGIT_RE = re.compile(r"\d,\s+\d")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--json", type=Path)
    args = parser.parse_args()

    persisted = 0
    spacing_drift = []
    comma_no_space_layout = []
    kinsoku_start = []
    digit_comma_space_digit = []

    for path in sorted((ROOT / "translations/korean/messages").glob("msgsec*.toml")):
        data = tomllib.loads(path.read_text(encoding="utf-8"))
        rel = path.relative_to(ROOT).as_posix()
        for rid, row in data.items():
            korean = row.get("korean")
            if isinstance(korean, str):
                for m in DIGIT_COMMA_SPACE_DIGIT_RE.finditer(korean):
                    digit_comma_space_digit.append({"path": rel, "id": str(rid), "match": m.group(0), "korean": korean})
            layout = row.get("layout")
            if not isinstance(korean, str) or not isinstance(layout, str) or not layout:
                continue
            persisted += 1
            if qa.semantic_spacing_key(korean) != qa.semantic_spacing_key(layout):
                spacing_drift.append({"path": rel, "id": str(rid), "korean": korean, "layout": layout})
            if COMMA_NO_SPACE_RE.search(layout):
                comma_no_space_layout.append({"path": rel, "id": str(rid), "korean": korean, "layout": layout})
            for line_no, line in enumerate(layout.split(qa.LINE_BREAK), start=1):
                visible = qa._visible_line(line).lstrip()
                if visible and visible[0] in qa.KINSOKU_LINE_START:
                    kinsoku_start.append({"path": rel, "id": str(rid), "line": line_no, "char": visible[0], "layout": layout})

    result = {
        "status": "READ_ONLY_AUDIT",
        "persisted_layout_count": persisted,
        "semantic_spacing_drift_count": len(spacing_drift),
        "comma_no_space_layout_count": len(comma_no_space_layout),
        "kinsoku_line_start_count": len(kinsoku_start),
        "digit_comma_space_digit_count": len(digit_comma_space_digit),
        "semantic_spacing_drift": spacing_drift,
        "comma_no_space_layout": comma_no_space_layout,
        "kinsoku_line_start": kinsoku_start,
        "digit_comma_space_digit": digit_comma_space_digit,
    }
    rendered = json.dumps(result, ensure_ascii=False, indent=2)
    print(json.dumps({k: v for k, v in result.items() if not isinstance(v, list)}, ensure_ascii=False, sort_keys=True))
    if args.json:
        args.json.write_text(rendered + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
