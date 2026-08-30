#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""Fail closed on structural drift from the upstream English contract surface.

This is intentionally a source-structure gate, not a runtime-safety proof. It
protects the project-wide English-first premise by making the current deliberate
Korean differences explicit instead of allowing silent validator drift.
"""
from __future__ import annotations

import pathlib
import re

ROOT = pathlib.Path(__file__).resolve().parents[2]


def read(path: str) -> str:
    return (ROOT / path).read_text(encoding="utf-8")


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit("ENGLISH_FIRST_PARITY_FAIL " + message)


english_validate = read("internal/layout/validate.go")
korean_validate = read("internal/layout/validate_korean_english_contract.go")
korean_c5 = read("internal/layout/validate_korean.go")
english_compile = read("internal/message/compile.go")
korean_compile = read("internal/message/compile_korean.go")
korean_materialize = read("internal/message/korean_materialize.go")
release_build = read("internal/release/korean_build.go")
rules = read("internal/layout/rules.go")
premise = read("AGENTS.md")

require("NON-NEGOTIABLE PROJECT PREMISE" in premise and "ENGLISH PATCH FIRST" in premise,
        "AGENTS.md no longer contains the English-first project premise")

consumer_re = re.compile(r"e\.consumers\.([A-Za-z0-9_]+)")
english_consumers = set(consumer_re.findall(english_validate))
korean_consumers = set(consumer_re.findall(korean_validate))
korean_c5_consumers = set(consumer_re.findall(korean_c5))

# C5 is deliberately handled in the Korean exact-materialization validator,
# rather than duplicated in ValidateKoreanEnglishConsumerContracts.
deliberate_split = {"C5IDs", "SinglePageC5IDs"}
missing = english_consumers - korean_consumers - deliberate_split
require(not missing, "Korean validator lost English consumer references: " + ",".join(sorted(missing)))
require(deliberate_split <= korean_c5_consumers,
        "documented C5 split is not backed by Korean C5 consumer membership")

# Both paths must use the same *engine constants defined in rules.go*. Do not
# mistake local helper parameter names such as bufferCapacityBytes for engine
# contracts merely because their spelling ends in CapacityBytes.
rule_constant_re = re.compile(r"^\s*([a-z][A-Za-z0-9]*(?:CapacityBytes|MaxPayloadBytes|MaxLineBytes|MaxPages))\s*=", re.M)
rule_constants = set(rule_constant_re.findall(rules))
english_constants = {name for name in rule_constants if re.search(r"\b" + re.escape(name) + r"\b", english_validate)}
korean_constants = {
    name for name in rule_constants
    if re.search(r"\b" + re.escape(name) + r"\b", korean_validate)
    or re.search(r"\b" + re.escape(name) + r"\b", korean_c5)
}
missing_constants = english_constants - korean_constants
require(not missing_constants,
        "Korean validation lost English capacity constants: " + ",".join(sorted(missing_constants)))

# Control traversal and lowering must remain shared with English. Korean is
# allowed to replace only the natural-text encoder/validation layer.
require("p.splitSemanticWith" in korean_materialize,
        "Korean semantic traversal no longer uses shared splitSemanticWith")
require("p.materializeValues" in korean_materialize,
        "Korean materialization no longer uses shared materializeValues")

# Bank representation/capacity remains an engine contract, not a language rule.
for anchor in ("RuntimeBankCapacity(bank.Section)", "binary.LittleEndian.PutUint32"):
    require(anchor in english_compile, f"English compiler anchor disappeared: {anchor}")
    require(anchor in korean_compile, f"Korean compiler drifted from English bank contract: {anchor}")

# Production desktop release path must carry both halves of the fixed-storage
# contract before Korean banks are compiled.
english_gate = release_build.find("ValidateKoreanEnglishConsumerContracts")
c5_gate = release_build.find("validateKoreanRuntimeStorage")
compile_gate = release_build.find("compileKoreanBanksWithPlan")
require(min(english_gate, c5_gate, compile_gate) >= 0,
        "production Korean release path is missing a required parity gate")
require(english_gate < compile_gate and c5_gate < compile_gate,
        "production Korean release compiles banks before storage parity validation")

print("ENGLISH_FIRST_PARITY_PASS")
print("english_consumers=" + ",".join(sorted(english_consumers)))
print("korean_direct_consumers=" + ",".join(sorted(korean_consumers)))
print("korean_c5_split=" + ",".join(sorted(deliberate_split)))
print("shared_capacity_constants=" + ",".join(sorted(english_constants)))
