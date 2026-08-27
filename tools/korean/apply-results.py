#!/usr/bin/env python3
"""Validate and apply Korean result JSONL files to sparse overlays.

Supported input rows:
  {"section":3,"id":"30000","korean":"..."}
  {"section":3,"start":30000,"korean":["...","...",...]}
  {"recover_from":{"commit":"<40-char sha>","path":"work/korean-results/old.jsonl"}}

The compact sequential form expands IDs from start and keeps large GPT result
packets small. Canonical Japanese is always loaded locally. Fixed runtime
controls are validated after expansion. ``<line-break>`` is forbidden in the
translator-owned ``korean`` field because wrapping belongs only in generated
``layout`` metadata. Historical recovery sources are normalized by removing
legacy ``<line-break>`` layout markers before semantic validation.

Historical recovery rows are deliberately fail-safe: the historical packet is
read from Git history, but an already-existing canonical Korean translation is
never overwritten or treated as a conflict. Only IDs that are still missing are
recovered. Ordinary result rows remain fail-closed on conflicts.

Ordinary existing/duplicate conflicts are collected and reported together. The
script does not write any overlay file when even one ordinary conflict exists,
so this remains fail-closed while avoiding one-conflict-per-run debugging.
"""
from __future__ import annotations

import argparse
import json
from pathlib import Path
import re
import subprocess
import sys
import tomllib

from control_tags import fixed_tokens

ROOT = Path(__file__).resolve().parents[2]
CANON = ROOT / "translations" / "messages"
KOREAN = ROOT / "translations" / "korean" / "messages"
SECTION_RE = re.compile(r"^msgsec(\d{3})")
FULL_COMMIT_RE = re.compile(r"^[0-9a-fA-F]{40}$")
AUTO_PART = 99
LINE_BREAK = "<line-break>"
RECOVERY_ROOT = ("work", "korean-results")


def q(s: str) -> str:
    return json.dumps(s, ensure_ascii=False)


def section_from_path(path: Path) -> int:
    match = SECTION_RE.match(path.stem)
    if not match:
        raise SystemExit(f"cannot infer section from Korean overlay filename: {path}")
    return int(match.group(1))


def auto_path(section: int) -> Path:
    return KOREAN / f"msgsec{section:03d}-part{AUTO_PART:02d}.toml"


def canonical_for(section: int) -> dict[str, dict[str, object]]:
    path = CANON / f"msgsec{section:03d}.toml"
    if not path.exists():
        raise SystemExit(f"missing canonical section {section}: {path}")
    with path.open("rb") as f:
        raw = tomllib.load(f)
    return {str(k): v for k, v in raw.items() if isinstance(v, dict)}


def load_existing() -> tuple[dict[tuple[int, str], str], dict[int, dict[str, dict[str, str]]]]:
    owners: dict[tuple[int, str], Path] = {}
    existing: dict[tuple[int, str], str] = {}
    auto: dict[int, dict[str, dict[str, str]]] = {}
    for path in sorted(KOREAN.glob("msgsec*.toml")):
        section = section_from_path(path)
        with path.open("rb") as f:
            data = tomllib.load(f)
        for rid, rec in data.items():
            if not isinstance(rec, dict) or not rec.get("korean"):
                continue
            key = (section, str(rid))
            ko = str(rec["korean"])
            if key in existing:
                if existing[key] != ko:
                    raise SystemExit(f"conflicting Korean id {section}/{rid}: {owners[key]} and {path}")
                continue
            existing[key] = ko
            owners[key] = path
            if path == auto_path(section):
                saved = {
                    "japanese": str(rec.get("japanese", "")),
                    "korean": ko,
                }
                if rec.get("layout"):
                    saved["layout"] = str(rec["layout"])
                auto.setdefault(section, {})[str(rid)] = saved
    return existing, auto


def render(records: dict[str, dict[str, str]]) -> str:
    out = ["# SPDX-License-Identifier: CC-BY-SA-4.0", ""]
    for rid in sorted(records, key=lambda x: int(x) if x.isdigit() else x):
        rec = records[rid]
        out += [f"[{q(rid)}]", f"japanese = {q(rec['japanese'])}", f"korean = {q(rec['korean'])}"]
        if rec.get("layout"):
            out.append(f"layout = {q(rec['layout'])}")
        out.append("")
    return "\n".join(out)


def expanded_rows(obj: dict, input_path: Path, lineno: int):
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


def validate_recovery_spec(spec: object, input_path: Path, lineno: int) -> tuple[str, str]:
    if not isinstance(spec, dict):
        raise SystemExit(f"{input_path}:{lineno}: recover_from must be an object")
    commit = str(spec.get("commit", "")).strip()
    path = str(spec.get("path", "")).strip()
    if not commit or not path:
        raise SystemExit(f"{input_path}:{lineno}: recover_from requires commit and path")
    if not FULL_COMMIT_RE.fullmatch(commit):
        raise SystemExit(f"{input_path}:{lineno}: recovery commit must be a full 40-character hex SHA")

    recovery_path = Path(path)
    if recovery_path.is_absolute() or ".." in recovery_path.parts:
        raise SystemExit(f"{input_path}:{lineno}: unsafe recovery path: {path}")
    if len(recovery_path.parts) < 3 or recovery_path.parts[:2] != RECOVERY_ROOT:
        raise SystemExit(
            f"{input_path}:{lineno}: recovery path must be under work/korean-results/: {path}"
        )
    if recovery_path.suffix != ".jsonl":
        raise SystemExit(f"{input_path}:{lineno}: recovery source must be a .jsonl result packet: {path}")
    return commit.lower(), recovery_path.as_posix()


def historical_rows(obj: dict, input_path: Path, lineno: int):
    if "recover_from" not in obj:
        for section, rid, ko in expanded_rows(obj, input_path, lineno):
            yield section, rid, ko, False, f"{input_path}:{lineno}"
        return

    commit, path = validate_recovery_spec(obj.get("recover_from"), input_path, lineno)
    proc = subprocess.run(
        ["git", "show", f"{commit}:{path}"],
        cwd=ROOT,
        text=True,
        capture_output=True,
        check=False,
    )
    if proc.returncode != 0:
        detail = proc.stderr.strip() or proc.stdout.strip()
        raise SystemExit(f"{input_path}:{lineno}: cannot read recovery source {commit}:{path}: {detail}")

    source_label = Path(f"{commit[:12]}-{Path(path).name}")
    for source_lineno, line in enumerate(proc.stdout.splitlines(), 1):
        if not line.strip():
            continue
        try:
            historical = json.loads(line)
        except json.JSONDecodeError as exc:
            raise SystemExit(
                f"{input_path}:{lineno}: invalid JSON in recovery source {path}:{source_lineno}: {exc}"
            ) from exc
        if not isinstance(historical, dict):
            raise SystemExit(
                f"{input_path}:{lineno}: recovery source {path}:{source_lineno} must contain JSON objects"
            )
        if "recover_from" in historical:
            raise SystemExit(
                f"{input_path}:{lineno}: nested recover_from is not allowed in {path}:{source_lineno}"
            )
        origin = f"{commit}:{path}:{source_lineno}"
        for section, rid, ko in expanded_rows(historical, source_label, source_lineno):
            # Recovery packets may predate the semantic/layout split. Layout
            # markers are deterministic presentation metadata, not translation
            # meaning, so normalize them only on the historical recovery path.
            ko = ko.replace(LINE_BREAK, "")
            yield section, rid, ko, True, origin


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("inputs", nargs="+", type=Path)
    args = ap.parse_args()

    existing, auto = load_existing()
    canonical_cache: dict[int, dict[str, dict[str, object]]] = {}
    seen_input: dict[tuple[int, str], str] = {}
    seen_origin: dict[tuple[int, str], str] = {}
    added = 0
    unchanged = 0
    recovery_skips: list[tuple[int, str, str, str]] = []
    ordinary_conflicts: list[tuple[int, str, str, str]] = []

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

                for section, rid, ko, recovery, origin in historical_rows(obj, input_path, lineno):
                    key = (section, rid)
                    if key in seen_input:
                        if seen_input[key] != ko:
                            if recovery:
                                recovery_skips.append(
                                    (section, rid, "earlier input preserved", f"{origin}; first={seen_origin[key]}")
                                )
                                continue
                            ordinary_conflicts.append(
                                (section, rid, "conflicting duplicate input", f"{origin}; first={seen_origin[key]}")
                            )
                            continue
                        continue

                    seen_input[key] = ko
                    seen_origin[key] = origin
                    if not ko:
                        raise SystemExit(f"{origin}: empty Korean translation {section}/{rid}")
                    if LINE_BREAK in ko:
                        raise SystemExit(
                            f"{origin}: semantic Korean {section}/{rid} contains {LINE_BREAK}; "
                            "omit layout breaks from translation output"
                        )

                    canon = canonical_cache.setdefault(section, canonical_for(section))
                    rec = canon.get(rid)
                    if rec is None or "japanese" not in rec:
                        raise SystemExit(f"{origin}: unknown canonical id {section}/{rid}")
                    ja = str(rec["japanese"])
                    if fixed_tokens(ja) != fixed_tokens(ko):
                        raise SystemExit(
                            f"{origin}: fixed-control mismatch {section}/{rid}: "
                            f"{fixed_tokens(ja)} != {fixed_tokens(ko)}"
                        )

                    if key in existing:
                        if existing[key] != ko:
                            if recovery:
                                recovery_skips.append((section, rid, "current corpus preserved", origin))
                                continue
                            ordinary_conflicts.append((section, rid, "existing translation differs", origin))
                            continue
                        unchanged += 1
                        continue

                    auto.setdefault(section, {})[rid] = {"japanese": ja, "korean": ko}
                    existing[key] = ko
                    added += 1

    if ordinary_conflicts:
        for section, rid, reason, origin in sorted(ordinary_conflicts):
            print(
                f"conflict {section}/{rid}: {reason}; source={origin}",
                file=sys.stderr,
            )
        raise SystemExit(
            f"{len(ordinary_conflicts)} conflicting ordinary translation rows; no overlays written"
        )

    KOREAN.mkdir(parents=True, exist_ok=True)
    for section, records in sorted(auto.items()):
        auto_path(section).write_text(render(records), encoding="utf-8")

    for section, rid, reason, origin in sorted(recovery_skips):
        print(
            f"recovery-skip {section}/{rid}: {reason}; source={origin}",
            file=sys.stderr,
        )
    print(
        f"applied {added} new translations; {unchanged} already identical; "
        f"{len(recovery_skips)} historical conflicts preserved current corpus"
    )


if __name__ == "__main__":
    main()
