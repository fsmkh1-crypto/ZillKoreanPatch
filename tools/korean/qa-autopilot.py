#!/usr/bin/env python3
"""Conservative Korean QA autopilot.

Automatically fixes only repeated-source inconsistencies that are provably safe:
1) Korean variants differ only by whitespace / punctuation presentation; or
2) the Japanese source is a standalone fixed label with an exact canonical
   Korean surface in section-0's authoritative label table.

Everything else is reported as blocked for contextual / speaker QA.
"""
from __future__ import annotations

import argparse
from collections import defaultdict, Counter
import json
from pathlib import Path
import re
import tomllib

ROOT = Path(__file__).resolve().parents[2]
KDIR = ROOT / "translations" / "korean" / "messages"
SEED = KDIR / "msgsec000-part99.toml"
FRE = re.compile(r"^msgsec(\d{3})(?:(?:-part\d+)|b)?\.toml$")
HDR = re.compile(r'^\["(?P<id>\d+)"\]$')
KORE = re.compile(r'^korean = (?P<v>".*")$')
WS = re.compile(r"[\s\u3000]+")
# Presentation punctuation only. Runtime controls are deliberately not included.
PUNCT_TRANS = str.maketrans("", "", " .,!?:;~…。！？、，．・…⋯—-―ー'\"“”‘’()[]{}<>")


def norm_source(s: str) -> str:
    return WS.sub("", s.replace("<line-break>", ""))


def norm_variant(s: str) -> str:
    return WS.sub(" ", s).strip()


def semantic_shape(s: str) -> str:
    # Keep letters, Hangul, digits and runtime-token payload characters; erase
    # only presentation whitespace/punctuation. This is intentionally strict.
    return WS.sub("", s).translate(PUNCT_TRANS)


def clean_seed_label(s: str) -> bool:
    if not s.endswith("<end>"):
        return False
    body = s[:-5]
    if not body or "<" in body or ">" in body or "\n" in body:
        return False
    return len(WS.sub("", body)) <= 24


def load_seed() -> dict[str, str]:
    with SEED.open("rb") as f:
        data = tomllib.load(f)
    candidates: dict[str, set[str]] = defaultdict(set)
    for rec in data.values():
        if not isinstance(rec, dict):
            continue
        ja, ko = rec.get("japanese"), rec.get("korean")
        if not isinstance(ja, str) or not isinstance(ko, str):
            continue
        if clean_seed_label(ja) and clean_seed_label(ko):
            candidates[norm_source(ja)].add(norm_variant(ko))
    return {ja: next(iter(kos)) for ja, kos in candidates.items() if len(kos) == 1}


def load_records() -> tuple[dict[str, list[dict[str, object]]], dict[int, tuple[Path, str, str]]]:
    groups: dict[str, list[dict[str, object]]] = defaultdict(list)
    by_id: dict[int, tuple[Path, str, str]] = {}
    seen: set[int] = set()
    for p in sorted(KDIR.glob("msgsec*.toml")):
        m = FRE.match(p.name)
        if not m:
            continue
        with p.open("rb") as f:
            data = tomllib.load(f)
        for rid, rec in data.items():
            try:
                n = int(rid)
            except ValueError:
                continue
            if n in seen:
                continue
            seen.add(n)
            if not isinstance(rec, dict):
                continue
            ja, ko = rec.get("japanese"), rec.get("korean")
            if not isinstance(ja, str) or not isinstance(ko, str):
                continue
            row = {"id": rid, "path": p.name, "japanese": ja, "korean": ko}
            groups[norm_source(ja)].append(row)
            by_id[n] = (p, ja, ko)
    return groups, by_id


def choose_safe_target(key: str, rows: list[dict[str, object]], seed: dict[str, str]) -> tuple[str | None, str]:
    variants = Counter(norm_variant(str(r["korean"])) for r in rows)
    if len(variants) < 2:
        return None, "consistent"

    # Exact standalone canonical label wins.
    canonical = seed.get(key)
    if canonical is not None:
        # Only use the seed if every variant is itself a short standalone label.
        if all(clean_seed_label(str(r["korean"])) for r in rows):
            return canonical, "section0_fixed_label"

    # If stripping only presentation punctuation/space makes every variant
    # identical, choose the most common exact surface; ties are lexical/stable.
    shapes = {semantic_shape(v) for v in variants}
    if len(shapes) == 1:
        target = sorted(variants.items(), key=lambda kv: (-kv[1], kv[0]))[0][0]
        return target, "presentation_only"

    return None, "context_required"


def apply_changes(changes: dict[int, str]) -> tuple[int, int]:
    if not changes:
        return 0, 0
    touched_files = 0
    changed = 0
    ids_by_path: dict[Path, dict[str, str]] = defaultdict(dict)
    for n, target in changes.items():
        # Resolve the canonical first-seen path again to avoid overlay edits.
        for p in sorted(KDIR.glob("msgsec*.toml")):
            if not FRE.match(p.name):
                continue
            with p.open("rb") as f:
                data = tomllib.load(f)
            if str(n) in data:
                ids_by_path[p][str(n)] = target
                break
    for p, wanted in ids_by_path.items():
        lines = p.read_text(encoding="utf-8").splitlines()
        cur = None
        dirty = False
        for i, line in enumerate(lines):
            m = HDR.match(line)
            if m:
                cur = m.group("id")
                continue
            if cur is None or cur not in wanted or not KORE.match(line):
                continue
            lines[i] = "korean = " + json.dumps(wanted[cur], ensure_ascii=False)
            changed += 1
            dirty = True
            cur = None
        if dirty:
            p.write_text("\n".join(lines) + "\n", encoding="utf-8")
            touched_files += 1
    return changed, touched_files


def scan() -> tuple[dict[str, object], dict[int, str]]:
    seed = load_seed()
    groups, _ = load_records()
    changes: dict[int, str] = {}
    safe_groups = []
    blocked = []
    inconsistent = 0
    inconsistent_records = 0
    for key, rows in groups.items():
        if len(rows) < 2:
            continue
        variants = {norm_variant(str(r["korean"])) for r in rows}
        if len(variants) < 2:
            continue
        inconsistent += 1
        inconsistent_records += len(rows)
        target, reason = choose_safe_target(key, rows, seed)
        entry = {
            "normalized_japanese": key,
            "japanese_example": rows[0]["japanese"],
            "occurrences": len(rows),
            "variants": sorted(variants),
            "reason": reason,
        }
        if target is None:
            blocked.append(entry)
            continue
        entry["target"] = target
        safe_groups.append(entry)
        for r in rows:
            if norm_variant(str(r["korean"])) != target:
                changes[int(str(r["id"]))] = target
    report = {
        "schema": 1,
        "inconsistent_groups": inconsistent,
        "inconsistent_records": inconsistent_records,
        "safe_groups": len(safe_groups),
        "safe_records_to_change": len(changes),
        "blocked_groups": len(blocked),
        "safe": safe_groups,
        "blocked": blocked,
    }
    return report, changes


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--apply", action="store_true")
    ap.add_argument("--json", type=Path)
    ap.add_argument("--max-passes", type=int, default=12)
    args = ap.parse_args()

    passes = []
    total_changed = 0
    total_files = 0
    for idx in range(1, max(args.max_passes, 1) + 1):
        report, changes = scan()
        report["pass"] = idx
        passes.append(report)
        print(f"autopilot pass {idx}: inconsistent={report['inconsistent_groups']} safe={report['safe_groups']} blocked={report['blocked_groups']} changes={len(changes)}")
        if not args.apply or not changes:
            break
        changed, files = apply_changes(changes)
        total_changed += changed
        total_files += files
        if changed == 0:
            break
    final, _ = scan()
    output = {
        "schema": 1,
        "applied": bool(args.apply),
        "passes": passes,
        "total_changed_records": total_changed,
        "total_touched_file_events": total_files,
        "final": final,
    }
    if args.json:
        args.json.parent.mkdir(parents=True, exist_ok=True)
        args.json.write_text(json.dumps(output, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"autopilot final: inconsistent={final['inconsistent_groups']} blocked={final['blocked_groups']} changed={total_changed}")


if __name__ == "__main__":
    main()
