#!/usr/bin/env python3
from __future__ import annotations
import argparse, json, re, tomllib
from pathlib import Path

PIN = "a98d9ce29f361d666ec23da0dcfd351f24537ffd"
SEC_RE = re.compile(r"msgsec(\d+)-part(\d+)\.toml$")

def key_path(p: Path):
    m=SEC_RE.search(p.name)
    return (int(m.group(1)), int(m.group(2))) if m else (10**9,10**9)

def key_id(x: str):
    return (0,int(x)) if x.isdigit() else (1,x)

def load_english(root: Path):
    out={}
    for p in sorted((root/"translations/messages").glob("msgsec*.toml"), key=key_path):
        d=tomllib.loads(p.read_text(encoding="utf-8"))
        for rid,row in d.items():
            if isinstance(row,dict): out[str(rid)]={"japanese":row.get("japanese",""),"english":row.get("english","")}
    return out

def main():
    ap=argparse.ArgumentParser()
    ap.add_argument("--root",type=Path,default=Path("."))
    ap.add_argument("--english-root",type=Path,required=True)
    ap.add_argument("--start-id",required=True)
    ap.add_argument("--count",type=int,default=430)
    ap.add_argument("--output",type=Path,required=True)
    ap.add_argument("--selection-output",type=Path,required=True)
    a=ap.parse_args(); root=a.root.resolve(); eroot=a.english_root.resolve()
    covered=set()
    ledger=root/"translations/korean/review-ledger.jsonl"
    for line in ledger.read_text(encoding="utf-8").splitlines():
        if line.strip():
            row=json.loads(line); covered.add(str(row["id"]))
    eng=load_english(eroot)
    start=int(a.start_id)
    chosen=[]
    for p in sorted((root/"translations/korean/messages").glob("msgsec*.toml"), key=key_path):
        d=tomllib.loads(p.read_text(encoding="utf-8"))
        for rid,row in sorted(((str(k),v) for k,v in d.items() if isinstance(v,dict) and isinstance(v.get("korean"),str)), key=lambda kv:key_id(kv[0])):
            if not rid.isdigit() or int(rid)<start or rid in covered: continue
            er=eng.get(rid,{"japanese":"","english":""})
            chosen.append({"id":rid,"source_path":str(p.relative_to(root)),"japanese":row.get("japanese",""),"english":er.get("english",""),"korean":row.get("korean",""),"jp_matches_pinned_english":row.get("japanese","")==er.get("japanese","")})
            if len(chosen)>=a.count: break
        if len(chosen)>=a.count: break
    if len(chosen)<a.count: raise SystemExit(f"only {len(chosen)} candidates available")
    out=["# Beta1 next-review candidate packet","",f"start_id: {a.start_id}",f"candidate_count: {len(chosen)}",f"pinned_english_sha: {PIN}","order: source-file then numeric ID","note: candidate packet only; creates no review coverage.",""]
    for r in chosen:
        q=lambda s:json.dumps(s,ensure_ascii=False,separators=(",",":"))
        out += [f"## {r['id']}",f"source: {r['source_path']}",f"jp_matches_pinned_english: {'yes' if r['jp_matches_pinned_english'] else 'NO'}",f"JP: {q(r['japanese'])}",f"EN: {q(r['english'])}",f"KO: {q(r['korean'])}",""]
    op=a.output if a.output.is_absolute() else root/a.output; op.parent.mkdir(parents=True,exist_ok=True); op.write_text("\n".join(out)+"\n",encoding="utf-8")
    sp=a.selection_output if a.selection_output.is_absolute() else root/a.selection_output; sp.parent.mkdir(parents=True,exist_ok=True); sp.write_text(json.dumps({"schema_version":1,"start_id":a.start_id,"candidate_count":len(chosen),"pinned_english_sha":PIN,"ids":[r["id"] for r in chosen],"source_paths":sorted(set(r["source_path"] for r in chosen))},ensure_ascii=False,indent=2)+"\n",encoding="utf-8")
    print(json.dumps({"status":"PASS","count":len(chosen),"first":chosen[0]["id"],"last":chosen[-1]["id"],"files":sorted(set(r["source_path"] for r in chosen))},ensure_ascii=False,indent=2))
if __name__=="__main__": main()
