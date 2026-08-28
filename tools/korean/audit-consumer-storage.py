#!/usr/bin/env python3
import glob
import re
import tomllib
from pathlib import Path

CONTROL = re.compile(r"<(?:\$[0-9A-F]{2}|color:[befho]|end|equal|greater-equal|if|less-equal|less|line-break|not-equal|select|value:\$[0-9A-F]{2})>")
VALUE = re.compile(r"<value:\$[0-9A-F]{2}>")
PRINTF = re.compile(r"%(?:[1-9][0-9]*)?[su]")
FORMAT_SIGNATURE_IDS = {20006, 170006, 170007, 170008, 170009, 1070022, 1070023}

CAP = {
    "bounded label": 28,
    "guild client": 17,
    "guild region": 152,
    "character-creation choice": 31,
    "equipment feedback": 109,
    "trap": 104,
    "chronicle entry": 765,  # Validate uses payload > 764.
    "c20 group": 768,
    "c22 total": 512,
    "c22 page": 256,
    "guild posting": 316,
}
C22_MAX_PAGES = 9
C22_MAX_LINE_BYTES = 56
C5_PAGE_CAPACITY = 256
C5_LINES_PER_PAGE = 3
C5_MAX_PAGES = 9
PLAYER_NAME_MAX_ENCODED_BYTES = 16
TRAP_VALUE_MAX_BYTES = 11
GUILD_POSTING_INTEGER_MAX_BYTES = 20
TRAP_ID = 1070079


def load_rows(pattern: str):
    rows = {}
    paths = {}
    for filename in sorted(glob.glob(pattern)):
        path = Path(filename)
        with path.open("rb") as f:
            data = tomllib.load(f)
        for key, row in data.items():
            rid = int(key)
            previous = rows.get(rid)
            if previous is not None and previous != row:
                raise RuntimeError(f"conflicting duplicate ID {rid}: {paths[rid]} vs {path}")
            rows[rid] = row
            paths[rid] = path
    return rows, paths


def normalize(text: str) -> str:
    return text.replace("\r\n", "\n").replace("\n", "<line-break>")


def plain_bytes(text: str) -> int:
    total = 0
    for ch in text:
        try:
            total += len(ch.encode("cp932"))
        except UnicodeEncodeError:
            # Every custom Korean/mobile glyph is assigned to one existing
            # two-byte renderer key. Its exact key does not affect storage size.
            total += 2
    return total


def expanded_bytes(text: str) -> int:
    text = normalize(text)
    total = 0
    cursor = 0
    for match in CONTROL.finditer(text):
        total += plain_bytes(text[cursor:match.start()])
        tag = match.group(0)
        if tag == "<line-break>":
            total += 1
        elif tag.startswith("<color:"):
            total += 3
        cursor = match.end()
    total += plain_bytes(text[cursor:])
    return total


def minimum_bytes(text: str, rid: int) -> int:
    if rid in FORMAT_SIGNATURE_IDS:
        text = PRINTF.sub("", text)
    return expanded_bytes(text)


def effective_text(rid: int, korean, source) -> str:
    row = korean.get(rid)
    if row is not None:
        return row.get("layout") or row.get("korean") or ""
    return source.get(rid, {}).get("japanese", "")


def category_for(rid: int, ranges):
    for r in ranges:
        if rid < r["first"]:
            break
        if rid <= r["last"]:
            return r["category"], r["basis"]
    return "", ""


def add_violation(out, label, rid, size, maximum, detail=""):
    out.append((label, rid, size, maximum, detail))


def main() -> int:
    root = Path(__file__).resolve().parents[2]
    korean, paths = load_rows(str(root / "translations/korean/messages/msgsec*.toml"))
    source, _ = load_rows(str(root / "translations/messages/msgsec*.toml"))
    with (root / "release/layout/categories.toml").open("rb") as f:
        categories_doc = tomllib.load(f)
    with (root / "release/layout/consumer-map.toml").open("rb") as f:
        consumers = tomllib.load(f)
    ranges = categories_doc["range"]
    translated = set(korean)
    violations = []

    def check_ids(label, ids, capacity):
        for rid in ids:
            if rid not in translated:
                continue
            size = expanded_bytes(effective_text(rid, korean, source))
            if size >= capacity:
                add_violation(violations, label, rid, size, capacity - 1)

    check_ids("bounded label", consumers.get("bounded_label_ids", []), CAP["bounded label"])
    check_ids("guild client", consumers.get("guild_client_ids", []), CAP["guild client"])
    check_ids("guild region", consumers.get("guild_region_ids", []), CAP["guild region"])

    for rid in sorted(translated):
        category, basis = category_for(rid, ranges)
        if basis != "verified":
            continue
        text = effective_text(rid, korean, source)
        size = expanded_bytes(text)
        if category == "character-creation-choice" and size >= CAP["character-creation choice"]:
            add_violation(violations, "character-creation choice", rid, size, CAP["character-creation choice"] - 1)
        if category == "equipment-feedback" and size >= CAP["equipment feedback"]:
            add_violation(violations, "equipment feedback", rid, size, CAP["equipment feedback"] - 1)
        if category == "chronicle-entry":
            maximum = 764
            worst = size + text.count("<value:$28>") * PLAYER_NAME_MAX_ENCODED_BYTES
            if worst > maximum:
                add_violation(violations, "chronicle entry", rid, worst, maximum, "includes max player-name expansion")

    if TRAP_ID in translated:
        text = effective_text(TRAP_ID, korean, source)
        worst = expanded_bytes(text) + len(VALUE.findall(text)) * TRAP_VALUE_MAX_BYTES
        if worst >= CAP["trap"]:
            add_violation(violations, "trap", TRAP_ID, worst, CAP["trap"] - 1, "includes max value expansion")

    for group in consumers.get("c20_group", []):
        ids = group.get("ids", [])
        if not any(rid in translated for rid in ids):
            continue
        total = sum(minimum_bytes(effective_text(rid, korean, source), rid) + 1 for rid in ids)
        if total >= CAP["c20 group"]:
            add_violation(violations, "C20 group", ids[0], total, CAP["c20 group"] - 1, f"ids={ids}")

    c22_no_layout = []
    for rid in consumers.get("c22_ids", []):
        if rid not in translated:
            continue
        row = korean[rid]
        text = row.get("layout") or row.get("korean") or ""
        runtime = normalize(text).split("<end>", 1)[0]
        lines = runtime.split("<line-break>")
        total = minimum_bytes(runtime, rid)
        pages = (len(lines) + 3) // 4
        if pages > C22_MAX_PAGES:
            add_violation(violations, "C22 pages", rid, pages, C22_MAX_PAGES)
        if total >= CAP["c22 total"]:
            add_violation(violations, "C22 total", rid, total, CAP["c22 total"] - 1)
        for index, line in enumerate(lines, 1):
            size = minimum_bytes(line, rid)
            if size > C22_MAX_LINE_BYTES:
                add_violation(violations, "C22 line", rid, size, C22_MAX_LINE_BYTES, f"line={index}")
        for start in range(0, len(lines), 4):
            page = lines[start:start + 4]
            size = minimum_bytes("<line-break>".join(page), rid)
            if size >= CAP["c22 page"]:
                add_violation(violations, "C22 page", rid, size, CAP["c22 page"] - 1, f"page={start // 4 + 1}")
        if not row.get("layout"):
            c22_no_layout.append(rid)

    posting_candidates = consumers.get("posting_candidates", {})
    role_ids = {
        "destination": posting_candidates.get("destination", []),
        "escorted role/name": posting_candidates.get("escorted role/name", []),
        "qualifier/title": posting_candidates.get("qualifier/title", []),
        "target item": posting_candidates.get("target item", []),
        "target monster": posting_candidates.get("target monster", []),
    }
    maxima = {}
    for role, ids in role_ids.items():
        maxima[role] = max((expanded_bytes(effective_text(rid, korean, source)) for rid in ids if rid in source or rid in korean), default=0)
    integer_roles = set(posting_candidates.get("integer_roles", []))
    for posting in consumers.get("posting", []):
        rid = posting["id"]
        if rid not in translated:
            continue
        text = effective_text(rid, korean, source)
        size = expanded_bytes(text)
        for tag, role in posting.get("bindings", {}).items():
            count = text.count(tag)
            size += count * (GUILD_POSTING_INTEGER_MAX_BYTES if role in integer_roles else maxima.get(role, 0))
        if size >= CAP["guild posting"]:
            add_violation(violations, "guild posting", rid, size, CAP["guild posting"] - 1)

    # C5 validation in the Go release code lowers source control flow before
    # checking branch-local pages. This Python forensic pass cannot reproduce
    # that exactly, so report its scope instead of pretending to prove safety.
    c5_ids = set(consumers.get("c5_ids", [])) | set(consumers.get("single_page_c5_ids", []))
    c5_translated = sorted(c5_ids & translated)

    print(f"translated_records={len(translated)}")
    print(f"consumer_storage_violations={len(violations)}")
    by_label = {}
    for label, *_ in violations:
        by_label[label] = by_label.get(label, 0) + 1
    for label in sorted(by_label):
        print(f"violation_count[{label}]={by_label[label]}")
    for label, rid, size, maximum, detail in sorted(violations, key=lambda x: (x[1], x[0], x[2])):
        path = paths.get(rid)
        rel = path.relative_to(root) if path else "?"
        suffix = f" {detail}" if detail else ""
        print(f"VIOLATION label={label!r} id={rid} bytes={size} maximum={maximum} file={rel}{suffix}")
    print(f"c22_translated_without_authored_layout={len(c22_no_layout)}")
    print("c22_translated_without_authored_layout_ids=" + ",".join(map(str, c22_no_layout)))
    print(f"c5_branch_accurate_audit_pending={len(c5_translated)}")
    print("c5_branch_accurate_audit_pending_ids=" + ",".join(map(str, c5_translated)))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
