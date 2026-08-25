#!/usr/bin/env python3
"""Build compact, GitHub-friendly JSONL source chunks for bulk translation.

The chunk files intentionally contain Japanese only (plus section/id), omitting the
English field and TOML syntax. They are small enough for repository fetch APIs to
return whole chunks without truncating giant canonical section files.
"""
from __future__ import annotations

import argparse
import json
from pathlib import Path
import shutil
import tomllib

ROOT = Path(__file__).resolve().parents[2]
MESSAGES = ROOT / "translations" / "messages"
DEFAULT_OUT = ROOT / "work" / "korean-source"


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--out", type=Path, default=DEFAULT_OUT)
    ap.add_argument("--max-bytes", type=int, default=48_000,
                    help="approximate maximum UTF-8 bytes per chunk")
    ap.add_argument("--section", type=int, action="append",
                    help="limit to section; may be repeated")
    args = ap.parse_args()

    if args.max_bytes < 8_000:
        raise SystemExit("--max-bytes must be >= 8000")

    if args.out.exists():
        shutil.rmtree(args.out)
    args.out.mkdir(parents=True)

    wanted = set(args.section or [])
    manifest: list[dict[str, object]] = []
    total_records = 0

    for path in sorted(MESSAGES.glob("msgsec*.toml")):
        section = int(path.stem.removeprefix("msgsec"))
        if wanted and section not in wanted:
            continue
        with path.open("rb") as f:
            data = tomllib.load(f)

        chunk_no = 1
        buf: list[str] = []
        size = 0
        first_id: str | None = None
        last_id: str | None = None

        def flush() -> None:
            nonlocal chunk_no, buf, size, first_id, last_id
            if not buf:
                return
            name = f"msgsec{section:03d}-chunk{chunk_no:03d}.jsonl"
            payload = "".join(buf)
            (args.out / name).write_text(payload, encoding="utf-8")
            manifest.append({
                "file": name,
                "section": section,
                "first_id": first_id,
                "last_id": last_id,
                "records": len(buf),
                "bytes": len(payload.encode("utf-8")),
            })
            chunk_no += 1
            buf = []
            size = 0
            first_id = last_id = None

        for rid, rec in data.items():
            ja = rec.get("japanese")
            if ja is None:
                continue
            line = json.dumps({"section": section, "id": rid, "japanese": ja},
                              ensure_ascii=False, separators=(",", ":")) + "\n"
            line_size = len(line.encode("utf-8"))
            if buf and size + line_size > args.max_bytes:
                flush()
            if first_id is None:
                first_id = rid
            last_id = rid
            buf.append(line)
            size += line_size
            total_records += 1
        flush()

    (args.out / "manifest.json").write_text(
        json.dumps({"records": total_records, "chunks": manifest}, ensure_ascii=False,
                   indent=2) + "\n",
        encoding="utf-8",
    )
    print(f"built {len(manifest)} chunks / {total_records} records in {args.out}")


if __name__ == "__main__":
    main()
