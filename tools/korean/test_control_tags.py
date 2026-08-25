#!/usr/bin/env python3
from __future__ import annotations

import unittest

from control_tags import fixed_tokens, runtime_tokens


class ControlTagTests(unittest.TestCase):
    def test_bracketed_natural_text_is_not_runtime_control(self) -> None:
        text = "<未使用><value:$15>本文<line-break><미사용><end>"
        self.assertEqual(
            runtime_tokens(text),
            ["<value:$15>", "<line-break>", "<end>"],
        )
        self.assertEqual(fixed_tokens(text), ["<value:$15>", "<end>"])

    def test_display_projection_forms_are_recognized(self) -> None:
        text = (
            "<if><select><call:12><jump:34><value:$1F><add><not-equal>"
            "<operator:$04:$7B><color:A><discard:C:$01><escape:$48>"
            "<separator><backspace><tab><$01><end>"
        )
        self.assertEqual(len(runtime_tokens(text)), 16)

    def test_line_break_is_layout_owned(self) -> None:
        source = "문장<line-break>다음<end>"
        translated = "문장 다음<end>"
        self.assertEqual(fixed_tokens(source), fixed_tokens(translated))


if __name__ == "__main__":
    unittest.main()
