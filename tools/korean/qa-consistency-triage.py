#!/usr/bin/env python3
"""Triage QA-3 repeated-source inconsistencies by risk."""
from __future__ import annotations

import argparse
import json
from pathlib import Path
import re
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[2]
QA3 = ROOT / "tools" / "korean" / "qa-consistency.py"
FIXED_TOKEN_RE = re.compile(r"<[^>]+>")
SPACE_RE = re.compile(r"[\s\u3000]+")
PUNCT_RE = re.compile(r"[\s\u3000\.,!?…。！？、・:：;；'\"“”‘’()（）\[\]{}<>〈〉《》「」『』…—―~～·]+")


def semantic_letters(text: str) -> str:
    text = FIXED_TOKEN_RE.sub("", text)
    return PUNCT_RE.sub("", text)


def source_visible_chars(text: str) -> int:
    text = FIXED_TOKEN_RE.sub("", text.replace("<line-break>", ""))
    return len(SPACE_RE.sub("", text))


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--json", type=Path)
    ap.add_argument("--max-examples", type=int, default=30)
    args = ap.parse_args()

    with tempfile.TemporaryDirectory() as td:
        raw = Path(td) / "qa3.json"
        subprocess.run(["python3", str(QA3), "--json", str(raw), "--max-examples", "0"], cwd=ROOT, check=True, stdout=subprocess.DEVNULL)
        report = json.loads(raw.read_text(encoding="utf-8"))

    buckets: dict[str, list[dict[str, object]]] = {"formatting_only": [], "short_label": [], "lexical": []}
    for group in report["groups"]:
        variants = [str(v["korean"]) for v in group["variants"]]
        lexical_forms = {semantic_letters(v) for v in variants}
        if len(lexical_forms) == 1:
            bucket = "formatting_only"
        elif source_visible_chars(str(group["japanese_example"])) <= 12:
            bucket = "short_label"
        else:
            bucket = "lexical"
        copy = dict(group)
        copy["triage"] = bucket
        buckets[bucket].append(copy)

    summary = {k: len(v) for k, v in buckets.items()}
    record_counts = {k: sum(int(g["occurrences"]) for g in v) for k, v in buckets.items()}
    out = {"schema": 1, "input_inconsistent_groups": report["inconsistent_groups"], "input_inconsistent_records": report["inconsistent_records"], "groups_by_triage": summary, "records_by_triage": record_counts, "buckets": buckets}
    if args.json:
        args.json.parent.mkdir(parents=True, exist_ok=True)
        args.json.write_text(json.dumps(out, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    print("Korean QA-3 consistency triage")
    print("  groups_by_triage: " + json.dumps(summary, ensure_ascii=False, sort_keys=True))
    print("  records_by_triage: " + json.dumps(record_counts, ensure_ascii=False, sort_keys=True))
    shown = 0
    for bucket in ("formatting_only", "short_label", "lexical"):
        for group in buckets[bucket]:
            if shown >= max(args.max_examples, 0):
                return
            compact = {"triage": bucket, "occurrences": group["occurrences"], "variant_count": group["variant_count"], "japanese_example": group["japanese_example"], "variants": [{"korean": v["korean"], "count": v["count"]} for v in group["variants"]]}
            print("  example: " + json.dumps(compact, ensure_ascii=False, sort_keys=True))
            shown += 1


if __name__ == "__main__":
    main()
