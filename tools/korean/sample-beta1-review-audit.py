#!/usr/bin/env python3
"""Select a reproducible, propagation-sensitive second-pass Beta1 KEEP sample."""
from __future__ import annotations

import argparse
import json
import random
from pathlib import Path


def key(r):
    rid = str(r["id"])
    return (0, int(rid)) if rid.isdigit() else (1, rid)


def choose(rng, rows, n):
    n = min(max(n, 0), len(rows))
    return rng.sample(rows, n) if n else []


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--ledger", type=Path, default=Path("translations/korean/review-ledger.jsonl"))
    ap.add_argument("--size", type=int, default=200)
    ap.add_argument("--seed", default="beta1-final-keep-audit-v2")
    ap.add_argument("--propagated-share", type=float, default=0.50,
                    help="target sample share reserved for propagated KEEP; deliberately above natural share when available")
    ap.add_argument("--second-reviewer", required=True,
                    help="identity of second-pass reviewer/model; must differ from every sampled first-pass reviewer when known")
    ap.add_argument("--output", type=Path)
    args = ap.parse_args()
    if args.size <= 0 or not 0 <= args.propagated_share <= 1:
        raise SystemExit("invalid size/share")

    direct = []; propagated = []
    for line in args.ledger.read_text(encoding="utf-8").splitlines():
        if not line.strip():
            continue
        row = json.loads(line)
        if row.get("context_state") != "CONTEXT_KEEP":
            continue
        first = row.get("reviewer")
        if first and first == args.second_reviewer:
            raise SystemExit(f"second reviewer equals first reviewer for eligible KEEP id={row['id']}")
        (propagated if row.get("evidence") == "propagated" else direct).append(row)
    direct.sort(key=key); propagated.sort(key=key)

    rng = random.Random(args.seed)
    prop_target = round(args.size * args.propagated_share)
    prop_pick = choose(rng, propagated, prop_target)
    remaining = args.size - len(prop_pick)
    direct_pick = choose(rng, direct, remaining)
    remaining = args.size - len(prop_pick) - len(direct_pick)
    if remaining:
        # If direct population is too small, fill from propagated without duplicates.
        rest = [r for r in propagated if r not in prop_pick]
        prop_pick += choose(rng, rest, remaining)

    chosen = sorted(direct_pick + prop_pick, key=key)
    out = {
        "schema_version": 2,
        "purpose": "Beta1 reproducible independent second-pass KEEP accuracy sample",
        "seed": args.seed,
        "second_reviewer": args.second_reviewer,
        "requested_size": args.size,
        "target_propagated_share": args.propagated_share,
        "eligible_direct_keep": len(direct),
        "eligible_propagated_keep": len(propagated),
        "sample_size": len(chosen),
        "sample_direct_keep": len(direct_pick),
        "sample_propagated_keep": len(prop_pick),
        "ids": [str(r["id"]) for r in chosen],
        "rule": "Second reviewer/model must differ from first pass. Re-read JP + pinned English + current KO independently. Propagated KEEP is deliberately over-sampled. More than 5% requiring correction triggers expanded KEEP re-review.",
    }
    text = json.dumps(out, ensure_ascii=False, indent=2) + "\n"
    if args.output:
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(text, encoding="utf-8")
    print(text, end="")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
