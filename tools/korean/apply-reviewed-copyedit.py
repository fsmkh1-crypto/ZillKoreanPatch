#!/usr/bin/env python3
import argparse
import hashlib
import json
import re
import tomllib
from pathlib import Path

SECTION_RE = re.compile(r'^\["([0-9]+)"\]$')
CONTROL_RE = re.compile(r"<[^>]+>")


def toml_quote(value: str) -> str:
    return json.dumps(value, ensure_ascii=False)


def controls(text: str) -> list[str]:
    return CONTROL_RE.findall(text)


def load_manifest(path: Path) -> dict:
    data = json.loads(path.read_text(encoding="utf-8"))
    required = {"korean_file", "records"}
    missing = required - data.keys()
    if missing:
        raise ValueError(f"manifest missing required fields: {sorted(missing)}")
    return data


def proposal_records(manifest: dict) -> list[dict]:
    out = []
    seen = set()
    for record in manifest["records"]:
        if record.get("decision") != "PROPOSED_CHANGE":
            continue
        rid = str(record["id"])
        if rid in seen:
            raise ValueError(f"duplicate proposed record id {rid}")
        seen.add(rid)
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


def validate_record_state(data: dict, record: dict, verify: bool) -> str:
    rid = str(record["id"])
    if rid not in data:
        raise ValueError(f"record {rid} missing from Korean overlay")
    row = data[rid]
    expected_japanese = record.get("japanese")
    if expected_japanese is not None and row.get("japanese") != expected_japanese:
        raise ValueError(
            f"record {rid} Japanese/source reference changed: "
            f"expected {expected_japanese!r}, found {row.get('japanese')!r}"
        )
    current = row.get("korean")
    before = record["before_korean"]
    after = record["proposed_korean"]
    if verify:
        if current != after:
            raise ValueError(f"record {rid} is not at reviewed after-value")
        return "verified"
    if current == before:
        return "apply"
    if current == after:
        return "already-applied"
    raise ValueError(
        f"record {rid} Korean no longer matches reviewed before/after values; "
        f"refusing blind overwrite"
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


def apply_manifest(root: Path, manifest_path: Path, verify: bool = False) -> dict:
    manifest = load_manifest(manifest_path)
    records = proposal_records(manifest)
    overlay_path = root / manifest["korean_file"]
    text, data = read_overlay(overlay_path)

    states = {}
    replacements = {}
    for record in records:
        rid = str(record["id"])
        state = validate_record_state(data, record, verify)
        states[rid] = state
        if state == "apply":
            replacements[rid] = record["proposed_korean"]

    if verify:
        return {
            "status": "VERIFIED_APPLIED",
            "proposed_records": len(records),
            "verified_records": len(records),
            "changed_records": 0,
            "korean_file": manifest["korean_file"],
        }

    if replacements:
        rendered = rewrite_korean_lines(text, replacements)
        overlay_path.write_text(rendered, encoding="utf-8", newline="")
        _, post = read_overlay(overlay_path)
        for record in records:
            rid = str(record["id"])
            if post[rid].get("korean") != record["proposed_korean"]:
                raise ValueError(f"record {rid} failed post-write verification")

    digest = hashlib.sha256(overlay_path.read_bytes()).hexdigest()
    return {
        "status": "APPLIED_OR_ALREADY_APPLIED",
        "proposed_records": len(records),
        "changed_records": len(replacements),
        "already_applied_records": len(records) - len(replacements),
        "korean_file": manifest["korean_file"],
        "korean_file_sha256": digest,
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", required=True, type=Path)
    parser.add_argument("--verify", action="store_true")
    parser.add_argument("--json", type=Path)
    args = parser.parse_args()

    root = Path(__file__).resolve().parents[2]
    manifest_path = args.manifest
    if not manifest_path.is_absolute():
        manifest_path = root / manifest_path
    result = apply_manifest(root, manifest_path, verify=args.verify)
    rendered = json.dumps(result, ensure_ascii=False, indent=2)
    print(rendered)
    if args.json:
        args.json.write_text(rendered + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
