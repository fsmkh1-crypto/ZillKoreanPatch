#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""Collect every Korean corpus character outside stock CP932 + installed glyphs.

This is intentionally aggregate/fail-closed: one bad character must not turn
release validation into a one-record-per-build chase.
"""
from __future__ import annotations

import collections
import pathlib
import re
import tomllib

ROOT = pathlib.Path(__file__).resolve().parents[2]
KOREAN = ROOT / "translations" / "korean" / "messages"
GLYPHS = ROOT / "release" / "korean" / "font" / "glyphs.toml"
KOREAN_FILE = re.compile(r"^msgsec\d{3}(?:(?:-part\d{2})|b)?\.toml$")
CONTROL = re.compile(r"<[^>]+>")


def stock_cp932(ch: str) -> bool:
    try:
        ch.encode("cp932")
        return True
    except UnicodeEncodeError:
        return False


def main() -> None:
    with GLYPHS.open("rb") as fh:
        catalog = set(tomllib.load(fh).get("glyphs", {}))

    rows: dict[int, tuple[str, str]] = {}
    for path in sorted(p for p in KOREAN.iterdir() if p.is_file() and KOREAN_FILE.fullmatch(p.name)):
        with path.open("rb") as fh:
            data = tomllib.load(fh)
        for key, value in data.items():
            try:
                ident = int(key)
            except ValueError:
                continue
            pair = (str(value.get("korean", "")), str(value.get("layout", "")))
            if ident in rows:
                if rows[ident] == pair:
                    continue
                raise SystemExit(f"KOREAN_GLYPH_REPERTOIRE_FAIL conflicting Korean ID {ident}")
            rows[ident] = pair

    failures: dict[tuple[int, str, str], int] = collections.Counter()
    bad_chars: collections.Counter[str] = collections.Counter()
    for ident, pair in sorted(rows.items()):
        for field, text in zip(("korean", "layout"), pair):
            if not text:
                continue
            visible = CONTROL.sub("", text)
            for ch in visible:
                if ch in "\r\n\t":
                    continue
                if stock_cp932(ch) or ch in catalog:
                    continue
                failures[(ident, field, ch)] += 1
                bad_chars[ch] += 1

    print(
        "KOREAN_GLYPH_REPERTOIRE_SUMMARY "
        f"accepted_rows={len(rows)} installed_custom={len(catalog)} "
        f"bad_characters={len(bad_chars)} bad_records={len({i for i, _, _ in failures})}"
    )
    for ch, count in sorted(bad_chars.items(), key=lambda item: (ord(item[0]), item[0])):
        print(f"KOREAN_GLYPH_BAD_CHAR char={ch!r} unicode=U+{ord(ch):04X} occurrences={count}")
    for (ident, field, ch), count in sorted(failures.items()):
        print(
            f"KOREAN_GLYPH_BAD_RECORD id={ident} field={field} char={ch!r} "
            f"unicode=U+{ord(ch):04X} occurrences={count}"
        )

    if len(rows) != 42016:
        raise SystemExit(f"KOREAN_GLYPH_REPERTOIRE_FAIL accepted rows={len(rows)} want 42016")
    if failures:
        raise SystemExit(
            f"KOREAN_GLYPH_REPERTOIRE_FAIL unsupported characters={len(bad_chars)} "
            f"records={len({i for i, _, _ in failures})}"
        )
    print("KOREAN_GLYPH_REPERTOIRE_PASS")


if __name__ == "__main__":
    main()
