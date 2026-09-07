#!/usr/bin/env python3
"""Create or verify deterministic Beta1 dense-review packets.

The packet is the exact byte stream a reviewer is expected to read. Its SHA-256
is therefore meaningful only if CI can regenerate the same bytes from
(reviewed_commit, pinned English SHA, ID set). This tool provides both paths.
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

CONTROL = re.compile(r"<[^>]+>")
PIN = "a98d9ce29f361d666ec23da0dcfd351f24537ffd"


def git(root: Path, *args: str) -> str:
    return subprocess.check_output(["git", "-C", str(root), *args], text=True, encoding="utf-8").strip()


def git_show(root: Path, commit: str, rel: str) -> str:
    return subprocess.check_output(["git", "-C", str(root), "show", f"{commit}:{rel}"], text=True, encoding="utf-8")


def sort_id(rid: str):
    return (0, int(rid)) if rid.isdigit() else (1, rid)


def english_rows(root: Path) -> dict[str, dict]:
    if git(root, "rev-parse", "HEAD") != PIN:
        raise SystemExit("English checkout is not pinned reference")
    out = {}
    for path in sorted((root / "translations/messages").glob("msgsec*.toml")):
        data = tomllib.loads(path.read_text(encoding="utf-8"))
        for rid, row in data.items():
            if isinstance(row, dict):
                rid = str(rid)
                if rid in out:
                    raise SystemExit(f"duplicate English id {rid}")
                out[rid] = {"japanese": row.get("japanese", ""), "english": row.get("english", "")}
    return out


def korean_at(root: Path, commit: str, ids: list[str], preferred_files: list[str]) -> dict[str, dict]:
    wanted = set(ids)
    found: dict[str, list[dict]] = defaultdict(list)
    files = list(preferred_files)
    if not files:
        files = [line for line in git(root, "ls-tree", "-r", "--name-only", commit, "--", "translations/korean/messages").splitlines() if line.endswith(".toml")]
    for rel in sorted(set(files)):
        try:
            data = tomllib.loads(git_show(root, commit, rel))
        except subprocess.CalledProcessError:
            continue
        for rid in wanted:
            row = data.get(rid)
            if isinstance(row, dict) and isinstance(row.get("korean"), str):
                found[rid].append({"path": rel, **row})
    missing = wanted - set(found)
    if missing and preferred_files:
        # A scope may contain IDs whose file moved since review. Fall back to the
        # full tree at reviewed_commit rather than silently using current paths.
        return korean_at(root, commit, ids, [])
    if missing:
        raise SystemExit(f"review packet IDs missing at {commit}: {sorted(missing, key=sort_id)[:20]}")
    out = {}
    for rid, rows in found.items():
        first = rows[0]
        for row in rows[1:]:
            for key in ("japanese", "korean", "layout", "consumer"):
                if row.get(key, "") != first.get(key, ""):
                    raise SystemExit(f"alias drift id={rid} key={key}")
        out[rid] = {
            "paths": sorted(r["path"] for r in rows), "japanese": first.get("japanese", ""),
            "korean": first.get("korean", ""), "layout": first.get("layout", ""), "consumer": first.get("consumer", ""),
        }
    return out


def q(s: str) -> str:
    # JSON quoting gives one deterministic, lossless display line while retaining
    # spaces and control tags exactly.
    return json.dumps(s, ensure_ascii=False, separators=(",", ":"))


def render(commit: str, english_sha: str, ids: list[str], rows: dict[str, dict], eng: dict[str, dict]) -> bytes:
    lines = [
        "# Beta1 Deterministic Context Review Packet",
        f"reviewed_commit: {commit}",
        f"english_sha: {english_sha}",
        f"id_count: {len(ids)}",
        "order: numeric-id",
        "encoding: UTF-8 LF",
        "",
    ]
    for rid in sorted(ids, key=sort_id):
        row = rows[rid]
        erow = eng.get(rid, {"japanese": "", "english": ""})
        # Korean source JP is authoritative for stale reconstruction. A mismatch
        # against pinned-English JP is exposed, never silently reconciled.
        jp_match = row["japanese"] == erow.get("japanese", "")
        controls = CONTROL.findall(row["korean"])
        lines += [
            f"## {rid}",
            f"paths: {json.dumps(row['paths'], ensure_ascii=False, separators=(',', ':'))}",
            f"jp_matches_pinned_english: {'yes' if jp_match else 'NO'}",
            f"persisted_layout: {'yes' if bool(row['layout']) else 'no'}",
            f"controls: {json.dumps(controls, ensure_ascii=False, separators=(',', ':'))}",
            f"JP: {q(row['japanese'])}",
            f"EN: {q(erow.get('english', ''))}",
            f"KO: {q(row['korean'])}",
            "",
        ]
    return ("\n".join(lines) + "\n").encode("utf-8")


def scope_inputs(scope: dict) -> tuple[str, list[str], list[str]]:
    ids = [str(x) for x in scope["ids"]]
    if len(ids) != len(set(ids)):
        raise SystemExit("scope contains duplicate IDs")
    files = [str(x) for x in scope.get("source_files", [])]
    return str(scope["reviewed_commit"]), ids, files


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--root", default=".", type=Path)
    ap.add_argument("--english-root", required=True, type=Path)
    ap.add_argument("--scope", type=Path)
    ap.add_argument("--reviewed-commit")
    ap.add_argument("--source-file", action="append", default=[])
    ap.add_argument("--ids", help="comma-separated IDs; omit with --source-file to use every Korean row in those files")
    ap.add_argument("--output", type=Path)
    ap.add_argument("--verify", action="store_true")
    args = ap.parse_args()

    root = args.root.resolve(); eroot = args.english_root.resolve(); eng = english_rows(eroot)
    expected_hash = None
    if args.scope:
        scope_path = args.scope if args.scope.is_absolute() else root / args.scope
        scope = json.loads(scope_path.read_text(encoding="utf-8"))
        if scope.get("english_sha") != PIN:
            raise SystemExit("scope English SHA drift")
        commit, ids, files = scope_inputs(scope)
        expected_hash = scope.get("packet_sha256")
    else:
        commit = args.reviewed_commit or git(root, "rev-parse", "HEAD")
        files = list(args.source_file)
        if args.ids:
            ids = [x.strip() for x in args.ids.split(",") if x.strip()]
        elif files:
            ids = []
            for rel in files:
                data = tomllib.loads(git_show(root, commit, rel))
                ids.extend(str(rid) for rid, row in data.items() if isinstance(row, dict) and isinstance(row.get("korean"), str))
            ids = sorted(set(ids), key=sort_id)
        else:
            raise SystemExit("provide --scope, --ids, or --source-file")

    rows = korean_at(root, commit, ids, files)
    packet = render(commit, PIN, ids, rows, eng)
    digest = hashlib.sha256(packet).hexdigest()

    if args.verify:
        if not expected_hash:
            raise SystemExit("--verify requires scope packet_sha256")
        if digest != expected_hash:
            raise SystemExit(f"packet hash mismatch expected={expected_hash} actual={digest}")
    if args.output:
        out = args.output if args.output.is_absolute() else root / args.output
        out.parent.mkdir(parents=True, exist_ok=True)
        out.write_bytes(packet)
    print(json.dumps({"status": "PASS", "reviewed_commit": commit, "english_sha": PIN, "id_count": len(ids), "packet_sha256": digest}, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
