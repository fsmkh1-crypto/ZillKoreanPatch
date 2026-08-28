#!/usr/bin/env python3
"""QA-4 speaker/voice hazard candidate scan.

Candidate generation only: nothing in this tool is auto-fixed. It combines
high-confidence lexical address mismatches with broader Japanese/Korean register
signals and ranks the resulting rows for human/LLM context review.

Multi-branch records are compared branch-by-branch only when the authoritative
fixed control-token skeleton matches. Genuine structural fallbacks are reported
rather than hidden; ordinary records without <end> are simply scanned whole.

Reviewed contextual exceptions are kept in qa-voice-exceptions.json so already
checked character/register choices do not keep consuming the actionable queue.
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
VOICE_EXCEPTIONS_PATH = ROOT / "tools" / "korean" / "qa-voice-exceptions.json"
SECTION_FILE_RE = re.compile(r"^msgsec(\d{3})(?:(?:-part\d+)|b)?\.toml$")
END = "<end>"
CONTROL_RE = re.compile(r"<[^<>]+>")

# These are candidate cues, not correctness rules. Japanese politeness markers
# do not map mechanically to Korean 존댓말; character voice and relationship
# context can legitimately differ. Broad register cues therefore rank below
# explicit hostile/address contradictions.
# Bare でしょう is deliberately excluded: in dialogue it is often a casual or
# feminine confirmation form rather than a reliable politeness signal.
JP_POLITE_RE = re.compile(r"(?:です(?:か|ね|よ)?|でした(?:か|ね|よ)?|ます(?:か|ね|よ)?|ました(?:か|ね|よ)?|ません|ください|下さい|ございます)(?:[。！？!?…]|$)")
# Bare かい is deliberately excluded: it is too easy to confuse with adjective
# endings such as 温かい, producing noisy false positives.
JP_BLUNT_RE = re.compile(r"(?:だぞ|だぜ|だな|だろ|だろう|じゃないか|ではないか|しろ|するな|くれ|だね)(?:[。！？!?…]|$)")
KO_FORMAL_RE = re.compile(r"(?:습니다|습니까|입니다|입니까|십시오|세요|셨어요|했어요|해요|예요|이에요|군요|네요|죠)(?:[.!?…]|$)")
# Avoid one-syllable suffix cues such as 니/네/군/해. They frequently occur at
# the tail of non-casual forms or quotative/lexical endings (e.g. 붕어하셨다니)
# and create far more noise than useful signal. Keep only distinctive endings.
KO_CASUAL_RE = re.compile(r"(?:해라|하지 마|하지마|해 줘|해줘|했어|할게|할까|하냐|구나|잖아|거야|거냐|마라|해 둬|해둬)(?:[.!?…]|$)")
JP_ROUGH_FIRST_RE = re.compile(r"(?:俺様|俺|オレ)(?:は|が|も|を|に|の|で|、|，|,|\s)")

# Korean first-person forms need token-aware matching. In particular, treating
# ``나가`` as ``나+가`` creates false positives in ordinary verbs such as
# ``전장에 나가 있다``. Standard subject forms are ``내가`` / ``제가``;
# explicit pronoun forms below deliberately exclude the ambiguous ``나가/저가``.
KO_ROUGH_FIRST_RE = re.compile(
    r"(?:^|[\s,，。!?！？])(?:나는|나도|나를|나에게|나한테|나의|내가)(?=$|[\s,，。!?！？])"
)
KO_POLITE_FIRST_RE = re.compile(
    r"(?:^|[\s,，。!?！？])(?:저는|저도|저를|저에게|저한테|저의|제가)(?=$|[\s,，。!?！？])"
)

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

    if JP_ROUGH_FIRST_RE.search(japanese) and KO_POLITE_FIRST_RE.search(ko_vis):
        findings.append("rough_first_person_polite_ko")
    if has_any(japanese, ("わたくし", "私め", "拙者")) and KO_ROUGH_FIRST_RE.search(ko_vis):
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


def load_reviewed_exceptions(path: Path = VOICE_EXCEPTIONS_PATH) -> dict[tuple[str, int, str], dict[str, str]]:
    if not path.exists():
        return {}
    raw = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(raw, list):
        raise ValueError(f"QA-4 exception registry must be a list: {path}")
    out: dict[tuple[str, int, str], dict[str, str]] = {}
    for i, item in enumerate(raw):
        if not isinstance(item, dict):
            raise ValueError(f"QA-4 exception #{i} is not an object")
        rid = str(item.get("id", ""))
        segment = item.get("segment", 0)
        kind = item.get("kind")
        category = item.get("category")
        reason = item.get("reason")
        if not rid.isdigit() or not isinstance(segment, int) or segment < 0:
            raise ValueError(f"QA-4 exception #{i} has invalid id/segment")
        if kind not in SEVERITY or not isinstance(category, str) or not category or not isinstance(reason, str) or not reason:
            raise ValueError(f"QA-4 exception #{i} has invalid kind/category/reason")
        key = (rid, segment, kind)
        if key in out:
            raise ValueError(f"duplicate QA-4 exception key: {key}")
        out[key] = {"category": category, "reason": reason}
    return out


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--json", type=Path)
    ap.add_argument("--max-examples", type=int, default=80)
    args = ap.parse_args()

    reviewed = load_reviewed_exceptions()
    matched_reviewed: set[tuple[str, int, str]] = set()
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
                    reviewed_kinds: list[str] = []
                    reviewed_details: list[dict[str, str]] = []
                    for kind in kinds:
                        key = (str(rid), segment, kind)
                        if key in reviewed:
                            matched_reviewed.add(key)
                            reviewed_kinds.append(kind)
                            reviewed_details.append({"kind": kind, **reviewed[key]})
                    actionable_kinds = [kind for kind in kinds if kind not in reviewed_kinds]
                    findings.append({
                        "id": str(rid),
                        "section": section,
                        "path": path.name,
                        "segment": segment,
                        "priority": row_priority(kinds),
                        "actionable_priority": row_priority(actionable_kinds),
                        "kinds": kinds,
                        "actionable_kinds": actionable_kinds,
                        "reviewed_exception_kinds": reviewed_kinds,
                        "reviewed_exceptions": reviewed_details,
                        "japanese": ja_seg + (END if END in ja else ""),
                        "korean": ko_seg + (END if END in ko else ""),
                    })

    stale_reviewed = sorted(set(reviewed) - matched_reviewed)
    if stale_reviewed:
        raise ValueError("stale QA-4 exception entries: " + ", ".join(map(str, stale_reviewed)))

    findings.sort(
        key=lambda row: (
            -int(row["actionable_priority"]),
            -int(row["priority"]),
            int(row["id"]),
            int(row["segment"]),
        )
    )
    all_kinds = sorted({k for row in findings for k in row["kinds"]})
    actionable_rows = [row for row in findings if row["actionable_kinds"]]
    priority_bands = {
        "tier1_80_plus": sum(int(row["priority"]) >= 80 for row in findings),
        "tier2_60_79": sum(60 <= int(row["priority"]) < 80 for row in findings),
        "tier3_under_60": sum(int(row["priority"]) < 60 for row in findings),
    }
    actionable_priority_bands = {
        "tier1_80_plus": sum(int(row["actionable_priority"]) >= 80 for row in actionable_rows),
        "tier2_60_79": sum(60 <= int(row["actionable_priority"]) < 80 for row in actionable_rows),
        "tier3_under_60": sum(int(row["actionable_priority"]) < 60 for row in actionable_rows),
    }
    report = {
        "schema": 5,
        "scanned_records": scanned,
        "finding_count": len(findings),
        "finding_record_count": len({str(row["id"]) for row in findings}),
        "actionable_finding_count": len(actionable_rows),
        "actionable_record_count": len({str(row["id"]) for row in actionable_rows}),
        "reviewed_exception_kind_count": sum(len(row["reviewed_exception_kinds"]) for row in findings),
        "finding_by_kind": {
            kind: sum(kind in row["kinds"] for row in findings)
            for kind in all_kinds
        },
        "priority_bands": priority_bands,
        "actionable_priority_bands": actionable_priority_bands,
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
    print(f"  actionable_finding_count: {report['actionable_finding_count']}")
    print(f"  actionable_record_count: {report['actionable_record_count']}")
    print(f"  reviewed_exception_kind_count: {report['reviewed_exception_kind_count']}")
    print(f"  fallback_count: {fallback_count}")
    print("  priority_bands: " + json.dumps(priority_bands, ensure_ascii=False, sort_keys=True))
    print("  actionable_priority_bands: " + json.dumps(actionable_priority_bands, ensure_ascii=False, sort_keys=True))
    print("  finding_by_kind: " + json.dumps(report["finding_by_kind"], ensure_ascii=False, sort_keys=True))
    for row in fallback_examples[: max(args.max_examples, 0)]:
        print("  fallback: " + json.dumps(row, ensure_ascii=False, sort_keys=True))
    for row in findings[: max(args.max_examples, 0)]:
        print("  example: " + json.dumps(row, ensure_ascii=False, sort_keys=True))


if __name__ == "__main__":
    main()
