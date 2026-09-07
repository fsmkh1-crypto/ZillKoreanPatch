#!/usr/bin/env python3
"""Evaluate completed blind KEEP audit results and decide the measured-cohort gate."""
from __future__ import annotations

import argparse
import csv
import json
import re
import subprocess
import sys
import tempfile
from collections import Counter, defaultdict
from pathlib import Path

ALLOWED_CLASSES = {
    "semantic",
    "speaker_register",
    "naturalness",
    "punctuation_spacing",
}
MEASURED_COHORT = "measured_022"
GATE_LIMITS = {
    "semantic": 0.02,
    "speaker_register": 0.02,
    "naturalness": 0.05,
}


def regenerate_mapping(root: Path, english_root: Path, config: Path) -> dict:
    builder = root / "tools/korean/build-beta1-blind-keep-audit.py"
    with tempfile.TemporaryDirectory(prefix="beta1-blind-audit-") as tmp:
        tmpdir = Path(tmp)
        mapping = tmpdir / "mapping.json"
        public = tmpdir / "public"
        cmd = [
            sys.executable,
            str(builder),
            "--root",
            str(root),
            "--english-root",
            str(english_root),
            "--config",
            str(config),
            "--output-dir",
            str(public),
            "--mapping-output",
            str(mapping),
        ]
        subprocess.run(cmd, check=True, stdout=subprocess.PIPE, text=True, encoding="utf-8")
        return json.loads(mapping.read_text(encoding="utf-8"))


def parse_classes(raw: str) -> set[str]:
    vals = {x.strip() for x in re.split(r"[;,]", raw or "") if x.strip()}
    unknown = vals - ALLOWED_CLASSES
    if unknown:
        raise ValueError(f"unknown error classes: {sorted(unknown)}")
    return vals


def load_results(path: Path, expected_ids: set[str]) -> dict[str, dict]:
    with path.open("r", encoding="utf-8-sig", newline="") as f:
        reader = csv.DictReader(f)
        required = {"anonymous_id", "verdict", "error_classes", "proposed_korean", "reason"}
        missing = required - set(reader.fieldnames or [])
        if missing:
            raise SystemExit(f"results CSV missing columns: {sorted(missing)}")
        rows: dict[str, dict] = {}
        for line_no, row in enumerate(reader, 2):
            anon = (row.get("anonymous_id") or "").strip()
            if not anon:
                raise SystemExit(f"blank anonymous_id at CSV line {line_no}")
            if anon in rows:
                raise SystemExit(f"duplicate anonymous_id {anon}")
            verdict = (row.get("verdict") or "").strip().upper()
            if verdict not in {"KEEP", "SHOULD_EDIT"}:
                raise SystemExit(f"invalid verdict for {anon}: {verdict!r}")
            try:
                classes = parse_classes(row.get("error_classes") or "")
            except ValueError as exc:
                raise SystemExit(f"{anon}: {exc}") from exc
            proposed = row.get("proposed_korean") or ""
            reason = row.get("reason") or ""
            if verdict == "KEEP":
                if classes:
                    raise SystemExit(f"KEEP row {anon} must not carry error classes")
                if proposed.strip():
                    raise SystemExit(f"KEEP row {anon} must not propose replacement Korean")
            else:
                if not classes:
                    raise SystemExit(f"SHOULD_EDIT row {anon} needs at least one error class")
                if not proposed.strip():
                    raise SystemExit(f"SHOULD_EDIT row {anon} needs proposed_korean")
                if not reason.strip():
                    raise SystemExit(f"SHOULD_EDIT row {anon} needs a reason")
            rows[anon] = {
                "verdict": verdict,
                "classes": sorted(classes),
                "proposed_korean": proposed,
                "reason": reason,
            }
    actual = set(rows)
    if actual != expected_ids:
        missing = sorted(expected_ids - actual)
        extra = sorted(actual - expected_ids)
        raise SystemExit(f"results row set mismatch missing={missing[:10]} extra={extra[:10]}")
    return rows


def summarize(mapping: dict, results: dict[str, dict]) -> dict:
    by_cohort: dict[str, list[tuple[dict, dict]]] = defaultdict(list)
    corrections = []
    for map_row in mapping["rows"]:
        anon = map_row["anonymous_id"]
        result = results[anon]
        cohort = map_row["cohort"]
        by_cohort[cohort].append((map_row, result))
        if result["verdict"] == "SHOULD_EDIT":
            corrections.append(
                {
                    "anonymous_id": anon,
                    "real_id": map_row["real_id"],
                    "cohort": cohort,
                    "batch_id": map_row.get("batch_id"),
                    "error_classes": result["classes"],
                    "proposed_korean": result["proposed_korean"],
                    "reason": result["reason"],
                }
            )

    cohort_summary = {}
    for cohort, pairs in sorted(by_cohort.items()):
        total = len(pairs)
        should_edit = sum(r["verdict"] == "SHOULD_EDIT" for _, r in pairs)
        class_counts = Counter()
        for _, result in pairs:
            for cls in result["classes"]:
                class_counts[cls] += 1
        cohort_summary[cohort] = {
            "total": total,
            "should_edit": should_edit,
            "should_edit_rate": should_edit / total if total else 0.0,
            "class_counts": {cls: class_counts.get(cls, 0) for cls in sorted(ALLOWED_CLASSES)},
            "class_rates": {
                cls: (class_counts.get(cls, 0) / total if total else 0.0)
                for cls in sorted(ALLOWED_CLASSES)
            },
        }

    measured = cohort_summary.get(MEASURED_COHORT)
    if not measured:
        raise SystemExit(f"measured cohort {MEASURED_COHORT} missing")
    checks = {}
    for cls, limit in GATE_LIMITS.items():
        rate = measured["class_rates"][cls]
        checks[cls] = {
            "rate": rate,
            "limit": limit,
            "pass": rate <= limit,
            "affected_rows": measured["class_counts"][cls],
            "total_rows": measured["total"],
        }
    expansion_gate = all(v["pass"] for v in checks.values())
    return {
        "schema_version": 1,
        "audit_id": mapping["audit_id"],
        "korean_head": mapping["korean_head"],
        "english_sha": mapping["english_sha"],
        "cohorts": cohort_summary,
        "measured_cohort_gate": {
            "cohort": MEASURED_COHORT,
            "checks": checks,
            "punctuation_spacing_is_diagnostic_only": True,
            "pass_for_023_batch_expansion": expansion_gate,
        },
        "corrections": corrections,
    }


def render_markdown(doc: dict) -> str:
    lines = [
        f"# Blind KEEP audit evaluation — {doc['audit_id']}",
        "",
        f"- Korean head: `{doc['korean_head']}`",
        f"- Pinned English: `{doc['english_sha']}`",
        "",
        "## Cohort results",
        "",
        "| Cohort | Rows | SHOULD_EDIT | Rate | Semantic | Speaker/register | Naturalness | Punctuation/spacing |",
        "| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |",
    ]
    for name, row in doc["cohorts"].items():
        rates = row["class_rates"]
        lines.append(
            f"| {name} | {row['total']} | {row['should_edit']} | {row['should_edit_rate']:.3%} | "
            f"{rates['semantic']:.3%} | {rates['speaker_register']:.3%} | "
            f"{rates['naturalness']:.3%} | {rates['punctuation_spacing']:.3%} |"
        )
    gate = doc["measured_cohort_gate"]
    lines += ["", "## Measured-cohort expansion gate", ""]
    for cls in ("semantic", "speaker_register", "naturalness"):
        c = gate["checks"][cls]
        lines.append(
            f"- {cls}: {c['affected_rows']}/{c['total_rows']} = {c['rate']:.3%}; "
            f"limit {c['limit']:.1%} — {'PASS' if c['pass'] else 'FAIL'}"
        )
    lines += [
        "- punctuation_spacing: tracked separately; not part of the expansion gate.",
        "",
        f"**023 expansion gate: {'PASS' if gate['pass_for_023_batch_expansion'] else 'FAIL'}**",
        "",
        f"Corrections requiring follow-up: **{len(doc['corrections'])}**",
    ]
    return "\n".join(lines) + "\n"


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--root", type=Path, default=Path("."))
    ap.add_argument("--english-root", type=Path, required=True)
    ap.add_argument("--config", type=Path, required=True)
    ap.add_argument("--results", type=Path, required=True)
    ap.add_argument("--json-output", type=Path)
    ap.add_argument("--markdown-output", type=Path)
    args = ap.parse_args()

    root = args.root.resolve()
    english_root = args.english_root.resolve()
    config = args.config if args.config.is_absolute() else root / args.config
    results_path = args.results if args.results.is_absolute() else root / args.results
    mapping = regenerate_mapping(root, english_root, config)
    expected = {row["anonymous_id"] for row in mapping["rows"]}
    results = load_results(results_path, expected)
    summary = summarize(mapping, results)

    if args.json_output:
        out = args.json_output if args.json_output.is_absolute() else root / args.json_output
        out.parent.mkdir(parents=True, exist_ok=True)
        out.write_text(json.dumps(summary, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    if args.markdown_output:
        out = args.markdown_output if args.markdown_output.is_absolute() else root / args.markdown_output
        out.parent.mkdir(parents=True, exist_ok=True)
        out.write_text(render_markdown(summary), encoding="utf-8")

    print(json.dumps(summary["measured_cohort_gate"], ensure_ascii=False, indent=2))
    return 0 if summary["measured_cohort_gate"]["pass_for_023_batch_expansion"] else 2


if __name__ == "__main__":
    raise SystemExit(main())
