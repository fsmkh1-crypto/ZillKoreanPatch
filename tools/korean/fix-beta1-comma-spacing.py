#!/usr/bin/env python3
import argparse
import json
import re
import tomllib
from pathlib import Path

SECTION_RE = re.compile(r'^\["([0-9]+)"\]$')
PATTERN = re.compile(r',(?=[^\s<])')
EXPECTED_OCCURRENCES = 221
EXPECTED_RECORDS = 184
EXPECTED_FILES = 78


def toml_quote(value: str) -> str:
    return json.dumps(value, ensure_ascii=False)


def population(root: Path):
    occurrences = 0
    rows = set()
    files = set()
    for path in sorted((root / 'translations/korean/messages').glob('*.toml')):
        data = tomllib.loads(path.read_text(encoding='utf-8'))
        rel = path.relative_to(root).as_posix()
        for rid, row in data.items():
            if not isinstance(row, dict) or not isinstance(row.get('korean'), str):
                continue
            n = len(PATTERN.findall(row['korean']))
            if n:
                occurrences += n
                rows.add((rel, str(rid)))
                files.add(rel)
    return occurrences, rows, files


def rewrite(path: Path):
    text = path.read_text(encoding='utf-8')
    data = tomllib.loads(text)
    replacements = {}
    for rid, row in data.items():
        if isinstance(row, dict) and isinstance(row.get('korean'), str):
            after = PATTERN.sub(', ', row['korean'])
            if after != row['korean']:
                replacements[str(rid)] = after
    if not replacements:
        return 0
    lines = text.splitlines(keepends=True)
    current = None
    changed = set()
    for i, line in enumerate(lines):
        stripped = line.rstrip('\r\n')
        m = SECTION_RE.match(stripped)
        if m:
            current = m.group(1)
            continue
        if current in replacements and stripped.startswith('korean = "'):
            newline = '\r\n' if line.endswith('\r\n') else ('\n' if line.endswith('\n') else '')
            lines[i] = f'korean = {toml_quote(replacements[current])}{newline}'
            changed.add(current)
    if changed != set(replacements):
        raise SystemExit(f'{path}: failed exact line rewrite')
    rendered = ''.join(lines)
    tomllib.loads(rendered)
    path.write_text(rendered, encoding='utf-8', newline='')
    return len(replacements)


def main() -> int:
    p = argparse.ArgumentParser()
    p.add_argument('--apply', action='store_true')
    p.add_argument('--json', type=Path)
    args = p.parse_args()
    root = Path(__file__).resolve().parents[2]
    occ, rows, files = population(root)
    if (occ, len(rows), len(files)) != (EXPECTED_OCCURRENCES, EXPECTED_RECORDS, EXPECTED_FILES):
        raise SystemExit(f'comma-spacing population drift: got occurrences={occ} records={len(rows)} files={len(files)}')
    before_files = sorted(files)
    if args.apply:
        for rel in before_files:
            rewrite(root / rel)
        residual, _, _ = population(root)
        if residual != 0:
            raise SystemExit(f'comma-spacing residual={residual}')
    result = {
        'mode': 'apply' if args.apply else 'verify-population',
        'occurrences': occ,
        'records': len(rows),
        'files': len(files),
        'changed_files': before_files,
    }
    rendered = json.dumps(result, ensure_ascii=False, indent=2)
    print(rendered)
    if args.json:
        args.json.write_text(rendered + '\n', encoding='utf-8')
    return 0

if __name__ == '__main__':
    raise SystemExit(main())
