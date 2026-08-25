#!/usr/bin/env python3
from __future__ import annotations

import importlib.util
from pathlib import Path
import unittest

from control_tags import fixed_tokens, runtime_tokens


def load_script(filename: str, module_name: str):
    path = Path(__file__).with_name(filename)
    spec = importlib.util.spec_from_file_location(module_name, path)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"cannot load {path}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


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

    def test_refresh_keeps_legitimate_korean_reflow(self) -> None:
        refresh = load_script("refresh-japanese-refs.py", "refresh_japanese_refs")
        japanese = "汝、無限のソウルを持つ者よ<line-break>我に応ぜよ<line-break>答えよ<end>"
        korean = "그대, 무한의 소울을 지닌 자여 나에게 응답하라<line-break>대답하라<end>"
        self.assertIsNone(refresh.record_problem(japanese, japanese, korean))

    def test_refresh_uses_full_fixed_control_contract(self) -> None:
        refresh = load_script("refresh-japanese-refs.py", "refresh_japanese_refs_full")
        japanese = "<call:12><operator:$04:$7B><value:$1F>本文<end>"
        korean = "<call:12><operator:$04:$7B><value:$1F>본문<end>"
        self.assertIsNone(refresh.record_problem(japanese, japanese, korean))
        bad = "<call:13><operator:$04:$7B><value:$1F>본문<end>"
        self.assertEqual(
            refresh.record_problem(japanese, japanese, bad),
            "fixed control-token mismatch",
        )

    def test_legacy_linebreak_migration_separates_semantics_and_layout(self) -> None:
        migrate = load_script("migrate-semantic-linebreaks.py", "migrate_semantic_linebreaks")
        legacy = "그대, 무한의 소울을 지닌 자여 <line-break> 나에게 응답하라<line-break>대답하라<end>"
        self.assertEqual(
            migrate.semanticize(legacy),
            "그대, 무한의 소울을 지닌 자여 나에게 응답하라 대답하라<end>",
        )
        self.assertEqual(
            migrate.layoutize(legacy),
            "그대, 무한의 소울을 지닌 자여<line-break>나에게 응답하라<line-break>대답하라<end>",
        )
        self.assertNotIn("<line-break>", migrate.semanticize(legacy))


if __name__ == "__main__":
    unittest.main()
