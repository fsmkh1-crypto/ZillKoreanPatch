#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""Fail closed if Korean renderer metrics drift from the English visual contract model."""
from __future__ import annotations

import pathlib
import re
import tomllib

ROOT = pathlib.Path(__file__).resolve().parents[2]

metrics = tomllib.loads((ROOT / "release/font/metrics.toml").read_text(encoding="utf-8"))
korean_metrics = (ROOT / "internal/zillfont/korean_metrics.go").read_text(encoding="utf-8")
mobile_plan = (ROOT / "cmd/zill/build_korean_mobile_plan.go").read_text(encoding="utf-8")
mobile_font = (ROOT / "internal/release/korean_font_mobile.go").read_text(encoding="utf-8")
visual = (ROOT / "internal/layout/validate_korean_visual.go").read_text(encoding="utf-8")
storage = (ROOT / "internal/layout/validate_korean_english_contract.go").read_text(encoding="utf-8")


def const(name: str) -> int:
    m = re.search(rf"\b{name}\s*=\s*(-?[0-9]+)", korean_metrics)
    if not m:
        raise SystemExit(f"KOREAN_LAYOUT_FONT_METRICS_FAIL missing {name}")
    return int(m.group(1))


advance = const("KoreanTargetAdvance")
raster_width = const("KoreanRasterWidth")
raster_height = const("KoreanRasterHeight")
if advance <= 0 or raster_width <= 0 or raster_height <= 0:
    raise SystemExit("KOREAN_LAYOUT_FONT_METRICS_FAIL non-positive Korean renderer geometry")
if raster_width > advance:
    raise SystemExit(
        f"KOREAN_LAYOUT_FONT_METRICS_FAIL raster width {raster_width} exceeds target advance {advance}"
    )

# The English metric table is authoritative for stock text. It legitimately
# contains a few narrow double-byte punctuation glyphs. The mobile Korean full
# repack may repurpose any installed double-byte key and rewrites mapped custom
# glyphs to KoreanTargetAdvance, so those nominal English widths must NOT be
# imposed on mapped Hangul. Instead, the Korean visual validator must explicitly
# use KoreanTargetAdvance for mapped runes and the English metric table for
# unmapped stock text.
glyphs = metrics.get("glyph", {})
if not isinstance(glyphs, dict) or not glyphs:
    raise SystemExit("KOREAN_LAYOUT_FONT_METRICS_FAIL empty English glyph metrics")

double_count = 0
nominal_mismatches = []
for raw_key, value in glyphs.items():
    key = int(raw_key, 0)
    lo, hi = key & 0xFF, (key >> 8) & 0xFF
    is_double = hi != 0 and ((0x81 <= lo <= 0x9F) or (0xE0 <= lo <= 0xFC)) and ((0x40 <= hi <= 0x7E) or (0x80 <= hi <= 0xFC))
    if not is_double:
        continue
    double_count += 1
    if value != advance:
        nominal_mismatches.append((raw_key, value))
if double_count == 0:
    raise SystemExit("KOREAN_LAYOUT_FONT_METRICS_FAIL no double-byte metrics found")

for anchor in (
    "installed := font.DoubleByteKeys()",
    "koreanslots.BuildPlan(texts, installed",
):
    if anchor not in mobile_plan:
        raise SystemExit("KOREAN_LAYOUT_FONT_METRICS_FAIL mobile planner drifted: " + anchor)
for anchor in (
    "FullRepackAuthenticatedRetailFont",
    "VerifyFullRepackSemantics",
):
    if anchor not in mobile_font:
        raise SystemExit("KOREAN_LAYOUT_FONT_METRICS_FAIL mobile font path drifted: " + anchor)
for anchor in (
    "ValidateKoreanEnglishVisualContracts",
    "zillfont.KoreanTargetAdvance",
    'e.category(row.ID, "character-profile")',
    "profileAdvance",
    "profileMaxLines",
):
    if anchor not in visual:
        raise SystemExit("KOREAN_LAYOUT_FONT_METRICS_FAIL Korean visual validator drifted: " + anchor)
if "e.ValidateKoreanEnglishVisualContracts(source, korean, layouts, mapping)" not in storage:
    raise SystemExit("KOREAN_LAYOUT_FONT_METRICS_FAIL Korean release storage gate no longer invokes hard English visual contracts")

print(
    "KOREAN_LAYOUT_FONT_METRICS_PASS "
    f"double_byte_metrics={double_count} target_advance={advance} raster={raster_width}x{raster_height} "
    f"nominal_double_byte_width_differences={len(nominal_mismatches)} mapped_override=true"
)
