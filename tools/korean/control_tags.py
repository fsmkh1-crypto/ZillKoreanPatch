#!/usr/bin/env python3
"""Recognize only canonical runtime-control projections in message text.

Angle-bracketed natural game text (for example ``<未使用>``) is translatable
and must not be confused with bytecode. ``<line-break>`` is deliberately
excluded from the fixed contract because Korean wrapping is build-owned layout.
"""
from __future__ import annotations

import re

RUNTIME_CONTROL_RE = re.compile(
    r"<(?:"
    r"if|select|call:[0-9]+|jump:[0-9]+|value:\$[0-9A-F]{2}|"
    r"add|subtract|multiply|divide|modulo|equal|not-equal|less|greater|"
    r"less-equal|greater-equal|and|or|operator:\$[0-9A-F]{2}:\$[0-9A-F]{2}|"
    r"color:[^<>]|discard:[^<>]:\$[0-9A-F]{2}|escape:\$[0-9A-F]{2}|"
    r"end|separator|backspace|tab|line-break|\$[0-9A-F]{2}"
    r")>"
)


def runtime_tokens(text: str) -> list[str]:
    return RUNTIME_CONTROL_RE.findall(text)


def fixed_tokens(text: str) -> list[str]:
    return [token for token in runtime_tokens(text) if token != "<line-break>"]
