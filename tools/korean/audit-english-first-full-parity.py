#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""Fail closed on structural drift from the upstream English contract surface."""
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
english_release = read("internal/release/build.go")
release_build = read("internal/release/korean_build.go")
mobile_build = read("internal/release/korean_mobile.go")
mobile_preflight = read("internal/release/korean_mobile_preflight.go")
korean_font = read("internal/release/korean_font.go")
korean_full_repack_verify = read("internal/zillfont/korean_full_repack_verify.go")
korean_fixed = read("internal/release/korean_fixed.go")
korean_fixeddata = read("internal/fixeddata/korean_eboot.go")
mobile_plan = read("cmd/zill/build_korean_mobile_plan.go")
slot_plan = read("internal/koreanslots/plan.go")
paa = read("internal/gamefmt/paa/paa.go")
disc = read("internal/release/disc.go")
english_font_manifest = read("release/font/manifest.toml")
executable_manifest = read("patches/executable/manifest.toml")
rules = read("internal/layout/rules.go")
premise = read("AGENTS.md")

require("NON-NEGOTIABLE PROJECT PREMISE" in premise and "ENGLISH PATCH FIRST" in premise,
        "AGENTS.md no longer contains the English-first project premise")

consumer_re = re.compile(r"e\.consumers\.([A-Za-z0-9_]+)")
english_consumers = set(consumer_re.findall(english_validate))
korean_consumers = set(consumer_re.findall(korean_validate))
korean_c5_consumers = set(consumer_re.findall(korean_c5))
deliberate_split = {"C5IDs", "SinglePageC5IDs"}
missing = english_consumers - korean_consumers - deliberate_split
require(not missing, "Korean validator lost English consumer references: " + ",".join(sorted(missing)))
require(deliberate_split <= korean_c5_consumers,
        "documented C5 split is not backed by Korean C5 consumer membership")

rule_constant_re = re.compile(r"^\s*([a-z][A-Za-z0-9]*(?:CapacityBytes|MaxPayloadBytes|MaxLineBytes|MaxPages))\s*=", re.M)
rule_constants = set(rule_constant_re.findall(rules))
english_constants = {n for n in rule_constants if re.search(r"\b" + re.escape(n) + r"\b", english_validate)}
korean_constants = {n for n in rule_constants if re.search(r"\b" + re.escape(n) + r"\b", korean_validate) or re.search(r"\b" + re.escape(n) + r"\b", korean_c5)}
missing_constants = english_constants - korean_constants
require(not missing_constants,
        "Korean validation lost English capacity constants: " + ",".join(sorted(missing_constants)))

require("p.splitSemanticWith" in korean_materialize,
        "Korean semantic traversal no longer uses shared splitSemanticWith")
require("p.materializeValues" in korean_materialize,
        "Korean materialization no longer uses shared materializeValues")

for anchor in ("RuntimeBankCapacity(bank.Section)", "binary.LittleEndian.PutUint32"):
    require(anchor in english_compile, f"English compiler anchor disappeared: {anchor}")
    require(anchor in korean_compile, f"Korean compiler drifted from English bank contract: {anchor}")

# Every production/preflight entry point must run the English-first storage gates
# before compiling a Korean bank. This prevents desktop/mobile/preflight drift.
def require_release_chain(label: str, text: str) -> None:
    english_gate = text.find("ValidateKoreanEnglishConsumerContracts")
    c5_gate = text.find("validateKoreanRuntimeStorage")
    compile_gate = text.find("compileKoreanBanksWithPlan")
    derive_english = text.find("DeriveKoreanEnglishConsumerLayouts")
    require(min(english_gate, c5_gate, compile_gate, derive_english) >= 0,
            f"{label} path is missing an English-first parity gate")
    require(derive_english < english_gate < compile_gate and c5_gate < compile_gate,
            f"{label} path compiles before completing English/C5 contract validation")


require_release_chain("desktop Korean release", release_build)
require_release_chain("mobile Korean ISO", mobile_build)
require_release_chain("mobile Korean preflight", mobile_preflight)

# Font inputs are an engine asset contract too. Reuse the exact retail source
# hashes already authenticated by the upstream English static-font manifest.
english_font_hashes = re.findall(r'^source_sha256\s*=\s*"([0-9a-f]{64})"', english_font_manifest, re.M)
require(len(english_font_hashes) == 2, "upstream English font manifest no longer exposes exactly two source fingerprints")
for digest in english_font_hashes:
    require(digest in korean_font,
            "Korean font path no longer pins upstream English retail source fingerprint " + digest)
require("prepareKoreanMobileFontReplacements" in mobile_build,
        "mobile Korean build no longer reaches the authenticated font path")
# English has a frozen result_sha256 for each complete font member. Korean's
# corpus-derived mapping cannot have one static result hash, so its verifier must
# enforce an exact mutation surface: only atlas image payloads and modeled PAF
# geometry/metric fields may differ from authenticated retail bytes.
for anchor in (
    "verifyFullRepackContainerMutationSurface(retailAtlas, retailPAF, patchedAtlas, patchedPAF)",
    "changed immutable atlas/container byte",
    "changed immutable PAF/container byte",
):
    require(anchor in korean_full_repack_verify,
            "Korean full-font verifier lost English-equivalent result boundary: " + anchor)

# BOOT/EBOOT is shared engine machinery. Korean must apply the same executable
# manifest first, authenticate the manifest-patched ELF just like English, then
# prove its sparse localization overlay did not clobber any runtime patch span.
manifest_source = re.search(r'^source_sha256\s*=\s*"([0-9a-f]{64})"', executable_manifest, re.M)
manifest_result = re.search(r'^result_sha256\s*=\s*"([0-9a-f]{64})"', executable_manifest, re.M)
require(manifest_source is not None and manifest_result is not None,
        "executable manifest lost source/result fingerprints")
for anchor in ("elfpatch.Apply(source, manifest)", "applyKoreanFixedEBOOT(root, patched, mapping)"):
    require(anchor in release_build, "Korean executable build drifted from shared manifest chain: " + anchor)
require("patchedELFSHA256" in korean_fixeddata,
        "Korean fixed EBOOT overlay lost patched-ELF fingerprint authentication")
require("elfpatch.VerifyApplied(result, manifest)" in korean_fixed,
        "Korean fixed EBOOT overlay lost executable manifest postcondition verification")

# Slot reuse is Korean-only, so it must be at least as conservative as the
# project-owned production BuildPlan contract: exact two-byte references in
# authenticated BOOT/EBOOT/bindata blobs are excluded before allocation, not
# merely reported afterwards.
require("ExcludeExactByteReferences" in slot_plan,
        "production slot planner lost exact-byte ownership exclusion")
require("koreanslots.BuildPlan(texts, installed, rendererKeySetSlice(reserved), boot, eboot, bindata)" in mobile_plan,
        "mobile slot planner no longer feeds authenticated BOOT/EBOOT/bindata into BuildPlan")
require("finalAudit.CandidateHits != 0" in mobile_plan and "finalAudit.MappedHits" in mobile_plan,
        "mobile slot planner no longer fails closed on post-allocation exact-byte collisions")

# Archive rebuilding and final ISO authoring are shared English/Korean engine
# boundaries. The shared PAA implementation must reject duplicate member
# replacements, reopen the rebuilt pair, and compare every member payload. Both
# Korean release paths must then use the same ISO authoring helper as English,
# which reopens the ISO and compares every staged PSP_GAME file byte-for-byte.
for anchor in (
    "member %d is selected by more than one replacement",
    "verifyRebuilt(p, indexTempPath, archiveTempPath, rebuiltIndex, resolved)",
    "rebuilt member %d %q payload differs",
):
    require(anchor in paa, "shared PAA rebuild lost fail-closed archive contract: " + anchor)
for label, text in (
    ("English release", english_release),
    ("desktop Korean release", release_build),
    ("mobile Korean ISO", mobile_build),
):
    require("archive.pair.Rebuild(" in text,
            f"{label} no longer rebuilds archives through the shared verified PAA path")
    require("authorTranslatedISO(" in text,
            f"{label} no longer authors ISO through the shared provenance path")
for anchor in (
    "verifyAuthoredPSPGame(outputPath, gameDir)",
    "compareExactReaders(got, want)",
    "FORENSIC ISO_PSP_GAME_PROVENANCE",
):
    require(anchor in disc, "shared ISO authoring lost staged-PSP_GAME provenance contract: " + anchor)

print("ENGLISH_FIRST_PARITY_PASS")
print("english_consumers=" + ",".join(sorted(english_consumers)))
print("korean_direct_consumers=" + ",".join(sorted(korean_consumers)))
print("korean_c5_split=" + ",".join(sorted(deliberate_split)))
print("shared_capacity_constants=" + ",".join(sorted(english_constants)))
print("release_entrypoints=desktop,mobile,preflight")
print("font_source_fingerprints=english_manifest_exact")
print("font_result_boundary=dynamic_exact_mutation_surface")
print("executable_manifest_chain=shared_and_postverified")
print("slot_ownership=authenticated_exact_byte_fail_closed")
print("archive_rebuild=shared_duplicate_reject_and_exact_payload_verify")
print("iso_provenance=shared_staged_psp_game_exact_verify")
