#!/usr/bin/env python3
"""Build the authoritative sparse Beta1 language-review ledger.

The ledger deliberately distinguishes automated coverage from human/contextual
review. Candidate discovery scans do NOT count as CONTEXT_KEEP. Only explicit
review scopes and approved copyedit manifests can advance contextual coverage.
"""
from __future__ import annotations

import argparse
import json
import re
import tomllib
from collections import defaultdict
from pathlib import Path

MANIFEST_RE = re.compile(r"beta1-contextual-copyedit-(\d{3})-reviewed\.json$")
STATUS_ORDER = {"AUTO_ONLY": 0, "CONTEXT_KEEP": 1, "CONTEXT_EDIT": 2}


def load_accepted(root: Path) -> dict[str, list[str]]:
    accepted: dict[str, list[str]] = defaultdict(list)
    for path in sorted((root / "translations/korean/messages").glob("*.toml")):
        data = tomllib.loads(path.read_text(encoding="utf-8"))
        rel = path.relative_to(root).as_posix()
        for rid, row in data.items():
            if isinstance(row, dict) and isinstance(row.get("korean"), str):
                accepted[str(rid)].append(rel)
    if not accepted:
        raise SystemExit("no accepted Korean records found")
    return dict(accepted)


def promote(entries: dict[str, dict], rid: str, status: str, batch: str, *, edited: bool) -> None:
    item = entries.setdefault(rid, {"status": "AUTO_ONLY", "review_batches": [], "edit_batches": []})
    if STATUS_ORDER[status] > STATUS_ORDER[item["status"]]:
        item["status"] = status
    if batch not in item["review_batches"]:
        item["review_batches"].append(batch)
    if edited and batch not in item["edit_batches"]:
        item["edit_batches"].append(batch)


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--root", default=".")
    ap.add_argument("--scopes", default="docs/audit/beta1-context-review-scopes.json")
    ap.add_argument("--ledger", default="docs/audit/beta1-review-ledger.json")
    ap.add_argument("--summary", default="docs/audit/beta1-review-coverage.md")
    args = ap.parse_args()

    root = Path(args.root).resolve()
    accepted = load_accepted(root)
    accepted_ids = set(accepted)
    scopes_path = root / args.scopes
    scopes = json.loads(scopes_path.read_text(encoding="utf-8"))

    entries: dict[str, dict] = {}
    scope_rows: list[dict] = []

    for scope in scopes.get("reviewed_scopes", []):
        batch = str(scope["batch"]).zfill(3)
        kind = scope["kind"]
        if kind == "numeric_range":
            start = int(scope["start"])
            end = int(scope["end"])
            ids = [str(i) for i in range(start, end + 1)]
        elif kind == "ids":
            ids = [str(x) for x in scope["ids"]]
        else:
            raise SystemExit(f"unknown scope kind {kind!r}")
        missing = [rid for rid in ids if rid not in accepted_ids]
        if missing:
            raise SystemExit(f"review scope batch {batch} contains non-accepted IDs: {missing[:10]}")
        expected = scope.get("expected_count")
        if expected is not None and len(ids) != int(expected):
            raise SystemExit(f"review scope batch {batch} expected {expected}, got {len(ids)}")
        for rid in ids:
            promote(entries, rid, "CONTEXT_KEEP", batch, edited=False)
        scope_rows.append({
            "batch": batch,
            "kind": kind,
            "count": len(ids),
            "evidence": scope.get("evidence", ""),
            "note": scope.get("note", ""),
        })

    manifests = []
    per_batch = []
    for path in sorted((root / "docs/audit").glob("beta1-contextual-copyedit-*-reviewed.json")):
        m = MANIFEST_RE.search(path.name)
        if not m:
            continue
        batch = m.group(1)
        doc = json.loads(path.read_text(encoding="utf-8"))
        record_ids = []
        for rec in doc.get("records", []):
            rid = str(rec["id"])
            if rid not in accepted_ids:
                raise SystemExit(f"manifest {path.name} references non-accepted ID {rid}")
            record_ids.append(rid)
            promote(entries, rid, "CONTEXT_EDIT", batch, edited=True)
        manifests.append(path.relative_to(root).as_posix())
        per_batch.append({"batch": batch, "records": len(record_ids), "unique_ids": len(set(record_ids)), "manifest": path.relative_to(root).as_posix()})

    # Sparse representation: AUTO_ONLY is implicit for every accepted ID not listed.
    sparse = {}
    for rid in sorted(entries, key=lambda x: (int(x) if x.isdigit() else 10**30, x)):
        item = entries[rid]
        if item["status"] == "AUTO_ONLY":
            continue
        item["review_batches"].sort()
        item["edit_batches"].sort()
        sparse[rid] = item

    context_edit = sum(1 for x in sparse.values() if x["status"] == "CONTEXT_EDIT")
    context_keep = sum(1 for x in sparse.values() if x["status"] == "CONTEXT_KEEP")
    context_reviewed = context_edit + context_keep
    auto_only = len(accepted_ids) - context_reviewed

    ledger = {
        "version": 1,
        "purpose": "Authoritative sparse Beta1 language-review coverage ledger",
        "accepted_total": len(accepted_ids),
        "default_status": "AUTO_ONLY",
        "status_semantics": {
            "AUTO_ONLY": "Accepted Korean record covered by automated/static QA, but no proven full contextual review is recorded here.",
            "CONTEXT_KEEP": "Japanese/English/Korean context was explicitly reviewed and no edit was required at that review point.",
            "CONTEXT_EDIT": "Japanese/English/Korean context was explicitly reviewed and at least one approved contextual copyedit was applied.",
        },
        "counting_policy": "Candidate scans and regex discovery do not count as contextual review unless an explicit review scope or approved manifest proves it.",
        "coverage": {
            "context_reviewed_unique": context_reviewed,
            "context_edit_unique": context_edit,
            "context_keep_unique": context_keep,
            "auto_only_unique": auto_only,
            "context_review_percent": round(context_reviewed * 100 / len(accepted_ids), 3),
        },
        "explicit_review_scopes": scope_rows,
        "manifests": manifests,
        "entries": sparse,
    }

    ledger_path = root / args.ledger
    ledger_path.write_text(json.dumps(ledger, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    total_manifest_records = sum(x["records"] for x in per_batch)
    lines = [
        "# Beta1 Language Review Coverage",
        "",
        "> Generated by `tools/korean/build-beta1-review-ledger.py`. Do not hand-edit the generated ledger or this summary.",
        "",
        "## Current coverage",
        "",
        f"- Accepted Korean IDs: **{len(accepted_ids):,}**",
        f"- Proven contextual review, unique IDs: **{context_reviewed:,} ({context_reviewed * 100 / len(accepted_ids):.3f}%)**",
        f"  - `CONTEXT_EDIT`: **{context_edit:,}**",
        f"  - `CONTEXT_KEEP`: **{context_keep:,}**",
        f"- `AUTO_ONLY`: **{auto_only:,}**",
        f"- Approved manifest records (not de-duplicated across batches): **{total_manifest_records:,}**",
        "",
        "## Counting rule",
        "",
        "A scanner hit, regex candidate, or automated QA pass is not counted as human/contextual coverage. An ID becomes context-reviewed only when an explicit reviewed scope proves it was read, or an approved contextual-copyedit manifest proves review through an applied edit. This intentionally understates historical coverage rather than inventing KEEP decisions.",
        "",
        "## Explicit full-review scopes",
        "",
    ]
    if scope_rows:
        lines += ["| Batch | Scope | IDs | Evidence |", "| --- | --- | ---: | --- |"]
        for row in scope_rows:
            lines.append(f"| {row['batch']} | {row['kind']} | {row['count']:,} | {row['evidence']} |")
    else:
        lines.append("None recorded.")
    lines += ["", "## Approved edit manifests", "", "| Batch | Records | Unique IDs | Manifest |", "| --- | ---: | ---: | --- |"]
    for row in per_batch:
        lines.append(f"| {row['batch']} | {row['records']:,} | {row['unique_ids']:,} | `{row['manifest']}` |")
    lines += [
        "",
        "## Interpretation",
        "",
        "This percentage is **proven unique contextual coverage**, not an estimate of translation quality. The true amount historically read may be higher; it is deliberately not credited unless auditable evidence exists. From the next sequential-review batches onward, KEEP IDs should be recorded explicitly so this percentage can become a true whole-corpus progress meter.",
        "",
    ]
    (root / args.summary).write_text("\n".join(lines), encoding="utf-8")


if __name__ == "__main__":
    main()
