#!/usr/bin/env python3
"""Cheap, read-only precheck for newly generated Korean result packets.

This is intentionally not authoritative. It reuses the exact fixed_tokens()
implementation used by apply-results.py, so agents can catch obvious control
signature mistakes before committing a packet. apply-results.py remains the
sole writer and authoritative gate.
"""
from __future__ import annotations

import argparse
import json
from pathlib import Path
import sys
import tomllib

from control_tags import fixed_tokens

ROOT = Path(__file__).resolve().parents[2]
CANON = ROOT / "translations" / "messages"
LINE_BREAK = "<line-break>"


def canonical_for(section: int) -> dict[str, dict[str, object]]:
    path = CANON / f"msgsec{section:03d}.toml"
    if not path.exists():
        raise SystemExit(f"missing canonical section {section}: {path}")
    with path.open("rb") as f:
        raw = tomllib.load(f)
    return {str(k): v for k, v in raw.items() if isinstance(v, dict)}


def expanded_rows(obj: dict, input_path: Path, lineno: int):
    if "recover_from" in obj:
        raise SystemExit(
            f"{input_path}:{lineno}: precheck is for fresh result rows only; "
            "historical recovery stays authoritative in apply-results.py"
        )
    try:
        section = int(obj["section"])
    except (KeyError, TypeError, ValueError) as exc:
        raise SystemExit(f"{input_path}:{lineno}: missing or invalid section") from exc
    if "id" in obj:
        if "korean" not in obj:
            raise SystemExit(f"{input_path}:{lineno}: id row requires korean")
        yield section, str(obj["id"]), str(obj["korean"])
        return
    if "start" in obj and isinstance(obj.get("korean"), list):
        try:
            start = int(obj["start"])
        except (TypeError, ValueError) as exc:
            raise SystemExit(f"{input_path}:{lineno}: invalid sequential start") from exc
        for offset, ko in enumerate(obj["korean"]):
            yield section, str(start + offset), str(ko)
        return
    raise SystemExit(f"{input_path}:{lineno}: expected id+korean or start+korean[]")


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("inputs", nargs="+", type=Path)
    args = ap.parse_args()

    canonical_cache: dict[int, dict[str, dict[str, object]]] = {}
    seen: set[tuple[int, str]] = set()
    checked = 0

    for input_path in args.inputs:
        with input_path.open(encoding="utf-8") as f:
            for lineno, line in enumerate(f, 1):
                if not line.strip():
                    continue
                try:
                    obj = json.loads(line)
                except json.JSONDecodeError as exc:
                    raise SystemExit(f"{input_path}:{lineno}: invalid JSON: {exc}") from exc
                if not isinstance(obj, dict):
                    raise SystemExit(f"{input_path}:{lineno}: expected a JSON object")

                for section, rid, ko in expanded_rows(obj, input_path, lineno):
                    key = (section, rid)
                    if key in seen:
                        raise SystemExit(f"{input_path}:{lineno}: duplicate result id {section}/{rid}")
                    seen.add(key)
                    if not ko:
                        raise SystemExit(f"{input_path}:{lineno}: empty Korean translation {section}/{rid}")
                    if LINE_BREAK in ko:
                        raise SystemExit(
                            f"{input_path}:{lineno}: semantic Korean {section}/{rid} contains {LINE_BREAK}"
                        )

                    canon = canonical_cache.setdefault(section, canonical_for(section))
                    rec = canon.get(rid)
                    if rec is None or "japanese" not in rec:
                        raise SystemExit(f"{input_path}:{lineno}: unknown canonical id {section}/{rid}")
                    ja = str(rec["japanese"])
                    ja_tokens = fixed_tokens(ja)
                    ko_tokens = fixed_tokens(ko)
                    if ja_tokens != ko_tokens:
                        raise SystemExit(
                            f"{input_path}:{lineno}: fixed-control mismatch {section}/{rid}: "
                            f"{ja_tokens} != {ko_tokens}"
                        )
                    checked += 1

    print(f"prechecked {checked} fresh Korean result rows; authoritative apply still required")


if __name__ == "__main__":
    try:
        main()
    except BrokenPipeError:
        sys.exit(1)
