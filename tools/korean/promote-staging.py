#!/usr/bin/env python3
"""Build a recovery-pointer promotion packet from staged Korean result files.

This tool is intentionally mechanical. It resolves the latest commit touching
an existing staged result path and emits ``recover_from`` rows. Canonical IDs,
Korean text, fixed tokens, and duplicate translation semantics remain owned by
``apply-results.py`` so validation cannot drift between two implementations.
"""
from __future__ import annotations

import argparse
import json
from pathlib import Path
import re
import subprocess
import sys


ROOT = Path(__file__).resolve().parents[2]
DEFAULT_REF = "origin/korean-manual-staging"
RESULT_ROOT = ("work", "korean-results")
FULL_COMMIT_RE = re.compile(r"^[0-9a-fA-F]{40}$")


def validate_staging_path(raw: str) -> str:
    path = Path(raw)
    if path.is_absolute() or ".." in path.parts:
        raise SystemExit(f"unsafe staging path: {raw}")
    if len(path.parts) < 3 or path.parts[:2] != RESULT_ROOT:
        raise SystemExit(f"staging path must be under work/korean-results/: {raw}")
    if path.suffix != ".jsonl":
        raise SystemExit(f"staging source must be a .jsonl result packet: {raw}")
    return path.as_posix()


def run_git(args: list[str]) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["git", *args],
        cwd=ROOT,
        text=True,
        capture_output=True,
        check=False,
    )


def resolve_latest_source_commit(ref: str, path: str) -> str:
    exists = run_git(["cat-file", "-e", f"{ref}:{path}"])
    if exists.returncode != 0:
        detail = exists.stderr.strip() or exists.stdout.strip()
        raise SystemExit(f"staging source does not exist at {ref}:{path}: {detail}")

    proc = run_git(["log", "-1", "--format=%H", ref, "--", path])
    commit = proc.stdout.strip()
    if proc.returncode != 0 or not FULL_COMMIT_RE.fullmatch(commit):
        detail = proc.stderr.strip() or proc.stdout.strip()
        raise SystemExit(f"cannot resolve latest staging commit for {ref}:{path}: {detail}")

    # Guard against a surprising history/path mismatch before emitting a pointer.
    source = run_git(["cat-file", "-e", f"{commit}:{path}"])
    if source.returncode != 0:
        detail = source.stderr.strip() or source.stdout.strip()
        raise SystemExit(f"resolved commit does not contain staging source {commit}:{path}: {detail}")
    return commit.lower()


def build_rows(ref: str, paths: list[str]) -> list[dict[str, dict[str, str]]]:
    rows: list[dict[str, dict[str, str]]] = []
    seen: set[str] = set()
    for raw in paths:
        path = validate_staging_path(raw)
        if path in seen:
            continue
        seen.add(path)
        commit = resolve_latest_source_commit(ref, path)
        rows.append({"recover_from": {"commit": commit, "path": path}})
    return rows


def render(rows: list[dict[str, dict[str, str]]]) -> str:
    return "".join(json.dumps(row, ensure_ascii=False, separators=(",", ":")) + "\n" for row in rows)


def main() -> None:
    ap = argparse.ArgumentParser(
        description="Resolve current staging files to immutable recover_from pointers."
    )
    ap.add_argument("paths", nargs="+", help="staged JSONL paths under work/korean-results/")
    ap.add_argument("--ref", default=DEFAULT_REF, help=f"staging ref (default: {DEFAULT_REF})")
    ap.add_argument("--out", type=Path, help="write packet here instead of stdout")
    args = ap.parse_args()

    rows = build_rows(args.ref, args.paths)
    text = render(rows)
    if args.out:
        args.out.parent.mkdir(parents=True, exist_ok=True)
        args.out.write_text(text, encoding="utf-8")
        print(f"wrote {len(rows)} recovery pointers to {args.out}", file=sys.stderr)
    else:
        sys.stdout.write(text)


if __name__ == "__main__":
    main()
