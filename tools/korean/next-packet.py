#!/usr/bin/env python3
"""Emit the next untranslated Korean work packet from canonical corpus.

The translated set is derived from the checked-in Korean overlays, so resume state
never depends on a manually maintained cursor. Output is compact JSONL containing
only section/id/Japanese source. A progress JSON file can be written alongside it.
"""
from __future__ import annotations

import argparse
import json
from pathlib import Path
import tomllib

ROOT = Path(__file__).resolve().parents[2]
CANON = ROOT / "translations" / "messages"
KOREAN = ROOT / "translations" / "korean" / "messages"


def load_translated() -> set[tuple[int, str]]:
    translated: set[tuple[int, str]] = set()
    owners: dict[tuple[int, str], Path] = {}
    for path in sorted(KOREAN.glob("msgsec*.toml")):
        stem = path.stem
        # Accept msgsec006.toml and msgsec003-part05.toml alike.
        prefix = stem.split("-", 1)[0]
        section = int(prefix.removeprefix("msgsec"))
        with path.open("rb") as f:
            data = tomllib.load(f)
        for rid, rec in data.items():
            if not isinstance(rec, dict) or not rec.get("korean"):
                continue
            key = (section, str(rid))
            if key in translated:
                raise SystemExit(
                    f"duplicate Korean id {section}/{rid}: {owners[key]} and {path}"
                )
            translated.add(key)
            owners[key] = path
    return translated


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


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--max-records", type=int, default=1500)
    ap.add_argument("--max-bytes", type=int, default=400_000)
    ap.add_argument("--out", type=Path)
    ap.add_argument("--progress", type=Path)
    args = ap.parse_args()
    if args.max_records < 1 or args.max_bytes < 1:
        raise SystemExit("packet limits must be positive")

    translated = load_translated()
    canonical = canonical_rows()
    packet: list[str] = []
    packet_bytes = 0
    first: tuple[int, str] | None = None
    last: tuple[int, str] | None = None

    for section, rid, ja in canonical:
        if (section, rid) in translated:
            continue
        line = json.dumps(
            {"section": section, "id": rid, "japanese": ja},
            ensure_ascii=False,
            separators=(",", ":"),
        ) + "\n"
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
    done = len(translated)
    # A translated ID not present in canonical indicates repository drift.
    canonical_keys = {(s, rid) for s, rid, _ in canonical}
    orphaned = sorted(translated - canonical_keys)
    if orphaned:
        raise SystemExit(f"Korean overlays contain {len(orphaned)} non-canonical ids; first={orphaned[0]}")
    progress = {
        "records_total": total,
        "records_done": done,
        "records_remaining": total - done,
        "percent_done": round(done * 100.0 / total, 4) if total else 100.0,
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
