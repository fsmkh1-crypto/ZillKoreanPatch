#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""Inventory source message rows that are still outside the Korean overlay.

English-first rule: this is discovery only. A missing row is not automatically
safe to translate. The report groups rows by upstream category and flags likely
system/settings, tutorial/help, and input/name-entry surfaces for consumer review.
"""
from __future__ import annotations

import collections
import pathlib
import re
import tomllib

ROOT = pathlib.Path(__file__).resolve().parents[2]
SOURCE = ROOT / "translations" / "messages"
KOREAN = ROOT / "translations" / "korean" / "messages"
CATEGORIES = ROOT / "release" / "layout" / "categories.toml"

SURFACES = {
    "settings-system": re.compile(r"(?:設定|環境|コンフィグ|オプション|音量|ＢＧＭ|BGM|ＳＥ|SE|振動|画面|明るさ|キー設定|操作設定|ロード|セーブ)", re.I),
    "tutorial-help": re.compile(r"(?:チュートリアル|ヘルプ|説明|操作方法|遊び方|について|とは|表示され|ボタン|押して|選択して|決定して|キャンセル)", re.I),
    "input-name": re.compile(r"(?:名前|入力|文字|ひらがな|カタカナ|英数|アルファベット|決定|削除|戻る|変換|スペース)", re.I),
}


def load_rows(directory: pathlib.Path) -> dict[int, dict]:
    rows: dict[int, dict] = {}
    for path in sorted(directory.glob("msgsec*.toml")):
        with path.open("rb") as fh:
            data = tomllib.load(fh)
        for key, value in data.items():
            try:
                ident = int(key)
            except ValueError:
                continue
            if ident in rows:
                raise SystemExit(f"UNTRANSLATED_SURFACE_FAIL duplicate message {ident}")
            rows[ident] = value
    return rows


def load_category_ranges() -> list[tuple[int, int, str, str]]:
    with CATEGORIES.open("rb") as fh:
        data = tomllib.load(fh)
    return [
        (int(row["first"]), int(row["last"]), str(row["category"]), str(row["basis"]))
        for row in data["range"]
    ]


def category_for(ident: int, ranges: list[tuple[int, int, str, str]]) -> tuple[str, str]:
    for first, last, category, basis in ranges:
        if first <= ident <= last:
            return category, basis
        if ident < first:
            break
    return "unclassified", "unknown"


def compact(text: str, limit: int = 96) -> str:
    text = re.sub(r"<[^>]+>", " ", text)
    text = re.sub(r"\s+", " ", text).strip()
    return text if len(text) <= limit else text[: limit - 1] + "…"


def main() -> None:
    source = load_rows(SOURCE)
    korean = load_rows(KOREAN)
    ranges = load_category_ranges()
    missing = sorted(set(source) - set(korean))

    by_category: collections.Counter[tuple[str, str]] = collections.Counter()
    by_surface: dict[str, list[tuple[int, str, str, str]]] = {name: [] for name in SURFACES}
    empty_source = 0
    english_available = 0

    for ident in missing:
        row = source[ident]
        japanese = str(row.get("japanese", ""))
        english = str(row.get("english", ""))
        category, basis = category_for(ident, ranges)
        by_category[(category, basis)] += 1
        if not japanese.strip():
            empty_source += 1
        if english.strip():
            english_available += 1
        searchable = japanese + "\n" + english
        for surface, pattern in SURFACES.items():
            if pattern.search(searchable):
                by_surface[surface].append((ident, category, basis, compact(japanese)))

    print(
        "UNTRANSLATED_SURFACE_SUMMARY "
        f"source_rows={len(source)} korean_rows={len(korean)} missing_rows={len(missing)} "
        f"english_available={english_available} empty_japanese={empty_source}"
    )
    for (category, basis), count in sorted(by_category.items(), key=lambda item: (-item[1], item[0])):
        print(f"UNTRANSLATED_CATEGORY category={category!r} basis={basis!r} count={count}")

    for surface in ("settings-system", "tutorial-help", "input-name"):
        rows = by_surface[surface]
        print(f"UNTRANSLATED_SURFACE surface={surface!r} candidates={len(rows)} heuristic_only=true")
        for ident, category, basis, japanese in rows[:80]:
            print(
                f"UNTRANSLATED_CANDIDATE surface={surface!r} id={ident} "
                f"category={category!r} basis={basis!r} japanese={japanese!r}"
            )
        if len(rows) > 80:
            print(f"UNTRANSLATED_CANDIDATE_OMITTED surface={surface!r} omitted={len(rows)-80}")

    # This census is intentionally non-blocking for missing rows. It does fail
    # closed on impossible accounting, so CI cannot silently stop examining part
    # of the source/Korean corpus.
    if len(korean) > len(source) or any(ident not in source for ident in korean):
        raise SystemExit("UNTRANSLATED_SURFACE_FAIL Korean overlay contains IDs outside source corpus")


if __name__ == "__main__":
    main()
