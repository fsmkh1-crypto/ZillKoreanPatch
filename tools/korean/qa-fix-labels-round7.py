#!/usr/bin/env python3
from __future__ import annotations
import json,re,tomllib
from pathlib import Path

ROOT=Path(__file__).resolve().parents[2]
KDIR=ROOT/'translations'/'korean'/'messages'
FRE=re.compile(r'^msgsec(\d{3})(?:(?:-part\d+)|b)?\.toml$')
HDR=re.compile(r'^\["(?P<id>\d+)"\]$')
KORE=re.compile(r'^korean = (?P<v>".*")$')
WS=re.compile(r'[\s\u3000]+')

def norm(s:str)->str:
    return WS.sub('',s.replace('<line-break>',''))

# Exact repeated-source punctuation / elongation normalization only.
# Preserve punctuation or elongation explicitly present in the Japanese source.
CAN={
    norm('あばよ、とっつぁん、<line-break>そして<value:$28>！<line-break>生きて、また会おうぜ！<end>'):
        '잘 있어, 아저씨, 그리고 <value:$28>! 살아서 다시 만나자!<end>',
    norm('ふぅーん。<end>'):'흐~응.<end>',
}

seen=set(); changed=files=0
for p in sorted(KDIR.glob('msgsec*.toml')):
    if not FRE.match(p.name): continue
    with p.open('rb') as f:data=tomllib.load(f)
    lines=p.read_text(encoding='utf-8').splitlines();cur=None;dirty=False
    for i,line in enumerate(lines):
        m=HDR.match(line)
        if m:cur=m.group('id');continue
        if cur is None or not KORE.match(line):continue
        try:n=int(cur)
        except ValueError:cur=None;continue
        if n in seen:cur=None;continue
        seen.add(n);rec=data.get(cur);cur=None
        if not isinstance(rec,dict):continue
        ja=rec.get('japanese');ko=rec.get('korean')
        if not isinstance(ja,str) or not isinstance(ko,str):continue
        target=CAN.get(norm(ja))
        if target is not None and ko!=target:
            lines[i]='korean = '+json.dumps(target,ensure_ascii=False);dirty=True;changed+=1
    if dirty:
        p.write_text('\n'.join(lines)+'\n',encoding='utf-8');files+=1
print(f'Label round 7: {changed} records across {files} files')
