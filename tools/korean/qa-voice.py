#!/usr/bin/env python3
"""QA-4 high-confidence speaker/voice hazard candidate scan.

This is deliberately a candidate generator, not a correctness gate. It only
flags lexical combinations where the Japanese address/register and Korean
rendering are strongly at odds. Human scene/context review decides fixes.
"""
from __future__ import annotations

import argparse
import json
from pathlib import Path
import re
import tomllib

ROOT = Path(__file__).resolve().parents[2]
KOREAN_DIR = ROOT / "translations" / "korean" / "messages"
SECTION_FILE_RE = re.compile(r"^msgsec(\d{3})(?:(?:-part\d+)|b)?\.toml$")


def has_any(text: str, needles: tuple[str, ...]) -> bool:
    return any(n in text for n in needles)


def classify(japanese: str, korean: str) -> list[str]:
    findings: list[str] = []

    # 貴様/てめえ are hostile or contemptuous second-person forms. A polite
    # Korean second person is almost always a voice/register regression.
    if ("貴様" in japanese or "てめえ" in japanese or "てめぇ" in japanese) and has_any(
        korean, ("당신", "귀하", "선생", "자네")
    ):
        findings.append("hostile_jp_polite_ko")

    # お前 is ordinarily familiar/blunt. 당신 is possible only in narrow
    # relationship contexts, so surface it for scene review rather than auto-fix.
    if "お前" in japanese and "당신" in korean:
        findings.append("blunt_jp_polite_ko")

    # あなた/貴方 is not compatible with an explicitly contemptuous 네놈 unless
    # the local dramatic context supplies that force.
    if ("あなた" in japanese or "貴方" in japanese) and has_any(korean, ("네놈", "이 자식", "이놈")):
        findings.append("neutral_jp_hostile_ko")

    # Explicit Japanese honorifics rendered with a hostile Korean addressee are
    # worth reviewing even when a title translation may legitimately replace 様.
    if "様" in japanese and has_any(korean, ("네놈", "이 자식", "이놈")):
        findings.append("honorific_jp_hostile_ko")

    return findings


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--json", type=Path)
    ap.add_argument("--max-examples", type=int, default=80)
    args = ap.parse_args()

    findings: list[dict[str, object]] = []
    seen_ids: set[int] = set()
    scanned = 0

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
            ja = rec.get("japanese")
            ko = rec.get("korean")
            if not isinstance(ja, str) or not isinstance(ko, str) or not ko:
                continue
            scanned += 1
            kinds = classify(ja, ko)
            if kinds:
                findings.append({
                    "id": str(rid),
                    "section": section,
                    "path": path.name,
                    "kinds": kinds,
                    "japanese": ja,
                    "korean": ko,
                })

    report = {
        "schema": 1,
        "scanned_records": scanned,
        "finding_count": len(findings),
        "finding_by_kind": {
            kind: sum(kind in row["kinds"] for row in findings)
            for kind in sorted({k for row in findings for k in row["kinds"]})
        },
        "findings": findings,
    }
    if args.json:
        args.json.parent.mkdir(parents=True, exist_ok=True)
        args.json.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    print("Korean QA-4 speaker/voice hazard scan")
    print(f"  scanned_records: {scanned}")
    print(f"  finding_count: {len(findings)}")
    print("  finding_by_kind: " + json.dumps(report["finding_by_kind"], ensure_ascii=False, sort_keys=True))
    for row in findings[: max(args.max_examples, 0)]:
        print("  example: " + json.dumps(row, ensure_ascii=False, sort_keys=True))


if __name__ == "__main__":
    main()
