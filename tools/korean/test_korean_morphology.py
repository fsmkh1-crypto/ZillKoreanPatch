import unittest

import korean_morphology as morphology


class KoreanMorphologyTests(unittest.TestCase):
    def test_vowel_final_particle_pairs(self):
        self.assertIn(("을", "를"), morphology.wrong_particle_pairs("발로르"))
        self.assertIn(("이", "가"), morphology.wrong_particle_pairs("발로르"))

    def test_consonant_final_particle_pairs(self):
        self.assertIn(("를", "을"), morphology.wrong_particle_pairs("로스톨"))
        self.assertIn(("와", "과"), morphology.wrong_particle_pairs("로스톨"))

    def test_rieul_final_uses_ro(self):
        self.assertIn(("으로", "로"), morphology.wrong_particle_pairs("로스톨"))

    def test_boundary_matches_auxiliary_particle_but_not_longer_hangul_word(self):
        pattern = morphology.wrong_particle_pattern("발로르", "이")
        self.assertIsNotNone(pattern.search("발로르이는"))
        self.assertIsNone(pattern.search("발로르이라는"))

    def test_scanner_and_legacy_fixer_use_same_boundary_constructor(self):
        a = morphology.wrong_particle_pattern("로스톨", "가")
        b = morphology.wrong_particle_pattern("로스톨", "가")
        fixtures = ["로스톨가", "로스톨가는", "로스톨가로", "로스톨가!", "앞 로스톨가 뒤"]
        self.assertEqual([bool(a.search(x)) for x in fixtures], [bool(b.search(x)) for x in fixtures])


if __name__ == "__main__":
    unittest.main()
