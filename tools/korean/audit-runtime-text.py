#!/usr/bin/env python3
import glob
import re
import tomllib
from pathlib import Path

VALUE_HANGUL = re.compile(r"<value:\$[0-9A-F]{2}>[가-힣]")
LINE_BREAK = "<line-break>"
TARGET_ID = "210065"


def load_rows(pattern: str):
    rows = {}
    locations = {}
    for filename in sorted(glob.glob(pattern)):
        path = Path(filename)
        with path.open("rb") as f:
            data = tomllib.load(f)
        for key, row in data.items():
            rows[key] = row
            locations[key] = path
    return rows, locations


def main() -> int:
    root = Path(__file__).resolve().parents[2]
    korean, korean_locations = load_rows(str(root / "translations/korean/messages/msgsec*.toml"))
    source, _ = load_rows(str(root / "translations/messages/msgsec*.toml"))

    translated = 0
    with_layout = 0
    without_layout = []
    without_layout_source_multiline = []
    value_adjacent = []

    for key, row in korean.items():
        text = row.get("korean", "")
        if not text:
            continue
        translated += 1
        layout = row.get("layout", "")
        if layout:
            with_layout += 1
        else:
            without_layout.append(key)
            japanese = source.get(key, {}).get("japanese", "")
            if LINE_BREAK in japanese:
                without_layout_source_multiline.append(key)

        for field in ("korean", "layout"):
            value = row.get(field, "")
            for match in VALUE_HANGUL.finditer(value):
                value_adjacent.append((key, field, match.group(0), korean_locations[key]))

    print(f"translated_records={translated}")
    print(f"records_with_layout={with_layout}")
    print(f"records_without_layout={len(without_layout)}")
    print(f"records_without_layout_but_source_multiline={len(without_layout_source_multiline)}")
    print(f"value_immediately_followed_by_hangul_occurrences={len(value_adjacent)}")
    print(f"value_immediately_followed_by_hangul_records={len(set(key for key, _, _, _ in value_adjacent))}")

    for key, field, token, path in value_adjacent:
        print(f"VALUE_HANGUL id={key} field={field} token={token} file={path.relative_to(root)}")

    print("missing_layout_source_multiline_ids=" + ",".join(without_layout_source_multiline))

    target = korean.get(TARGET_ID)
    if target is None:
        print(f"target_{TARGET_ID}=MISSING_FROM_KOREAN_CORPUS")
    else:
        path = korean_locations[TARGET_ID].relative_to(root)
        print(f"target_{TARGET_ID}_file={path}")
        print(f"target_{TARGET_ID}_korean={target.get('korean', '')}")
        print(f"target_{TARGET_ID}_layout={target.get('layout', '')}")
        print(f"target_{TARGET_ID}_has_layout={bool(target.get('layout', ''))}")
        print(f"target_{TARGET_ID}_source_multiline={LINE_BREAK in source.get(TARGET_ID, {}).get('japanese', '')}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
