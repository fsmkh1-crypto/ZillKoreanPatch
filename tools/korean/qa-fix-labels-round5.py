#!/usr/bin/env python3
from __future__ import annotations
import json,re,tomllib
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
KDIR=ROOT/'translations'/'korean'/'messages'
FRE=re.compile(r'^msgsec(\d{3})(?:(?:-part\d+)|b)?\.toml$')
HDR=re.compile(r'^\["(?P<id>\d+)"\]$')
KORE=re.compile(r'^korean = (?P<v>".*")$')
# Low-risk only: identical repeated Japanese source with punctuation/whitespace drift.
TARGETS={
 '560017':('くっ…。<end>','큭….<end>'),
 '560130':('うっ…。<end>','윽….<end>'),
 '560159':('うう…。<end>','으으….<end>'),
 '1340296':('う…う…。<end>','으…으….<end>'),
}
seen=set(); changed=files=0
for p in sorted(KDIR.glob('msgsec*.toml')):
 if not FRE.match(p.name): continue
 with p.open('rb') as f: data=tomllib.load(f)
 lines=p.read_text(encoding='utf-8').splitlines(); cur=None; dirty=False
 for i,line in enumerate(lines):
  m=HDR.match(line)
  if m: cur=m.group('id'); continue
  if cur is None or not KORE.match(line): continue
  rid=cur; cur=None
  if rid in seen: continue
  seen.add(rid)
  spec=TARGETS.get(rid)
  if spec is None: continue
  rec=data.get(rid)
  if not isinstance(rec,dict): raise SystemExit(f'{p.name}:{rid}: missing record')
  ja=rec.get('japanese'); ko=rec.get('korean'); expected_ja,target=spec
  if ja!=expected_ja: raise SystemExit(f'{p.name}:{rid}: source changed: {ja!r}')
  if ko!=target:
   lines[i]='korean = '+json.dumps(target,ensure_ascii=False); dirty=True; changed+=1
 if dirty:
  p.write_text('\n'.join(lines)+'\n',encoding='utf-8'); files+=1
missing=set(TARGETS)-seen
if missing: raise SystemExit(f'missing target ids: {sorted(missing)}')
print(f'Label round 5: {changed} records across {files} files')
