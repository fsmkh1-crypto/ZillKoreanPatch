#!/usr/bin/env python3
"""Synthetic canonical message markers that must stay source-owned.

These strings are metadata/placeholders in the retail corpus, not player-visible
natural text. They must never enter Korean translation packets or translation
memory because the Korean materializer intentionally rejects unknown angle-
bracket markup in semantic payloads.
"""
from __future__ import annotations

SYNTHETIC_SOURCE_TEXTS = frozenset({"<未使用><end>"})


def is_synthetic_source(text: str) -> bool:
    return text in SYNTHETIC_SOURCE_TEXTS
