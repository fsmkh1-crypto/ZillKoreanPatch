import importlib.util
from pathlib import Path
import sys
import unittest

MODULE_PATH = Path(__file__).with_name("qa-voice.py")
sys.path.insert(0, str(MODULE_PATH.parent))
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

    def test_branches_are_aligned_when_control_skeleton_matches(self):
        ja = "<if><value:$01><equal>1お前だ<end>あなたです<end>"
        ko = "<if><value:$01><equal>1너야<end>당신입니다<end>"
        self.assertTrue(mod.branch_alignment_safe(ja, ko))
        self.assertEqual(
            mod.paired_segments(ja, ko),
            [(0, "<if><value:$01><equal>1お前だ", "<if><value:$01><equal>1너야"),
             (1, "あなたです", "당신입니다")],
        )

    def test_equal_end_count_does_not_override_control_mismatch(self):
        ja = "<if><value:$01><equal>1お前だ<end>あなたです<end>"
        ko = "<if><value:$02><equal>1너야<end>당신입니다<end>"
        self.assertFalse(mod.branch_alignment_safe(ja, ko))
        self.assertEqual(mod.paired_segments(ja, ko), [(0, ja, ko)])

    def test_unaligned_end_count_falls_back_to_whole_record(self):
        ja = "お前だ<end>あなたです<end>"
        ko = "너야<end>"
        self.assertFalse(mod.branch_alignment_safe(ja, ko))
        self.assertEqual(mod.paired_segments(ja, ko), [(0, ja, ko)])


if __name__ == "__main__":
    unittest.main()
