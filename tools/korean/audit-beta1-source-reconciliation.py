#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""Account for every source ID; do not equate the packet filter with localization.

Read-only unless --json is provided. Categories describe repository contents,
not retail reachability. Numeric ranges in the report are inclusive and exact.
"""
import argparse
import collections
import importlib.util
import json
from pathlib import Path
import re
import tomllib
import unicodedata

ROOT = Path(__file__).resolve().parents[2]
spec = importlib.util.spec_from_file_location('next_packet', Path(__file__).with_name('next-packet.py'))
packet = importlib.util.module_from_spec(spec)
spec.loader.exec_module(packet)


def load(folder, field):
    rows, owners, duplicates = {}, {}, []
    for path in sorted(folder.glob('msgsec*.toml')):
        section = packet.section_from_path(path)
        for rid, row in tomllib.loads(path.read_text()).items():
            if not isinstance(row, dict) or field not in row:
                continue
            key = (section, str(rid))
            if key in rows:
                if rows[key] != row:
                    raise ValueError(f'conflicting duplicate: {key}')
                duplicates.append({'section': section, 'id': int(rid), 'first': owners[key], 'alias': str(path.relative_to(ROOT))})
            rows[key], owners[key] = row, str(path.relative_to(ROOT))
    return rows, owners, duplicates


def ranges(ids):
    out = []
    for rid in sorted(ids):
        if out and rid == out[-1][1] + 1:
            out[-1][1] = rid
        else:
            out.append([rid, rid])
    return out


def classify(section, rid, ja):
    # These exact synthetic markers are explicitly excluded by source_markers.py.
    if packet.is_synthetic_source(ja):
        return 'synthetic_unused_marker'
    if not ja.strip():
        return 'no_visible_text'
    visible = packet.visible_source_text(ja)
    if not visible:
        return 'runtime_control_only'
    normalized = unicodedata.normalize('NFKC', visible)
    if normalized.startswith('\\DATA\\') or normalized.startswith('\\TIM\\'):
        return 'technical_resource_path'
    if section == 196 and (rid == 1960476 or 1960484 <= rid <= 1960496):
        return 'technical_effect_label'
    if normalized == 'Dummy':
        return 'technical_dummy_label'
    if visible == 'ＺＺＺＺＺＺ…。':
        return 'intentional_sleep_symbol'
    if re.fullmatch(r'[\W\d_]+', normalized):
        return 'intentional_punctuation_numeric_passthrough'
    if packet.JAPANESE_SCRIPT_RE.search(visible):
        return 'untranslated_japanese_review'
    # Latin-only prose/title is NOT automatically declared a safe skip.
    return 'latin_title_or_text_review'


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--json', type=Path)
    args = ap.parse_args()
    source, owners, source_duplicates = load(ROOT/'translations/messages', 'japanese')
    korean, _, korean_duplicates = load(ROOT/'translations/korean/messages', 'korean')
    assert not korean.keys() - source.keys(), 'orphan Korean IDs'
    groups = collections.defaultdict(list)
    missing = []
    for key, row in sorted(source.items()):
        section, sid = key
        rid = int(sid)
        if key in korean:
            assert korean[key]['japanese'] == row['japanese'], f'source drift {key}'
            assert korean[key]['korean'], f'empty Korean {key}'
            category = 'accepted_korean'
        else:
            category = classify(section, rid, row['japanese'])
            missing.append({'section': section, 'id': rid, 'category': category, 'japanese': row['japanese'], 'english': row.get('english', ''), 'source_file': owners[key]})
        groups[category].append(rid)
    counts = {k: len(v) for k, v in sorted(groups.items())}
    assert sum(counts.values()) == len(source)
    all_ids = [i for v in groups.values() for i in v]
    assert len(set(all_ids)) == len(source), 'cross-section ID collision'
    old_no_text = sum(not packet.visible_source_text(r['japanese']) for r in source.values())
    old_passthrough = sum(bool(packet.visible_source_text(r['japanese'])) and not packet.JAPANESE_SCRIPT_RE.search(packet.visible_source_text(r['japanese'])) for r in source.values())
    accepted_filter_skips = sum(not packet.needs_translation(source[k]['japanese']) for k in korean)
    report = {'schema': 1, 'source_records': len(source), 'accepted_korean': len(korean),
              'missing_overlay': len(missing), 'counts': counts,
              'category_id_ranges_inclusive': {k: ranges(v) for k, v in sorted(groups.items())},
              'missing_records': missing, 'source_duplicates': source_duplicates,
              'identical_korean_aliases': korean_duplicates,
              'historical_packet_filter_on_current_source': {'no_text': old_no_text, 'passthrough': old_passthrough, 'already_accepted_overlap': accepted_filter_skips},
              'evidence_boundary': 'Repository contents and packet-filter reconciliation only. REVIEW titles require reachability/localization decision. No retail or runtime reachability claim.'}
    if args.json:
        args.json.parent.mkdir(parents=True, exist_ok=True)
        args.json.write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps({k: report[k] for k in ('source_records', 'accepted_korean', 'missing_overlay', 'counts', 'historical_packet_filter_on_current_source')}, ensure_ascii=False))
    print('exact_total_reconciliation=PASS; runtime_reachability=PENDING')


if __name__ == '__main__':
    main()
