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
manifest_path = ROOT / "android-patcher/app/src/main/AndroidManifest.xml"
manifest_text = manifest_path.read_text(encoding="utf-8")

# Both planners must authenticate the same engine-owned blobs and hand those
# exact bytes to BuildPlan. Literal CP932 scans are only supplementary evidence;
# exact-byte exclusion is the production ownership boundary.
for label, text in (("desktop", desktop), ("mobile", mobile)):
    for anchor in (
        "loadAuthenticatedRetailBOOT(gameDir)",
        "loadAuthenticatedRetailEBOOT(gameDir)",
        "loadRetailBindata(gameDir)",
        "fixeddata.ApplyEquipment(bindata, equipment)",
        "boot, eboot, bindata",
    ):
        require(anchor in text, f"{label} planner lost authenticated ownership input: {anchor}")

require(
    "koreanslots.BuildPlan(texts, font.KoreanCompatibleKeys(), rendererKeySetSlice(reserved), boot, eboot, bindata)"
    in desktop,
    "desktop planner no longer gives authenticated BOOT/EBOOT/BINDATA bytes to BuildPlan",
)
require(
    "koreanslots.BuildPlan(texts, installed, rendererKeySetSlice(reserved), boot, eboot, bindata)"
    in mobile,
    "mobile planner no longer gives authenticated BOOT/EBOOT/BINDATA bytes to BuildPlan",
)

# Desktop is the older atlas-only path: it must stay on geometry-compatible
# slots and must not mutate the chosen mapping after BuildPlan. Mobile performs
# a full atlas+PAF repack and may relocate private 0x87 keys, but therefore must
# re-audit exact-byte ownership after that mutation.
require("font.KoreanCompatibleKeys()" in desktop,
        "desktop atlas-only planner lost geometry-compatible slot restriction")
require("plan.Mapping =" not in desktop,
        "desktop planner now mutates mapping after exact-byte ownership audit; add an explicit post-mutation audit")
require("installed := font.DoubleByteKeys()" in mobile,
        "mobile full-repack planner lost explicit double-byte slot universe")
require("plan.Mapping = mapping" in mobile,
        "mobile private-key relocation disappeared; reclassify the post-mutation audit requirement")
require("finalAudit, err := auditMobileExactByteReuse(plan, blobs...)" in mobile,
        "mobile planner mutates mapping without final authenticated exact-byte audit")
require("finalAudit.CandidateHits != 0 || len(finalAudit.MappedHits) != 0" in mobile,
        "mobile final exact-byte audit no longer fails closed")

# Command routing must have exactly one production mobile build route and one
# preflight route, both using the bound mobile planner and the release package
# whose English-first gates are audited separately.
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
print("desktop_slot_ownership=authenticated_exact_bytes_no_post_buildplan_mutation")
print("mobile_slot_ownership=authenticated_exact_bytes_post_relocation_reaudited")
print("mobile_command=bound_planner_to_preflight_or_iso_only_release")
print("android_launcher=mainactivity_only_freeze_capture_nonexported")
