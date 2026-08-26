#!/usr/bin/env python3
from __future__ import annotations

import importlib.util
from pathlib import Path
import tempfile
import unittest

MODULE_PATH = Path(__file__).with_name("auto-safe.py")
spec = importlib.util.spec_from_file_location("auto_safe", MODULE_PATH)
assert spec and spec.loader
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)


class AutoSafeTests(unittest.TestCase):
    def test_safe_numeric_patterns(self) -> None:
        self.assertEqual(mod.safe_pattern("8月26日<end>"), "8월 26일<end>")
        self.assertEqual(mod.safe_pattern("第12章<end>"), "제12장<end>")
        self.assertIsNone(mod.safe_pattern("王都8月<end>"))

    def test_tm_only_propagates_unambiguous_sources(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "msgsec001.toml").write_text(
                '# SPDX-License-Identifier: CC-BY-SA-4.0\n\n'
                '["10000"]\n'
                'japanese = "同じ<end>"\n'
                'korean = "같음<end>"\n\n'
                '["10001"]\n'
                'japanese = "曖昧<end>"\n'
                'korean = "첫째<end>"\n',
                encoding="utf-8",
            )
            (root / "msgsec002.toml").write_text(
                '# SPDX-License-Identifier: CC-BY-SA-4.0\n\n'
                '["20000"]\n'
                'japanese = "同じ<end>"\n'
                'korean = "같음<end>"\n\n'
                '["20001"]\n'
                'japanese = "曖昧<end>"\n'
                'korean = "둘째<end>"\n',
                encoding="utf-8",
            )
            old = mod.KOREAN
            try:
                mod.KOREAN = root
                tm = mod.load_translation_memory()
            finally:
                mod.KOREAN = old
            self.assertEqual(tm["同じ<end>"], "같음<end>")
            self.assertNotIn("曖昧<end>", tm)


if __name__ == "__main__":
    unittest.main()
