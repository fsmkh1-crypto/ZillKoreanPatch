#!/usr/bin/env python3
import argparse
import hashlib
import json
import re
import subprocess
import tomllib
from collections import defaultdict
from copy import deepcopy
from pathlib import Path

SECTION_RE = re.compile(r'^\["([0-9]+)"\]$')
CONTROL_RE = re.compile(r"<[^>]+>")


def toml_quote(value: str) -> str:
    return json.dumps(value, ensure_ascii=False)


def controls(text: str) -> list[str]:
    return CONTROL_RE.findall(text)


def manifest_digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def current_head(root: Path) -> str:
    return subprocess.check_output(["git", "-C", str(root), "rev-parse", "HEAD"], text=True).strip()


def is_ancestor(root: Path, ancestor: str, descendant: str) -> bool:
    proc = subprocess.run(
        ["git", "-C", str(root), "merge-base", "--is-ancestor", ancestor, descendant],
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )
    return proc.returncode == 0


def load_manifest(path: Path) -> dict:
    data = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(data.get("records"), list):
        raise ValueError("manifest must contain a records list")
    if not data.get("korean_file") and not all(
        isinstance(record, dict) and (record.get("path") or record.get("paths"))
        for record in data["records"]
        if isinstance(record, dict) and record.get("decision") == "PROPOSED_CHANGE"
    ):
        raise ValueError("manifest needs korean_file or per-record path/paths")
    return data


def validate_manifest_base(root: Path, manifest: dict, require_base_sha: bool) -> tuple[str, str]:
    base = str(manifest.get("korean_sha") or manifest.get("base_sha") or "")
    if not base:
        if require_base_sha:
            raise ValueError("manifest missing korean_sha/base_sha")
        # Legacy/tests without an audited base SHA do not need a Git repository.
        return "", ""
    head = current_head(root)
    if not is_ancestor(root, base, head):
        raise ValueError(f"manifest base {base} is not an ancestor of current HEAD {head}")
    return base, head


def load_overrides(path: Path | None) -> dict[str, dict]:
    if path is None:
        return {}
    data = json.loads(path.read_text(encoding="utf-8"))
    entries = data.get("overrides")
    if not isinstance(entries, list):
        raise ValueError("override file must contain an 'overrides' list")
    out = {}
    for entry in entries:
        rid = str(entry["id"])
        if rid in out:
            raise ValueError(f"duplicate override record id {rid}")
        required = {"reviewed_proposed_korean", "replacement_proposed_korean", "reason"}
        missing = required - entry.keys()
        if missing:
            raise ValueError(f"override {rid} missing required fields: {sorted(missing)}")
        out[rid] = entry
    return out


def apply_overrides(manifest: dict, overrides: dict[str, dict]) -> tuple[dict, int]:
    if not overrides:
        return manifest, 0
    effective = deepcopy(manifest)
    records = {str(record["id"]): record for record in effective["records"]}
    for rid, override in overrides.items():
        if rid not in records:
            raise ValueError(f"override {rid} does not exist in reviewed manifest")
        record = records[rid]
        if record.get("decision") != "PROPOSED_CHANGE":
            raise ValueError(f"override {rid} does not target a PROPOSED_CHANGE record")
        reviewed = override["reviewed_proposed_korean"]
        replacement = override["replacement_proposed_korean"]
        if record.get("proposed_korean") != reviewed:
            raise ValueError(
                f"override {rid} reviewed proposal mismatch: "
                f"manifest={record.get('proposed_korean')!r} override={reviewed!r}"
            )
        if reviewed == replacement:
            raise ValueError(f"override {rid} replacement equals reviewed proposal")
        if controls(reviewed) != controls(replacement):
            raise ValueError(
                f"override {rid} changes runtime control/substitution topology: "
                f"{controls(reviewed)!r} -> {controls(replacement)!r}"
            )
        record["proposed_korean"] = replacement
        record["override_reason"] = override["reason"]
    return effective, len(overrides)


def record_paths(manifest: dict, record: dict) -> list[str]:
    if isinstance(record.get("paths"), list) and record["paths"]:
        paths = [str(path) for path in record["paths"]]
    elif record.get("path"):
        paths = [str(record["path"])]
    elif manifest.get("korean_file"):
        paths = [str(manifest["korean_file"])]
    else:
        raise ValueError(f"record {record.get('id')} has no target path")
    if len(paths) != len(set(paths)):
        raise ValueError(f"record {record.get('id')} repeats target paths")
    return paths


def proposal_records(manifest: dict) -> list[dict]:
    out = []
    seen = set()
    for record in manifest["records"]:
        if record.get("decision") != "PROPOSED_CHANGE":
            continue
        rid = str(record["id"])
        paths = tuple(record_paths(manifest, record))
        key = (rid, paths)
        if key in seen:
            raise ValueError(f"duplicate proposed record id/path set {rid} {paths}")
        seen.add(key)
        before = record["before_korean"]
        after = record["proposed_korean"]
        if before == after:
            raise ValueError(f"record {rid} is marked changed but before == after")
        if controls(before) != controls(after):
            raise ValueError(
                f"record {rid} changes runtime control/substitution topology: "
                f"{controls(before)!r} -> {controls(after)!r}"
            )
        out.append(record)
    if not out:
        raise ValueError("manifest contains no PROPOSED_CHANGE records")
    return out


def read_overlay(path: Path) -> tuple[str, dict]:
    text = path.read_text(encoding="utf-8")
    data = tomllib.loads(text)
    return text, data


def validate_record_state(data: dict, record: dict, verify: bool, path: str) -> str:
    rid = str(record["id"])
    if rid not in data:
        raise ValueError(f"record {rid} missing from Korean overlay {path}")
    row = data[rid]
    expected_japanese = record.get("japanese")
    if expected_japanese is not None and row.get("japanese") != expected_japanese:
        raise ValueError(
            f"record {rid} Japanese/source reference changed in {path}: "
            f"expected {expected_japanese!r}, found {row.get('japanese')!r}"
        )
    current = row.get("korean")
    before = record["before_korean"]
    after = record["proposed_korean"]
    if verify:
        if current != after:
            raise ValueError(f"record {rid} in {path} is not at reviewed after-value")
        return "verified"
    if current == before:
        return "apply"
    if current == after:
        return "already-applied"
    raise ValueError(
        f"record {rid} in {path} no longer matches reviewed before/after values; refusing blind overwrite"
    )


def rewrite_korean_lines(text: str, replacements: dict[str, str]) -> str:
    lines = text.splitlines(keepends=True)
    current = None
    changed = set()
    for index, line in enumerate(lines):
        stripped = line.rstrip("\r\n")
        match = SECTION_RE.match(stripped)
        if match:
            current = match.group(1)
            continue
        if current in replacements and stripped.startswith('korean = "'):
            newline = "\r\n" if line.endswith("\r\n") else ("\n" if line.endswith("\n") else "")
            lines[index] = f"korean = {toml_quote(replacements[current])}{newline}"
            changed.add(current)
    missing = set(replacements) - changed
    if missing:
        raise ValueError(f"failed to locate Korean lines for IDs: {sorted(missing)}")
    rendered = "".join(lines)
    tomllib.loads(rendered)
    return rendered


def apply_manifest(
    root: Path,
    manifest_path: Path,
    verify: bool = False,
    overrides_path: Path | None = None,
    require_base_sha: bool = False,
) -> dict:
    manifest = load_manifest(manifest_path)
    base_sha, head_sha = validate_manifest_base(root, manifest, require_base_sha)
    overrides = load_overrides(overrides_path)
    manifest, override_count = apply_overrides(manifest, overrides)
    records = proposal_records(manifest)

    per_path_records: dict[str, list[dict]] = defaultdict(list)
    for record in records:
        for path in record_paths(manifest, record):
            per_path_records[path].append(record)

    loaded: dict[str, tuple[str, dict]] = {}
    states = {}
    replacements_by_path: dict[str, dict[str, str]] = defaultdict(dict)
    logical_changed = set()
    physical_apply_count = 0

    for rel_path in sorted(per_path_records):
        overlay_path = root / rel_path
        text, data = read_overlay(overlay_path)
        loaded[rel_path] = (text, data)
        seen_ids = set()
        for record in per_path_records[rel_path]:
            rid = str(record["id"])
            if rid in seen_ids:
                raise ValueError(f"duplicate target record {rid} within {rel_path}")
            seen_ids.add(rid)
            state = validate_record_state(data, record, verify, rel_path)
            states[f"{rel_path}:{rid}"] = state
            if state == "apply":
                replacements_by_path[rel_path][rid] = record["proposed_korean"]
                logical_changed.add(rid)
                physical_apply_count += 1

    if verify:
        return {
            "status": "VERIFIED_APPLIED",
            "manifest_sha256": manifest_digest(manifest_path),
            "manifest_base_sha": base_sha,
            "current_head_sha": head_sha,
            "proposed_records": len(records),
            "verified_physical_targets": len(states),
            "changed_records": 0,
            "override_records": override_count,
            "target_files": sorted(per_path_records),
        }

    rendered_by_path = {}
    for rel_path, replacements in replacements_by_path.items():
        if replacements:
            text, _ = loaded[rel_path]
            rendered_by_path[rel_path] = rewrite_korean_lines(text, replacements)

    for rel_path, rendered in rendered_by_path.items():
        (root / rel_path).write_text(rendered, encoding="utf-8", newline="")

    file_hashes = {}
    for rel_path in sorted(per_path_records):
        overlay_path = root / rel_path
        _, post = read_overlay(overlay_path)
        for record in per_path_records[rel_path]:
            rid = str(record["id"])
            if post[rid].get("korean") != record["proposed_korean"]:
                raise ValueError(f"record {rid} in {rel_path} failed post-write verification")
        file_hashes[rel_path] = hashlib.sha256(overlay_path.read_bytes()).hexdigest()

    return {
        "status": "APPLIED_OR_ALREADY_APPLIED",
        "manifest_sha256": manifest_digest(manifest_path),
        "manifest_base_sha": base_sha,
        "current_head_sha": head_sha,
        "proposed_records": len(records),
        "changed_records": len(logical_changed),
        "physical_targets_changed": physical_apply_count,
        "already_applied_physical_targets": len(states) - physical_apply_count,
        "override_records": override_count,
        "target_files": sorted(per_path_records),
        "target_file_sha256": file_hashes,
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", required=True, type=Path)
    parser.add_argument("--overrides", type=Path)
    parser.add_argument("--verify", action="store_true")
    parser.add_argument("--require-base-sha", action="store_true")
    parser.add_argument("--json", type=Path)
    args = parser.parse_args()

    root = Path(__file__).resolve().parents[2]
    manifest_path = args.manifest
    if not manifest_path.is_absolute():
        manifest_path = root / manifest_path
    overrides_path = args.overrides
    if overrides_path is not None and not overrides_path.is_absolute():
        overrides_path = root / overrides_path
    result = apply_manifest(
        root,
        manifest_path,
        verify=args.verify,
        overrides_path=overrides_path,
        require_base_sha=args.require_base_sha,
    )
    rendered = json.dumps(result, ensure_ascii=False, indent=2)
    print(rendered)
    if args.json:
        args.json.write_text(rendered + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
