#!/usr/bin/env python3
"""Measure Beta1 review workload before setting dense-review throughput targets."""
from __future__ import annotations

import argparse
import json
import re
import tomllib
from collections import Counter, defaultdict
from pathlib import Path

CONTROL = re.compile(r"<[^>]+>")
SENTENCE_PUNCT = re.compile(r"[.!?…。！？、,;:：；…]")
BINS = [(10, "1-10"), (20, "11-20"), (40, "21-40"), (80, "41-80"), (160, "81-160"), (10**9, "161+")]


def visible(s: str) -> str:
    return CONTROL.sub("", s).replace("\r", "").replace("\n", "")


def bucket(n: int) -> str:
    for limit, name in BINS:
        if n <= limit:
            return name
    raise AssertionError


def load_rows(root: Path):
    multi = defaultdict(list)
    for path in sorted((root / "translations/korean/messages").glob("*.toml")):
        data = tomllib.loads(path.read_text(encoding="utf-8"))
        for rid, row in data.items():
            if isinstance(row, dict) and isinstance(row.get("korean"), str):
                multi[str(rid)].append(row)
    out = {}
    for rid, rows in multi.items():
        first = rows[0]
        for row in rows[1:]:
            for key in ("japanese", "korean", "layout", "consumer"):
                if row.get(key, "") != first.get(key, ""):
                    raise SystemExit(f"alias drift id={rid} key={key}")
        out[rid] = first
    return out


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--root", default=".", type=Path)
    ap.add_argument("--consumer-evidence", default="docs/audit/beta1-review-consumer-evidence.json", type=Path)
    ap.add_argument("--groups", default="docs/audit/beta1-review-groups.json", type=Path)
    ap.add_argument("--json", default="docs/audit/beta1-review-distribution.json", type=Path)
    ap.add_argument("--markdown", default="docs/audit/beta1-review-distribution.md", type=Path)
    a = ap.parse_args(); root = a.root.resolve()
    rows = load_rows(root)
    cmap = json.loads((root / a.consumer_evidence).read_text(encoding="utf-8"))
    groups_doc = json.loads((root / a.groups).read_text(encoding="utf-8"))
    infos = cmap["entries"]
    if set(rows) != set(infos):
        raise SystemExit("consumer evidence population differs from accepted Korean corpus")

    group_size = defaultdict(lambda: 1)
    for g in groups_doc.get("groups", []):
        ids = [str(x) for x in g["ids"]]
        for rid in ids:
            group_size[rid] = len(ids)

    ko_bins = Counter(); jp_bins = Counter(); cross = Counter(); consumer_counts = Counter(); category_counts = Counter()
    duplicate_sizes = Counter(); fixed_by_ko = Counter(); control_count = punct_count = rapid_candidate = 0
    rapid_duplicate_extra = 0

    for rid, row in rows.items():
        ko = visible(row.get("korean", "")); jp = visible(row.get("japanese", ""))
        kb = bucket(len(ko)); jb = bucket(len(jp)); ko_bins[kb] += 1; jp_bins[jb] += 1; cross[(jb, kb)] += 1
        info = infos[rid]; consumer_counts[info.get("consumer", "unproven")] += 1; category_counts[info.get("category", "uncategorized")] += 1
        if info.get("fixed_buffer"):
            fixed_by_ko[kb] += 1
        controls = [x for x in CONTROL.findall(row.get("korean", "")) if x != "<end>"]
        has_controls = bool(controls); has_punct = bool(SENTENCE_PUNCT.search(ko))
        control_count += int(has_controls); punct_count += int(has_punct)
        gsize = group_size[rid]; duplicate_sizes[gsize] += 1

        # This is only a workload candidate, never contextual credit: both JP and
        # KO are short, no semantic/runtime controls beyond <end>, and no sentence
        # punctuation. It deliberately does NOT use FIXED_BUFFER as a proxy.
        if len(ko) <= 20 and len(jp) <= 20 and not has_controls and not has_punct:
            rapid_candidate += 1
            if gsize > 1:
                rapid_duplicate_extra += 1

    order = [name for _, name in BINS]
    result = {
        "accepted_ids": len(rows),
        "ko_visible_length_bins": {k: ko_bins[k] for k in order},
        "jp_visible_length_bins": {k: jp_bins[k] for k in order},
        "fixed_buffer_by_ko_length": {k: fixed_by_ko[k] for k in order},
        "rows_with_non_end_controls": control_count,
        "rows_with_sentence_punctuation": punct_count,
        "language_duplicate_group_size_population": {str(k): duplicate_sizes[k] for k in sorted(duplicate_sizes)},
        "rapid_scan_candidate": {
            "definition": "visible KO<=20 AND visible JP<=20 AND no non-<end> controls AND no sentence punctuation",
            "ids": rapid_candidate,
            "duplicate_member_ids_within_candidate": rapid_duplicate_extra,
            "note": "workload triage only; never review credit",
        },
        "top_consumers": consumer_counts.most_common(20),
        "top_categories": category_counts.most_common(30),
    }
    (root / a.json).write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    lines = ["# Beta1 Review Workload Distribution", "", "> Generated workload measurement. These classes do not grant review coverage.", "",
             f"- Accepted IDs: **{len(rows):,}**", f"- Rows with non-`<end>` controls: **{control_count:,}**",
             f"- Rows with sentence punctuation: **{punct_count:,}**", f"- Conservative rapid-scan candidates: **{rapid_candidate:,}**", "",
             "Rapid-scan candidate means visible KO <= 20 and visible JP <= 20, no non-`<end>` controls, and no sentence punctuation. `FIXED_BUFFER` is intentionally not used as a shortcut.", "",
             "## Visible length distribution", "", "| Length | KO IDs | JP IDs | FIXED_BUFFER within KO bin |", "| --- | ---: | ---: | ---: |"]
    for k in order:
        lines.append(f"| {k} | {ko_bins[k]:,} | {jp_bins[k]:,} | {fixed_by_ko[k]:,} |")
    lines += ["", "## Language duplicate-group population", "", "| Group size | IDs belonging to groups of this size |", "| ---: | ---: |"]
    for size in sorted(duplicate_sizes):
        lines.append(f"| {size} | {duplicate_sizes[size]:,} |")
    lines += ["", "## Top consumers", "", "| Consumer | IDs |", "| --- | ---: |"]
    for name, count in consumer_counts.most_common(20):
        lines.append(f"| `{name}` | {count:,} |")
    lines += ["", "## Top categories", "", "| Category | IDs |", "| --- | ---: |"]
    for name, count in category_counts.most_common(30):
        lines.append(f"| `{name}` | {count:,} |")
    lines += ["", "## Planning rule", "", "Do not set IDs/hour from `FIXED_BUFFER` population or from this heuristic alone. Run the first dense-review pilot and replace estimates with measured seconds/ID for rapid-scan, ordinary dialogue, and high-risk rows.", ""]
    (root / a.markdown).write_text("\n".join(lines), encoding="utf-8")
    print(json.dumps({"status": "PASS", "accepted": len(rows), "rapid_scan_candidate": rapid_candidate, "controls": control_count, "punctuation": punct_count}, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
