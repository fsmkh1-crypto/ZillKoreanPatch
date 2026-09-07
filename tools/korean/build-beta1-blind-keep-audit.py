#!/usr/bin/env python3
"""Build a deterministic blind second-pass KEEP audit packet.

The public packet hides real IDs, batch/scope labels, and source paths. Selection
and anonymization are deterministic so the original cohort can be reconstructed
later without publishing a mapping file to the reviewer.
"""
from __future__ import annotations

import argparse
import csv
import hashlib
import io
import json
import subprocess
import tomllib
from collections import defaultdict
from pathlib import Path

PINNED_ENGLISH_SHA = "a98d9ce29f361d666ec23da0dcfd351f24537ffd"
ALLOWED_ERROR_CLASSES = (
    "semantic",
    "speaker_register",
    "naturalness",
    "punctuation_spacing",
)


def git(root: Path, *args: str) -> str:
    return subprocess.check_output(
        ["git", "-C", str(root), *args], text=True, encoding="utf-8"
    ).strip()


def git_show(root: Path, commit: str, rel: str) -> str:
    return subprocess.check_output(
        ["git", "-C", str(root), "show", f"{commit}:{rel}"],
        text=True,
        encoding="utf-8",
    )


def digest_key(seed: str, namespace: str, rid: str, basis: str = "") -> bytes:
    payload = f"{seed}\0{namespace}\0{rid}\0{basis}".encode("utf-8")
    return hashlib.sha256(payload).digest()


def q(text: str) -> str:
    return json.dumps(text, ensure_ascii=False, separators=(",", ":"))


def load_config(path: Path) -> dict:
    doc = json.loads(path.read_text(encoding="utf-8"))
    required = {
        "audit_id",
        "korean_head",
        "english_sha",
        "seed",
        "context_radius",
        "dense_control_count",
        "legacy_control_count",
    }
    missing = sorted(required - set(doc))
    if missing:
        raise SystemExit(f"audit config missing keys: {missing}")
    if doc["english_sha"] != PINNED_ENGLISH_SHA:
        raise SystemExit("audit config English SHA is not the pinned reference")
    if int(doc["context_radius"]) < 0:
        raise SystemExit("context_radius must be non-negative")
    return doc


def load_english(root: Path) -> dict[str, dict[str, str]]:
    if git(root, "rev-parse", "HEAD") != PINNED_ENGLISH_SHA:
        raise SystemExit("English checkout is not at pinned SHA")
    out: dict[str, dict[str, str]] = {}
    for path in sorted((root / "translations/messages").glob("msgsec*.toml")):
        data = tomllib.loads(path.read_text(encoding="utf-8"))
        for raw_id, row in data.items():
            if not isinstance(row, dict):
                continue
            rid = str(raw_id)
            if rid in out:
                raise SystemExit(f"duplicate English id {rid}")
            out[rid] = {
                "japanese": str(row.get("japanese", "")),
                "english": str(row.get("english", "")),
            }
    return out


def load_ledger(root: Path, korean_head: str) -> list[dict]:
    text = git_show(root, korean_head, "translations/korean/review-ledger.jsonl")
    rows = []
    seen = set()
    for line_no, line in enumerate(text.splitlines(), 1):
        if not line.strip():
            continue
        row = json.loads(line)
        rid = str(row["id"])
        if rid in seen:
            raise SystemExit(f"duplicate ledger id {rid} at line {line_no}")
        seen.add(rid)
        row["id"] = rid
        rows.append(row)
    return rows


def select_rows(ledger: list[dict], cfg: dict) -> tuple[list[dict], dict[str, list[dict]]]:
    keeps = [r for r in ledger if r.get("context_state") == "CONTEXT_KEEP"]
    measured = [
        r for r in keeps
        if r.get("evidence") == "scope_full_read" and str(r.get("batch_id")) == "022"
    ]
    dense = [
        r for r in keeps
        if r.get("evidence") == "scope_full_read"
        and str(r.get("batch_id")) in {"018", "019", "020", "021"}
    ]
    legacy = [r for r in keeps if r.get("evidence") == "full_read"]

    expected_measured = cfg.get("expected_measured_count")
    if expected_measured is not None and len(measured) != int(expected_measured):
        raise SystemExit(
            f"measured cohort drift: expected={expected_measured} actual={len(measured)}"
        )
    if len(dense) < int(cfg["dense_control_count"]):
        raise SystemExit("dense control pool is smaller than requested sample")
    if len(legacy) < int(cfg["legacy_control_count"]):
        raise SystemExit("legacy control pool is smaller than requested sample")

    seed = str(cfg["seed"])
    dense = sorted(
        dense,
        key=lambda r: digest_key(
            seed, "dense-control", r["id"], str(r.get("language_basis_sha256", ""))
        ),
    )[: int(cfg["dense_control_count"])]
    legacy = sorted(
        legacy,
        key=lambda r: digest_key(
            seed, "legacy-control", r["id"], str(r.get("language_basis_sha256", ""))
        ),
    )[: int(cfg["legacy_control_count"])]

    cohorts = {
        "measured_022": measured,
        "dense_018_021": dense,
        "legacy_001_017": legacy,
    }
    selected = [*measured, *dense, *legacy]
    if len({r["id"] for r in selected}) != len(selected):
        raise SystemExit("selected audit rows overlap across cohorts")

    selected = sorted(
        selected,
        key=lambda r: digest_key(
            seed, "blind-order", r["id"], str(r.get("language_basis_sha256", ""))
        ),
    )
    return selected, cohorts


def load_korean_files(root: Path, korean_head: str) -> tuple[dict[str, list[tuple[str, dict]]], dict[str, list[str]]]:
    files = [
        rel
        for rel in git(
            root,
            "ls-tree",
            "-r",
            "--name-only",
            korean_head,
            "--",
            "translations/korean/messages",
        ).splitlines()
        if rel.endswith(".toml")
    ]
    file_rows: dict[str, list[tuple[str, dict]]] = {}
    id_paths: dict[str, list[str]] = defaultdict(list)
    for rel in sorted(files):
        data = tomllib.loads(git_show(root, korean_head, rel))
        ordered: list[tuple[str, dict]] = []
        for raw_id, row in data.items():
            if not isinstance(row, dict) or not isinstance(row.get("korean"), str):
                continue
            rid = str(raw_id)
            ordered.append((rid, row))
            id_paths[rid].append(rel)
        file_rows[rel] = ordered
    return file_rows, id_paths


def canonical_occurrence(
    rid: str,
    file_rows: dict[str, list[tuple[str, dict]]],
    id_paths: dict[str, list[str]],
) -> tuple[str, int, dict]:
    paths = sorted(id_paths.get(rid, []))
    if not paths:
        raise SystemExit(f"selected id {rid} missing from Korean tree")
    canonical = None
    chosen: tuple[str, int, dict] | None = None
    for rel in paths:
        rows = file_rows[rel]
        hits = [(idx, row) for idx, (row_id, row) in enumerate(rows) if row_id == rid]
        if len(hits) != 1:
            raise SystemExit(f"unexpected occurrence count id={rid} path={rel} count={len(hits)}")
        idx, row = hits[0]
        signature = (
            str(row.get("japanese", "")),
            str(row.get("korean", "")),
            str(row.get("layout", "")),
            str(row.get("consumer", "")),
        )
        if canonical is None:
            canonical = signature
            chosen = (rel, idx, row)
        elif signature != canonical:
            raise SystemExit(f"alias drift for selected id {rid}")
    assert chosen is not None
    return chosen


def render_context_row(label: str, rid: str, row: dict, eng: dict[str, dict[str, str]]) -> list[str]:
    erow = eng.get(rid, {"english": ""})
    return [
        f"**{label}**",
        f"- JP: {q(str(row.get('japanese', '')))}",
        f"- EN: {q(str(erow.get('english', '')))}",
        f"- KO: {q(str(row.get('korean', '')))}",
    ]


def build_public_packet(
    selected: list[dict],
    cfg: dict,
    eng: dict[str, dict[str, str]],
    file_rows: dict[str, list[tuple[str, dict]]],
    id_paths: dict[str, list[str]],
) -> tuple[str, list[dict]]:
    radius = int(cfg["context_radius"])
    lines = [
        "# Blind KEEP Accuracy Audit Packet",
        "",
        "This packet intentionally hides original IDs, batch/scope labels, and source paths.",
        "Review only the TARGET row in each item. Nearby rows exist only to establish scene, speaker/register, and local wording context.",
        "Japanese is the semantic source. Pinned English is reference evidence for how the established patch handled meaning/structure; do not translate English mechanically.",
        "",
        f"Audit items: {len(selected)}",
        f"Pinned English SHA: `{PINNED_ENGLISH_SHA}`",
        "",
    ]
    mapping: list[dict] = []
    for ordinal, ledger_row in enumerate(selected, 1):
        anon = f"A{ordinal:03d}"
        rid = ledger_row["id"]
        rel, idx, target = canonical_occurrence(rid, file_rows, id_paths)
        ordered = file_rows[rel]
        start = max(0, idx - radius)
        stop = min(len(ordered), idx + radius + 1)
        lines += [f"## {anon}", ""]
        for pos in range(start, stop):
            ctx_id, ctx_row = ordered[pos]
            if pos == idx:
                lines += render_context_row("TARGET", ctx_id, ctx_row, eng)
            elif pos < idx:
                lines += render_context_row(f"CONTEXT BEFORE {idx - pos}", ctx_id, ctx_row, eng)
            else:
                lines += render_context_row(f"CONTEXT AFTER {pos - idx}", ctx_id, ctx_row, eng)
            lines.append("")
        lines.append("---")
        lines.append("")
        mapping.append(
            {
                "anonymous_id": anon,
                "real_id": rid,
                "batch_id": ledger_row.get("batch_id"),
                "evidence": ledger_row.get("evidence"),
                "language_basis_sha256": ledger_row.get("language_basis_sha256"),
                "source_path": rel,
            }
        )
    return "\n".join(lines).rstrip() + "\n", mapping


def reviewer_instructions(audit_id: str) -> str:
    classes = ", ".join(f"`{x}`" for x in ALLOWED_ERROR_CLASSES)
    return f"""# Blind KEEP audit reviewer instructions\n\nAudit ID: `{audit_id}`\n\nYou are performing an independent second-pass accuracy audit of previously accepted Korean KEEP rows. Do not search the repository for original IDs, batch membership, prior decisions, or suspected weak areas. Use only the supplied packet.\n\nFor every anonymous item:\n\n1. Review only the row marked **TARGET**. Nearby rows are context only.\n2. Treat Japanese as semantic authority.\n3. Use pinned English as evidence for meaning, structure, and the established patch's solution, but do not translate English mechanically.\n4. Return `KEEP` only if the current Korean target needs no correction.\n5. Return `SHOULD_EDIT` if you would change the target for any material reason.\n6. If `SHOULD_EDIT`, assign one or more error classes from: {classes}.\n7. If `SHOULD_EDIT`, provide the full corrected Korean target string, preserving required control tags exactly.\n8. Give a short reason tied to Japanese meaning/context and, where useful, the pinned English handling.\n\nError-class definitions:\n\n- `semantic`: mistranslation, wrong subject/object, polarity, fact, nuance, omission/addition, or other meaning error.\n- `speaker_register`: speaker voice, politeness level, relationship-dependent address, honorific/register, or character-consistency error.\n- `naturalness`: meaning is essentially correct but the Korean is materially awkward, calqued, or unnatural enough to warrant an edit.\n- `punctuation_spacing`: punctuation, spacing, or typography-only correction.\n\nDo not apply any pass/fail threshold yourself. Judge each item independently. Fill the provided CSV template without reordering or deleting rows.\n"""


def results_template(mapping: list[dict]) -> str:
    buf = io.StringIO(newline="")
    writer = csv.writer(buf, lineterminator="\n")
    writer.writerow(
        [
            "anonymous_id",
            "verdict",
            "error_classes",
            "proposed_korean",
            "reason",
        ]
    )
    for row in mapping:
        writer.writerow([row["anonymous_id"], "", "", "", ""])
    return buf.getvalue()


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--root", type=Path, default=Path("."))
    ap.add_argument("--english-root", type=Path, required=True)
    ap.add_argument("--config", type=Path, required=True)
    ap.add_argument("--output-dir", type=Path, required=True)
    ap.add_argument("--mapping-output", type=Path)
    args = ap.parse_args()

    root = args.root.resolve()
    eroot = args.english_root.resolve()
    cfg_path = args.config if args.config.is_absolute() else root / args.config
    cfg = load_config(cfg_path)
    korean_head = str(cfg["korean_head"])
    git(root, "cat-file", "-e", f"{korean_head}^{{commit}}")

    ledger = load_ledger(root, korean_head)
    selected, cohorts = select_rows(ledger, cfg)
    eng = load_english(eroot)
    file_rows, id_paths = load_korean_files(root, korean_head)
    packet, mapping = build_public_packet(selected, cfg, eng, file_rows, id_paths)

    cohort_by_real = {}
    for cohort, rows in cohorts.items():
        for row in rows:
            cohort_by_real[row["id"]] = cohort
    for row in mapping:
        row["cohort"] = cohort_by_real[row["real_id"]]

    out_dir = args.output_dir if args.output_dir.is_absolute() else root / args.output_dir
    out_dir.mkdir(parents=True, exist_ok=True)
    (out_dir / "packet.md").write_text(packet, encoding="utf-8", newline="\n")
    (out_dir / "reviewer-instructions.md").write_text(
        reviewer_instructions(str(cfg["audit_id"])), encoding="utf-8", newline="\n"
    )
    (out_dir / "results-template.csv").write_text(
        results_template(mapping), encoding="utf-8", newline="\n"
    )
    packet_sha = hashlib.sha256(packet.encode("utf-8")).hexdigest()
    public_manifest = {
        "schema_version": 1,
        "audit_id": cfg["audit_id"],
        "korean_head": korean_head,
        "english_sha": PINNED_ENGLISH_SHA,
        "item_count": len(mapping),
        "context_radius": int(cfg["context_radius"]),
        "packet_sha256": packet_sha,
        "blinding": {
            "original_ids_hidden": True,
            "batch_scope_hidden": True,
            "source_paths_hidden": True,
            "reviewer_should_not_search_repository": True,
        },
    }
    (out_dir / "manifest.json").write_text(
        json.dumps(public_manifest, ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
        newline="\n",
    )

    if args.mapping_output:
        mapping_path = args.mapping_output if args.mapping_output.is_absolute() else root / args.mapping_output
        mapping_path.parent.mkdir(parents=True, exist_ok=True)
        private_doc = {
            "schema_version": 1,
            "audit_id": cfg["audit_id"],
            "korean_head": korean_head,
            "english_sha": PINNED_ENGLISH_SHA,
            "seed": cfg["seed"],
            "cohort_counts": {name: len(rows) for name, rows in cohorts.items()},
            "rows": mapping,
        }
        mapping_path.write_text(
            json.dumps(private_doc, ensure_ascii=False, indent=2) + "\n",
            encoding="utf-8",
            newline="\n",
        )

    print(
        json.dumps(
            {
                "status": "PASS",
                "audit_id": cfg["audit_id"],
                "korean_head": korean_head,
                "english_sha": PINNED_ENGLISH_SHA,
                "cohort_counts": {name: len(rows) for name, rows in cohorts.items()},
                "item_count": len(mapping),
                "packet_sha256": packet_sha,
            },
            ensure_ascii=False,
            indent=2,
        )
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
