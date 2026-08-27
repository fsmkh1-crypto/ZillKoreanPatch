#!/usr/bin/env python3
"""High-signal advisory Korean text-sanity scan.

Structural token checks cannot catch typography residue such as full-width
spaces in translated prose or sentence boundaries accidentally glued together.
This scan is advisory-only and deliberately excludes token-boundary artifacts,
blank/passthrough rows, and common short reduplication.
"""
from __future__ import annotations

import argparse
from collections import Counter
import json
from pathlib import Path
import re
import tomllib

ROOT = Path(__file__).resolve().parents[2]
KOREAN_DIR = ROOT / "translations" / "korean" / "messages"
SECTION_FILE_RE = re.compile(r"^msgsec(\d{3})(?:(?:-part\d+)|b)?\.toml$")
HANGUL_RE = re.compile(r"[가-힣]")
TOKEN_RE = re.compile(r"<[^>]+>")
GLUED_SENTENCE_RE = re.compile(r"[.!?。！？][가-힣]")
SPACE_BEFORE_END_RE = re.compile(r"(?<=[가-힣0-9.!?…。！？])([ \u3000]+)(?=<end>)")
# Three or more syllables repeated immediately is unusual enough to review.
ADJACENT_REPEAT_RE = re.compile(r"([가-힣]{3,8})\1")


def issue(kind: str, section: int, rid: str, path: str, ja: str, ko: str, detail: str) -> dict[str, object]:
    return {"kind": kind, "section": section, "id": rid, "path": path, "japanese": ja, "korean": ko, "detail": detail}


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--json", type=Path)
    ap.add_argument("--max-examples", type=int, default=60)
    args = ap.parse_args()

    findings: list[dict[str, object]] = []
    seen_ids: set[int] = set()
    unique_records = 0

    for path in sorted(KOREAN_DIR.glob("msgsec*.toml")):
        m = SECTION_FILE_RE.match(path.name)
        if not m:
            continue
        section = int(m.group(1))
        with path.open("rb") as f:
            data = tomllib.load(f)
        for rid, rec in data.items():
            try:
                numeric = int(rid)
            except ValueError:
                continue
            if numeric in seen_ids:
                continue
            seen_ids.add(numeric)
            unique_records += 1
            ja = rec.get("japanese")
            ko = rec.get("korean")
            if not isinstance(ja, str) or not isinstance(ko, str):
                continue

            # Ignore blank/layout-only rows and exact passthrough rows.
            semantic = TOKEN_RE.sub(" ", ko)
            if not HANGUL_RE.search(semantic):
                continue

            if "\u3000" in ko and ko != ja:
                findings.append(issue("fullwidth_space_in_korean", section, rid, path.name, ja, ko, f"count={ko.count(chr(0x3000))}"))

            for match in SPACE_BEFORE_END_RE.finditer(ko):
                findings.append(issue("space_before_end", section, rid, path.name, ja, ko, repr(match.group(1))))

            # Tokens become separators, never disappear: branch boundaries must
            # not create false glued-sentence findings.
            separated = TOKEN_RE.sub(" ", ko)
            for match in GLUED_SENTENCE_RE.finditer(separated):
                findings.append(issue("glued_sentence", section, rid, path.name, ja, ko, match.group(0)))

            for match in ADJACENT_REPEAT_RE.finditer(separated):
                findings.append(issue("long_adjacent_duplicate_fragment", section, rid, path.name, ja, ko, match.group(1)))

    counts = Counter(str(x["kind"]) for x in findings)
    report = {
        "schema": 2,
        "unique_records": unique_records,
        "finding_count": len(findings),
        "finding_by_kind": dict(sorted(counts.items())),
        "findings": findings,
    }
    if args.json:
        args.json.parent.mkdir(parents=True, exist_ok=True)
        args.json.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    print("Korean text sanity QA")
    print(f"  unique_records: {unique_records}")
    print(f"  finding_count: {len(findings)}")
    print("  finding_by_kind: " + json.dumps(report["finding_by_kind"], ensure_ascii=False))
    for row in findings[: max(args.max_examples, 0)]:
        print("  example: " + json.dumps(row, ensure_ascii=False, sort_keys=True))


if __name__ == "__main__":
    main()
