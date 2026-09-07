#!/usr/bin/env python3
"""Attach engine-derived consumer/runtime structure to Beta1 review ledger v3.

This script MUST NOT change language equivalence or CONTEXT_* coverage.
Consumer/storage/runtime data is audit metadata and structural QA state only.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import tomllib
from collections import defaultdict
from pathlib import Path


def row_signature(info: dict) -> str:
    payload = {
        "category": info.get("category", ""), "basis": info.get("basis", ""), "consumer": info.get("consumer", ""),
        "fixed_buffer": bool(info.get("fixed_buffer")), "fixed_contracts": info.get("fixed_contracts", []),
        "runtime_pending": bool(info.get("runtime_pending")), "runtime_pending_reason": info.get("runtime_pending_reason", ""),
    }
    return hashlib.sha256(json.dumps(payload, sort_keys=True, separators=(",", ":")).encode()).hexdigest()


def load_current(root: Path):
    multi = defaultdict(list)
    for path in sorted((root / "translations/korean/messages").glob("*.toml")):
        data = tomllib.loads(path.read_text(encoding="utf-8")); rel = path.relative_to(root).as_posix()
        for rid, row in data.items():
            if isinstance(row, dict) and isinstance(row.get("korean"), str):
                multi[str(rid)].append((rel, row))
    out = {}
    for rid, items in multi.items():
        first = items[0][1]
        for _, row in items[1:]:
            for k in ("japanese", "korean", "layout", "consumer"):
                if row.get(k, "") != first.get(k, ""):
                    raise SystemExit(f"alias drift id={rid} key={k}")
        out[rid] = {
            "japanese": first.get("japanese", ""), "korean": first.get("korean", ""), "layout": first.get("layout", ""),
            "raw_consumer": first.get("consumer", ""), "paths": sorted(p for p, _ in items), "alias_count": len(items),
        }
    return out


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--consumer-map", required=True, type=Path)
    ap.add_argument("--root", default=".", type=Path)
    ap.add_argument("--ledger", default="translations/korean/review-ledger.jsonl", type=Path)
    ap.add_argument("--groups", default="docs/audit/beta1-review-groups.json", type=Path)
    ap.add_argument("--summary", default="docs/audit/beta1-review-coverage.md", type=Path)
    a = ap.parse_args()
    root = a.root.resolve(); cmap = json.loads(a.consumer_map.read_text(encoding="utf-8")); infos = cmap["entries"]
    current = load_current(root)
    if len(current) != int(cmap["accepted_ids"]):
        raise SystemExit("consumer map accepted population mismatch")

    ledger_path = root / a.ledger
    ledger = []
    flag_counts = defaultdict(int)
    for line in ledger_path.read_text(encoding="utf-8").splitlines():
        if not line.strip():
            continue
        row = json.loads(line); rid = str(row["id"]); info = infos.get(rid)
        if info is None:
            raise SystemExit(f"consumer map lacks ledger id {rid}")
        flags = [f for f in row.get("flags", []) if f not in ("FIXED_BUFFER", "RUNTIME_PENDING")]
        if info.get("fixed_buffer"):
            flags.append("FIXED_BUFFER")
        if info.get("runtime_pending"):
            flags.append("RUNTIME_PENDING")
        row["flags"] = sorted(set(flags))
        row["engine_consumer"] = {
            "category": info["category"], "basis": info["basis"], "consumer": info["consumer"],
            "fixed_contracts": info.get("fixed_contracts", []), "runtime_pending_reason": info.get("runtime_pending_reason", ""),
        }
        row["english_consumer_contract_sha256"] = cmap["english_contract_sha256"]
        row["engine_consumer_signature"] = row_signature(info)
        for f in row["flags"]:
            flag_counts[f] += 1
        ledger.append(row)
    ledger_path.write_text("".join(json.dumps(r, ensure_ascii=False, sort_keys=True, separators=(",", ":")) + "\n" for r in ledger), encoding="utf-8")

    # Preserve language-equivalence groups produced by build-beta1-review-ledger.py.
    # Enrichment adds structural population metadata, but must never tighten or
    # loosen the language propagation signature.
    gp = root / a.groups
    gd = json.loads(gp.read_text(encoding="utf-8"))
    gd["english_consumer_contract_sha256"] = cmap["english_contract_sha256"]
    gd["fixed_buffer_ids"] = int(cmap["fixed_buffer_ids"])
    gd["runtime_pending_ids"] = int(cmap["runtime_pending_ids"])
    gd["structural_metadata_policy"] = "consumer/alias/layout/runtime fields are retained for audit and structural QA, not language-equivalence grouping"
    gp.write_text(json.dumps(gd, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    sp = root / a.summary
    text = sp.read_text(encoding="utf-8")
    marker = "## Orthogonal flags currently derived"
    if marker not in text:
        raise SystemExit("coverage summary missing orthogonal flag section")
    start = text.index(marker)
    end = text.index("## Language propagation candidates", start)
    replacement = """## Orthogonal flags and structural QA\n\n- `PERSISTED_LAYOUT`: **%d** among ledger rows\n- `ALIAS_GROUP`: **%d** among ledger rows\n- `SOURCE_ANOMALY`: **%d** among ledger rows\n- `LAYOUT_RECHECK`: **%d** among ledger rows; does not invalidate language coverage\n- `FIXED_BUFFER`: **%d** among ledger rows; full accepted population **%d**\n- `RUNTIME_PENDING`: **%d** among ledger rows; full accepted population **%d**\n- English consumer/category contract SHA-256: `%s`\n\nConsumer/storage/runtime metadata remains in each ledger row for traceability, but it is not part of language propagation equivalence.\n\n""" % (
        flag_counts["PERSISTED_LAYOUT"], flag_counts["ALIAS_GROUP"], flag_counts["SOURCE_ANOMALY"], flag_counts["LAYOUT_RECHECK"],
        flag_counts["FIXED_BUFFER"], int(cmap["fixed_buffer_ids"]), flag_counts["RUNTIME_PENDING"], int(cmap["runtime_pending_ids"]),
        cmap["english_contract_sha256"],
    )
    sp.write_text(text[:start] + replacement + text[end:], encoding="utf-8")
    print(json.dumps({
        "status": "PASS", "ledger_rows": len(ledger), "fixed_buffer_population": cmap["fixed_buffer_ids"],
        "runtime_pending_population": cmap["runtime_pending_ids"], "language_groups_preserved": gd.get("duplicate_language_groups", 0),
        "language_propagation_candidate_extra_ids": gd.get("propagation_candidate_extra_ids", 0),
    }, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
