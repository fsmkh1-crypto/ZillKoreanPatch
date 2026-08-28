import importlib.util
from pathlib import Path
import sys
import tempfile
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

    def test_humble_first_person_matches_real_na_pronoun(self):
        self.assertIn("humble_first_person_blunt_ko", mod.classify("わたくしは行くわ", "나는 갈게."))

    def test_humble_first_person_does_not_match_naga_verb(self):
        self.assertNotIn(
            "humble_first_person_blunt_ko",
            mod.classify("わたくしの仲間です", "제 동료가 전장에 나가 있습니다."),
        )

    def test_rough_first_person_matches_real_je_pronoun(self):
        self.assertIn("rough_first_person_polite_ko", mod.classify("俺がやるっす", "제가 하겠습니다."))

    def test_ore_sama_inside_third_person_title_is_not_first_person(self):
        self.assertNotIn(
            "rough_first_person_polite_ko",
            mod.classify("俺様天下のレムオン様にはわからない", "천하의 레몬 님께서는 모르시겠지만 저는 압니다."),
        )

    def test_adjective_kai_is_not_blunt_question_particle(self):
        self.assertNotIn(
            "blunt_jp_formal_ko",
            mod.classify("あなたの手は温かい。安らぎを与えてくれます。", "당신의 손은 따뜻하군요. 안식을 줍니다."),
        )

    def test_question_particle_kai_still_counts(self):
        self.assertIn(
            "blunt_jp_formal_ko",
            mod.classify("冒険者登録かい？", "모험자 등록입니까?"),
        )

    def test_exception_registry_validation(self):
        with tempfile.TemporaryDirectory() as td:
            path = Path(td) / "exceptions.json"
            path.write_text(
                '[{"id":"123","segment":0,"kind":"rough_first_person_polite_ko",'
                '"category":"context","reason":"reviewed"}]',
                encoding="utf-8",
            )
            got = mod.load_reviewed_exceptions(path)
            self.assertEqual(got[("123", 0, "rough_first_person_polite_ko")]["category"], "context")

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
