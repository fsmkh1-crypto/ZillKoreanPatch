#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""Collect every Korean corpus character absent from the installed renderer font.

This mirrors layout.Engine.measureKoreanRenderer and koreanslots.RendererRune:
- Korean custom-atlas runes are accepted from the reviewed glyph catalog;
- the two explicit typography aliases are normalized before lookup;
- every other rune is encoded as stock CP932 and its exact renderer key must
  exist in release/font/metrics.toml.
"""
from __future__ import annotations

import collections
import pathlib
import re
import tomllib

ROOT = pathlib.Path(__file__).resolve().parents[2]
KOREAN = ROOT / "translations" / "korean" / "messages"
GLYPHS = ROOT / "release" / "korean" / "font" / "glyphs.toml"
METRICS = ROOT / "release" / "font" / "metrics.toml"
KOREAN_FILE = re.compile(r"^msgsec\d{3}(?:(?:-part\d{2})|b)?\.toml$")
CONTROL = re.compile(r"<[^>]+>")
ALIASES = {"~": "～", "‘": "'"}


def renderer_char(ch: str) -> str:
    return ALIASES.get(ch, ch)


def cp932_renderer_key(ch: str) -> int | None:
    try:
        encoded = ch.encode("cp932")
    except UnicodeEncodeError:
        return None
    if len(encoded) == 1:
        return encoded[0]
    if len(encoded) == 2:
        return encoded[0] | (encoded[1] << 8)
    return None


def parse_metric_keys() -> set[int]:
    with METRICS.open("rb") as fh:
        raw = tomllib.load(fh).get("glyph", {})
    out: set[int] = set()
    for key in raw:
        try:
            out.add(int(str(key), 0))
        except ValueError as exc:
            raise SystemExit(f"KOREAN_GLYPH_REPERTOIRE_FAIL invalid metric key {key!r}") from exc
    return out


def main() -> None:
    with GLYPHS.open("rb") as fh:
        catalog = set(tomllib.load(fh).get("glyphs", {}))
    installed = parse_metric_keys()

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

    failures: dict[tuple[int, str, str, str], int] = collections.Counter()
    bad_chars: collections.Counter[tuple[str, str]] = collections.Counter()
    aliases_used: collections.Counter[tuple[str, str]] = collections.Counter()
    for ident, pair in sorted(rows.items()):
        for field, text in zip(("korean", "layout"), pair):
            if not text:
                continue
            visible = CONTROL.sub("", text)
            for raw in visible:
                if raw in "\r\n\t":
                    continue
                ch = renderer_char(raw)
                if ch != raw:
                    aliases_used[(raw, ch)] += 1
                if ch in catalog:
                    continue
                renderer_key = cp932_renderer_key(ch)
                if renderer_key is None:
                    reason = "not_cp932_and_not_custom"
                elif renderer_key not in installed:
                    reason = f"missing_installed_metric_0x{renderer_key:04x}"
                else:
                    continue
                failures[(ident, field, raw, reason)] += 1
                bad_chars[(raw, reason)] += 1

    bad_record_ids = {ident for ident, _, _, _ in failures}
    print(
        "KOREAN_GLYPH_REPERTOIRE_SUMMARY "
        f"accepted_rows={len(rows)} installed_metric_keys={len(installed)} installed_custom={len(catalog)} "
        f"aliases={sum(aliases_used.values())} bad_characters={len(bad_chars)} bad_records={len(bad_record_ids)}"
    )
    for (raw, normalized), count in sorted(aliases_used.items()):
        print(
            f"KOREAN_GLYPH_ALIAS source={raw!r} unicode=U+{ord(raw):04X} "
            f"renderer={normalized!r} renderer_unicode=U+{ord(normalized):04X} occurrences={count}"
        )
    for (ch, reason), count in sorted(bad_chars.items(), key=lambda item: (ord(item[0][0]), item[0][1])):
        print(f"KOREAN_GLYPH_BAD_CHAR char={ch!r} unicode=U+{ord(ch):04X} reason={reason} occurrences={count}")
    for (ident, field, ch, reason), count in sorted(failures.items()):
        print(
            f"KOREAN_GLYPH_BAD_RECORD id={ident} field={field} char={ch!r} "
            f"unicode=U+{ord(ch):04X} reason={reason} occurrences={count}"
        )

    if len(rows) != 42016:
        raise SystemExit(f"KOREAN_GLYPH_REPERTOIRE_FAIL accepted rows={len(rows)} want 42016")
    if failures:
        raise SystemExit(
            f"KOREAN_GLYPH_REPERTOIRE_FAIL unsupported installed-font characters={len(bad_chars)} records={len(bad_record_ids)}"
        )
    print("KOREAN_GLYPH_REPERTOIRE_PASS")


if __name__ == "__main__":
    main()
