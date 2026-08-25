#!/usr/bin/env python3
"""Emit high-confidence exact Korean translations from a work packet.

This intentionally handles only context-free literals whose Korean rendering is
stable. Everything else remains for LLM translation. Runtime control tokens are
kept verbatim by mapping the complete canonical string.
"""
from __future__ import annotations

import argparse
import json
from pathlib import Path

EXACT = {
    "未使用<end>": "미사용<end>",
    "<未使用><end>": "<미사용><end>",
    "＜無＞<end>": "＜없음＞<end>",
    "終了<end>": "종료<end>",
    "　予備<end>": "　예비<end>",
    "予備<end>": "예비<end>",
    "宿屋<end>": "여관<end>",
    " 魔   法<end>": " 마   법<end>",
    " 能力値<end>": " 능력치<end>",
    " 装   備<end>": " 장   비<end>",
    " スキル<end>": " 스킬<end>",
    " ソウル<end>": " 소울<end>",
    " 情   報<end>": " 정   보<end>",
}


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("packet", type=Path)
    ap.add_argument("--out", type=Path)
    args = ap.parse_args()

    rows: list[str] = []
    with args.packet.open(encoding="utf-8") as f:
        for lineno, line in enumerate(f, 1):
            if not line.strip():
                continue
            obj = json.loads(line)
            ja = str(obj["japanese"])
            ko = EXACT.get(ja)
            if ko is None:
                continue
            rows.append(json.dumps(
                {"section": int(obj["section"]), "id": str(obj["id"]), "korean": ko},
                ensure_ascii=False,
                separators=(",", ":"),
            ) + "\n")

    payload = "".join(rows)
    if args.out:
        args.out.parent.mkdir(parents=True, exist_ok=True)
        args.out.write_text(payload, encoding="utf-8")
    else:
        print(payload, end="")
    print(f"auto-trivial: {len(rows)} rows", file=__import__("sys").stderr)


if __name__ == "__main__":
    main()
