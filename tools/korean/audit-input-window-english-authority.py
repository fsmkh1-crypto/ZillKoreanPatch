#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""English-first audit for the two stock name/input command strips.

The upstream English patch localizes these surfaces as ordinary guarded EBOOT
fixed strings. It does not declare a keyboard/input executable patch feature.
This audit deliberately does not claim anything about the retail keyboard's
internal character tables; it only hard-gates the part that upstream actually
owns.
"""
from __future__ import annotations

import pathlib
import tomllib

ROOT = pathlib.Path(__file__).resolve().parents[2]
ENGLISH = ROOT / "release" / "strings" / "eboot.toml"
KOREAN = ROOT / "release" / "korean" / "strings" / "eboot.toml"
MANIFEST = ROOT / "patches" / "executable" / "manifest.toml"

EXPECTED = {
    0x243D50: {
        "source": "カナ　　かな　　英数　　ゆだねる　　空白　　取消　　入力完了",
        "english": "Kana  kana  ABC  auto  space  cancel  done",
        "korean": "가타  히라  영문  자동  공백  취소  완료",
        "commands": 7,
    },
    0x2465F8: {
        "source": "  カナ   かな   記号   ゆだねる   空白   取消   中止   完了",
        "english": "  Kana   kana   sym   auto   space   cancel   quit   done",
        "korean": "  가타   히라   기호   자동   공백   취소   중지   완료",
        "commands": 8,
    },
}


def load(path: pathlib.Path) -> dict:
    with path.open("rb") as fh:
        return tomllib.load(fh)


def entry(table: dict, offset: int) -> dict:
    # TOML integer-looking bare keys are represented as strings by tomllib.
    key = hex(offset)
    if key not in table:
        raise SystemExit(f"INPUT_WINDOW_FAIL missing field {key} in {table!r}")
    return table[key]


def main() -> None:
    english = load(ENGLISH)
    korean = load(KOREAN)
    manifest = load(MANIFEST)

    features = sorted({str(p.get("feature", "")) for p in manifest.get("patch", [])})
    forbidden = [f for f in features if any(word in f.lower() for word in ("keyboard", "input", "name-entry", "name_input"))]
    if forbidden:
        raise SystemExit(f"INPUT_WINDOW_FAIL unexpected executable input feature(s): {forbidden}")

    for offset, want in EXPECTED.items():
        en = entry(english, offset)
        ko = entry(korean, offset)
        if en.get("source") != want["source"] or en.get("replacement") != want["english"]:
            raise SystemExit(f"INPUT_WINDOW_FAIL upstream-English authority drift at {offset:#x}")
        if ko.get("source") != want["source"] or ko.get("replacement") != want["korean"]:
            raise SystemExit(f"INPUT_WINDOW_FAIL Korean command-strip drift at {offset:#x}")
        # Command count is semantic ordering, not layout geometry. Double/triple
        # spaces delimit the labels in the authenticated upstream strings.
        labels = [x for x in ko["replacement"].strip().split("  ") if x.strip()]
        if len(labels) != want["commands"]:
            raise SystemExit(
                f"INPUT_WINDOW_FAIL command count {len(labels)} at {offset:#x}; want {want['commands']}"
            )

    print(
        "INPUT_WINDOW_ENGLISH_AUTHORITY_PASS "
        "fields=2 executable_input_features=0 mechanics=stock-or-unproven "
        "scope=fixed-command-strips-only"
    )


if __name__ == "__main__":
    main()
