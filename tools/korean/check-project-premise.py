#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""Fail CI if the repository-level English-patch parity premise disappears."""

from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
AGENTS = ROOT / "AGENTS.md"
PR_TEMPLATE = ROOT / ".github" / "pull_request_template.md"

REQUIRED_AGENTS = (
    "NON-NEGOTIABLE PROJECT PREMISE — ENGLISH PATCH FIRST",
    "primary reference for the Korean patch",
    "A Korean-specific heuristic MUST NOT supersede an established English-patch contract",
    "PASS",
    "DIFFERENT-BY-DESIGN",
    "MISSING",
    "UNKNOWN",
    "Unexplained divergence from an established English-patch engine contract is a release blocker.",
)

REQUIRED_PR = (
    "English patch parity checked: YES / N/A",
    "English reference/consumer contract:",
    "Korean divergence, if any:",
    "Evidence for divergence:",
)


def require(path: Path, needles: tuple[str, ...]) -> None:
    if not path.is_file():
        raise SystemExit(f"PROJECT_PREMISE_MISSING file={path.relative_to(ROOT)}")
    text = path.read_text(encoding="utf-8")
    missing = [needle for needle in needles if needle not in text]
    if missing:
        raise SystemExit(
            f"PROJECT_PREMISE_INCOMPLETE file={path.relative_to(ROOT)} "
            + "missing="
            + repr(missing)
        )


def main() -> None:
    require(AGENTS, REQUIRED_AGENTS)
    require(PR_TEMPLATE, REQUIRED_PR)
    print("PROJECT_PREMISE_OK english_patch_first=true pr_evidence_required=true")


if __name__ == "__main__":
    main()
