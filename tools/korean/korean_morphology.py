#!/usr/bin/env python3
"""Shared Korean particle-boundary helpers for candidate detection and legacy audits."""

import re

# A particle match may end at string/end punctuation, or be followed by a small
# set of auxiliary particles that were intentionally supported by the Beta1
# refined scanner. Keep this in one module so scanners/fixers cannot drift.
AUX_FOLLOW = r"(?=$|[^가-힣]|(?:는|도|만|부터|까지)(?=[^가-힣]|$))"


def jongseong_index(syllable: str) -> int:
    code = ord(syllable)
    if not (0xAC00 <= code <= 0xD7A3):
        raise ValueError(f"not a Hangul syllable: {syllable!r}")
    return (code - 0xAC00) % 28


def wrong_particle_pairs(word: str) -> list[tuple[str, str]]:
    jong = jongseong_index(word[-1])
    if jong == 0:
        return [("이", "가"), ("을", "를"), ("과", "와"), ("은", "는"), ("으로", "로")]
    pairs = [("가", "이"), ("를", "을"), ("와", "과"), ("는", "은")]
    if jong == 8:  # ㄹ-final takes 로
        pairs.append(("으로", "로"))
    else:
        pairs.append(("로", "으로"))
    return pairs


def wrong_particle_pattern(word: str, wrong: str) -> re.Pattern[str]:
    return re.compile(re.escape(word + wrong) + AUX_FOLLOW)


def explicit_particle_patterns(words: list[str]):
    patterns = []
    for word in words:
        for wrong, right in wrong_particle_pairs(word):
            patterns.append((
                f"wrong_particle_{word}_{wrong}",
                wrong_particle_pattern(word, wrong),
                word + right,
            ))
    return patterns
