#!/usr/bin/env python3
"""Build review-only consumer/runtime metadata from pinned English contracts.

The fixed-buffer population mirrors HK47196/zill internal/layout/validate.go at the
pinned English reference. Runtime-pending IDs are consumed from the existing
repository contract audit log; they are never inferred from token spelling here.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import re
import tomllib
from pathlib import Path

RUNTIME_RE = re.compile(r"FORENSIC KOREAN_DIALOGUE_RUNTIME_PENDING reason=([^ ]+) count=(\d+) ids=\[([^\]]*)\] status=PENDING_NOT_PASS")
TRAP_ID = 1070079  # pinned English internal/layout/rules.go
FIXED_CATEGORIES = {"character-creation-choice", "equipment-feedback", "chronicle-entry"}
ITEM_DESCRIPTION_CATEGORIES = {"equipment-description", "item-effect-description", "quest-item-description"}
NARROW_CATEGORIES = {"dialogue", "in-world-guidance"}


def read_bytes(path: Path) -> bytes:
    return path.read_bytes()


def ids(value) -> set[int]:
    return {int(x) for x in (value or [])}


def category_for(ranges, rid: int) -> tuple[str, str]:
    for r in ranges:
        if rid < int(r["first"]):
            break
        if rid <= int(r["last"]):
            return str(r["category"]), str(r["basis"])
    return "uncategorized", "unknown"


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument('--english-root', required=True, type=Path)
    ap.add_argument('--korean-root', default='.', type=Path)
    ap.add_argument('--runtime-log', required=True, type=Path)
    ap.add_argument('--output', required=True, type=Path)
    args = ap.parse_args()

    eroot=args.english_root.resolve(); kroot=args.korean_root.resolve()
    rels=['release/layout/consumer-map.toml','release/layout/categories.toml']
    contract_hash=hashlib.sha256()
    for rel in rels:
        eb=read_bytes(eroot/rel); kb=read_bytes(kroot/rel)
        if eb != kb:
            raise SystemExit(f'Korean layout authority drift from pinned English: {rel}')
        contract_hash.update(rel.encode()); contract_hash.update(b'\0'); contract_hash.update(eb); contract_hash.update(b'\0')

    consumers=tomllib.loads((eroot/rels[0]).read_text(encoding='utf-8'))
    categories=tomllib.loads((eroot/rels[1]).read_text(encoding='utf-8')).get('range',[])
    bounded=ids(consumers.get('bounded_label_ids')); c5=ids(consumers.get('c5_ids'))
    c5portrait=ids(consumers.get('c5_portrait_ids')); c22=ids(consumers.get('c22_ids'))
    single=ids(consumers.get('single_page_c5_ids')); gclient=ids(consumers.get('guild_client_ids'))
    gcomment=ids(consumers.get('guild_commentary_ids')); gregion=ids(consumers.get('guild_region_ids'))
    c20=set(); postings=set()
    for g in consumers.get('c20_group',[]): c20.update(ids(g.get('ids')))
    for p in consumers.get('posting',[]): postings.add(int(p['id']))

    runtime_pending={}
    text=args.runtime_log.read_text(encoding='utf-8',errors='replace')
    for m in RUNTIME_RE.finditer(text):
        reason=m.group(1); listed=[int(x) for x in m.group(3).split() if x.strip()]
        if len(listed) != int(m.group(2)):
            raise SystemExit(f'runtime pending log count mismatch for {reason}')
        for rid in listed: runtime_pending[rid]=reason

    # Accepted IDs are supplied by Korean overlays, not by the English source population.
    accepted=set()
    for path in sorted((kroot/'translations/korean/messages').glob('*.toml')):
        data=tomllib.loads(path.read_text(encoding='utf-8'))
        for rid,row in data.items():
            if isinstance(row,dict) and isinstance(row.get('korean'),str): accepted.add(int(rid))

    entries={}
    for rid in sorted(accepted):
        category,basis=category_for(categories,rid)
        contracts=[]
        if rid in bounded: contracts.append('bounded-label')
        if rid in gclient: contracts.append('guild-client')
        if rid in gregion: contracts.append('guild-region')
        if rid == TRAP_ID: contracts.append('trap')
        if category in FIXED_CATEGORIES: contracts.append(category)
        if rid in c20: contracts.append('c20-group')
        if rid in c22: contracts.append('c22')
        if rid in c5 or rid in single: contracts.append('c5')
        if rid in postings: contracts.append('guild-posting')

        if category in ITEM_DESCRIPTION_CATEGORIES: consumer='item-description'
        elif rid in c22: consumer='c22'
        elif rid in c5portrait: consumer='c5-portrait'
        elif rid in c5: consumer='c5'
        elif rid in single: consumer='c5-single-page'
        elif rid in bounded: consumer='bounded-label'
        elif rid in gclient: consumer='guild-client'
        elif rid in gregion: consumer='guild-region'
        elif rid in gcomment: consumer='guild-commentary'
        elif basis == 'verified' and category in NARROW_CATEGORIES: consumer='verified-narrow-dialogue'
        else: consumer='unproven'

        entries[str(rid)]={
            'category':category,'basis':basis,'consumer':consumer,
            'fixed_buffer':bool(contracts),'fixed_contracts':sorted(set(contracts)),
            'runtime_pending':rid in runtime_pending,
            'runtime_pending_reason':runtime_pending.get(rid,'')
        }

    out={
        'schema_version':1,
        'authority':'HK47196/zill pinned English consumer/category contracts + current Korean runtime coverage audit',
        'english_contract_sha256':contract_hash.hexdigest(),
        'accepted_ids':len(accepted),
        'fixed_buffer_ids':sum(1 for v in entries.values() if v['fixed_buffer']),
        'runtime_pending_ids':sum(1 for v in entries.values() if v['runtime_pending']),
        'entries':entries,
    }
    args.output.parent.mkdir(parents=True,exist_ok=True)
    args.output.write_text(json.dumps(out,ensure_ascii=False,sort_keys=True,separators=(',',':'))+'\n',encoding='utf-8')
    print(json.dumps({k:v for k,v in out.items() if k!='entries'},ensure_ascii=False,indent=2))
    return 0

if __name__=='__main__':
    raise SystemExit(main())
