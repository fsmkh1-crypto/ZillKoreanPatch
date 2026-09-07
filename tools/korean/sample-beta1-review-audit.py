#!/usr/bin/env python3
"""Select a reproducible second-pass sample from valid Beta1 CONTEXT_KEEP rows."""
import argparse
import json
import random
from pathlib import Path


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument('--ledger', type=Path, default=Path('translations/korean/review-ledger.jsonl'))
    ap.add_argument('--size', type=int, default=200)
    ap.add_argument('--seed', default='beta1-final-keep-audit-v1')
    ap.add_argument('--output', type=Path)
    args = ap.parse_args()
    if args.size <= 0:
        raise SystemExit('--size must be positive')

    rows=[]
    for line in args.ledger.read_text(encoding='utf-8').splitlines():
        if not line.strip():
            continue
        row=json.loads(line)
        if row.get('context_state') == 'CONTEXT_KEEP' and row.get('evidence') != 'propagated':
            rows.append(row)
    rows.sort(key=lambda r: (0,int(r['id'])) if str(r['id']).isdigit() else (1,str(r['id'])))
    rng=random.Random(args.seed)
    n=min(args.size,len(rows))
    chosen=rng.sample(rows,n) if n else []
    chosen.sort(key=lambda r: (0,int(r['id'])) if str(r['id']).isdigit() else (1,str(r['id'])))
    out={
        'schema_version':1,
        'purpose':'Beta1 reproducible second-pass accuracy sample',
        'seed':args.seed,
        'requested_size':args.size,
        'eligible_direct_keep':len(rows),
        'sample_size':len(chosen),
        'ids':[str(r['id']) for r in chosen],
        'rule':'Review JP + pinned English + current KO/layout independently; record corrections separately. More than 5% requiring correction triggers expanded KEEP re-review.'
    }
    text=json.dumps(out,ensure_ascii=False,indent=2)+'\n'
    if args.output:
        args.output.parent.mkdir(parents=True,exist_ok=True)
        args.output.write_text(text,encoding='utf-8')
    print(text,end='')
    return 0

if __name__ == '__main__':
    raise SystemExit(main())
