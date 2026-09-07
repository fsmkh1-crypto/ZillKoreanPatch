#!/usr/bin/env python3
"""Materialize reviewed Beta1 ID + proposed Korean rows into the existing queue manifest.

This tool performs no translation decisions and grants no review coverage. It reads the
exact Korean before-values and overlay paths from a reviewed base commit, verifies
alias consistency and runtime control-token topology, then emits the multifile manifest
consumed by the existing Beta1 reviewed-copyedit queue.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import re
import subprocess
import tomllib
from pathlib import Path

PINNED_ENGLISH_SHA = "a98d9ce29f361d666ec23da0dcfd351f24537ffd"
CONTROL_RE = re.compile(r"<[^>]+>")
OVERLAY_PREFIX = "translations/korean/messages/"


def git(root: Path, *args: str) -> str:
    return subprocess.check_output(
        ["git", "-C", str(root), *args],
        text=True,
        encoding="utf-8",
    ).strip()


def canonical_commit(root: Path, ref: str) -> str:
    return git(root, "rev-parse", "--verify", f"{ref}^{{commit}}")


def require_ancestor(root: Path, base_sha: str, head_sha: str) -> None:
    proc = subprocess.run(
        ["git", "-C", str(root), "merge-base", "--is-ancestor", base_sha, head_sha],
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )
    if proc.returncode != 0:
        raise ValueError(f"base {base_sha} is not an ancestor of current HEAD {head_sha}")


def controls(text: str) -> list[str]:
    return CONTROL_RE.findall(text)


def load_json(path: Path) -> dict:
    data = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(data, dict):
        raise ValueError("proposal input must be a JSON object")
    return data


def proposal_records(doc: dict) -> list[dict]:
    records = doc.get("records")
    if not isinstance(records, list) or not records:
        raise ValueError("proposal input must contain a non-empty records list")
    seen: set[str] = set()
    out: list[dict] = []
    for pos, record in enumerate(records, start=1):
        if not isinstance(record, dict):
            raise ValueError(f"proposal record #{pos} is not an object")
        rid = str(record.get("id", ""))
        proposed = record.get("proposed_korean")
        if not rid or not rid.isdigit():
            raise ValueError(f"proposal record #{pos} has invalid numeric id: {rid!r}")
        if rid in seen:
            raise ValueError(f"duplicate proposal id {rid}")
        if not isinstance(proposed, str):
            raise ValueError(f"proposal id {rid} missing string proposed_korean")
        seen.add(rid)
        out.append(record)
    return out


def base_overlay_paths(root: Path, base_sha: str) -> list[str]:
    listing = git(
        root,
        "ls-tree",
        "-r",
        "--name-only",
        base_sha,
        "--",
        OVERLAY_PREFIX.rstrip("/"),
    )
    return sorted(
        line
        for line in listing.splitlines()
        if line.startswith(OVERLAY_PREFIX) and line.endswith(".toml")
    )


def read_git_text(root: Path, base_sha: str, rel_path: str) -> str:
    return subprocess.check_output(
        ["git", "-C", str(root), "show", f"{base_sha}:{rel_path}"],
        text=True,
        encoding="utf-8",
    )


def build_index(root: Path, base_sha: str) -> dict[str, dict]:
    index: dict[str, dict] = {}
    for rel_path in base_overlay_paths(root, base_sha):
        data = tomllib.loads(read_git_text(root, base_sha, rel_path))
        for raw_id, row in data.items():
            rid = str(raw_id)
            if not isinstance(row, dict) or not isinstance(row.get("korean"), str):
                continue
            japanese = row.get("japanese", "")
            korean = row["korean"]
            prev = index.get(rid)
            if prev is None:
                index[rid] = {
                    "japanese": japanese,
                    "before_korean": korean,
                    "paths": [rel_path],
                }
                continue
            if (prev["japanese"], prev["before_korean"]) != (japanese, korean):
                raise ValueError(
                    f"alias drift at base {base_sha}: id={rid} "
                    f"existing={prev['paths']} new={rel_path}"
                )
            prev["paths"].append(rel_path)
    for row in index.values():
        row["paths"].sort()
    return index


def build_manifest(proposals: dict, index: dict[str, dict], base_sha: str) -> dict:
    supplied_english = proposals.get("english_reference_sha", PINNED_ENGLISH_SHA)
    if supplied_english != PINNED_ENGLISH_SHA:
        raise ValueError(
            f"English reference mismatch: expected {PINNED_ENGLISH_SHA}, "
            f"found {supplied_english}"
        )
    supplied_source = proposals.get("source_head")
    if supplied_source is not None and supplied_source != base_sha:
        raise ValueError(
            f"proposal source_head {supplied_source} does not match base {base_sha}"
        )

    materialized: list[dict] = []
    for record in proposal_records(proposals):
        rid = str(record["id"])
        source = index.get(rid)
        if source is None:
            raise ValueError(f"proposal id {rid} not found in Korean overlays at {base_sha}")

        expected_japanese = record.get("expected_japanese")
        if expected_japanese is not None and expected_japanese != source["japanese"]:
            raise ValueError(f"proposal id {rid} Japanese source changed")

        expected_before = record.get("expected_before_korean")
        if expected_before is not None and expected_before != source["before_korean"]:
            raise ValueError(f"proposal id {rid} Korean before-value changed")

        before = source["before_korean"]
        proposed = record["proposed_korean"]
        if before == proposed:
            raise ValueError(f"proposal id {rid} before == proposed")
        if controls(before) != controls(proposed):
            raise ValueError(
                f"proposal id {rid} changes runtime control/substitution topology: "
                f"{controls(before)!r} -> {controls(proposed)!r}"
            )

        materialized.append(
            {
                "id": rid,
                "paths": source["paths"],
                "before_korean": before,
                "proposed_korean": proposed,
            }
        )

    return {
        "version": 1,
        "purpose": proposals.get(
            "purpose",
            "Reviewed Beta1 contextual copyedit materialized from ID + proposed Korean",
        ),
        "source_head": base_sha,
        "english_reference_sha": PINNED_ENGLISH_SHA,
        "records": materialized,
    }


def render_manifest(manifest: dict) -> str:
    return json.dumps(manifest, ensure_ascii=False, separators=(",", ":")) + "\n"


def main() -> int:
    parser = argparse.ArgumentParser(
        description=(
            "Materialize a reviewed Beta1 ID + proposed_korean list into the "
            "fail-closed multifile copyedit manifest consumed by the existing queue."
        )
    )
    parser.add_argument("--proposals", required=True, type=Path)
    parser.add_argument("--output", required=True, type=Path)
    parser.add_argument("--base-sha", help="Reviewed corpus commit; defaults to current HEAD")
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[2])
    parser.add_argument(
        "--overwrite",
        action="store_true",
        help="Allow replacing an existing output file",
    )
    args = parser.parse_args()

    root = args.root.resolve()
    proposals_path = args.proposals if args.proposals.is_absolute() else root / args.proposals
    output_path = args.output if args.output.is_absolute() else root / args.output

    head_sha = canonical_commit(root, "HEAD")
    base_sha = canonical_commit(root, args.base_sha or head_sha)
    require_ancestor(root, base_sha, head_sha)

    proposals = load_json(proposals_path)
    index = build_index(root, base_sha)
    manifest = build_manifest(proposals, index, base_sha)
    rendered = render_manifest(manifest)

    if output_path.exists() and not args.overwrite:
        raise ValueError(f"output already exists: {output_path}; pass --overwrite to replace")
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(rendered, encoding="utf-8")

    result = {
        "status": "PASS",
        "source_head": base_sha,
        "record_count": len(manifest["records"]),
        "target_files": sorted(
            {path for record in manifest["records"] for path in record["paths"]}
        ),
        "output": str(output_path.relative_to(root))
        if output_path.is_relative_to(root)
        else str(output_path),
        "manifest_sha256": hashlib.sha256(rendered.encode("utf-8")).hexdigest(),
    }
    print(json.dumps(result, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
