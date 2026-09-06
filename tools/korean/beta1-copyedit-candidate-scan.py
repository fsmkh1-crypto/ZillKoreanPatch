#!/usr/bin/env python3
import argparse
import json
import re
import tomllib
from collections import Counter
from pathlib import Path

CONTROL_RE = re.compile(r"<[^>]+>")
HANGUL = re.compile(r"[가-힣]")
AUX_FOLLOW = r"(?=$|[^가-힣]|(?:는|도|만|부터|까지)(?=[^가-힣]|$))"

# User-approved canonical names/terms plus Balor, whose repeated 을/를 error was
# independently identified during contextual review. This is candidate detection,
# never blind mutation.
PARTICLE_TARGETS = [
    "로스톨", "페름", "레무온", "아트레이아", "콘스", "기어", "석화수",
    "녹사", "파르셴", "플린트", "소도", "주작장군", "현무장군", "발로르",
]


def jongseong_index(syllable: str) -> int:
    code = ord(syllable)
    if not (0xAC00 <= code <= 0xD7A3):
        raise ValueError(f"not a Hangul syllable: {syllable!r}")
    return (code - 0xAC00) % 28


def wrong_particle_pairs(word: str) -> list[tuple[str, str]]:
    jong = jongseong_index(word[-1])
    if jong == 0:  # vowel ending
        return [("이", "가"), ("을", "를"), ("과", "와"), ("은", "는"), ("으로", "로")]
    pairs = [("가", "이"), ("를", "을"), ("와", "과"), ("는", "은")]
    # ㄹ-final takes 로; other consonants take 으로.
    if jong == 8:
        pairs.append(("으로", "로"))
    else:
        pairs.append(("로", "으로"))
    return pairs


def explicit_patterns():
    patterns = []
    for word in PARTICLE_TARGETS:
        for wrong, right in wrong_particle_pairs(word):
            patterns.append((
                f"wrong_particle_{word}_{wrong}",
                re.compile(re.escape(word + wrong) + AUX_FOLLOW),
                word + right,
            ))
    patterns.append(("spacing_moyang_ijiman", re.compile(r"모양 이지만"), "모양이지만"))
    return patterns


EXPLICIT_PATTERNS = explicit_patterns()

# Broad heuristics are contextual candidates only; they intentionally do not fail CI
# and must not be auto-applied.
REVIEW_PATTERNS = [
    ("comma_before_common_connective", re.compile(r",\s+(?:하지만|그러나|그래도|그리고|그런데|그러므로|따라서)")),
    ("space_before_common_particle", re.compile(r"[가-힣]\s+(?:은|는|이|가|을|를|와|과|에|의|도|만|부터|까지|에게|으로|로)(?=[^가-힣]|$)")),
    ("duplicate_terminal_punctuation", re.compile(r"(?:[!?]{3,}|\.{4,}|,{2,})")),
]


def strip_controls(text: str) -> str:
    return CONTROL_RE.sub(" ", text)


def iter_records(root: Path):
    base = root / "translations/korean/messages"
    for path in sorted(base.glob("*.toml")):
        data = tomllib.loads(path.read_text(encoding="utf-8"))
        for rid, row in data.items():
            if isinstance(row, dict) and isinstance(row.get("korean"), str):
                yield path.relative_to(root).as_posix(), str(rid), row


def scan(root: Path) -> dict:
    findings = []
    counts = Counter()
    scanned = 0
    for path, rid, row in iter_records(root):
        scanned += 1
        korean = row["korean"]
        visible = strip_controls(korean)
        if not HANGUL.search(visible):
            continue
        for kind, pattern, suggested in EXPLICIT_PATTERNS:
            for match in pattern.finditer(visible):
                findings.append({
                    "severity": "HIGH_CONFIDENCE_REVIEW",
                    "kind": kind,
                    "path": path,
                    "id": rid,
                    "japanese": row.get("japanese", ""),
                    "korean": korean,
                    "matched": match.group(0),
                    "suggested_fragment": suggested,
                })
                counts[kind] += 1
        for kind, pattern in REVIEW_PATTERNS:
            if pattern.search(visible):
                findings.append({
                    "severity": "CONTEXT_REVIEW",
                    "kind": kind,
                    "path": path,
                    "id": rid,
                    "japanese": row.get("japanese", ""),
                    "korean": korean,
                })
                counts[kind] += 1
    return {
        "status": "CANDIDATE_SCAN_ONLY_NO_MUTATION",
        "records_scanned": scanned,
        "finding_count": len(findings),
        "finding_by_kind": dict(sorted(counts.items())),
        "findings": findings,
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--json", type=Path)
    parser.add_argument("--max-print", type=int, default=80)
    args = parser.parse_args()
    root = Path(__file__).resolve().parents[2]
    result = scan(root)
    print(f"BETA1_COPYEDIT_CANDIDATE_SUMMARY records={result['records_scanned']} findings={result['finding_count']} by_kind={json.dumps(result['finding_by_kind'], ensure_ascii=False, sort_keys=True)}")
    for finding in result["findings"][: args.max_print]:
        print(json.dumps(finding, ensure_ascii=False, sort_keys=True))
    if args.json:
        args.json.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
