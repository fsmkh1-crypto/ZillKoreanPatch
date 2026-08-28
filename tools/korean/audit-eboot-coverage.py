#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later

"""Audit sparse Korean EBOOT fixed-string coverage against the English manifest.

This is diagnostic for translation coverage, but structural mistakes are fatal:
- every Korean offset must exist in the canonical EBOOT string table;
- the guarded Japanese source must match exactly;
- the Korean runtime byte estimate must not exceed the retail fixed-width field.

Missing Korean offsets are reported, not treated as failure, so coverage can grow
incrementally while still making omissions measurable.
"""

from __future__ import annotations

import pathlib
import sys
import tomllib

ROOT = pathlib.Path(__file__).resolve().parents[2]
ENGLISH = ROOT / "release" / "strings" / "eboot.toml"
KOREAN = ROOT / "release" / "korean" / "strings" / "eboot.toml"


def load(path: pathlib.Path) -> dict[str, dict[str, str]]:
    with path.open("rb") as f:
        data = tomllib.load(f)
    return data


def source_bytes(text: str) -> int:
    return len(text.encode("cp932"))


def runtime_bytes(text: str) -> int:
    total = 0
    for ch in text:
        code = ord(ch)
        if 0xAC00 <= code <= 0xD7A3:  # Hangul syllable -> custom double-byte renderer key
            total += 2
            continue
        try:
            total += len(ch.encode("cp932"))
        except UnicodeEncodeError:
            # Any other custom rune would also require one double-byte renderer key.
            total += 2
    return total


def main() -> int:
    english = load(ENGLISH)
    korean = load(KOREAN)
    errors: list[str] = []

    for offset, field in sorted(korean.items(), key=lambda item: int(item[0], 0)):
        base = english.get(offset)
        if base is None:
            errors.append(f"{offset}: Korean offset not present in release/strings/eboot.toml")
            continue
        if field.get("source") != base.get("source"):
            errors.append(f"{offset}: guarded source differs from canonical EBOOT source")
            continue
        capacity = source_bytes(field["source"])
        used = runtime_bytes(field["replacement"])
        if used > capacity:
            errors.append(f"{offset}: Korean replacement uses {used} bytes; capacity is {capacity}")

    missing = sorted(set(english) - set(korean), key=lambda value: int(value, 0))
    print(
        "Korean EBOOT coverage: "
        f"translated={len(korean)} total={len(english)} missing={len(missing)} "
        f"coverage={len(korean) / len(english) * 100:.1f}%"
    )
    if missing:
        preview = ", ".join(missing[:20])
        suffix = " ..." if len(missing) > 20 else ""
        print(f"Korean EBOOT missing offsets (first 20): {preview}{suffix}")

    if errors:
        for error in errors:
            print(f"ERROR: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
