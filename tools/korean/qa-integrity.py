#!/usr/bin/env python3
"""Whole-corpus QA-1 integrity/anomaly scanner for Korean localization.

Critical findings are structural defects that can make the patch unsafe and
cause a non-zero exit status. Advisory findings are triage signals for later
human QA; they are reported but do not fail CI.
"""
from __future__ import annotations

import argparse
from collections import Counter
import json
from pathlib import Path
import re
import sys
import tomllib

ROOT = Path(__file__).resolve().parents[2]
KOREAN_DIR = ROOT / "translations" / "korean" / "messages"
CANONICAL_DIR = ROOT / "translations" / "messages"

sys.path.insert(0, str(Path(__file__).resolve().parent))
from control_tags import fixed_tokens  # noqa: E402

JP_RE = re.compile(r"[\u3040-\u30ff]")
ANGLE_RE = re.compile(r"<[^<>]*>")
LINE_BREAK = "<line-break>"


def visible_text(text: str) -> str:
    return ANGLE_RE.sub("", text).strip()


def issue(kind: str, section: int, rid: str, **extra: object) -> dict[str, object]:
    out: dict[str, object] = {"kind": kind, "section": section, "id": rid}
    out.update(extra)
    return out


def load_toml(path: Path) -> dict[str, dict[str, object]]:
    with path.open("rb") as f:
        data = tomllib.load(f)
    return {str(k): v for k, v in data.items() if isinstance(v, dict)}


def scan() -> dict[str, object]:
    critical: list[dict[str, object]] = []
    advisory: list[dict[str, object]] = []
    counts: Counter[str] = Counter()

    paths = sorted(KOREAN_DIR.glob("msgsec*.toml"))
    for path in paths:
        section = int(path.stem.removeprefix("msgsec"))
        canonical_path = CANONICAL_DIR / path.name
        korean_data = load_toml(path)
        canonical_data = load_toml(canonical_path) if canonical_path.exists() else {}

        counts["files"] += 1
        counts["records"] += len(korean_data)

        for rid, rec in korean_data.items():
            ja = rec.get("japanese")
            ko = rec.get("korean")
            layout = rec.get("layout")

            if not isinstance(ja, str):
                critical.append(issue("missing_japanese", section, rid))
                continue

            canonical = canonical_data.get(rid, {}).get("japanese")
            if canonical is None:
                critical.append(issue("missing_canonical_record", section, rid))
            elif canonical != ja:
                critical.append(issue("canonical_japanese_mismatch", section, rid))

            if ko is None:
                counts["without_korean"] += 1
                continue
            if not isinstance(ko, str):
                critical.append(issue("non_string_korean", section, rid))
                continue

            counts["with_korean"] += 1
            if not ko.strip():
                critical.append(issue("empty_korean", section, rid))
                continue

            if LINE_BREAK in ko:
                critical.append(issue("semantic_line_break_in_korean", section, rid))

            ja_tokens = fixed_tokens(ja)
            ko_tokens = fixed_tokens(ko)
            if ja_tokens != ko_tokens:
                critical.append(
                    issue(
                        "fixed_token_mismatch",
                        section,
                        rid,
                        japanese_tokens=ja_tokens,
                        korean_tokens=ko_tokens,
                    )
                )

            if isinstance(layout, str):
                if fixed_tokens(layout) != ko_tokens:
                    critical.append(issue("layout_fixed_token_mismatch", section, rid))
            elif layout is not None:
                critical.append(issue("non_string_layout", section, rid))

            jp_chars = JP_RE.findall(ko)
            if jp_chars:
                advisory.append(
                    issue(
                        "japanese_script_residue",
                        section,
                        rid,
                        count=len(jp_chars),
                        sample="".join(jp_chars[:12]),
                    )
                )

            ja_visible = visible_text(ja)
            ko_visible = visible_text(ko)
            if ja_visible and ko_visible:
                ratio = len(ko_visible) / len(ja_visible)
                if ratio < 0.18 or ratio > 2.40:
                    advisory.append(
                        issue(
                            "extreme_length_ratio",
                            section,
                            rid,
                            ratio=round(ratio, 3),
                            japanese_chars=len(ja_visible),
                            korean_chars=len(ko_visible),
                        )
                    )

    by_kind_critical = Counter(x["kind"] for x in critical)
    by_kind_advisory = Counter(x["kind"] for x in advisory)
    return {
        "schema": 1,
        "scope": "translations/korean/messages/msgsec*.toml",
        "counts": dict(sorted(counts.items())),
        "critical_count": len(critical),
        "advisory_count": len(advisory),
        "critical_by_kind": dict(sorted(by_kind_critical.items())),
        "advisory_by_kind": dict(sorted(by_kind_advisory.items())),
        "critical": critical,
        "advisory": advisory,
    }


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--json", type=Path, help="write full machine-readable report")
    ap.add_argument("--max-examples", type=int, default=20)
    args = ap.parse_args()

    report = scan()
    if args.json:
        args.json.parent.mkdir(parents=True, exist_ok=True)
        args.json.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    print("Korean QA-1 integrity scan")
    for key, value in report["counts"].items():
        print(f"  {key}: {value}")
    print(f"  critical: {report['critical_count']}")
    print(f"  advisory: {report['advisory_count']}")
    print(f"  critical_by_kind: {json.dumps(report['critical_by_kind'], ensure_ascii=False, sort_keys=True)}")
    print(f"  advisory_by_kind: {json.dumps(report['advisory_by_kind'], ensure_ascii=False, sort_keys=True)}")

    examples = report["critical"] + report["advisory"]
    for item in examples[: max(args.max_examples, 0)]:
        print("  example: " + json.dumps(item, ensure_ascii=False, sort_keys=True))

    raise SystemExit(1 if report["critical_count"] else 0)


if __name__ == "__main__":
    main()
