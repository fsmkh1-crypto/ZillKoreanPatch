#!/usr/bin/env python3
"""Drop already-translated ordinary rows from pending Korean result packets.

This is a race guard for the serialized apply pipeline. A result packet can be
created from a helper snapshot that becomes stale before its workflow starts,
because an earlier apply/TM run may translate some of the same IDs first.

Rules:
- Current corpus always wins for rows whose ID is already translated.
- Conflicting duplicate rows inside the pending inputs remain fatal.
- Recovery specs are left untouched; apply-results.py already handles them
  fail-safely.
- Compact sequential rows are expanded to ordinary id rows so individual stale
  IDs can be removed safely.
- This script does not validate controls or write overlays; apply-results.py
  remains the authoritative validator/applicator.
"""
from __future__ import annotations

import argparse
import json
from pathlib import Path
import re
import sys
import tomllib

ROOT = Path(__file__).resolve().parents[2]
KOREAN = ROOT / "translations" / "korean" / "messages"
SECTION_RE = re.compile(r"^msgsec(\d{3})")


def section_from_path(path: Path) -> int:
    match = SECTION_RE.match(path.stem)
    if not match:
        raise SystemExit(f"cannot infer section from Korean overlay filename: {path}")
    return int(match.group(1))


def load_existing_ids() -> set[tuple[int, str]]:
    existing: set[tuple[int, str]] = set()
    for path in sorted(KOREAN.glob("msgsec*.toml")):
        section = section_from_path(path)
        with path.open("rb") as f:
            data = tomllib.load(f)
        for rid, rec in data.items():
            if isinstance(rec, dict) and rec.get("korean"):
                existing.add((section, str(rid)))
    return existing


def expand_ordinary(obj: dict, path: Path, lineno: int) -> list[dict]:
    try:
        section = int(obj["section"])
    except (KeyError, TypeError, ValueError) as exc:
        raise SystemExit(f"{path}:{lineno}: missing or invalid section") from exc

    if "id" in obj:
        if "korean" not in obj:
            raise SystemExit(f"{path}:{lineno}: id row requires korean")
        return [{"section": section, "id": str(obj["id"]), "korean": str(obj["korean"])}]

    if "start" in obj and isinstance(obj.get("korean"), list):
        try:
            start = int(obj["start"])
        except (TypeError, ValueError) as exc:
            raise SystemExit(f"{path}:{lineno}: invalid sequential start") from exc
        return [
            {"section": section, "id": str(start + offset), "korean": str(ko)}
            for offset, ko in enumerate(obj["korean"])
        ]

    raise SystemExit(f"{path}:{lineno}: expected id+korean, start+korean[], or recover_from")


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("inputs", nargs="+", type=Path)
    args = ap.parse_args()

    existing = load_existing_ids()
    seen: dict[tuple[int, str], tuple[str, str]] = {}
    total_kept = 0
    total_stale = 0

    for path in args.inputs:
        kept: list[dict] = []
        stale = 0
        with path.open(encoding="utf-8") as f:
            for lineno, line in enumerate(f, 1):
                if not line.strip():
                    continue
                try:
                    obj = json.loads(line)
                except json.JSONDecodeError as exc:
                    raise SystemExit(f"{path}:{lineno}: invalid JSON: {exc}") from exc
                if not isinstance(obj, dict):
                    raise SystemExit(f"{path}:{lineno}: expected a JSON object")

                if "recover_from" in obj:
                    kept.append(obj)
                    continue

                for row in expand_ordinary(obj, path, lineno):
                    key = (int(row["section"]), str(row["id"]))
                    ko = str(row["korean"])
                    origin = f"{path}:{lineno}"
                    if key in seen:
                        prev_ko, prev_origin = seen[key]
                        if prev_ko != ko:
                            raise SystemExit(
                                f"pending conflict {key[0]}/{key[1]}: {prev_origin} != {origin}"
                            )
                        continue
                    seen[key] = (ko, origin)

                    if key in existing:
                        stale += 1
                        total_stale += 1
                        print(f"stale-skip {key[0]}/{key[1]}: current corpus already translated; source={origin}")
                        continue

                    kept.append(row)
                    total_kept += 1

        text = "".join(json.dumps(row, ensure_ascii=False, separators=(",", ":")) + "\n" for row in kept)
        path.write_text(text, encoding="utf-8")
        print(f"{path}: kept={len(kept)} stale-skipped={stale}")

    print(f"preflight complete: kept={total_kept} stale-skipped={total_stale}")


if __name__ == "__main__":
    main()
