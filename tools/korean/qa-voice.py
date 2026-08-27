#!/usr/bin/env python3
"""QA-4 high-confidence speaker/voice hazard candidate scan.

This is deliberately a candidate generator, not a correctness gate. It flags
lexical combinations where Japanese address/register and Korean rendering are
strongly at odds. Multi-branch records are compared branch-by-branch so a polite
third-party title in one branch cannot contaminate a hostile address in another.
Human scene/context review decides fixes.
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
END = "<end>"


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

    # あなた/貴方 is not normally compatible with explicitly contemptuous
    # Korean address. Local dramatic context can still justify it, so review only.
    if ("あなた" in japanese or "貴方" in japanese) and has_any(korean, ("네놈", "이 자식", "이놈")):
        findings.append("neutral_jp_hostile_ko")

    return findings


def paired_segments(japanese: str, korean: str) -> list[tuple[int, str, str]]:
    """Return aligned <end>-terminated semantic branches when structurally safe."""
    ja = japanese.split(END)
    ko = korean.split(END)
    # Both well-formed message strings normally end in <end>, yielding a final
    # empty item. Only branch-align when both have the same segment count.
    if len(ja) == len(ko) and len(ja) > 1:
        return [(i, ja[i], ko[i]) for i in range(len(ja) - 1)]
    return [(0, japanese, korean)]


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
            for segment, ja_seg, ko_seg in paired_segments(ja, ko):
                kinds = classify(ja_seg, ko_seg)
                if kinds:
                    findings.append({
                        "id": str(rid),
                        "section": section,
                        "path": path.name,
                        "segment": segment,
                        "kinds": kinds,
                        "japanese": ja_seg + END,
                        "korean": ko_seg + END,
                    })

    all_kinds = sorted({k for row in findings for k in row["kinds"]})
    report = {
        "schema": 1,
        "scanned_records": scanned,
        "finding_count": len(findings),
        "finding_by_kind": {
            kind: sum(kind in row["kinds"] for row in findings)
            for kind in all_kinds
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
