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
english_engine = read("internal/layout/engine.go")
korean_validate = read("internal/layout/validate_korean_english_contract.go")
korean_c5 = read("internal/layout/validate_korean.go")
korean_visual = read("internal/layout/validate_korean_visual.go")
korean_warnings = read("internal/layout/validate_korean_warnings.go")
english_compile = read("internal/message/compile.go")
korean_compile = read("internal/message/compile_korean.go")
korean_materialize = read("internal/message/korean_materialize.go")
english_release = read("internal/release/build.go")
release_build = read("internal/release/korean_build.go")
mobile_build = read("internal/release/korean_mobile.go")
mobile_preflight = read("internal/release/korean_mobile_preflight.go")
contract_chain = read("internal/release/korean_contract_chain.go")
korean_font = read("internal/release/korean_font.go")
korean_fixed = read("internal/release/korean_fixed.go")
korean_fixeddata = read("internal/fixeddata/korean_eboot.go")
equipment_fixeddata = read("internal/fixeddata/equipment.go")
desktop_plan = read("cmd/zill/build_korean.go")
mobile_plan = read("cmd/zill/build_korean_mobile_plan.go")
slot_plan = read("internal/koreanslots/plan.go")
paa = read("internal/gamefmt/paa/paa.go")
disc = read("internal/release/disc.go")
english_font_manifest = read("release/font/manifest.toml")
executable_manifest = read("patches/executable/manifest.toml")
rules = read("internal/layout/rules.go")
premise = read("AGENTS.md")
android_activity = read("android-patcher/app/src/main/java/com/fsmkh1/zillfontdump/MainActivity.java")
android_payload_integrity = read("android-patcher/app/src/main/java/com/fsmkh1/zillfontdump/ProjectAssetIntegrity.java")
android_application = read("android-patcher/app/src/main/java/com/fsmkh1/zillfontdump/PayloadRepairApplication.java")
android_workflow = read(".github/workflows/android-korean-a054-rc.yml")

require("NON-NEGOTIABLE PROJECT PREMISE" in premise and "ENGLISH PATCH FIRST" in premise,
        "AGENTS.md no longer contains the English-first project premise")

# Consumer and capacity parity.
consumer_re = re.compile(r"e\.consumers\.([A-Za-z0-9_]+)")
english_consumers = set(consumer_re.findall(english_validate))
korean_consumers = set(consumer_re.findall(korean_validate))
korean_c5_consumers = set(consumer_re.findall(korean_c5))
deliberate_split = {"C5IDs", "SinglePageC5IDs"}
missing = english_consumers - korean_consumers - deliberate_split
require(not missing, "Korean validator lost English consumer references: " + ",".join(sorted(missing)))
require(deliberate_split <= korean_c5_consumers, "documented C5 split lost Korean C5 consumer membership")

rule_constant_re = re.compile(r"^\s*([a-z][A-Za-z0-9]*(?:CapacityBytes|MaxPayloadBytes|MaxLineBytes|MaxPages))\s*=", re.M)
rule_constants = set(rule_constant_re.findall(rules))
english_constants = {n for n in rule_constants if re.search(r"\b" + re.escape(n) + r"\b", english_validate)}
korean_constants = {n for n in rule_constants if re.search(r"\b" + re.escape(n) + r"\b", korean_validate) or re.search(r"\b" + re.escape(n) + r"\b", korean_c5)}
require(not (english_constants - korean_constants), "Korean validation lost English capacity constants")

# Upstream English Reflow has two hard character-profile visual conditions and
# non-blocking warning families in engine.go. Korean must mirror both severity
# classes; only renderer measurement is allowed to differ.
for anchor in ("profileAdvance", "profileMaxLines"):
    require(anchor in korean_visual, "Korean hard visual parity lost " + anchor)
for code in ("line_exceeds_authoring_ceiling", "item_description_single_line_overflow",
             "chronicle_vertical_overflow", "runtime_substitution_unbounded",
             "guild_job_client_overflow"):
    require(code in english_engine and code in korean_warnings,
            "Korean warning parity lost upstream code " + code)

# Semantic/materialization/compiler parity.
require("p.splitSemanticWith" in korean_materialize, "Korean semantic traversal lost shared parser")
require("p.materializeValues" in korean_materialize, "Korean materialization lost shared value lowering")
for anchor in ("RuntimeBankCapacity(bank.Section)", "binary.LittleEndian.PutUint32"):
    require(anchor in english_compile and anchor in korean_compile, "bank compiler contract drift: " + anchor)

# All release entry points must reach one shared ordered contract chain before
# compilation. Do not require duplicated direct gates in each release file: that
# was precisely the structure that allowed desktop/mobile/preflight drift.
def require_release_chain(label: str, text: str, call: str) -> None:
    shared = text.find(call)
    compile_gate = text.find("compileKoreanBanksWithPlan")
    require(shared >= 0 and compile_gate >= 0, f"{label} missing shared parity chain")
    require(shared < compile_gate, f"{label} compiles before shared English-first validation")


require_release_chain("desktop", release_build,
                      'runKoreanEnglishContractChain(root, "desktop", source, korean, plan.Mapping)')
require_release_chain("mobile", mobile_build,
                      'runKoreanEnglishContractChain(root, "mobile", source, korean, plan.Mapping)')
require_release_chain("preflight", mobile_preflight,
                      'runKoreanEnglishContractChain(root, "preflight", source, korean, plan.Mapping)')

ordered_contract_stages = (
    "DeriveKoreanC5StorageLayouts",
    "DeriveKoreanEnglishConsumerLayouts",
    "DeriveKoreanEnglishVisualLayouts",
    "DeriveKoreanC22RetailScannerLayouts",
    "ValidateKoreanEnglishConsumerContracts",
    "AuditKoreanEnglishVisualWarnings",
    "validateKoreanRuntimeStorage",
)
positions = [contract_chain.find(stage) for stage in ordered_contract_stages]
require(all(p >= 0 for p in positions), "shared Korean contract chain lost a required stage")
require(positions == sorted(positions), "shared Korean contract-chain order drifted")

# Font source authentication remains pinned to upstream English retail assets.
english_font_hashes = re.findall(r'^source_sha256\s*=\s*"([0-9a-f]{64})"', english_font_manifest, re.M)
require(len(english_font_hashes) == 2, "English font manifest source fingerprints drifted")
for digest in english_font_hashes:
    require(digest in korean_font, "Korean font path lost English retail source fingerprint " + digest)

# Executable and fixed-data contracts.
require(re.search(r'^source_sha256\s*=\s*"[0-9a-f]{64}"', executable_manifest, re.M) is not None,
        "executable source fingerprint missing")
require(re.search(r'^result_sha256\s*=\s*"[0-9a-f]{64}"', executable_manifest, re.M) is not None,
        "executable result fingerprint missing")
require("elfpatch.VerifyApplied(result, manifest)" in korean_fixed,
        "Korean EBOOT overlay lost executable patch postcondition")
for anchor in ("sha256.Sum256(source) != patchedELFSHA256", "len(encoded) > len(expected)",
               "source[start+len(expected)] != 0", "translation fields overlap"):
    require(anchor in korean_fixeddata, "Korean fixed EBOOT guard disappeared: " + anchor)
for anchor in ("sha256.Sum256(source) != bindataSHA256", "equipmentRecordCount = 132", "equipmentNameSize    = 17"):
    require(anchor in equipment_fixeddata, "BINDATA guard disappeared: " + anchor)

# Renderer ownership must be based on evidence that the engine interprets as
# text. Whole-blob exact two-byte occurrence was experimentally disproven as an
# ownership rule and is prohibited by the shared planner API itself.
for label, text in (("desktop", desktop_plan), ("mobile", mobile_plan)):
    for anchor in ("loadAuthenticatedRetailBOOT(gameDir)", "loadAuthenticatedRetailEBOOT(gameDir)",
                   "loadRetailBindata(gameDir)", "slotaudit.ScanCP932Literals(boot)",
                   "slotaudit.ScanCP932Literals(bindata)", "mergeRendererKeys(reserved, bootScan.Keys)",
                   "mergeRendererKeys(reserved, bindataScan.Keys)"):
        require(anchor in text, f"{label} renderer ownership lost structured evidence: {anchor}")
require("koreanslots.BuildPlan(texts, font.KoreanCompatibleKeys(), rendererKeySetSlice(reserved))" in desktop_plan,
        "desktop planner regressed from structured ownership")
require("koreanslots.BuildPlan(texts, installed, rendererKeySetSlice(reserved))" in mobile_plan,
        "mobile planner regressed from structured ownership")
require("authenticatedBlobs ...[]byte" not in slot_plan and "ExcludeExactByteReferences" not in slot_plan,
        "shared planner API reintroduced arbitrary whole-blob ownership")
require("finalAudit.CandidateHits != 0 || len(finalAudit.MappedHits) != 0" not in mobile_plan,
        "mobile whole-blob aliases became a hard failure again")

# Archive and ISO provenance.
for anchor in ("member %d is selected by more than one replacement", "verifyRebuilt(", "payload differs"):
    require(anchor in paa, "PAA rebuild contract weakened: " + anchor)
for text, label in ((english_release, "English"), (release_build, "Korean desktop"), (mobile_build, "Korean mobile")):
    require("archive.pair.Rebuild(" in text, label + " release bypasses verified archive rebuild")
    require("authorTranslatedISO(" in text, label + " release bypasses verified ISO authoring")
for anchor in ("verifyAuthoredPSPGame(outputPath, gameDir)", "compareExactReaders(got, want)"):
    require(anchor in disc, "ISO provenance contract weakened: " + anchor)

# Android payload/cache/export provenance.
for anchor in ("python3 tools/korean/audit-english-first-full-parity.py", "payload-manifest.sha256", "sha256sum -c payload-manifest.sha256"):
    require(anchor in android_workflow, "Android release provenance lost: " + anchor)
for anchor in ("static void verifyPayload(File root)", "payload manifest digest mismatch", "payload file set mismatch"):
    require(anchor in android_payload_integrity, "Android payload integrity weakened: " + anchor)
require("ProjectAssetIntegrity.verifyPayload(root)" in android_application, "Android startup cache verification disappeared")
for anchor in ('"build-korean-iso"', '"--preflight-only"', "ProjectAssetIntegrity.verifyPayload(root)",
               "verifyFileMatchesUri(output, outputUri)", 'MessageDigest.getInstance("SHA-256")'):
    require(anchor in android_activity, "Android build/export path lost provenance anchor: " + anchor)

print("ENGLISH_FIRST_PARITY_PASS")
print("english_consumers=" + ",".join(sorted(english_consumers)))
print("korean_c5_split=" + ",".join(sorted(deliberate_split)))
print("shared_capacity_constants=" + ",".join(sorted(english_constants)))
print("visual_warning_parity=hard_profile_plus_upstream_warning_codes")
print("release_entrypoints=shared_chain_desktop,mobile,preflight")
print("slot_ownership=structured_cp932_fixed_renderer_evidence_api_enforced")
print("whole_blob_exact_byte_aliases=diagnostic_only")
print("archive_iso_android_provenance=verified")
