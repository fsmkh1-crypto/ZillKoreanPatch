#!/usr/bin/env python3
"""QA-2 canonical terminology analyzer.

For each seeded Japanese→Korean canonical term, find every translated record
whose Japanese source contains that term.  If the Korean semantic text does not
contain the canonical Korean surface, emit a mismatch for human review.

This is intentionally advisory-only: inflection, paraphrase, or nested terms can
create legitimate exceptions, so this tool reports candidates rather than
rewriting corpus data.
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
TERMS = ROOT / "translations" / "terminology" / "korean-canonical.toml"
SECTION_FILE_RE = re.compile(r"^msgsec(\d{3})(?:(?:-part\d+)|b)?\.toml$")


def load_terms() -> list[tuple[str, str]]:
    with TERMS.open("rb") as f:
        data = tomllib.load(f)
    out: list[tuple[str, str]] = []
    seen: set[str] = set()
    for row in data.get("entry", []):
        ja = str(row["japanese"])
        ko = str(row["korean"])
        if ja in seen:
            raise SystemExit(f"duplicate Japanese canonical term: {ja}")
        seen.add(ja)
        out.append((ja, ko))
    # Longest terms first makes the report easier to inspect when terms nest.
    return sorted(out, key=lambda x: (-len(x[0]), x[0]))


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--json", type=Path)
    ap.add_argument("--max-examples", type=int, default=60)
    args = ap.parse_args()

    terms = load_terms()
    hits: Counter[str] = Counter()
    mismatches: list[dict[str, object]] = []
    seen_ids: set[int] = set()

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
            # Identical legacy overlays are already accepted by the main loader;
            # only analyze each global message ID once.
            if numeric in seen_ids:
                continue
            seen_ids.add(numeric)
            ja = rec.get("japanese")
            ko = rec.get("korean")
            if not isinstance(ja, str) or not isinstance(ko, str):
                continue
            for ja_term, ko_term in terms:
                if ja_term not in ja:
                    continue
                hits[ja_term] += 1
                if ko_term not in ko:
                    mismatches.append({
                        "section": section,
                        "id": str(rid),
                        "path": path.name,
                        "japanese_term": ja_term,
                        "canonical_korean": ko_term,
                        "japanese": ja,
                        "korean": ko,
                    })

    mismatch_terms = Counter(str(x["japanese_term"]) for x in mismatches)
    report = {
        "schema": 1,
        "canonical_terms": len(terms),
        "matched_occurrences": sum(hits.values()),
        "terms_with_hits": sum(1 for v in hits.values() if v),
        "mismatch_count": len(mismatches),
        "mismatch_by_term": dict(sorted(mismatch_terms.items(), key=lambda kv: (-kv[1], kv[0]))),
        "hit_count_by_term": dict(sorted(hits.items(), key=lambda kv: (-kv[1], kv[0]))),
        "mismatches": mismatches,
    }
    if args.json:
        args.json.parent.mkdir(parents=True, exist_ok=True)
        args.json.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    print("Korean QA-2 canonical terminology scan")
    print(f"  canonical_terms: {report['canonical_terms']}")
    print(f"  terms_with_hits: {report['terms_with_hits']}")
    print(f"  matched_occurrences: {report['matched_occurrences']}")
    print(f"  mismatch_count: {report['mismatch_count']}")
    print("  mismatch_by_term: " + json.dumps(report["mismatch_by_term"], ensure_ascii=False))
    for row in mismatches[: max(args.max_examples, 0)]:
        print("  example: " + json.dumps(row, ensure_ascii=False, sort_keys=True))


if __name__ == "__main__":
    main()
