#!/usr/bin/env python3
"""High-signal advisory Korean text-sanity scan.

Structural token checks cannot catch typography residue. This scan intentionally
prefers precision over recall: it reports only ASCII sentence punctuation glued
to Hangul plus full-width spaces in translated Korean. Deliberate full-width
title punctuation and expressive reduplication are not treated as errors.
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
# ASCII only. Full-width punctuation is commonly intentional in copied titles/UI.
GLUED_ASCII_SENTENCE_RE = re.compile(r"[.!?][가-힣]")


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

            separated = TOKEN_RE.sub(" ", ko)
            if not HANGUL_RE.search(separated):
                continue

            if "\u3000" in ko and ko != ja:
                findings.append(issue("fullwidth_space_in_korean", section, rid, path.name, ja, ko, f"count={ko.count(chr(0x3000))}"))

            # Tokens are separators so branch boundaries cannot create false hits.
            for match in GLUED_ASCII_SENTENCE_RE.finditer(separated):
                findings.append(issue("glued_ascii_sentence", section, rid, path.name, ja, ko, match.group(0)))

    counts = Counter(str(x["kind"]) for x in findings)
    report = {
        "schema": 3,
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
