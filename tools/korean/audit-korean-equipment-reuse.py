#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""Find evidence-backed Korean equipment-name candidates already present in the reviewed corpus."""
from __future__ import annotations

import glob
import pathlib
import tomllib

ROOT = pathlib.Path(__file__).resolve().parents[2]
MAX_PAYLOAD = 16


def load_toml(path: pathlib.Path):
    with path.open("rb") as f:
        return tomllib.load(f)


def renderer_bytes(text: str, custom: set[str]) -> int | None:
    total = 0
    for ch in text:
        if ch in custom:
            total += 2
            continue
        try:
            total += len(ch.encode("cp932"))
        except UnicodeEncodeError:
            return None
    return total


def main() -> int:
    equipment = load_toml(ROOT / "release/strings/equipment.toml")
    glyphs = set(load_toml(ROOT / "release/korean/font/glyphs.toml")["glyphs"])

    by_source: dict[str, set[str]] = {}
    source_records: dict[str, int] = {}
    for filename in sorted(glob.glob(str(ROOT / "translations/korean/messages/msgsec*.toml"))):
        rows = load_toml(pathlib.Path(filename))
        for row in rows.values():
            japanese = row.get("japanese", "")
            korean = row.get("korean", "")
            if not japanese.endswith("<end>") or not korean.endswith("<end>"):
                continue
            if any(tag in japanese[:-5] for tag in ("<line-break>", "<next-page>", "<value:", "<if>", "<select>")):
                continue
            if any(tag in korean[:-5] for tag in ("<line-break>", "<next-page>", "<value:", "<if>", "<select>")):
                continue
            source = japanese[:-5]
            candidate = korean[:-5]
            by_source.setdefault(source, set()).add(candidate)
            source_records[source] = source_records.get(source, 0) + 1

    exact_unique = 0
    exact_ambiguous = 0
    missing = 0
    fit = 0
    oversize = 0
    unencodable = 0

    for selector in range(1, 133):
        row = equipment[str(selector)]
        source = row["source"]
        candidates = sorted(by_source.get(source, set()))
        if not candidates:
            missing += 1
            print(f'EQUIPMENT_REUSE selector={selector} source={source!r} status=NO_EXACT_CORPUS_CANDIDATE')
            continue
        if len(candidates) != 1:
            exact_ambiguous += 1
            print(f'EQUIPMENT_REUSE selector={selector} source={source!r} status=AMBIGUOUS candidates={candidates!r}')
            continue
        exact_unique += 1
        candidate = candidates[0]
        size = renderer_bytes(candidate, glyphs)
        if size is None:
            unencodable += 1
            status = "UNENCODABLE"
        elif size <= MAX_PAYLOAD:
            fit += 1
            status = "FIT"
        else:
            oversize += 1
            status = "OVERSIZE"
        print(
            f'EQUIPMENT_REUSE selector={selector} source={source!r} candidate={candidate!r} '
            f'bytes={size} status={status} corpus_occurrences={source_records.get(source, 0)}'
        )

    print(f"equipment_total={len(equipment)}")
    print(f"exact_unique_candidates={exact_unique}")
    print(f"exact_ambiguous_candidates={exact_ambiguous}")
    print(f"no_exact_corpus_candidate={missing}")
    print(f"unique_candidates_fit_16_bytes={fit}")
    print(f"unique_candidates_oversize_16_bytes={oversize}")
    print(f"unique_candidates_unencodable={unencodable}")
    print("policy=report_only_no_unreviewed_bindata_mutation")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
