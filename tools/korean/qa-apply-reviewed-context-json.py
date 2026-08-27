#!/usr/bin/env python3
from pathlib import Path
import json
import re
import tomllib
import unicodedata

ROOT = Path("translations/korean/messages")
FIX_ROOT = Path("tools/korean")
FIX_GLOB = "reviewed-context-fixes*.json"


def same_text(a: str, b: str) -> bool:
    return unicodedata.normalize("NFC", a) == unicodedata.normalize("NFC", b)


def decode_korean_line(line: str) -> str:
    """Decode a single TOML basic-string assignment semantically."""
    try:
        parsed = tomllib.loads("[record]\n" + line + "\n")
        value = parsed["record"]["korean"]
    except (tomllib.TOMLDecodeError, KeyError, TypeError) as exc:
        raise ValueError(f"invalid korean TOML assignment: {line!r}") from exc
    if not isinstance(value, str):
        raise ValueError(f"korean value is not a string: {line!r}")
    return value


def encode_korean_line(value: str) -> str:
    """Emit a TOML basic string without rewriting surrounding file formatting."""
    # JSON string escaping is a valid subset of TOML basic-string escaping for
    # the values used here and correctly handles quotes, backslashes and controls.
    return "korean = " + json.dumps(value, ensure_ascii=False)


def load_fixes() -> dict[str, dict[str, str]]:
    fixes: dict[str, dict[str, str]] = {}
    for fix_path in sorted(FIX_ROOT.glob(FIX_GLOB)):
        data = json.loads(fix_path.read_text(encoding="utf-8"))
        overlap = set(fixes) & set(data)
        if overlap:
            raise SystemExit(f"duplicate reviewed-context ids in {fix_path}: {sorted(overlap)}")
        fixes.update(data)
    return fixes


def apply_fixes(fixes: dict[str, dict[str, str]]) -> tuple[int, int, set[str]]:
    record_re = re.compile(r'^\["(\d+)"\]$')
    changed = 0
    already = 0
    files_changed: set[str] = set()
    found: set[str] = set()

    for path in ROOT.glob("msgsec*-part99.toml"):
        lines = path.read_text(encoding="utf-8").splitlines()
        current = None
        dirty = False
        for i, line in enumerate(lines):
            m = record_re.match(line)
            if m:
                current = m.group(1)
                continue
            if current in fixes and line.startswith("korean = "):
                item = fixes[current]
                old = item["old"]
                new = item["new"]
                try:
                    actual = decode_korean_line(line)
                except ValueError as exc:
                    raise SystemExit(f"id={current} path={path}: {exc}") from exc
                if same_text(actual, new):
                    already += 1
                    found.add(current)
                    current = None
                    continue
                if not same_text(actual, old):
                    raise SystemExit(
                        f"guard mismatch id={current} path={path}: "
                        f"actual={actual!r} old={old!r} new={new!r}"
                    )
                lines[i] = encode_korean_line(new)
                changed += 1
                found.add(current)
                files_changed.add(str(path))
                dirty = True
                current = None
        if dirty:
            path.write_text("\n".join(lines) + "\n", encoding="utf-8")

    missing = set(fixes) - found
    if missing:
        raise SystemExit(f"missing ids: {sorted(missing)}")
    return changed, already, files_changed


def main() -> None:
    fixes = load_fixes()
    changed, already, files_changed = apply_fixes(fixes)
    print(
        f"reviewed-context-json: changed={changed} already={already} "
        f"files={len(files_changed)} shards={len(list(FIX_ROOT.glob(FIX_GLOB)))}"
    )


if __name__ == "__main__":
    main()
