#!/usr/bin/env python3
"""Advisory Korean text-sanity scan for localization artifacts.

Finds suspicious typography/automation residue that structural token checks do
not catch. This tool is advisory-only by design: dialogue and stylized UI text
can legitimately violate normal prose spacing.
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
HANGUL = r"가-힣"
TOKEN_RE = re.compile(r"<[^>]+>")
# Strong signals: sentence-final punctuation glued directly to Korean text.
GLUED_SENTENCE_RE = re.compile(rf"[.!?。！？][{HANGUL}]")
# Space immediately before a runtime/fixed token is often accidental, especially <end>.
SPACE_BEFORE_TOKEN_RE = re.compile(r"[ \u3000]+(?=<[^>]+>)")
# Same 2+ Hangul syllables duplicated adjacently, e.g. 눈동자동자.
ADJACENT_REPEAT_RE = re.compile(rf"([{HANGUL}]{{2,8}})\1")


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

            if "\u3000" in ko:
                findings.append(issue("fullwidth_space", section, rid, path.name, ja, ko, f"count={ko.count(chr(0x3000))}"))

            for match in SPACE_BEFORE_TOKEN_RE.finditer(ko):
                findings.append(issue("space_before_token", section, rid, path.name, ja, ko, repr(match.group(0))))

            semantic = TOKEN_RE.sub("", ko)
            for match in GLUED_SENTENCE_RE.finditer(semantic):
                findings.append(issue("glued_sentence", section, rid, path.name, ja, ko, match.group(0)))

            # Restrict duplicated-fragment warning to suspicious longer repeats;
            # exclude laughter/onomatopoeia-like one-syllable repeats by construction.
            for match in ADJACENT_REPEAT_RE.finditer(semantic):
                frag = match.group(1)
                if len(frag) >= 2:
                    findings.append(issue("adjacent_duplicate_fragment", section, rid, path.name, ja, ko, frag))

    counts = Counter(str(x["kind"]) for x in findings)
    report = {
        "schema": 1,
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
