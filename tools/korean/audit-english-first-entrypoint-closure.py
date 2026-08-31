#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""Fail closed if any production Korean entry point can bypass audited engine contracts."""
from __future__ import annotations

import pathlib
import xml.etree.ElementTree as ET

ROOT = pathlib.Path(__file__).resolve().parents[2]
ANDROID_NS = "{http://schemas.android.com/apk/res/android}"


def read(path: str) -> str:
    return (ROOT / path).read_text(encoding="utf-8")


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit("ENTRYPOINT_PARITY_FAIL " + message)


desktop = read("cmd/zill/build_korean.go")
mobile = read("cmd/zill/build_korean_mobile_plan.go")
mobile_command = read("cmd/zill/build_korean_iso.go")
zill_main = read("cmd/zill/main.go")
release_desktop = read("internal/release/korean_build.go")
release_mobile = read("internal/release/korean_mobile.go")
release_preflight = read("internal/release/korean_mobile_preflight.go")
release_mobile_prepare = read("internal/release/korean_mobile_prepare.go")
contract_chain = read("internal/release/korean_contract_chain.go")
slot_plan = read("internal/koreanslots/plan.go")
manifest_path = ROOT / "android-patcher/app/src/main/AndroidManifest.xml"
manifest_text = manifest_path.read_text(encoding="utf-8")

# Both planners must authenticate the same engine-owned sources. Renderer-slot
# reservations may come only from structured evidence the engine interprets as
# text: fixed strings and CP932 literal scans, never arbitrary whole-blob pairs.
for label, text in (("desktop", desktop), ("mobile", mobile)):
    for anchor in (
        "loadAuthenticatedRetailBOOT(gameDir)",
        "loadAuthenticatedRetailEBOOT(gameDir)",
        "loadRetailBindata(gameDir)",
        "fixeddata.ApplyEquipment(bindata, equipment)",
        "slotaudit.ScanCP932Literals(boot)",
        "slotaudit.ScanCP932Literals(bindata)",
        "mergeRendererKeys(reserved, bootScan.Keys)",
        "mergeRendererKeys(reserved, bindataScan.Keys)",
    ):
        require(anchor in text, f"{label} planner lost structured renderer ownership input: {anchor}")

require(
    "koreanslots.BuildPlan(texts, font.KoreanCompatibleKeys(), rendererKeySetSlice(reserved))" in desktop,
    "desktop planner no longer allocates from structured renderer reservations",
)
require(
    "koreanslots.BuildPlan(texts, installed, rendererKeySetSlice(reserved))" in mobile,
    "mobile planner no longer allocates from structured renderer reservations",
)
require("authenticatedBlobs ...[]byte" not in slot_plan,
        "shared slot planner API again accepts arbitrary whole-blob ownership evidence")
require("ExcludeExactByteReferences" not in slot_plan,
        "shared slot planner again performs whole-blob exact-byte exclusion")

# Desktop is the older atlas-only path: it must stay on geometry-compatible
# slots and must not mutate the chosen mapping after BuildPlan. Mobile performs
# a full atlas+PAF repack and may relocate private 0x87 keys. Whole-blob exact
# byte scans after allocation are telemetry only, never a hard gate.
require("font.KoreanCompatibleKeys()" in desktop,
        "desktop atlas-only planner lost geometry-compatible slot restriction")
require("plan.Mapping =" not in desktop,
        "desktop planner unexpectedly mutates mapping after allocation")
require("installed := font.DoubleByteKeys()" in mobile,
        "mobile full-repack planner lost explicit double-byte slot universe")
require("plan.Mapping = mapping" in mobile,
        "mobile private-key relocation disappeared; reclassify mapping policy")
require('printMobileSlotAudit("final-diagnostic-only"' in mobile,
        "mobile whole-blob collision telemetry disappeared; reclassify evidence boundary")
require("finalAudit.CandidateHits != 0 || len(finalAudit.MappedHits) != 0" not in mobile,
        "mobile whole-blob byte aliases became a release-blocking ownership rule again")

# Production releases must reach one shared English-first chain. Desktop calls
# it directly. Mobile build and preflight intentionally converge first through
# one deterministic preparation helper; requiring direct chain calls in both
# callers would force the exact duplicated orchestration this audit should ban.
desktop_call = 'runKoreanEnglishContractChain(root, "desktop", source, korean, plan.Mapping)'
require(desktop_call in release_desktop,
        "desktop release bypasses shared English-first contract chain")

for label, text, mode in (
    ("mobile", release_mobile, "mobile"),
    ("preflight", release_preflight, "preflight"),
):
    prepare_call = f'prepareKoreanMobilePayload(root, gameDir, version, "{mode}", planBuilder)'
    require(prepare_call in text,
            f"{label} release bypasses shared deterministic mobile preparation")
    require("runKoreanEnglishContractChain(" not in text,
            f"{label} release reintroduced a direct contract-chain implementation outside shared preparation")
    require("compileKoreanBanksWithPlan(" not in text,
            f"{label} release reintroduced direct bank compilation outside shared preparation")

mobile_contract = release_mobile_prepare.find(
    "runKoreanEnglishContractChain(root, mode, source, korean, plan.Mapping)"
)
mobile_compile = release_mobile_prepare.find("compileKoreanBanksWithPlan(")
require(mobile_contract >= 0,
        "shared mobile preparation bypasses shared English-first contract chain")
require(mobile_compile >= 0,
        "shared mobile preparation lost Korean bank compilation")
require(mobile_contract < mobile_compile,
        "shared mobile preparation compiles before English-first contract validation")

# None of the release wrappers/preparation helpers may independently reproduce
# the derivation/validation stages owned by korean_contract_chain.go.
for label, text in (
    ("desktop", release_desktop),
    ("mobile", release_mobile),
    ("preflight", release_preflight),
    ("mobile_prepare", release_mobile_prepare),
):
    require("DeriveKoreanC5StorageLayouts(" not in text,
            f"{label} release duplicated C5 derivation outside shared chain")
    require("DeriveKoreanEnglishConsumerLayouts(" not in text,
            f"{label} release duplicated consumer derivation outside shared chain")
    require("DeriveKoreanEnglishVisualLayouts(" not in text,
            f"{label} release duplicated visual derivation outside shared chain")
    require("ValidateKoreanEnglishConsumerContracts(" not in text,
            f"{label} release duplicated hard contract validation outside shared chain")

for anchor in (
    "DeriveKoreanC5StorageLayouts(",
    "DeriveKoreanEnglishConsumerLayouts(",
    "DeriveKoreanEnglishVisualLayouts(",
    "DeriveKoreanC22RetailScannerLayouts(",
    "ValidateKoreanEnglishConsumerContracts(",
    "AuditKoreanEnglishVisualWarnings(",
    "validateKoreanRuntimeStorage(",
):
    require(anchor in contract_chain, f"shared Korean contract chain lost required stage: {anchor}")

# Command routing must have exactly one production mobile build route and one
# preflight route, both using the bound mobile planner and release package.
require('case "build-korean":' in zill_main and
        'return runBuildKorean(root, args[1:], stdout, stderr)' in zill_main,
        "desktop build command no longer routes through audited runBuildKorean")
require('case "build-korean-iso":' in zill_main and
        'return runBuildKoreanISO(root, args[1:], stdout, stderr)' in zill_main,
        "mobile build command no longer routes through audited runBuildKoreanISO")
require("return buildKoreanAlphaPlanMobile(root, gameDir, source, korean)" in mobile_command,
        "mobile command no longer binds source/Korean project to audited mobile planner")
require("release.PreflightKoreanAlphaISOOnly(" in mobile_command,
        "mobile preflight no longer reaches audited release preflight")
require("release.BuildKoreanAlphaISOOnly(" in mobile_command,
        "mobile production build no longer reaches audited ISO-only release path")

# Android must expose MainActivity as the sole launcher. FreezeCaptureActivity is
# a forensic utility and must remain non-exported and without a launcher filter.
root = ET.parse(manifest_path).getroot()
app = root.find("application")
require(app is not None, "Android manifest has no application element")
activities = app.findall("activity")
by_name = {a.get(ANDROID_NS + "name"): a for a in activities}
main = by_name.get(".MainActivity")
freeze = by_name.get(".FreezeCaptureActivity")
require(main is not None, "Android MainActivity disappeared")
require(freeze is not None, "Android forensic FreezeCaptureActivity disappeared; reclassify tracer boundary")
require(main.get(ANDROID_NS + "exported") == "true", "MainActivity is no longer exported launcher")
require(freeze.get(ANDROID_NS + "exported") == "false", "FreezeCaptureActivity became externally exported")


def is_launcher(activity: ET.Element) -> bool:
    for filt in activity.findall("intent-filter"):
        actions = {x.get(ANDROID_NS + "name") for x in filt.findall("action")}
        categories = {x.get(ANDROID_NS + "name") for x in filt.findall("category")}
        if "android.intent.action.MAIN" in actions and "android.intent.category.LAUNCHER" in categories:
            return True
    return False


launchers = [name for name, activity in by_name.items() if is_launcher(activity)]
require(launchers == [".MainActivity"],
        "Android launcher set drifted from sole audited MainActivity route: " + repr(launchers))
require(".FreezeCaptureActivity" in manifest_text,
        "manifest text unexpectedly lost forensic activity declaration")

print("ENTRYPOINT_PARITY_PASS")
print("desktop_slot_ownership=structured_cp932_and_fixed_renderer_evidence")
print("mobile_slot_ownership=structured_cp932_and_fixed_renderer_evidence_post_relocation_whole_blob_diagnostic_only")
print("release_contract_chain=desktop_direct_mobile_shared_prepare_preflight_shared_prepare")
print("mobile_command=bound_planner_to_preflight_or_iso_only_release")
print("android_launcher=mainactivity_only_freeze_capture_nonexported")
