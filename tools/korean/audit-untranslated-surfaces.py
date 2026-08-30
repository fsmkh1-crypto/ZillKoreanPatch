#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""Inventory source rows that are still outside the Korean overlay.

English-first rule: discovery only. A missing row is not automatically safe to
translate. Candidate surfaces are narrowed by verified upstream categories plus
strong Japanese UI wording so ordinary dialogue does not swamp the report.
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
SOURCE_FILE = re.compile(r"^msgsec\d{3}\.toml$")
KOREAN_FILE = re.compile(r"^msgsec\d{3}(?:(?:-part\d{2})|b)?\.toml$")

SETTINGS_CATEGORIES = {"general-ui-menu", "system-help", "controller-label", "confirmation-prompt", "notification"}
TUTORIAL_CATEGORIES = {"system-help", "inventory-and-status-help", "in-world-guidance"}
INPUT_CATEGORIES = {"general-ui-menu", "system-help", "controller-label", "confirmation-prompt", "in-world-guidance"}

SETTINGS_JA = re.compile(r"(?:環境設定|オプション|メッセージ速度|音量|ＢＧＭ|BGM|ＳＥ|振動|画面位置|画面|明るさ|キー設定|操作設定|ロード選択|セーブ|ロード|インストール)")
TUTORIAL_JA = re.compile(r"(?:チュートリアル|クイックヘルプ|ヘルプ|操作方法|遊び方|カーソル移動|ボタンで決定|ボタンで|方向キーで|装備画面|ステータス画面)")
INPUT_JA = re.compile(r"(?:名前を入力|名前を書|文字まで入力|カーソル移動|ひらがな|カタカナ|英数|アルファベット|入力完了|入力してください|文字入力|削除|空白|スペース)")


def load_source_rows() -> dict[int, dict]:
    rows: dict[int, dict] = {}
    for path in sorted(p for p in SOURCE.iterdir() if p.is_file() and SOURCE_FILE.fullmatch(p.name)):
        with path.open("rb") as fh:
            data = tomllib.load(fh)
        for key, value in data.items():
            try:
                ident = int(key)
            except ValueError:
                continue
            if ident in rows:
                raise SystemExit(f"UNTRANSLATED_SURFACE_FAIL duplicate canonical source ID {ident} in {path.name}")
            rows[ident] = value
    return rows


def load_korean_rows() -> dict[int, dict]:
    # Mirror corpus.LoadKoreanProject: multipart/legacy filenames participate,
    # byte-identical duplicate rows are tolerated, conflicting duplicates fail.
    rows: dict[int, dict] = {}
    for path in sorted(p for p in KOREAN.iterdir() if p.is_file() and KOREAN_FILE.fullmatch(p.name)):
        with path.open("rb") as fh:
            data = tomllib.load(fh)
        for key, value in data.items():
            try:
                ident = int(key)
            except ValueError:
                continue
            normalized = {
                "japanese": str(value.get("japanese", "")),
                "korean": str(value.get("korean", "")),
                "layout": str(value.get("layout", "")),
            }
            if ident in rows:
                if rows[ident] == normalized:
                    continue
                raise SystemExit(f"UNTRANSLATED_SURFACE_FAIL conflicting Korean ID {ident} in {path.name}")
            rows[ident] = normalized
    return rows


def load_category_ranges() -> list[tuple[int, int, str, str]]:
    with CATEGORIES.open("rb") as fh:
        data = tomllib.load(fh)
    return [(int(r["first"]), int(r["last"]), str(r["category"]), str(r["basis"])) for r in data["range"]]


def category_for(ident: int, ranges: list[tuple[int, int, str, str]]) -> tuple[str, str]:
    for first, last, category, basis in ranges:
        if first <= ident <= last:
            return category, basis
        if ident < first:
            break
    return "unclassified", "unknown"


def compact(text: str, limit: int = 120) -> str:
    text = re.sub(r"<[^>]+>", " ", text)
    text = re.sub(r"\s+", " ", text).strip()
    return text if len(text) <= limit else text[: limit - 1] + "…"


def candidate(surface: str, category: str, japanese: str) -> bool:
    if surface == "settings-system":
        return category in SETTINGS_CATEGORIES and bool(SETTINGS_JA.search(japanese))
    if surface == "tutorial-help":
        return category in TUTORIAL_CATEGORIES and bool(TUTORIAL_JA.search(japanese))
    if surface == "input-name":
        return category in INPUT_CATEGORIES and bool(INPUT_JA.search(japanese))
    return False


def main() -> None:
    source = load_source_rows()
    korean = load_korean_rows()
    ranges = load_category_ranges()
    missing = sorted(set(source) - set(korean))

    by_category: collections.Counter[tuple[str, str]] = collections.Counter()
    by_surface: dict[str, list[tuple[int, str, str, str]]] = {
        "settings-system": [], "tutorial-help": [], "input-name": []
    }
    empty_source = 0
    english_available = 0

    for ident in missing:
        row = source[ident]
        japanese = str(row.get("japanese", ""))
        english = str(row.get("english", ""))
        category, basis = category_for(ident, ranges)
        by_category[(category, basis)] += 1
        empty_source += int(not japanese.strip())
        english_available += int(bool(english.strip()))
        for surface in by_surface:
            if candidate(surface, category, japanese):
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
        for ident, category, basis, japanese in rows:
            print(
                f"UNTRANSLATED_CANDIDATE surface={surface!r} id={ident} "
                f"category={category!r} basis={basis!r} japanese={japanese!r}"
            )

    if len(source) != 43116:
        raise SystemExit(f"UNTRANSLATED_SURFACE_FAIL canonical source rows={len(source)} want 43116")
    if len(korean) != 42016:
        raise SystemExit(f"UNTRANSLATED_SURFACE_FAIL Korean accepted rows={len(korean)} want 42016")
    if any(ident not in source for ident in korean):
        raise SystemExit("UNTRANSLATED_SURFACE_FAIL Korean overlay contains IDs outside source corpus")
    if len(missing) != 1100:
        raise SystemExit(f"UNTRANSLATED_SURFACE_FAIL source-minus-Korean rows={len(missing)} want 1100")


if __name__ == "__main__":
    main()
