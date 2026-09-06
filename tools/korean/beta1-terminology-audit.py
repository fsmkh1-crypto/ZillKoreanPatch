#!/usr/bin/env python3
import argparse
import json
import tomllib
from collections import Counter
from pathlib import Path


def load_decisions(path: Path) -> list[dict]:
    data = json.loads(path.read_text(encoding="utf-8"))
    decisions = data.get("decisions")
    if not isinstance(decisions, list) or not decisions:
        raise ValueError("decisions file has no decisions")
    seen = set()
    for decision in decisions:
        jp = decision["japanese"]
        if jp in seen:
            raise ValueError(f"duplicate Japanese decision: {jp}")
        seen.add(jp)
    return decisions


def iter_records(root: Path):
    base = root / "translations/korean/messages"
    for path in sorted(base.glob("*.toml")):
        data = tomllib.loads(path.read_text(encoding="utf-8"))
        for rid, row in data.items():
            if not isinstance(row, dict):
                continue
            yield path.relative_to(root).as_posix(), str(rid), row


def classify(korean: str, canonical: str, legacy: list[str]) -> tuple[str, str | None]:
    if canonical in korean:
        return "canonical", canonical
    for value in legacy:
        if value in korean:
            return "legacy", value
    return "unrecognized", None


def audit(root: Path, decisions_path: Path) -> dict:
    decisions = load_decisions(decisions_path)
    records = list(iter_records(root))
    report = {
        "status": "AUDIT_ONLY_NO_MUTATION",
        "record_occurrences_scanned": len(records),
        "decisions": [],
    }
    for decision in decisions:
        jp = decision["japanese"]
        canonical = decision["korean"]
        legacy = decision.get("legacy", [])
        counts = Counter()
        legacy_counts = Counter()
        examples = []
        source_hits = 0
        for path, rid, row in records:
            japanese = row.get("japanese", "")
            if jp not in japanese:
                continue
            source_hits += 1
            korean = row.get("korean", "")
            kind, value = classify(korean, canonical, legacy)
            counts[kind] += 1
            if kind == "legacy":
                legacy_counts[value] += 1
            if kind != "canonical" and len(examples) < 40:
                examples.append({
                    "id": rid,
                    "path": path,
                    "japanese": japanese,
                    "korean": korean,
                    "classification": kind,
                    "matched_legacy": value if kind == "legacy" else None,
                })
        report["decisions"].append({
            "japanese": jp,
            "english": decision.get("english"),
            "canonical_korean": canonical,
            "legacy": legacy,
            "source_hits": source_hits,
            "canonical_hits": counts["canonical"],
            "legacy_hits": counts["legacy"],
            "legacy_by_value": dict(sorted(legacy_counts.items())),
            "unrecognized_hits": counts["unrecognized"],
            "examples": examples,
        })
    return report


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--decisions",
        type=Path,
        default=Path("docs/audit/beta1-approved-terminology.json"),
    )
    parser.add_argument("--json", type=Path)
    args = parser.parse_args()
    root = Path(__file__).resolve().parents[2]
    decisions = args.decisions if args.decisions.is_absolute() else root / args.decisions
    result = audit(root, decisions)
    rendered = json.dumps(result, ensure_ascii=False, indent=2)
    print(rendered)
    if args.json:
        args.json.write_text(rendered + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
