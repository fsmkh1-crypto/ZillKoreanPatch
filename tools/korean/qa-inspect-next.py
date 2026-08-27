#!/usr/bin/env python3
from __future__ import annotations
from collections import Counter, defaultdict
from pathlib import Path
import re, tomllib

ROOT=Path(__file__).resolve().parents[2]
KDIR=ROOT/'translations'/'korean'/'messages'
FRE=re.compile(r'^msgsec(\d{3})(?:(?:-part\d+)|b)?\.toml$')
TERMS=['バルザー','ファドル','アーギルシャイア','ゾフォル','ダイダロ','ジンガ']
seen=set(); term_rows=defaultdict(list)
for p in sorted(KDIR.glob('msgsec*.toml')):
    if not FRE.match(p.name): continue
    with p.open('rb') as f: data=tomllib.load(f)
    for rid,rec in data.items():
        try: n=int(rid)
        except: continue
        if n in seen: continue
        seen.add(n)
        ja=rec.get('japanese'); ko=rec.get('korean')
        if not isinstance(ja,str) or not isinstance(ko,str): continue
        for t in TERMS:
            if t in ja:
                term_rows[t].append((rid,p.name,ja,ko))
for t in TERMS:
    rows=term_rows[t]
    print(f'## TERM {t} occurrences={len(rows)}')
    c=Counter()
    for _,_,ja,ko in rows:
        c[ko]+=1
    for ko,n in c.most_common(20): print(f'{n:4d} | {ko}')
    for rid,p,ja,ko in rows[:30]: print(f'  {rid} {p} | {ja} | {ko}')

# Print top current QA-3 groups in compact full form.
import subprocess, json, tempfile
out=Path('/tmp/qa3.json')
subprocess.run(['python3',str(ROOT/'tools/korean/qa-consistency.py'),'--json',str(out),'--max-examples','0'],check=True)
r=json.loads(out.read_text(encoding='utf-8'))
print(f"## QA3 groups={r['inconsistent_groups']} records={r['inconsistent_records']}")
for i,g in enumerate(r['groups'][:80],1):
    print(f"### {i} occ={g['occurrences']} variants={g['variant_count']} JA={g['japanese_example']}")
    for v in g['variants']:
        ids=','.join(str(x['id']) for x in v['records'][:12])
        print(f"  {v['count']} | {v['korean']} | ids={ids}")
