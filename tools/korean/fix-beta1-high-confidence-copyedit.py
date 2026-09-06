#!/usr/bin/env python3
import argparse
import json
import re
import tomllib
from collections import Counter
from pathlib import Path

from korean_morphology import wrong_particle_pattern

SECTION_RE = re.compile(r'^\["([0-9]+)"\]$')

# LEGACY ONE-OFF CHECKPOINT TOOL. Do not use this to create future Beta1 edits;
# new mutations must be exact reviewed manifests through apply-reviewed-copyedit.py.
# These historical expected counts document the already-applied refined batch.
RULE_SPECS = [
    ("발로르", "을", "발로르를", 42),
    ("발로르", "이", "발로르가", 16),
    ("로스톨", "가", "로스톨이", 3),
    ("로스톨", "를", "로스톨을", 3),
    ("로스톨", "와", "로스톨과", 2),
    ("로스톨", "는", "로스톨은", 1),
    ("소도", "으로", "소도로", 1),
]
RULES = [
    (f"wrong_particle_{word}_{wrong}", wrong_particle_pattern(word, wrong), replacement, expected)
    for word, wrong, replacement, expected in RULE_SPECS
]
RULES.append(("spacing_moyang_ijiman", re.compile(r"모양 이지만"), "모양이지만", 1))


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
        for kind, pattern, replacement, _expected in RULES:
            after, n = pattern.subn(replacement, after)
            if n:
                counts[kind] += n
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


def current_population(root: Path) -> tuple[Counter, set[tuple[str, str]], set[str]]:
    total = Counter()
    touched = set()
    touched_files = set()
    for path in sorted((root / "translations/korean/messages").glob("*.toml")):
        data = tomllib.loads(path.read_text(encoding="utf-8"))
        rel = path.relative_to(root).as_posix()
        for rid, row in data.items():
            if not isinstance(row, dict) or not isinstance(row.get("korean"), str):
                continue
            text = row["korean"]
            row_hit = False
            for kind, pattern, _replacement, _expected in RULES:
                n = len(list(pattern.finditer(text)))
                if n:
                    total[kind] += n
                    row_hit = True
            if row_hit:
                touched.add((rel, str(rid)))
                touched_files.add(rel)
    return total, touched, touched_files


def scan(root: Path, apply: bool) -> dict:
    before_total, before_touched, before_files = current_population(root)
    expected = {kind: expected for kind, _pattern, _replacement, expected in RULES}
    if dict(before_total) != expected:
        raise SystemExit(f"high-confidence population drift: expected={expected} actual={dict(before_total)}")
    if sum(before_total.values()) != 69 or len(before_touched) != 64 or len(before_files) != 29:
        raise SystemExit(
            f"unexpected population: findings={sum(before_total.values())}/69 "
            f"records={len(before_touched)}/64 files={len(before_files)}/29"
        )

    if apply:
        applied_total = Counter()
        applied_touched = set()
        applied_files = set()
        for path in sorted((root / "translations/korean/messages").glob("*.toml")):
            counts, ids = rewrite_file(path, True)
            applied_total.update(counts)
            if ids:
                rel = path.relative_to(root).as_posix()
                applied_files.add(rel)
                applied_touched.update((rel, rid) for rid in ids)
        if applied_total != before_total or applied_touched != before_touched or applied_files != before_files:
            raise SystemExit("applied population differs from verified pre-apply population")
        residual, _rows, _files = current_population(root)
        if residual:
            raise SystemExit(f"approved high-confidence residuals remain: {dict(residual)}")

    return {
        "mode": "apply" if apply else "verify-population",
        "finding_count": sum(before_total.values()),
        "record_count": len(before_touched),
        "file_count": len(before_files),
        "by_kind": dict(before_total),
        "files": sorted(before_files),
    }


def main() -> int:
    parser = argparse.ArgumentParser(description="LEGACY: historical Beta1 high-confidence batch only")
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
