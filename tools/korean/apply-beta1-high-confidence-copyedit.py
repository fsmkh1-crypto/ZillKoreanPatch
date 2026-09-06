#!/usr/bin/env python3
import argparse
import json
import re
import tomllib
from collections import defaultdict
from pathlib import Path

SECTION_RE = re.compile(r'^\["([0-9]+)"\]$')
CONTROL_RE = re.compile(r"<[^>]+>")


def toml_quote(value: str) -> str:
    return json.dumps(value, ensure_ascii=False)


def controls(text: str) -> list[str]:
    return CONTROL_RE.findall(text)


def load_manifest(path: Path) -> dict:
    data = json.loads(path.read_text(encoding="utf-8"))
    records = data.get("records")
    if not isinstance(records, list) or not records:
        raise ValueError("manifest must contain a non-empty records list")
    if int(data.get("record_count", -1)) != len(records):
        raise ValueError("manifest record_count does not match records length")
    return data


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


def apply_manifest(root: Path, manifest_path: Path, verify: bool) -> dict:
    manifest = load_manifest(manifest_path)
    by_file: dict[str, list[dict]] = defaultdict(list)
    seen = set()
    for record in manifest["records"]:
        required = {"path", "id", "japanese", "before", "after", "kinds"}
        missing = required - record.keys()
        if missing:
            raise ValueError(f"record missing fields: {sorted(missing)}")
        key = (record["path"], str(record["id"]))
        if key in seen:
            raise ValueError(f"duplicate manifest target: {key}")
        seen.add(key)
        if record["before"] == record["after"]:
            raise ValueError(f"record {key} before == after")
        if controls(record["before"]) != controls(record["after"]):
            raise ValueError(f"record {key} changes runtime control topology")
        by_file[record["path"]].append(record)

    changed_records = 0
    already_applied = 0
    changed_files = []
    for relpath, records in sorted(by_file.items()):
        path = root / relpath
        text = path.read_text(encoding="utf-8")
        data = tomllib.loads(text)
        replacements = {}
        for record in records:
            rid = str(record["id"])
            if rid not in data:
                raise ValueError(f"{relpath}:{rid} missing")
            row = data[rid]
            if row.get("japanese") != record["japanese"]:
                raise ValueError(f"{relpath}:{rid} Japanese/source changed")
            current = row.get("korean")
            if verify:
                if current != record["after"]:
                    raise ValueError(f"{relpath}:{rid} is not at approved after-value")
                continue
            if current == record["before"]:
                replacements[rid] = record["after"]
                changed_records += 1
            elif current == record["after"]:
                already_applied += 1
            else:
                raise ValueError(f"{relpath}:{rid} no longer matches exact before/after; refusing overwrite")
        if replacements:
            rendered = rewrite_korean_lines(text, replacements)
            path.write_text(rendered, encoding="utf-8", newline="")
            changed_files.append(relpath)
            post = tomllib.loads(rendered)
            for rid, expected in replacements.items():
                if post[rid].get("korean") != expected:
                    raise ValueError(f"{relpath}:{rid} failed post-write verification")

    if verify:
        status = "VERIFIED_APPLIED"
        changed_records = 0
        already_applied = len(manifest["records"])
    else:
        status = "APPLIED_OR_ALREADY_APPLIED"

    return {
        "status": status,
        "source_scan_head": manifest.get("source_scan_head"),
        "finding_count": manifest.get("finding_count"),
        "record_count": len(manifest["records"]),
        "target_file_count": len(by_file),
        "changed_file_count": len(changed_files),
        "changed_files": changed_files,
        "changed_records": changed_records,
        "already_applied_records": already_applied,
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", required=True, type=Path)
    parser.add_argument("--verify", action="store_true")
    parser.add_argument("--json", type=Path)
    args = parser.parse_args()
    root = Path(__file__).resolve().parents[2]
    manifest = args.manifest if args.manifest.is_absolute() else root / args.manifest
    result = apply_manifest(root, manifest, args.verify)
    rendered = json.dumps(result, ensure_ascii=False, indent=2)
    print(rendered)
    if args.json:
        args.json.write_text(rendered + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
