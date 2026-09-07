#!/usr/bin/env python3
"""Attach engine-derived consumer/runtime flags and strict group statistics to ledger v2."""
from __future__ import annotations

import argparse, hashlib, json, tomllib
from collections import defaultdict
from pathlib import Path


def row_signature(info: dict) -> str:
    payload={
        'category':info.get('category',''),'basis':info.get('basis',''),'consumer':info.get('consumer',''),
        'fixed_buffer':bool(info.get('fixed_buffer')),'fixed_contracts':info.get('fixed_contracts',[]),
        'runtime_pending':bool(info.get('runtime_pending')),'runtime_pending_reason':info.get('runtime_pending_reason','')
    }
    return hashlib.sha256(json.dumps(payload,sort_keys=True,separators=(',',':')).encode()).hexdigest()


def load_current(root: Path):
    multi=defaultdict(list)
    for path in sorted((root/'translations/korean/messages').glob('*.toml')):
        data=tomllib.loads(path.read_text(encoding='utf-8')); rel=path.relative_to(root).as_posix()
        for rid,row in data.items():
            if isinstance(row,dict) and isinstance(row.get('korean'),str):
                multi[str(rid)].append((rel,row))
    out={}
    for rid,items in multi.items():
        first=items[0][1]
        for _,row in items[1:]:
            for k in ('japanese','korean','layout','consumer'):
                if row.get(k,'') != first.get(k,''): raise SystemExit(f'alias drift id={rid} key={k}')
        out[rid]={
            'japanese':first.get('japanese',''),'korean':first.get('korean',''),'layout':first.get('layout',''),
            'raw_consumer':first.get('consumer',''),'paths':sorted(p for p,_ in items),'alias_count':len(items)
        }
    return out


def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--consumer-map',required=True,type=Path); ap.add_argument('--root',default='.',type=Path)
    ap.add_argument('--ledger',default='translations/korean/review-ledger.jsonl',type=Path)
    ap.add_argument('--groups',default='docs/audit/beta1-review-groups.json',type=Path)
    ap.add_argument('--summary',default='docs/audit/beta1-review-coverage.md',type=Path); a=ap.parse_args()
    root=a.root.resolve(); cmap=json.loads(a.consumer_map.read_text(encoding='utf-8')); infos=cmap['entries']
    current=load_current(root)
    if len(current)!=int(cmap['accepted_ids']): raise SystemExit('consumer map accepted population mismatch')

    ledger_path=root/a.ledger; ledger=[]; flag_counts=defaultdict(int)
    for line in ledger_path.read_text(encoding='utf-8').splitlines():
        if not line.strip(): continue
        row=json.loads(line); rid=str(row['id']); info=infos.get(rid)
        if info is None: raise SystemExit(f'consumer map lacks ledger id {rid}')
        flags=[f for f in row.get('flags',[]) if f not in ('FIXED_BUFFER','RUNTIME_PENDING')]
        if info.get('fixed_buffer'): flags.append('FIXED_BUFFER')
        if info.get('runtime_pending'): flags.append('RUNTIME_PENDING')
        row['flags']=sorted(set(flags)); row['engine_consumer']={
            'category':info['category'],'basis':info['basis'],'consumer':info['consumer'],
            'fixed_contracts':info.get('fixed_contracts',[]),'runtime_pending_reason':info.get('runtime_pending_reason','')
        }
        row['english_consumer_contract_sha256']=cmap['english_contract_sha256']
        row['engine_consumer_signature']=row_signature(info)
        for f in row['flags']: flag_counts[f]+=1
        ledger.append(row)
    ledger_path.write_text(''.join(json.dumps(r,ensure_ascii=False,sort_keys=True,separators=(',',':'))+'\n' for r in ledger),encoding='utf-8')

    groups=defaultdict(list)
    for rid,row in current.items():
        info=infos[rid]
        engine_sig=row_signature(info)
        storage_sig=hashlib.sha256(json.dumps({
            'paths':row['paths'],'alias_count':row['alias_count'],'raw_consumer':row['raw_consumer'],
            'engine_consumer_signature':engine_sig
        },sort_keys=True,separators=(',',':')).encode()).hexdigest()
        key=(row['japanese'],row['korean'],row['layout'],storage_sig)
        groups[key].append(rid)
    dups=[v for v in groups.values() if len(v)>1]
    gd={
        'version':3,'authority':'exact JP+KO+layout + pinned-English engine consumer signature + physical alias/storage signature',
        'english_consumer_contract_sha256':cmap['english_contract_sha256'],'accepted_ids':len(current),
        'unique_strict_signatures':len(groups),'strict_duplicate_groups':len(dups),
        'strict_duplicate_member_ids':sum(len(v) for v in dups),
        'strict_propagation_candidate_extra_ids':sum(len(v)-1 for v in dups),
        'fixed_buffer_ids':int(cmap['fixed_buffer_ids']),'runtime_pending_ids':int(cmap['runtime_pending_ids']),
        'counting_policy':'Candidate-only. Propagation remains zero until a directly reviewed representative and exact group match are recorded.'
    }
    (root/a.groups).write_text(json.dumps(gd,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')

    sp=root/a.summary; text=sp.read_text(encoding='utf-8')
    start=text.index('## Orthogonal flags currently derived')
    end=text.index('## Historical edit manifests')
    replacement='''## Orthogonal flags and strict propagation candidates\n\n- `PERSISTED_LAYOUT`: **%d** among ledger rows\n- `ALIAS_GROUP`: **%d** among ledger rows\n- `SOURCE_ANOMALY`: **%d** among ledger rows\n- `FIXED_BUFFER`: **%d** among ledger rows; full accepted population **%d**\n- `RUNTIME_PENDING`: **%d** among ledger rows; full accepted population **%d**\n- English consumer/category contract SHA-256: `%s`\n\nStrict propagation candidates are derived from exact JP + exact KO + exact persisted layout + pinned-English engine consumer signature + physical alias/storage signature. They are candidates only and contribute zero coverage until explicitly propagated from a directly reviewed representative.\n\n- Unique strict signatures: **%d**\n- Duplicate strict-signature groups: **%d**\n- IDs inside such groups: **%d**\n- Potential extra IDs: **%d**\n\n''' % (
        flag_counts['PERSISTED_LAYOUT'],flag_counts['ALIAS_GROUP'],flag_counts['SOURCE_ANOMALY'],
        flag_counts['FIXED_BUFFER'],int(cmap['fixed_buffer_ids']),flag_counts['RUNTIME_PENDING'],int(cmap['runtime_pending_ids']),
        cmap['english_contract_sha256'],len(groups),len(dups),sum(len(v) for v in dups),sum(len(v)-1 for v in dups))
    sp.write_text(text[:start]+replacement+text[end:],encoding='utf-8')
    print(json.dumps({'status':'PASS','ledger_rows':len(ledger),'fixed_buffer_population':cmap['fixed_buffer_ids'],
                      'runtime_pending_population':cmap['runtime_pending_ids'],'strict_duplicate_groups':len(dups),
                      'strict_propagation_candidate_extra_ids':sum(len(v)-1 for v in dups)},indent=2))
    return 0

if __name__=='__main__': raise SystemExit(main())
