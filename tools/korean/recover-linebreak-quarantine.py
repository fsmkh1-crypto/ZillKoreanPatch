#!/usr/bin/env python3
"""Recover Korean wording deleted by the legacy line-break-sensitive quarantine.

Commit 7620fdd ran the old refresh-japanese-refs.py whose control comparison
incorrectly treated Japanese <line-break> positions as fixed. Eight valid Korean
rows were removed. Their wording is recovered from the parent commit, while the
old visual breaks are moved to build-owned ``layout`` and current canonical
Japanese is used as the authoritative reference.

The script is intentionally one-shot/idempotent: existing IDs are skipped, and a
row is restored only if its full fixed runtime-control sequence still matches the
current canonical source.
"""
from __future__ import annotations

import json
from pathlib import Path
import re
import subprocess
import tomllib

from control_tags import fixed_tokens

ROOT = Path(__file__).resolve().parents[2]
CANON = ROOT / "translations" / "messages"
KOREAN = ROOT / "translations" / "korean" / "messages"
HISTORICAL_REF = "74a1bf1b049e9fa480cfdac730db5cb10f906570"
LINE_BREAK = "<line-break>"
BREAK_WITH_SPACE_RE = re.compile(r"\s*<line-break>\s*")

RECOVER = {
    3: {
        "30000": "translations/korean/messages/msgsec003.toml",
        "30007": "translations/korean/messages/msgsec003-part01.toml",
        "30017": "translations/korean/messages/msgsec003-part04.toml",
        "30018": "translations/korean/messages/msgsec003-part04.toml",
        "30028": "translations/korean/messages/msgsec003-part06.toml",
        "30029": "translations/korean/messages/msgsec003-part06.toml",
        "30030": "translations/korean/messages/msgsec003-part06.toml",
    },
    6: {
        "60011": "translations/korean/messages/msgsec006-part04.toml",
    },
}


def q(text: str) -> str:
    return json.dumps(text, ensure_ascii=False)


def historical_file(path: str) -> dict[str, object]:
    raw = subprocess.check_output(
        ["git", "show", f"{HISTORICAL_REF}:{path}"], cwd=ROOT
    )
    return tomllib.loads(raw.decode("utf-8"))


def canonical(section: int) -> dict[str, object]:
    with (CANON / f"msgsec{section:03d}.toml").open("rb") as f:
        return tomllib.load(f)


def existing_ids() -> set[str]:
    result: set[str] = set()
    for path in sorted(KOREAN.glob("msgsec*.toml")):
        with path.open("rb") as f:
            data = tomllib.load(f)
        for rid, rec in data.items():
            if isinstance(rec, dict) and rec.get("korean"):
                result.add(str(rid))
    return result


def semanticize(text: str) -> str:
    return BREAK_WITH_SPACE_RE.sub(" ", text)


def layoutize(text: str) -> str:
    return BREAK_WITH_SPACE_RE.sub(LINE_BREAK, text)


def render(records: dict[str, tuple[str, str, str]]) -> str:
    out = ["# SPDX-License-Identifier: CC-BY-SA-4.0", ""]
    for rid in sorted(records, key=int):
        japanese, korean, layout = records[rid]
        out += [
            f"[{q(rid)}]",
            f"japanese = {q(japanese)}",
            f"korean = {q(korean)}",
            f"layout = {q(layout)}",
            "",
        ]
    return "\n".join(out)


def main() -> None:
    existing = existing_ids()
    history_cache: dict[str, dict[str, object]] = {}
    restored = 0

    for section, ids in sorted(RECOVER.items()):
        current = canonical(section)
        output: dict[str, tuple[str, str, str]] = {}
        out_path = KOREAN / f"msgsec{section:03d}-part98.toml"
        if out_path.exists():
            with out_path.open("rb") as f:
                old = tomllib.load(f)
            for rid, rec in old.items():
                if isinstance(rec, dict):
                    output[str(rid)] = (
                        str(rec["japanese"]), str(rec["korean"]), str(rec["layout"])
                    )

        for rid, historical_path in ids.items():
            if rid in existing:
                continue
            history = history_cache.setdefault(historical_path, historical_file(historical_path))
            historical = history.get(rid)
            canonical_row = current.get(rid)
            if not isinstance(historical, dict) or not isinstance(canonical_row, dict):
                raise SystemExit(f"cannot recover canonical/historical row {section}/{rid}")
            old_korean = str(historical.get("korean", ""))
            current_japanese = str(canonical_row.get("japanese", ""))
            if not old_korean or not current_japanese:
                raise SystemExit(f"empty canonical/historical row {section}/{rid}")
            if fixed_tokens(current_japanese) != fixed_tokens(old_korean):
                raise SystemExit(
                    f"refuse recovery {section}/{rid}: fixed controls differ: "
                    f"{fixed_tokens(current_japanese)} != {fixed_tokens(old_korean)}"
                )
            semantic = semanticize(old_korean)
            layout = layoutize(old_korean)
            if LINE_BREAK in semantic:
                raise SystemExit(f"recovery left semantic line break in {section}/{rid}")
            output[rid] = (current_japanese, semantic, layout)
            existing.add(rid)
            restored += 1

        if output:
            out_path.write_text(render(output), encoding="utf-8")

    print(f"restored {restored} Korean rows quarantined by legacy line-break validation")


if __name__ == "__main__":
    main()
