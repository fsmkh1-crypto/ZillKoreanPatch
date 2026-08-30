#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""Audit Korean fixed-data handling against the upstream English engine contract."""
from __future__ import annotations

import pathlib
import re
import tomllib

ROOT = pathlib.Path(__file__).resolve().parents[2]


def read(path: str) -> str:
    return (ROOT / path).read_text(encoding="utf-8")


def load_toml(path: str) -> dict:
    with (ROOT / path).open("rb") as fh:
        return tomllib.load(fh)


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit("FIXED_DATA_PARITY_FAIL " + message)


def executable_newlines(text: str) -> str:
    # Mirrors fixeddata.executableNewlines: TOML decoding has already produced
    # literal CR/LF characters from escaped manifest sequences.
    return text


def cp932_size(text: str, label: str) -> int:
    try:
        return len(executable_newlines(text).encode("cp932"))
    except UnicodeEncodeError as exc:
        raise SystemExit(f"FIXED_DATA_PARITY_FAIL {label} is not stock CP932 encodable: {exc}") from exc


def korean_renderer_size(text: str, custom_glyphs: set[str], label: str) -> int:
    total = 0
    for ch in executable_newlines(text):
        try:
            total += len(ch.encode("cp932"))
            continue
        except UnicodeEncodeError:
            pass
        require(ch in custom_glyphs,
                f"{label} uses {ch!r} U+{ord(ch):04X} but Korean glyph catalog does not contain it")
        total += 2
    return total


def offset_fields(raw: dict, label: str) -> dict[int, dict]:
    out: dict[int, dict] = {}
    for key, value in raw.items():
        try:
            offset = int(key, 0) if isinstance(key, str) else int(key)
        except (TypeError, ValueError) as exc:
            raise SystemExit(f"FIXED_DATA_PARITY_FAIL {label} has invalid offset {key!r}") from exc
        require(offset not in out, f"{label} duplicates offset {offset:#x}")
        out[offset] = value
    return out


english_release = read("internal/release/build.go")
korean_release = read("internal/release/korean_build.go")
korean_mobile = read("internal/release/korean_mobile.go")
korean_preflight = read("internal/release/korean_mobile_preflight.go")
english_eboot = read("internal/fixeddata/eboot.go")
korean_eboot = read("internal/fixeddata/korean_eboot.go")
equipment = read("internal/fixeddata/equipment.go")
desktop_plan = read("cmd/zill/build_korean.go")
mobile_plan = read("cmd/zill/build_korean_mobile_plan.go")
korean_fixed_manifest = read("release/korean/strings/eboot.toml")
english_equipment_manifest = read("release/strings/equipment.toml")

# EBOOT: English translates the complete fixed-string table; Korean intentionally
# uses a sparse reviewed overlay, but every field it does touch must preserve the
# same authenticated fixed-width/NUL/overlap contract.
for anchor in (
    "sha256.Sum256(source) != patchedELFSHA256",
    "len(encoded) > len(expected)",
    "source[start+len(expected)] != 0",
    "translation fields overlap",
):
    require(anchor in english_eboot, "English EBOOT contract lost: " + anchor)
    require(anchor in korean_eboot, "Korean EBOOT contract lost: " + anchor)

require("const ebootFieldCount = 557" in english_eboot,
        "English complete EBOOT field authority no longer has 557 fields")
require("KoreanEBOOTTranslations is a sparse Korean overlay" in korean_eboot,
        "Korean sparse-overlay policy lost explicit classification")
require(len(re.findall(r"^0x[0-9a-f]+\s*=", korean_fixed_manifest, re.M)) > 0,
        "Korean fixed EBOOT overlay became empty")
require("elfpatch.VerifyApplied(result, manifest)" in read("internal/release/korean_fixed.go"),
        "Korean EBOOT overlay lost executable-manifest postverification")

# Manifest-level parity against the upstream English fixed-string authority.
# This catches bad offsets, source drift, absent Korean glyphs, and byte growth
# before an authenticated retail EBOOT is available. The device/desktop build
# repeats the exact check against real bytes in ApplyKoreanEBOOT.
english_fields = offset_fields(load_toml("release/strings/eboot.toml"), "English EBOOT manifest")
korean_fields = offset_fields(load_toml("release/korean/strings/eboot.toml"), "Korean EBOOT manifest")
glyph_catalog = load_toml("release/korean/font/glyphs.toml")
custom_glyphs = set(glyph_catalog.get("glyphs", {}))
require(len(english_fields) == 557,
        f"English fixed EBOOT manifest has {len(english_fields)} fields, expected 557")
require(bool(korean_fields), "Korean fixed EBOOT manifest is empty")

fixed_headroom: list[tuple[int, int, int]] = []
for offset, korean_field in sorted(korean_fields.items()):
    require(offset in english_fields,
            f"Korean fixed EBOOT offset {offset:#x} is absent from English authority")
    english_field = english_fields[offset]
    require(korean_field.get("source") == english_field.get("source"),
            f"Korean fixed EBOOT offset {offset:#x} source differs from English authority")
    source_size = cp932_size(str(korean_field["source"]), f"EBOOT {offset:#x} source")
    replacement_size = korean_renderer_size(
        str(korean_field["replacement"]), custom_glyphs, f"EBOOT {offset:#x} replacement"
    )
    require(replacement_size <= source_size,
            f"Korean fixed EBOOT offset {offset:#x} replacement uses {replacement_size} bytes; capacity is {source_size}")
    fixed_headroom.append((offset, source_size, replacement_size))

minimum_headroom = min((capacity - used for _, capacity, used in fixed_headroom), default=0)
print(
    "FIXED_EBOOT_KOREAN_CENSUS "
    f"english_authority={len(english_fields)} korean_fields={len(korean_fields)} "
    f"source_match=PASS capacity=PASS glyph_catalog=PASS minimum_headroom={minimum_headroom}"
)

# BINDATA/equipment: the English contract authenticates the exact retail member,
# all 132 records, 17-byte fields (16-byte payload + NUL), source text, and CD
# padding before mutation. Korean currently preserves retail equipment names, so
# this is a localization-completeness gap rather than a storage-safety bypass.
for anchor in (
    "equipmentRecordCount = 132",
    "equipmentRecordSize  = 0x24",
    "equipmentNameOffset  = 0x11",
    "equipmentNameSize    = 17",
    "sha256.Sum256(source) != bindataSHA256",
    "source guard does not match",
    "invalid source field padding",
):
    require(anchor in equipment, "BINDATA fixed-width guard lost: " + anchor)

require(len(re.findall(r"^\d+\s*=", english_equipment_manifest, re.M)) == 132,
        "English equipment authority no longer contains exactly 132 selectors")
require("addEquipment(root, archives)" in english_release,
        "English release no longer materializes guarded equipment names")
require("addEquipment(root, archives)" not in korean_release,
        "Korean desktop unexpectedly started applying English equipment text")
require("addEquipment(root, archives)" not in korean_mobile,
        "Korean mobile unexpectedly started applying English equipment text")
require("addEquipment(root, archives)" not in korean_preflight,
        "Korean preflight unexpectedly started applying English equipment text")

# Even while Korean leaves the equipment fields retail, both production planners
# must authenticate the exact BINDATA layout through ApplyEquipment and reserve
# structured CP932 literals so custom glyph allocation cannot steal live keys.
for label, text in (("desktop", desktop_plan), ("mobile", mobile_plan)):
    for anchor in (
        "loadRetailBindata(gameDir)",
        "fixeddata.ApplyEquipment(bindata, equipment)",
        "slotaudit.ScanCP932Literals(bindata)",
        "mergeRendererKeys(reserved, bindataScan.Keys)",
    ):
        require(anchor in text, f"{label} planner lost BINDATA authentication/renderer ownership: {anchor}")

print("FIXED_DATA_PARITY_PASS")
print("eboot=english_complete_557_korean_sparse_same_fixed_width_contract")
print("bindata_storage_contract=authenticated_132x17_source_guarded")
print("bindata_korean_localization=MISSING_retail_preserved")
print("freeze_relevance=static_storage_surface_closed_localization_gap_only")
