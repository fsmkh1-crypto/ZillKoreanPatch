#!/usr/bin/env python3
import argparse
import json
import re
import tomllib
from collections import Counter
from pathlib import Path

SECTION_RE = re.compile(r'^\["([0-9]+)"\]$')

RULES = [
    ("wrong_particle_발로르_을", "발로르을", "발로르를", 42),
    ("wrong_particle_발로르_이", "발로르이", "발로르가", 16),
    ("wrong_particle_로스톨_가", "로스톨가", "로스톨이", 3),
    ("wrong_particle_로스톨_를", "로스톨를", "로스톨을", 3),
    ("wrong_particle_로스톨_와", "로스톨와", "로스톨과", 2),
    ("wrong_particle_로스톨_는", "로스톨는", "로스톨은", 1),
    ("wrong_particle_소도_으로", "소도으로", "소도로", 1),
    ("spacing_moyang_ijiman", "모양 이지만", "모양이지만", 1),
]


def toml_quote(value: str) -> str:
    return json.dumps(value, ensure_ascii=False)


def rewrite_file(path: Path, apply: bool) -> tuple[Counter, set[str]]:
    text = path.read_text(encoding="utf-8")
    data = tomllib.loads(text)
    replacements = {}
    counts = Counter()
    touched_ids = set()
    for rid, row in data.items():
        if not isinstance(row, dict) or not isinstance(row.get("korean"), str):
            continue
        before = row["korean"]
        after = before
        for kind, old, new, _expected in RULES:
            n = after.count(old)
            if n:
                counts[kind] += n
                after = after.replace(old, new)
        if after != before:
            replacements[str(rid)] = after
            touched_ids.add(str(rid))
    if not apply or not replacements:
        return counts, touched_ids

    lines = text.splitlines(keepends=True)
    current = None
    changed = set()
    for index, line in enumerate(lines):
        stripped = line.rstrip("\r\n")
        m = SECTION_RE.match(stripped)
        if m:
            current = m.group(1)
            continue
        if current in replacements and stripped.startswith('korean = "'):
            newline = "\r\n" if line.endswith("\r\n") else ("\n" if line.endswith("\n") else "")
            lines[index] = f"korean = {toml_quote(replacements[current])}{newline}"
            changed.add(current)
    missing = set(replacements) - changed
    if missing:
        raise ValueError(f"{path}: failed to locate Korean lines for IDs {sorted(missing)}")
    rendered = "".join(lines)
    tomllib.loads(rendered)
    path.write_text(rendered, encoding="utf-8", newline="")
    return counts, touched_ids


def scan(root: Path, apply: bool) -> dict:
    total = Counter()
    touched = set()
    touched_files = set()
    for path in sorted((root / "translations/korean/messages").glob("*.toml")):
        counts, ids = rewrite_file(path, apply)
        total.update(counts)
        if ids:
            rel = path.relative_to(root).as_posix()
            touched_files.add(rel)
            touched.update((rel, rid) for rid in ids)

    expected = {kind: expected for kind, _old, _new, expected in RULES}
    if not apply:
        if dict(total) != expected:
            raise SystemExit(f"high-confidence population drift: expected={expected} actual={dict(total)}")
        if sum(total.values()) != 69 or len(touched) != 64 or len(touched_files) != 29:
            raise SystemExit(
                f"unexpected population: findings={sum(total.values())}/69 records={len(touched)}/64 files={len(touched_files)}/29"
            )
    else:
        # Re-scan post-write. All approved high-confidence patterns must be gone.
        residual = Counter()
        for path in sorted((root / "translations/korean/messages").glob("*.toml")):
            data = tomllib.loads(path.read_text(encoding="utf-8"))
            for row in data.values():
                if not isinstance(row, dict) or not isinstance(row.get("korean"), str):
                    continue
                text = row["korean"]
                for kind, old, _new, _expected in RULES:
                    residual[kind] += text.count(old)
        residual = Counter({k: v for k, v in residual.items() if v})
        if residual:
            raise SystemExit(f"approved high-confidence residuals remain: {dict(residual)}")

    return {
        "mode": "apply" if apply else "verify-population",
        "finding_count": sum(total.values()),
        "record_count": len(touched),
        "file_count": len(touched_files),
        "by_kind": dict(total),
        "files": sorted(touched_files),
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--apply", action="store_true")
    parser.add_argument("--json", type=Path)
    args = parser.parse_args()
    root = Path(__file__).resolve().parents[2]
    result = scan(root, args.apply)
    rendered = json.dumps(result, ensure_ascii=False, indent=2)
    print(rendered)
    if args.json:
        args.json.write_text(rendered + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
