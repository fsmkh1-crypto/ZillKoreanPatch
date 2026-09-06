#!/usr/bin/env python3
import argparse,json,re,tomllib
from pathlib import Path
SECTION_RE=re.compile(r'^\["([0-9]+)"\]$')
CONTROL_RE=re.compile(r'<[^>]+>')
def q(s): return json.dumps(s,ensure_ascii=False)
def controls(s): return CONTROL_RE.findall(s)
def main():
 p=argparse.ArgumentParser();p.add_argument('--manifest',required=True,type=Path);p.add_argument('--apply',action='store_true');p.add_argument('--json',type=Path);a=p.parse_args();root=Path(__file__).resolve().parents[2];mp=a.manifest if a.manifest.is_absolute() else root/a.manifest;m=json.loads(mp.read_text(encoding='utf-8'));byfile={}
 for r in m['records']:
  if controls(r['before_korean'])!=controls(r['proposed_korean']): raise SystemExit(f"control topology differs id={r['id']}")
  for path in r['paths']: byfile.setdefault(path,[]).append(r)
 touched=[];changed=0;mismatches=[];file_work={}
 for rel,recs in byfile.items():
  path=root/rel;text=path.read_text(encoding='utf-8');data=tomllib.loads(text);repl={}
  for r in recs:
   rid=str(r['id']);row=data.get(rid)
   if not row: mismatches.append({'id':rid,'path':rel,'error':'missing'});continue
   cur=row.get('korean');before=r['before_korean'];after=r['proposed_korean']
   if cur==after: continue
   if cur!=before:
    mismatches.append({'id':rid,'path':rel,'expected_before':before,'actual':cur});continue
   repl[rid]=after
  file_work[rel]=(path,text,repl)
 if mismatches:
  rendered=json.dumps({'status':'BEFORE_MISMATCH','mismatches':mismatches},ensure_ascii=False,indent=2);print(rendered)
  if a.json: a.json.write_text(rendered+'\n',encoding='utf-8')
  raise SystemExit(1)
 for rel,(path,text,repl) in file_work.items():
  if a.apply and repl:
   lines=text.splitlines(keepends=True);current=None;done=set()
   for i,line in enumerate(lines):
    s=line.rstrip('\r\n');mm=SECTION_RE.match(s)
    if mm: current=mm.group(1);continue
    if current in repl and s.startswith('korean = '):
     nl='\r\n' if line.endswith('\r\n') else ('\n' if line.endswith('\n') else '')
     lines[i]=f'korean = {q(repl[current])}{nl}';done.add(current)
   if done!=set(repl): raise SystemExit(f'write miss path={rel} ids={sorted(set(repl)-done)}')
   rendered=''.join(lines);tomllib.loads(rendered);path.write_text(rendered,encoding='utf-8',newline='');changed+=len(repl);touched.append(rel)
 out={'mode':'apply' if a.apply else 'verify','record_count':len(m['records']),'changed_records':changed,'touched_files':sorted(touched)};s=json.dumps(out,ensure_ascii=False,indent=2);print(s);a.json and a.json.write_text(s+'\n',encoding='utf-8')
if __name__=='__main__': main()
