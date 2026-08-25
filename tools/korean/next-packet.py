#!/usr/bin/env python3
"""Emit the next untranslated Korean work packet from canonical corpus.

The translated set is derived from checked-in Korean overlays, so resume state
never depends on a manually maintained cursor. Output is compact JSONL containing
only section/id/Japanese source. A progress JSON file can be written alongside it.

By default untranslated rows are ordered by encoded JSONL size (shortest first).
This keeps packet fetches dense for LLM translation while preserving deterministic
selection, exact IDs, and fully resumable imports. Canonical rows whose Japanese
source is the empty string are not translatable under the non-empty Korean overlay
contract, so they are counted separately as blank/skipped and never emitted.
Use --order canonical to restore strict source order when needed for debugging/review.
"""
from __future__ import annotations

import argparse
import json
from pathlib import Path
import re
import tomllib

ROOT = Path(__file__).resolve().parents[2]
CANON = ROOT / "translations" / "messages"
KOREAN = ROOT / "translations" / "korean" / "messages"
SECTION_RE = re.compile(r"^msgsec(\d{3})")


def section_from_path(path: Path) -> int:
    match = SECTION_RE.match(path.stem)
    if not match:
        raise SystemExit(f"cannot infer section from Korean overlay filename: {path}")
    return int(match.group(1))


def load_translated() -> set[tuple[int, str]]:
    values: dict[tuple[int, str], str] = {}
    owners: dict[tuple[int, str], Path] = {}
    for path in sorted(KOREAN.glob("msgsec*.toml")):
        section = section_from_path(path)
        with path.open("rb") as f:
            data = tomllib.load(f)
        for rid, rec in data.items():
            if not isinstance(rec, dict) or not rec.get("korean"):
                continue
            key = (section, str(rid))
            ko = str(rec["korean"])
            if key in values:
                if values[key] != ko:
                    raise SystemExit(
                        f"conflicting Korean id {section}/{rid}: {owners[key]} and {path}"
                    )
                continue
            values[key] = ko
            owners[key] = path
    return set(values)


def canonical_rows() -> list[tuple[int, str, str]]:
    rows: list[tuple[int, str, str]] = []
    for path in sorted(CANON.glob("msgsec*.toml")):
        section = int(path.stem.removeprefix("msgsec"))
        with path.open("rb") as f:
            data = tomllib.load(f)
        for rid, rec in data.items():
            ja = rec.get("japanese") if isinstance(rec, dict) else None
            if ja is not None:
                rows.append((section, str(rid), str(ja)))
    return rows


def encoded_line(row: tuple[int, str, str]) -> str:
    section, rid, ja = row
    return json.dumps(
        {"section": section, "id": rid, "japanese": ja},
        ensure_ascii=False,
        separators=(",", ":"),
    ) + "\n"


def numeric_id_key(rid: str) -> tuple[int, str]:
    try:
        return (int(rid), "")
    except ValueError:
        return (2**63 - 1, rid)


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--max-records", type=int, default=2000)
    ap.add_argument("--max-bytes", type=int, default=400_000)
    ap.add_argument("--order", choices=("shortest", "canonical"), default="shortest")
    ap.add_argument("--out", type=Path)
    ap.add_argument("--progress", type=Path)
    args = ap.parse_args()
    if args.max_records < 1 or args.max_bytes < 1:
        raise SystemExit("packet limits must be positive")

    translated = load_translated()
    canonical = canonical_rows()
    canonical_keys = {(s, rid) for s, rid, _ in canonical}
    orphaned = sorted(translated - canonical_keys)
    if orphaned:
        raise SystemExit(f"Korean overlays contain {len(orphaned)} non-canonical ids; first={orphaned[0]}")

    blank_keys = {(s, rid) for s, rid, ja in canonical if ja == ""}
    untranslated = [
        row for row in canonical
        if (row[0], row[1]) not in translated and row[2] != ""
    ]
    if args.order == "shortest":
        untranslated.sort(
            key=lambda row: (
                len(encoded_line(row).encode("utf-8")),
                row[0],
                numeric_id_key(row[1]),
            )
        )

    packet: list[str] = []
    packet_bytes = 0
    first: tuple[int, str] | None = None
    last: tuple[int, str] | None = None
    for section, rid, ja in untranslated:
        line = encoded_line((section, rid, ja))
        n = len(line.encode("utf-8"))
        if packet and (len(packet) >= args.max_records or packet_bytes + n > args.max_bytes):
            break
        packet.append(line)
        packet_bytes += n
        if first is None:
            first = (section, rid)
        last = (section, rid)

    payload = "".join(packet)
    if args.out:
        args.out.parent.mkdir(parents=True, exist_ok=True)
        args.out.write_text(payload, encoding="utf-8")
    else:
        print(payload, end="")

    total = len(canonical)
    blank = len(blank_keys)
    translated_nonblank = len(translated - blank_keys)
    done_effective = translated_nonblank + blank
    remaining_effective = total - done_effective
    progress = {
        "records_total": total,
        "records_translated": translated_nonblank,
        "records_blank_skipped": blank,
        "records_done_effective": done_effective,
        "records_remaining": remaining_effective,
        "percent_done": round(done_effective * 100.0 / total, 4) if total else 100.0,
        "packet_order": args.order,
        "packet_records": len(packet),
        "packet_bytes": packet_bytes,
        "packet_first": {"section": first[0], "id": first[1]} if first else None,
        "packet_last": {"section": last[0], "id": last[1]} if last else None,
    }
    if args.progress:
        args.progress.parent.mkdir(parents=True, exist_ok=True)
        args.progress.write_text(json.dumps(progress, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(progress, ensure_ascii=False), file=__import__("sys").stderr)


if __name__ == "__main__":
    main()
