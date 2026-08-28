#!/usr/bin/env python3
"""QA-4 speaker/voice hazard candidate scan.

Candidate generation only: nothing in this tool is auto-fixed. It combines
high-confidence lexical address mismatches with broader Japanese/Korean register
signals and ranks the resulting rows for human/LLM context review.

Multi-branch records are compared branch-by-branch only when the authoritative
fixed control-token skeleton matches. Genuine structural fallbacks are reported
rather than hidden; ordinary records without <end> are simply scanned whole.
"""
from __future__ import annotations

import argparse
import json
from pathlib import Path
import re
import tomllib

from control_tags import fixed_tokens

ROOT = Path(__file__).resolve().parents[2]
KOREAN_DIR = ROOT / "translations" / "korean" / "messages"
SECTION_FILE_RE = re.compile(r"^msgsec(\d{3})(?:(?:-part\d+)|b)?\.toml$")
END = "<end>"
CONTROL_RE = re.compile(r"<[^<>]+>")

# These are candidate cues, not correctness rules. Japanese politeness markers
# do not map mechanically to Korean 존댓말; character voice and relationship
# context can legitimately differ. Broad register cues therefore rank below
# explicit hostile/address contradictions.
JP_POLITE_RE = re.compile(r"(?:です(?:か|ね|よ)?|ます(?:か|ね|よ)?|ません|ください|下さい|ございます|でしょう)(?:[。！？!?…]|$)")
JP_BLUNT_RE = re.compile(r"(?:だぞ|だぜ|だな|だろ|だろう|じゃないか|ではないか|しろ|するな|くれ|かい|だね)(?:[。！？!?…]|$)")
KO_FORMAL_RE = re.compile(r"(?:습니다|습니까|입니다|입니까|십시오|세요|셨어요|했어요|해요|예요|이에요|군요|네요|죠)(?:[.!?…]|$)")
KO_CASUAL_RE = re.compile(r"(?:해라|하지 마|하지마|해 줘|해줘|해|했어|할게|할까|하냐|하니|냐|니|구나|군|네|잖아|거야|거냐|마라|해 둬|해둬)(?:[.!?…]|$)")

SEVERITY = {
    "hostile_jp_polite_ko": 100,
    "neutral_jp_hostile_ko": 95,
    "blunt_jp_polite_ko": 80,
    "blunt_jp_formal_ko": 70,
    "rough_first_person_polite_ko": 60,
    "humble_first_person_blunt_ko": 60,
    "polite_jp_casual_ko": 50,
}


def has_any(text: str, needles: tuple[str, ...]) -> bool:
    return any(n in text for n in needles)


def visible(text: str) -> str:
    return CONTROL_RE.sub("", text).strip()


def classify(japanese: str, korean: str) -> list[str]:
    findings: list[str] = []
    ja_vis = visible(japanese)
    ko_vis = visible(korean)

    if ("貴様" in japanese or "てめえ" in japanese or "てめぇ" in japanese) and has_any(
        korean, ("당신", "귀하", "선생", "자네")
    ):
        findings.append("hostile_jp_polite_ko")

    # お前 -> 당신 is ambiguous because 당신 can be intimate/spousal Korean.
    if "お前" in japanese and "당신" in korean:
        findings.append("blunt_jp_polite_ko")

    if ("あなた" in japanese or "貴方" in japanese) and has_any(korean, ("네놈", "이 자식", "이놈")):
        findings.append("neutral_jp_hostile_ko")

    # Broad sentence-register cues. Review-only and deliberately lower ranked.
    if JP_POLITE_RE.search(ja_vis) and KO_CASUAL_RE.search(ko_vis):
        findings.append("polite_jp_casual_ko")
    if JP_BLUNT_RE.search(ja_vis) and KO_FORMAL_RE.search(ko_vis):
        findings.append("blunt_jp_formal_ko")

    if has_any(japanese, ("俺", "オレ", "俺様")) and re.search(r"(?:^|[\s,，。!?！？])저(?:는|가|도|를|에게|한테|의|$)", ko_vis):
        findings.append("rough_first_person_polite_ko")
    if has_any(japanese, ("わたくし", "私め", "拙者")) and re.search(r"(?:^|[\s,，。!?！？])나(?:는|가|도|를|에게|한테|의|$)", ko_vis):
        findings.append("humble_first_person_blunt_ko")

    return list(dict.fromkeys(findings))


def branch_alignment_safe(japanese: str, korean: str) -> bool:
    """Whether positional <end>-segment pairing is structurally trustworthy."""
    ja = japanese.split(END)
    ko = korean.split(END)
    return (
        len(ja) == len(ko)
        and len(ja) > 1
        and fixed_tokens(japanese) == fixed_tokens(korean)
    )


def needs_structural_fallback(japanese: str, korean: str) -> bool:
    """True only when an <end>/control-bearing record cannot be safely aligned."""
    has_branch_structure = END in japanese or END in korean
    return has_branch_structure and not branch_alignment_safe(japanese, korean)


def paired_segments(japanese: str, korean: str) -> list[tuple[int, str, str]]:
    if branch_alignment_safe(japanese, korean):
        ja = japanese.split(END)
        ko = korean.split(END)
        return [(i, ja[i], ko[i]) for i in range(len(ja) - 1)]
    return [(0, japanese, korean)]


def row_priority(kinds: list[str]) -> int:
    return max((SEVERITY.get(kind, 1) for kind in kinds), default=0)


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--json", type=Path)
    ap.add_argument("--max-examples", type=int, default=80)
    args = ap.parse_args()

    findings: list[dict[str, object]] = []
    seen_ids: set[int] = set()
    scanned = 0
    fallback_count = 0
    fallback_examples: list[dict[str, object]] = []

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
            if needs_structural_fallback(ja, ko):
                fallback_count += 1
                if len(fallback_examples) < 20:
                    fallback_examples.append({
                        "id": str(rid),
                        "section": section,
                        "path": path.name,
                        "japanese_fixed_tokens": fixed_tokens(ja),
                        "korean_fixed_tokens": fixed_tokens(ko),
                        "japanese_end_count": ja.count(END),
                        "korean_end_count": ko.count(END),
                    })
            for segment, ja_seg, ko_seg in paired_segments(ja, ko):
                kinds = classify(ja_seg, ko_seg)
                if kinds:
                    findings.append({
                        "id": str(rid),
                        "section": section,
                        "path": path.name,
                        "segment": segment,
                        "priority": row_priority(kinds),
                        "kinds": kinds,
                        "japanese": ja_seg + (END if END in ja else ""),
                        "korean": ko_seg + (END if END in ko else ""),
                    })

    findings.sort(key=lambda row: (-int(row["priority"]), int(row["id"]), int(row["segment"])))
    all_kinds = sorted({k for row in findings for k in row["kinds"]})
    priority_bands = {
        "tier1_80_plus": sum(int(row["priority"]) >= 80 for row in findings),
        "tier2_60_79": sum(60 <= int(row["priority"]) < 80 for row in findings),
        "tier3_under_60": sum(int(row["priority"]) < 60 for row in findings),
    }
    report = {
        "schema": 4,
        "scanned_records": scanned,
        "finding_count": len(findings),
        "finding_record_count": len({str(row["id"]) for row in findings}),
        "finding_by_kind": {
            kind: sum(kind in row["kinds"] for row in findings)
            for kind in all_kinds
        },
        "priority_bands": priority_bands,
        "fallback_count": fallback_count,
        "fallback_examples": fallback_examples,
        "findings": findings,
    }
    if args.json:
        args.json.parent.mkdir(parents=True, exist_ok=True)
        args.json.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    print("Korean QA-4 speaker/voice hazard scan")
    print(f"  scanned_records: {scanned}")
    print(f"  finding_count: {len(findings)}")
    print(f"  finding_record_count: {report['finding_record_count']}")
    print(f"  fallback_count: {fallback_count}")
    print("  priority_bands: " + json.dumps(priority_bands, ensure_ascii=False, sort_keys=True))
    print("  finding_by_kind: " + json.dumps(report["finding_by_kind"], ensure_ascii=False, sort_keys=True))
    for row in fallback_examples[: max(args.max_examples, 0)]:
        print("  fallback: " + json.dumps(row, ensure_ascii=False, sort_keys=True))
    for row in findings[: max(args.max_examples, 0)]:
        print("  example: " + json.dumps(row, ensure_ascii=False, sort_keys=True))


if __name__ == "__main__":
    main()
