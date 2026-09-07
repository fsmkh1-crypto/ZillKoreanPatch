#!/usr/bin/env python3
"""Build Beta1 review ledger v3 with ID-granular stale-proof language evidence.

Language coverage and runtime/layout structure are deliberately separate:
- language_basis = SHA256(JP + NUL + pinned EN + NUL + KO)
- structural_basis = SHA256(layout + NUL + physical consumer signature)

Only language_basis mismatch demotes contextual coverage to CONTEXT_STALE.
Structural mismatch sets LAYOUT_RECHECK and is closed by layout/consumer QA, not
by making a reviewer reread unchanged language.

Historical 001-017 evidence is reconstructed at the actual semantic commit.
New dense-read scope files are also reconstructed per ID at reviewed_commit.
Scanner/mechanical-only evidence never becomes contextual coverage.
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
    return subprocess.check_output(
        ["git", "-C", str(root), "show", f"{commit}:{rel}"], text=True, encoding="utf-8", stderr=subprocess.DEVNULL
    )


def git_paths_for_id(root: Path, commit: str, rid: str) -> list[str]:
    needle = f'["{rid}"]'
    proc = subprocess.run(
        ["git", "-C", str(root), "grep", "-l", "-F", needle, commit, "--", "translations/korean/messages"],
        text=True, encoding="utf-8", stdout=subprocess.PIPE, stderr=subprocess.PIPE
    )
    if proc.returncode not in (0, 1):
        raise SystemExit(f"git grep failed id={rid} commit={commit}: {proc.stderr.strip()}")
    prefix = commit + ":"
    paths = []
    for line in proc.stdout.splitlines():
        line = line.strip()
        if not line:
            continue
        paths.append(line[len(prefix):] if line.startswith(prefix) else line.split(":", 1)[-1])
    return sorted(set(paths))


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
            out[rid] = {"english": row.get("english", ""), "japanese": row.get("japanese", ""), "path": rel}
    return out


def historical_row(root: Path, commit: str, rid: str, preferred_paths: list[str]) -> dict:
    found = []
    for rel in preferred_paths:
        try:
            data = tomllib.loads(git_show_text(root, commit, rel))
        except subprocess.CalledProcessError:
            continue
        row = data.get(rid)
        if isinstance(row, dict) and isinstance(row.get("korean"), str):
            found.append({"path": rel, **row})
    if not found:
        for rel in git_paths_for_id(root, commit, rid):
            data = tomllib.loads(git_show_text(root, commit, rel))
            row = data.get(rid)
            if isinstance(row, dict) and isinstance(row.get("korean"), str):
                found.append({"path": rel, **row})
    if not found:
        raise SystemExit(f"review basis missing id={rid} commit={commit}")
    return canonical_rows(rid, found)


def physical_consumer_signature(row: dict) -> str:
    payload = {
        "paths": row["paths"],
        "alias_count": row["alias_count"],
        "raw_consumer": row.get("consumer", ""),
    }
    return hashlib.sha256(json.dumps(payload, ensure_ascii=False, sort_keys=True, separators=(",", ":")).encode()).hexdigest()


def language_basis(japanese: str, english: str, korean: str) -> str:
    return hashlib.sha256("\0".join([japanese, english, korean]).encode("utf-8")).hexdigest()


def structural_basis(layout: str, physical_sig: str) -> str:
    return hashlib.sha256("\0".join([layout, physical_sig]).encode("utf-8")).hexdigest()


def load_source_anomalies(root: Path) -> set[str]:
    path = root / "tools/korean/known-source-anomalies.toml"
    if not path.exists():
        return set()
    data = tomllib.loads(path.read_text(encoding="utf-8"))
    return {str(k) for k, v in data.items() if isinstance(v, dict)}


def event_key(event: dict):
    # Dense scope IDs and manifest edits both have an explicit reviewed_commit.
    # manifest_edit wins a tie so an edited ID is never mislabeled KEEP.
    evidence_order = {"full_read": 0, "scope_full_read": 1, "manifest_edit": 2}
    return (int(event.get("batch", "999999")) if str(event.get("batch", "")).isdigit() else 999999,
            evidence_order.get(event["evidence"], 9))


def scope_files(root: Path) -> list[Path]:
    base = root / "docs/audit/review"
    return sorted(base.glob("scope-*.json")) if base.exists() else []


def validate_scope(doc: dict, path: Path, english_pin: str, accepted_ids: set[str]) -> tuple[list[str], set[str], set[str]]:
    required = {"schema_version", "scope_id", "kind", "reviewed_commit", "english_sha", "packet_sha256", "reviewer", "ids", "edit_ids", "exclusions"}
    missing = sorted(required - set(doc))
    if missing:
        raise SystemExit(f"scope {path}: missing fields {missing}")
    if doc["schema_version"] != 1 or doc["kind"] != "sequential_full_read":
        raise SystemExit(f"scope {path}: unsupported schema/kind")
    if doc["english_sha"] != english_pin:
        raise SystemExit(f"scope {path}: English SHA drift")
    ids = [str(x) for x in doc["ids"]]
    if len(ids) != len(set(ids)):
        raise SystemExit(f"scope {path}: duplicate IDs")
    bad = sorted(set(ids) - accepted_ids, key=sort_id)
    if bad:
        raise SystemExit(f"scope {path}: non-accepted IDs {bad[:20]}")
    edits = {str(x) for x in doc["edit_ids"]}
    exclusions = {str(x) for x in doc["exclusions"]}
    if not edits <= set(ids):
        raise SystemExit(f"scope {path}: edit_ids outside scope")
    if not exclusions <= set(ids):
        raise SystemExit(f"scope {path}: exclusions outside scope")
    if edits & exclusions:
        raise SystemExit(f"scope {path}: edit/exclusion overlap")
    return ids, edits, exclusions


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
    if git_sha(english_root) != english_pin:
        raise SystemExit("English checkout drift")
    semantic_commits = {str(k).zfill(3): str(v) for k, v in commit_map_doc["semantic_commits"].items()}

    current_rows_multi = load_korean_current(root)
    current = {rid: canonical_rows(rid, rows) for rid, rows in current_rows_multi.items()}
    english = load_english(english_root)
    accepted_ids = set(current)
    anomalies = load_source_anomalies(root)

    events: dict[str, list[dict]] = defaultdict(list)
    pending_batches = set()

    # Historical range/ID scopes (currently 001) are conservative proven full reads.
    legacy_scopes = json.loads((root / args.scopes).read_text(encoding="utf-8"))
    for scope in legacy_scopes.get("reviewed_scopes", []):
        batch = str(scope["batch"]).zfill(3)
        if batch not in semantic_commits:
            pending_batches.add(batch)
            continue
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
            events[rid].append({"batch": batch, "evidence": "full_read", "state": "CONTEXT_KEEP",
                                "reviewed_commit": semantic_commits[batch], "paths": current[rid]["paths"]})

    # New dense-read scopes. Stale is ALWAYS evaluated per ID by reconstructing
    # reviewed_commit, never at scope/packet granularity. edit_ids are omitted here
    # because their later semantic manifest supplies CONTEXT_EDIT evidence.
    dense_scope_meta = []
    for path in scope_files(root):
        doc = json.loads(path.read_text(encoding="utf-8"))
        ids, edits, exclusions = validate_scope(doc, path, english_pin, accepted_ids)
        keep_ids = [rid for rid in ids if rid not in edits and rid not in exclusions]
        for rid in keep_ids:
            events[rid].append({"batch": doc.get("batch_id", "999"), "evidence": "scope_full_read", "state": "CONTEXT_KEEP",
                                "reviewed_commit": doc["reviewed_commit"], "paths": current[rid]["paths"],
                                "scope_id": doc["scope_id"], "reviewer": doc["reviewer"], "packet_sha256": doc["packet_sha256"]})
        dense_scope_meta.append({"scope_id": doc["scope_id"], "ids": len(ids), "keeps": len(keep_ids),
                                 "edits": len(edits), "exclusions": len(exclusions), "path": path.relative_to(root).as_posix()})

    per_batch = []
    for path in sorted((root / "docs/audit").glob("beta1-contextual-copyedit-*-reviewed.json")):
        m = MANIFEST_RE.search(path.name)
        if not m:
            continue
        batch = m.group(1)
        doc = json.loads(path.read_text(encoding="utf-8"))
        ids = [str(rec["id"]) for rec in doc.get("records", [])]
        if batch not in semantic_commits:
            pending_batches.add(batch)
            per_batch.append({"batch": batch, "records": len(ids), "unique_ids": len(set(ids)), "manifest": path.relative_to(root).as_posix(), "status": "PENDING_BASIS"})
            continue
        for rec in doc.get("records", []):
            rid = str(rec["id"])
            if rid not in accepted_ids:
                raise SystemExit(f"manifest {path.name} non-accepted id {rid}")
            events[rid].append({"batch": batch, "evidence": "manifest_edit", "state": "CONTEXT_EDIT",
                                "reviewed_commit": semantic_commits[batch], "paths": list(rec.get("paths") or current[rid]["paths"])})
        per_batch.append({"batch": batch, "records": len(ids), "unique_ids": len(set(ids)), "manifest": path.relative_to(root).as_posix(), "status": "REGISTERED"})

    ledger_rows = []
    stale_ids = []
    layout_recheck_ids = []
    valid_full_read = valid_scope = valid_manifest = 0
    flag_counts = defaultdict(int)

    for rid in sorted(events, key=sort_id):
        latest = sorted(events[rid], key=event_key)[-1]
        hist = historical_row(root, latest["reviewed_commit"], rid, latest["paths"])
        eng = english.get(rid, {}).get("english", "")
        hist_phys = physical_consumer_signature(hist)
        cur = current[rid]
        cur_phys = physical_consumer_signature(cur)
        hist_lang = language_basis(hist["japanese"], eng, hist["korean"])
        cur_lang = language_basis(cur["japanese"], eng, cur["korean"])
        hist_struct = structural_basis(hist["layout"], hist_phys)
        cur_struct = structural_basis(cur["layout"], cur_phys)
        stale = hist_lang != cur_lang
        layout_recheck = hist_struct != cur_struct
        state = "CONTEXT_STALE" if stale else latest["state"]

        flags = []
        if cur["layout"]:
            flags.append("PERSISTED_LAYOUT")
        if cur["alias_count"] > 1:
            flags.append("ALIAS_GROUP")
        if rid in anomalies:
            flags.append("SOURCE_ANOMALY")
        if layout_recheck:
            flags.append("LAYOUT_RECHECK")
            layout_recheck_ids.append(rid)
        for flag in flags:
            flag_counts[flag] += 1

        if stale:
            stale_ids.append(rid)
        elif latest["evidence"] == "full_read":
            valid_full_read += 1
        elif latest["evidence"] == "scope_full_read":
            valid_scope += 1
        else:
            valid_manifest += 1

        ledger_rows.append({
            "id": rid, "context_state": state, "evidence": latest["evidence"], "batch_id": latest.get("batch"),
            "scope_id": latest.get("scope_id"), "reviewer": latest.get("reviewer"), "reviewed_commit": latest["reviewed_commit"],
            "packet_sha256": latest.get("packet_sha256"), "english_sha": english_pin,
            "language_basis_sha256": hist_lang, "current_language_basis_sha256": cur_lang,
            "structural_basis_sha256": hist_struct, "current_structural_basis_sha256": cur_struct,
            "physical_consumer_signature": cur_phys, "flags": flags, "propagated_from": None,
        })

    valid_context = valid_full_read + valid_scope + valid_manifest
    unreviewed = len(accepted_ids) - valid_context - len(stale_ids)

    # Language-equivalence candidates: exact JP + pinned EN + exact KO only.
    # Structural/runtime metadata is retained in ledger but not used to decide
    # linguistic equivalence. EN mismatch therefore ALWAYS splits a group.
    groups = defaultdict(list)
    for rid, row in current.items():
        groups[(row["japanese"], english.get(rid, {}).get("english", ""), row["korean"])].append(rid)
    duplicate_groups = [sorted(ids, key=sort_id) for ids in groups.values() if len(ids) > 1]
    duplicate_groups.sort(key=lambda ids: sort_id(ids[0]))
    group_doc = {
        "version": 4, "korean_head": current_sha, "english_sha": english_pin, "accepted_ids": len(accepted_ids),
        "authority": "language-equivalence candidate = exact JP + exact pinned EN + exact KO; EN mismatch forbids propagation",
        "unique_language_signatures": len(groups), "duplicate_language_groups": len(duplicate_groups),
        "duplicate_member_ids": sum(len(ids) for ids in duplicate_groups),
        "propagation_candidate_extra_ids": sum(len(ids)-1 for ids in duplicate_groups),
        "pending_review_basis_batches": sorted(pending_batches),
        "counting_policy": "Candidate-only. KEEP propagation requires representative review with every member context displayed; propagated KEEP is over-sampled in second-pass QA. EDIT uses the same language-equivalence signature plus exact-before manifest gates.",
        "groups": [{"representative": ids[0], "ids": ids} for ids in duplicate_groups],
    }
    (root / args.groups).write_text(json.dumps(group_doc, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    out_path = root / args.jsonl
    out_path.parent.mkdir(parents=True, exist_ok=True)
    with out_path.open("w", encoding="utf-8") as fh:
        for row in ledger_rows:
            fh.write(json.dumps(row, ensure_ascii=False, sort_keys=True, separators=(",", ":")) + "\n")

    total_manifest_records = sum(x["records"] for x in per_batch if x["status"] == "REGISTERED")
    lines = [
        "# Beta1 Language Review Coverage v3", "",
        "> Generated. Do not hand-edit `translations/korean/review-ledger.jsonl` or this summary.", "",
        "## Authoritative current coverage", "",
        f"- Accepted Korean IDs: **{len(accepted_ids):,}**",
        f"- Valid contextual review: **{valid_context:,} ({valid_context * 100 / len(accepted_ids):.3f}%)**",
        f"  - legacy direct `full_read`: **{valid_full_read:,}**",
        f"  - dense `scope_full_read`: **{valid_scope:,}**",
        f"  - direct `manifest_edit`: **{valid_manifest:,}**",
        "  - propagated: **0** (not yet credited)",
        f"- `CONTEXT_STALE`: **{len(stale_ids):,}**",
        f"- `LAYOUT_RECHECK`: **{len(layout_recheck_ids):,}** (orthogonal; does not erase language coverage)",
        f"- `UNREVIEWED` for contextual purposes: **{unreviewed:,}**",
        f"- Approved registered manifest records (historical, non-deduplicated): **{total_manifest_records:,}**",
        f"- Pending batches lacking a registered review basis: **{', '.join(sorted(pending_batches)) if pending_batches else 'none'}**", "",
        "Language stale is ID-granular and uses `SHA256(JP + NUL + pinned EN + NUL + KO)`.",
        "Structural drift uses `SHA256(layout + NUL + physical_consumer_signature)` and sets `LAYOUT_RECHECK` only.",
        "For dense scopes, the ledger reconstructs each ID at `reviewed_commit`; one changed ID never invalidates the rest of its scope.", "",
        "## Orthogonal flags currently derived", "",
        f"- `PERSISTED_LAYOUT`: **{flag_counts['PERSISTED_LAYOUT']:,}** among ledger rows",
        f"- `ALIAS_GROUP`: **{flag_counts['ALIAS_GROUP']:,}** among ledger rows",
        f"- `SOURCE_ANOMALY`: **{flag_counts['SOURCE_ANOMALY']:,}** among ledger rows",
        f"- `LAYOUT_RECHECK`: **{flag_counts['LAYOUT_RECHECK']:,}** among ledger rows", "",
        "## Language propagation candidates (not coverage)", "",
        "Candidate signature is exact Japanese + exact pinned English + exact Korean. EN mismatch is an unconditional split.",
        f"- Unique language signatures: **{len(groups):,}**",
        f"- Duplicate groups: **{len(duplicate_groups):,}**",
        f"- IDs inside duplicate groups: **{sum(len(x) for x in duplicate_groups):,}**",
        f"- Potential extra IDs: **{sum(len(x)-1 for x in duplicate_groups):,}**", "",
        "## Dense review scopes", "",
    ]
    if dense_scope_meta:
        lines += ["| Scope | IDs | KEEP | EDIT | Excluded | File |", "| --- | ---: | ---: | ---: | ---: | --- |"]
        for s in dense_scope_meta:
            lines.append(f"| {s['scope_id']} | {s['ids']:,} | {s['keeps']:,} | {s['edits']:,} | {s['exclusions']:,} | `{s['path']}` |")
    else:
        lines.append("- none yet")
    lines += ["", "## Historical edit manifests", "", "| Batch | Status | Records | Unique IDs | Manifest |", "| --- | --- | ---: | ---: | --- |"]
    for row in per_batch:
        lines.append(f"| {row['batch']} | {row['status']} | {row['records']:,} | {row['unique_ids']:,} | `{row['manifest']}` |")
    lines += ["", "## Completion/quality rule", "",
              "Coverage and accuracy are separate. Final Beta1 uses a reproducible second-pass sample of valid KEEP rows.",
              "The second-pass reviewer/model must differ from first-pass review, and propagated KEEP must be sampled above its population share.",
              "A correction rate above 5% requires expanded re-review. Scanner-zero alone is never whole-corpus proof.", ""]
    (root / args.summary).write_text("\n".join(lines), encoding="utf-8")

    print(json.dumps({"status": "PASS", "schema_version": 3, "accepted": len(accepted_ids), "valid_context": valid_context,
                      "stale": len(stale_ids), "layout_recheck": len(layout_recheck_ids), "unreviewed": unreviewed,
                      "legacy_full_read": valid_full_read, "scope_full_read": valid_scope, "manifest_edit": valid_manifest,
                      "propagated": 0, "pending_review_basis_batches": sorted(pending_batches),
                      "duplicate_language_groups": len(duplicate_groups),
                      "propagation_candidate_extra_ids": sum(len(x)-1 for x in duplicate_groups),
                      "stale_ids": stale_ids}, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
