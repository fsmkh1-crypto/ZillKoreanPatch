#!/usr/bin/env python3
import argparse
import json
import re
import tomllib
from collections import Counter, defaultdict
from pathlib import Path

CONTROL_RE = re.compile(r"<[^>]+>")
SECTION_RE = re.compile(r'^\["([0-9]+)"\]$')


def controls(text: str) -> list[str]:
    return CONTROL_RE.findall(text)


def load_decisions(path: Path) -> list[dict]:
    data = json.loads(path.read_text(encoding="utf-8"))
    decisions = data.get("decisions")
    if not isinstance(decisions, list) or not decisions:
        raise ValueError("decisions file has no decisions")
    seen = set()
    for decision in decisions:
        jp = decision["japanese"]
        ko = decision["korean"]
        legacy = decision.get("legacy", [])
        if jp in seen:
            raise ValueError(f"duplicate Japanese decision: {jp}")
        seen.add(jp)
        if not ko:
            raise ValueError(f"empty canonical Korean for {jp}")
        if ko in legacy:
            raise ValueError(f"canonical Korean repeated in legacy list for {jp}")
        if len(legacy) != len(set(legacy)):
            raise ValueError(f"duplicate legacy Korean values for {jp}")
    return decisions


def classify_text(korean: str, canonical: str, legacy: list[str]) -> tuple[str, str | None]:
    canonical_hit = canonical in korean
    legacy_hits = [value for value in legacy if value in korean]
    if canonical_hit and legacy_hits:
        raise ValueError(
            f"ambiguous row contains both canonical {canonical!r} and legacy {legacy_hits!r}: {korean!r}"
        )
    if len(legacy_hits) > 1:
        raise ValueError(f"ambiguous row contains multiple legacy forms {legacy_hits!r}: {korean!r}")
    if canonical_hit:
        return "canonical", canonical
    if legacy_hits:
        return "legacy", legacy_hits[0]
    return "unrecognized", None


def rewrite_overlay(path: Path, decisions: list[dict], apply: bool) -> dict:
    text = path.read_text(encoding="utf-8")
    data = tomllib.loads(text)
    replacements: dict[str, str] = {}
    stats = Counter()
    term_counts = defaultdict(Counter)
    changes = []

    for rid, row in data.items():
        if not isinstance(row, dict):
            continue
        japanese = row.get("japanese", "")
        korean = row.get("korean", "")
        if not isinstance(japanese, str) or not isinstance(korean, str):
            continue
        current = korean
        matched_any = False
        for decision in decisions:
            jp = decision["japanese"]
            if jp not in japanese:
                continue
            matched_any = True
            canonical = decision["korean"]
            legacy = decision.get("legacy", [])
            kind, value = classify_text(current, canonical, legacy)
            term_counts[jp][kind] += 1
            if kind == "unrecognized":
                raise ValueError(
                    f"{path}:{rid}: source contains {jp!r} but Korean contains neither "
                    f"canonical {canonical!r} nor known legacy {legacy!r}: {current!r}"
                )
            if kind == "legacy":
                before = current
                current = current.replace(value, canonical)
                if controls(before) != controls(current):
                    raise ValueError(f"{path}:{rid}: control topology changed during {jp} migration")
                changes.append({
                    "id": str(rid),
                    "japanese_term": jp,
                    "legacy": value,
                    "canonical": canonical,
                    "before_korean": before,
                    "after_korean": current,
                })
                stats["term_replacements"] += 1
        if matched_any:
            stats["source_matched_records"] += 1
        if current != korean:
            replacements[str(rid)] = current
            stats["changed_records"] += 1

    if apply and replacements:
        lines = text.splitlines(keepends=True)
        current_id = None
        rewritten = set()
        for index, line in enumerate(lines):
            stripped = line.rstrip("\r\n")
            match = SECTION_RE.match(stripped)
            if match:
                current_id = match.group(1)
                continue
            if current_id in replacements and stripped.startswith('korean = "'):
                newline = "\r\n" if line.endswith("\r\n") else ("\n" if line.endswith("\n") else "")
                lines[index] = f"korean = {json.dumps(replacements[current_id], ensure_ascii=False)}{newline}"
                rewritten.add(current_id)
        missing = set(replacements) - rewritten
        if missing:
            raise ValueError(f"{path}: failed to rewrite IDs {sorted(missing)}")
        rendered = "".join(lines)
        tomllib.loads(rendered)
        path.write_text(rendered, encoding="utf-8", newline="")

    return {
        "path": str(path),
        "source_matched_records": stats["source_matched_records"],
        "changed_records": stats["changed_records"],
        "term_replacements": stats["term_replacements"],
        "term_counts": {key: dict(value) for key, value in term_counts.items()},
        "changes": changes,
    }


def load_canonical(path: Path) -> tuple[str, dict[str, str]]:
    text = path.read_text(encoding="utf-8")
    data = tomllib.loads(text)
    entries = {}
    for entry in data.get("entry", []):
        jp = entry["japanese"]
        if jp in entries:
            raise ValueError(f"duplicate canonical entry for {jp}")
        entries[jp] = entry["korean"]
    return text, entries


def rewrite_canonical(path: Path, decisions: list[dict], apply: bool) -> dict:
    text, entries = load_canonical(path)
    updated = []
    added = []
    effective = dict(entries)
    for decision in decisions:
        jp = decision["japanese"]
        ko = decision["korean"]
        old = effective.get(jp)
        if old is None:
            effective[jp] = ko
            added.append({"japanese": jp, "korean": ko})
        elif old != ko:
            legacy = decision.get("legacy", [])
            if old not in legacy:
                raise ValueError(
                    f"canonical table has unexpected value for {jp}: {old!r}; "
                    f"expected {ko!r} or known legacy {legacy!r}"
                )
            effective[jp] = ko
            updated.append({"japanese": jp, "before": old, "after": ko})

    if apply and (updated or added):
        lines = text.splitlines(keepends=True)
        current_jp = None
        seen = set()
        for i, line in enumerate(lines):
            stripped = line.rstrip("\r\n")
            if stripped.startswith('japanese = '):
                current_jp = json.loads(stripped.split("=", 1)[1].strip())
                continue
            if current_jp in effective and stripped.startswith('korean = '):
                if current_jp in {item["japanese"] for item in updated}:
                    newline = "\r\n" if line.endswith("\r\n") else ("\n" if line.endswith("\n") else "")
                    lines[i] = f"korean = {json.dumps(effective[current_jp], ensure_ascii=False)}{newline}"
                seen.add(current_jp)
                current_jp = None
        rendered = "".join(lines)
        for item in added:
            if not rendered.endswith("\n"):
                rendered += "\n"
            rendered += (
                "\n[[entry]]\n"
                f"japanese = {json.dumps(item['japanese'], ensure_ascii=False)}\n"
                f"korean = {json.dumps(item['korean'], ensure_ascii=False)}\n"
            )
        parsed = tomllib.loads(rendered)
        final_entries = {entry["japanese"]: entry["korean"] for entry in parsed.get("entry", [])}
        for decision in decisions:
            if final_entries.get(decision["japanese"]) != decision["korean"]:
                raise ValueError(f"canonical post-write verification failed for {decision['japanese']}")
        path.write_text(rendered, encoding="utf-8", newline="")

    return {"updated": updated, "added": added, "unchanged": len(decisions) - len(updated) - len(added)}


def migrate(root: Path, decisions_path: Path, apply: bool) -> dict:
    decisions = load_decisions(decisions_path)
    canonical_path = root / "translations/terminology/korean-canonical.toml"
    canonical_result = rewrite_canonical(canonical_path, decisions, apply)

    file_reports = []
    total = Counter()
    all_changes = []
    for path in sorted((root / "translations/korean/messages").glob("*.toml")):
        report = rewrite_overlay(path, decisions, apply)
        if report["source_matched_records"] or report["changed_records"]:
            file_reports.append(report)
        total["source_matched_records"] += report["source_matched_records"]
        total["changed_records"] += report["changed_records"]
        total["term_replacements"] += report["term_replacements"]
        for change in report["changes"]:
            change["path"] = str(path.relative_to(root))
            all_changes.append(change)

    return {
        "status": "APPLIED" if apply else "DRY_RUN",
        "canonical": canonical_result,
        "source_matched_records": total["source_matched_records"],
        "changed_records": total["changed_records"],
        "term_replacements": total["term_replacements"],
        "changed_files": sum(1 for report in file_reports if report["changed_records"]),
        "files": file_reports,
        "changes": all_changes,
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--decisions", type=Path, default=Path("docs/audit/beta1-approved-terminology.json"))
    parser.add_argument("--apply", action="store_true")
    parser.add_argument("--json", type=Path)
    args = parser.parse_args()
    root = Path(__file__).resolve().parents[2]
    decisions = args.decisions if args.decisions.is_absolute() else root / args.decisions
    result = migrate(root, decisions, args.apply)
    rendered = json.dumps(result, ensure_ascii=False, indent=2)
    print(rendered)
    if args.json:
        args.json.write_text(rendered + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
