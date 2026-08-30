#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""Fail closed if Korean PARAM.SFO diverges from the upstream English engine contract."""
from __future__ import annotations

import pathlib
import re

ROOT = pathlib.Path(__file__).resolve().parents[2]


def read(path: str) -> str:
    return (ROOT / path).read_text(encoding="utf-8")


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit("PARAM_SFO_PARITY_FAIL " + message)


english = read("internal/release/build.go")
korean = read("internal/release/korean_build.go")
manifest = read("patches/system/param-sfo.toml")

# Both release paths must transform the authenticated retail PARAM.SFO through
# the same shared manifest and sfo.Apply implementation. Only the display title
# is language/version specific.
for label, text, fn in (
    ("English", english, "func buildSFO"),
    ("Korean", korean, "func buildKoreanAlphaSFO"),
):
    start = text.find(fn)
    require(start >= 0, f"{label} SFO builder disappeared")
    body = text[start : start + 1800]
    for anchor in (
        'os.ReadFile(filepath.Join(gameDir, "PARAM.SFO"))',
        'read(root, "patches", "system", "param-sfo.toml")',
        "sfo.ParseManifest(manifestData)",
        "sfo.Apply(source, manifest,",
    ):
        require(anchor in body, f"{label} SFO builder lost shared contract anchor: {anchor}")

require('"Zill O\'ll Infinite Plus [English %s]"' in english,
        "English localized title contract drifted")
require('"Zill O\'ll Infinite Plus [Korean Beta %s]"' in korean,
        "Korean localized title contract drifted")

# The shared manifest is itself pinned to the supported retail SFO and appends
# MEMSIZE without weakening the original source structure.
for pattern, label in (
    (r'^source_sha256\s*=\s*"[0-9a-f]{64}"$', "retail source SHA-256"),
    (r'^source_size\s*=\s*472$', "retail source size"),
    (r'^magic\s*=\s*0x46535000$', "PSF magic"),
    (r'^sfo_version\s*=\s*0x00000101$', "PSF version"),
    (r'^expected_absent_key\s*=\s*"MEMSIZE"$', "MEMSIZE absence guard"),
    (r'^append_key\s*=\s*"MEMSIZE"$', "MEMSIZE append key"),
    (r'^entry_format\s*=\s*0x0404$', "MEMSIZE format"),
    (r'^value\s*=\s*1$', "MEMSIZE value"),
    (r'^alignment\s*=\s*16$', "MEMSIZE alignment"),
):
    require(re.search(pattern, manifest, re.M) is not None, "manifest lost " + label)

print("PARAM_SFO_PARITY_PASS")
print("source_manifest=shared_authenticated")
print("transform=sfo.Apply_shared")
print("difference=localized_title_only")
