#!/usr/bin/env python3
"""Build Beta1 review ledger v2 with stale-proof historical review bases.

Authoritative principles:
- contextual state is distinct from automated QA and orthogonal flags;
- historical review bases are reconstructed at the actual semantic commit;
- any change to JP/EN/KO/layout/consumer signature invalidates that review;
- scanner/mechanical-only evidence never becomes CONTEXT_KEEP;
- propagation is never credited automatically; only strict candidate counts are reported.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import re
import subprocess
import tomllib
from collections import defaultdict
from pathlib import Path

MANIFEST_RE = re.compile(r"beta1-contextual-copyedit-(\d{3})-reviewed\.json$")
ENGLISH_PIN_DEFAULT = "a98d9ce29f361d666ec23da0dcfd351f24537ffd"


def git_sha(root: Path) -> str:
    return subprocess.check_output(["git", "-C", str(root), "rev-parse", "HEAD"], text=True).strip()


def git_show_text(root: Path, commit: str, rel: str) -> str:
    try:
        return subprocess.check_output(
            ["git", "-C", str(root), "show", f"{commit}:{rel}"], text=True, encoding="utf-8"
        )
    except subprocess.CalledProcessError as exc:
        raise SystemExit(f"cannot reconstruct {rel} at {commit}: {exc}") from exc


def sort_id(rid: str):
    return (0, int(rid)) if rid.isdigit() else (1, rid)


def load_korean_current(root: Path) -> dict[str, list[dict]]:
    out: dict[str, list[dict]] = defaultdict(list)
    base = root / "translations/korean/messages"
    for path in sorted(base.glob("*.toml")):
        data = tomllib.loads(path.read_text(encoding="utf-8"))
        rel = path.relative_to(root).as_posix()
        for rid, row in data.items():
            if isinstance(row, dict) and isinstance(row.get("korean"), str):
                out[str(rid)].append({"path": rel, **row})
    if not out:
        raise SystemExit("no Korean accepted rows")
    return dict(out)


def canonical_rows(rid: str, rows: list[dict]) -> dict:
    first = rows[0]
    for row in rows[1:]:
        for key in ("japanese", "korean", "layout", "consumer"):
            if row.get(key, "") != first.get(key, ""):
                raise SystemExit(f"non-identical alias id={rid} key={key}")
    return {
        "id": rid,
        "paths": sorted(row["path"] for row in rows),
        "japanese": first.get("japanese", ""),
        "korean": first.get("korean", ""),
        "layout": first.get("layout", ""),
        "consumer": first.get("consumer", ""),
        "alias_count": len(rows),
    }


def load_english(root: Path) -> dict[str, dict]:
    out = {}
    for path in sorted((root / "translations/messages").glob("msgsec*.toml")):
        data = tomllib.loads(path.read_text(encoding="utf-8"))
        rel = path.relative_to(root).as_posix()
        for rid, row in data.items():
            if not isinstance(row, dict):
                continue
            rid = str(rid)
            if rid in out:
                raise SystemExit(f"duplicate English id {rid}")
            out[rid] = {
                "english": row.get("english", ""),
                "japanese": row.get("japanese", ""),
                "path": rel,
            }
    return out


def historical_row(root: Path, commit: str, rid: str, paths: list[str]) -> dict:
    found = []
    for rel in paths:
        data = tomllib.loads(git_show_text(root, commit, rel))
        row = data.get(rid)
        if isinstance(row, dict) and isinstance(row.get("korean"), str):
            found.append({"path": rel, **row})
    if not found:
        raise SystemExit(f"review basis missing id={rid} commit={commit} paths={paths}")
    return canonical_rows(rid, found)


def consumer_signature(row: dict) -> str:
    """Conservative repository-visible storage/consumer signature.

    Includes physical overlay paths and alias multiplicity. The raw consumer field
    is included when present. Persisted layout content is hashed separately, while
    its presence is also represented here. This never grants propagation by itself.
    """
    payload = {
        "paths": row["paths"],
        "alias_count": row["alias_count"],
        "raw_consumer": row.get("consumer", ""),
        "persisted_layout": bool(row.get("layout", "")),
    }
    return hashlib.sha256(
        json.dumps(payload, ensure_ascii=False, sort_keys=True, separators=(",", ":")).encode("utf-8")
    ).hexdigest()


def basis_sha(japanese: str, english: str, korean: str, layout: str, consumer_sig: str) -> str:
    return hashlib.sha256("\0".join([japanese, english, korean, layout, consumer_sig]).encode("utf-8")).hexdigest()


def load_source_anomalies(root: Path) -> set[str]:
    path = root / "tools/korean/known-source-anomalies.toml"
    if not path.exists():
        return set()
    data = tomllib.loads(path.read_text(encoding="utf-8"))
    return {str(k) for k, v in data.items() if isinstance(v, dict)}


def event_key(event: dict):
    return (int(event["batch"]) if str(event["batch"]).isdigit() else 10**9, event["evidence"])


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--root", default=".")
    ap.add_argument("--english-root", required=True, type=Path)
    ap.add_argument("--scopes", default="docs/audit/beta1-context-review-scopes.json")
    ap.add_argument("--commits", default="docs/audit/beta1-context-review-commits.json")
    ap.add_argument("--jsonl", default="translations/korean/review-ledger.jsonl")
    ap.add_argument("--summary", default="docs/audit/beta1-review-coverage.md")
    ap.add_argument("--groups", default="docs/audit/beta1-review-groups.json")
    args = ap.parse_args()

    root = Path(args.root).resolve()
    english_root = args.english_root.resolve()
    current_sha = git_sha(root)

    commit_map_doc = json.loads((root / args.commits).read_text(encoding="utf-8"))
    english_pin = commit_map_doc.get("english_reference_sha", ENGLISH_PIN_DEFAULT)
    actual_english_sha = git_sha(english_root)
    if actual_english_sha != english_pin:
        raise SystemExit(f"English checkout drift: got {actual_english_sha}, expected {english_pin}")
    semantic_commits = {str(k).zfill(3): str(v) for k, v in commit_map_doc["semantic_commits"].items()}

    current_rows_multi = load_korean_current(root)
    current = {rid: canonical_rows(rid, rows) for rid, rows in current_rows_multi.items()}
    english = load_english(english_root)
    accepted_ids = set(current)
    anomalies = load_source_anomalies(root)

    events: dict[str, list[dict]] = defaultdict(list)
    scopes = json.loads((root / args.scopes).read_text(encoding="utf-8"))
    for scope in scopes.get("reviewed_scopes", []):
        batch = str(scope["batch"]).zfill(3)
        if batch not in semantic_commits:
            raise SystemExit(f"no semantic commit recorded for scope batch {batch}")
        if scope["kind"] == "numeric_range":
            ids = [str(i) for i in range(int(scope["start"]), int(scope["end"]) + 1)]
        elif scope["kind"] == "ids":
            ids = [str(x) for x in scope["ids"]]
        else:
            raise SystemExit(f"unknown scope kind {scope['kind']!r}")
        if scope.get("expected_count") is not None and len(ids) != int(scope["expected_count"]):
            raise SystemExit(f"scope count mismatch batch {batch}")
        for rid in ids:
            if rid not in accepted_ids:
                raise SystemExit(f"scope {batch} non-accepted id {rid}")
            events[rid].append({
                "batch": batch,
                "evidence": "full_read",
                "state": "CONTEXT_KEEP",
                "reviewed_commit": semantic_commits[batch],
                "paths": current[rid]["paths"],
            })

    per_batch = []
    for path in sorted((root / "docs/audit").glob("beta1-contextual-copyedit-*-reviewed.json")):
        m = MANIFEST_RE.search(path.name)
        if not m:
            continue
        batch = m.group(1)
        if batch not in semantic_commits:
            raise SystemExit(f"no semantic commit mapping for manifest batch {batch}")
        doc = json.loads(path.read_text(encoding="utf-8"))
        ids = []
        for rec in doc.get("records", []):
            rid = str(rec["id"])
            if rid not in accepted_ids:
                raise SystemExit(f"manifest {path.name} non-accepted id {rid}")
            ids.append(rid)
            events[rid].append({
                "batch": batch,
                "evidence": "manifest_edit",
                "state": "CONTEXT_EDIT",
                "reviewed_commit": semantic_commits[batch],
                "paths": list(rec.get("paths") or current[rid]["paths"]),
            })
        per_batch.append({
            "batch": batch,
            "records": len(ids),
            "unique_ids": len(set(ids)),
            "manifest": path.relative_to(root).as_posix(),
        })

    ledger_rows = []
    stale_ids = []
    valid_direct = 0
    valid_manifest = 0
    flag_counts = defaultdict(int)

    for rid in sorted(events, key=sort_id):
        latest = sorted(events[rid], key=event_key)[-1]
        hist = historical_row(root, latest["reviewed_commit"], rid, latest["paths"])
        eng = english.get(rid, {}).get("english", "")
        hist_sig = consumer_signature(hist)
        hist_basis = basis_sha(hist["japanese"], eng, hist["korean"], hist["layout"], hist_sig)

        cur = current[rid]
        cur_sig = consumer_signature(cur)
        cur_basis = basis_sha(cur["japanese"], eng, cur["korean"], cur["layout"], cur_sig)
        stale = hist_basis != cur_basis
        state = "CONTEXT_STALE" if stale else latest["state"]

        flags = []
        if cur["layout"]:
            flags.append("PERSISTED_LAYOUT")
        if cur["alias_count"] > 1:
            flags.append("ALIAS_GROUP")
        if rid in anomalies:
            flags.append("SOURCE_ANOMALY")
        for flag in flags:
            flag_counts[flag] += 1

        if stale:
            stale_ids.append(rid)
        elif latest["evidence"] == "full_read":
            valid_direct += 1
        elif latest["evidence"] == "manifest_edit":
            valid_manifest += 1

        ledger_rows.append({
            "id": rid,
            "context_state": state,
            "evidence": latest["evidence"],
            "batch_id": latest["batch"],
            "reviewed_commit": latest["reviewed_commit"],
            "basis_sha256": hist_basis,
            "current_basis_sha256": cur_basis,
            "english_sha": english_pin,
            "consumer_signature": cur_sig,
            "flags": flags,
            "propagated_from": None,
        })

    valid_context = valid_direct + valid_manifest
    unreviewed = len(accepted_ids) - valid_context - len(stale_ids)

    groups = defaultdict(list)
    for rid, row in current.items():
        key = (row["japanese"], row["korean"], row["layout"], consumer_signature(row))
        groups[key].append(rid)
    duplicate_groups = [ids for ids in groups.values() if len(ids) > 1]
    propagation_candidate_extra = sum(len(ids) - 1 for ids in duplicate_groups)
    duplicate_member_ids = sum(len(ids) for ids in duplicate_groups)

    group_doc = {
        "version": 2,
        "korean_head": current_sha,
        "english_sha": english_pin,
        "accepted_ids": len(accepted_ids),
        "unique_strict_signatures": len(groups),
        "strict_duplicate_groups": len(duplicate_groups),
        "strict_duplicate_member_ids": duplicate_member_ids,
        "strict_propagation_candidate_extra_ids": propagation_candidate_extra,
        "counting_policy": "Candidate-only. No propagation is credited until a representative is directly reviewed and every member matches exact JP+KO+layout+consumer_signature at that review basis."
    }
    (root / args.groups).write_text(json.dumps(group_doc, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    out_path = root / args.jsonl
    out_path.parent.mkdir(parents=True, exist_ok=True)
    with out_path.open("w", encoding="utf-8") as fh:
        for row in ledger_rows:
            fh.write(json.dumps(row, ensure_ascii=False) + "\n")

    total_manifest_records = sum(x["records"] for x in per_batch)
    lines = [
        "# Beta1 Language Review Coverage v2", "",
        "> Generated. Do not hand-edit `translations/korean/review-ledger.jsonl` or this summary.", "",
        "## Authoritative current coverage", "",
        f"- Accepted Korean IDs: **{len(accepted_ids):,}**",
        f"- Valid contextual review: **{valid_context:,} ({valid_context * 100 / len(accepted_ids):.3f}%)**",
        f"  - direct `full_read`: **{valid_direct:,}**",
        f"  - direct `manifest_edit`: **{valid_manifest:,}**",
        "  - propagated: **0** (not yet credited)",
        f"- `CONTEXT_STALE`: **{len(stale_ids):,}**",
        f"- `UNREVIEWED` for contextual purposes: **{unreviewed:,}**",
        f"- Approved manifest records (historical, non-deduplicated): **{total_manifest_records:,}**", "",
        "A row is valid only when its historical review basis still matches current",
        "`SHA256(JP + NUL + pinned EN + NUL + KO + NUL + layout + NUL + consumer_signature)`.",
        "Mismatch automatically reports `CONTEXT_STALE` and removes the row from valid coverage.", "",
        "## Orthogonal flags currently derived", "",
        f"- `PERSISTED_LAYOUT`: **{flag_counts['PERSISTED_LAYOUT']:,}** among ledger rows",
        f"- `ALIAS_GROUP`: **{flag_counts['ALIAS_GROUP']:,}** among ledger rows",
        f"- `SOURCE_ANOMALY`: **{flag_counts['SOURCE_ANOMALY']:,}** among ledger rows",
        "- `FIXED_BUFFER` / `RUNTIME_PENDING`: **not guessed**; pending integration with a repository-derived consumer/runtime map.", "",
        "## Strict propagation candidates (not coverage)", "",
        f"- Unique strict signatures: **{len(groups):,}**",
        f"- Duplicate strict-signature groups: **{len(duplicate_groups):,}**",
        f"- IDs inside such groups: **{duplicate_member_ids:,}**",
        f"- Potential extra IDs after one representative review per group: **{propagation_candidate_extra:,}**", "",
        "Strict signature requires exact Japanese + exact Korean + exact persisted layout +",
        "repository-visible consumer signature (paths, alias multiplicity, raw consumer metadata, persisted-layout state).",
        "These are candidates only; no automatic KEEP propagation is credited.", "",
        "## Historical edit manifests", "",
        "| Batch | Records | Unique IDs | Manifest |", "| --- | ---: | ---: | --- |",
    ]
    for row in per_batch:
        lines.append(f"| {row['batch']} | {row['records']:,} | {row['unique_ids']:,} | `{row['manifest']}` |")
    lines += ["", "## Completion/quality rule", "",
        "Coverage and accuracy are separate. Final Beta1 must additionally run a reproducible",
        "random second-pass audit of the valid KEEP population (target sample: 200); a correction",
        "rate above 5% invalidates the assumption that KEEP coverage is reliable and requires expanded re-review.",
        "Final scanners must include scanners introduced after the reviewed batches; scanner-zero alone is not",
        "evidence of whole-corpus correctness.", ""]
    (root / args.summary).write_text("\n".join(lines), encoding="utf-8")

    print(json.dumps({
        "status": "PASS", "schema_version": 2, "accepted": len(accepted_ids),
        "valid_context": valid_context, "stale": len(stale_ids), "unreviewed": unreviewed,
        "direct_full_read": valid_direct, "manifest_edit": valid_manifest, "propagated": 0,
        "strict_duplicate_groups": len(duplicate_groups),
        "strict_propagation_candidate_extra_ids": propagation_candidate_extra,
        "stale_ids": stale_ids,
    }, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
