import importlib.util
from pathlib import Path
import unittest

MODULE_PATH = Path(__file__).with_name("qa-voice.py")
spec = importlib.util.spec_from_file_location("qa_voice", MODULE_PATH)
mod = importlib.util.module_from_spec(spec)
assert spec.loader is not None
spec.loader.exec_module(mod)


class VoiceHazardTest(unittest.TestCase):
    def test_hostile_polite(self):
        self.assertIn("hostile_jp_polite_ko", mod.classify("貴様は何者だ", "당신은 누구지"))

    def test_hostile_hostile_is_clean(self):
        self.assertEqual(mod.classify("貴様は何者だ", "네놈은 누구냐"), [])

    def test_neutral_hostile(self):
        self.assertIn("neutral_jp_hostile_ko", mod.classify("あなたは誰？", "네놈은 누구냐?"))

    def test_branches_are_aligned(self):
        pairs = mod.paired_segments("お前だ<end>あなたです<end>", "너야<end>당신입니다<end>")
        self.assertEqual(pairs, [(0, "お前だ", "너야"), (1, "あなたです", "당신입니다")])

    def test_unaligned_falls_back_to_whole_record(self):
        pairs = mod.paired_segments("お前だ<end>あなたです<end>", "너야<end>")
        self.assertEqual(len(pairs), 1)


if __name__ == "__main__":
    unittest.main()
