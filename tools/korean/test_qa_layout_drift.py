import importlib.util
import unittest
from pathlib import Path


MODULE_PATH = Path(__file__).with_name("qa-layout-drift.py")
SPEC = importlib.util.spec_from_file_location("qa_layout_drift", MODULE_PATH)
qa_layout_drift = importlib.util.module_from_spec(SPEC)
assert SPEC.loader is not None
SPEC.loader.exec_module(qa_layout_drift)


class SemanticSpacingTests(unittest.TestCase):
    def test_line_break_counts_as_semantic_space(self):
        self.assertEqual(
            qa_layout_drift.semantic_spacing_key("앞 문장 다음 문장<end>"),
            qa_layout_drift.semantic_spacing_key("앞 문장<line-break>다음 문장<end>"),
        )

    def test_missing_space_after_comma_is_detected(self):
        self.assertNotEqual(
            qa_layout_drift.semantic_spacing_key("평온하게 살아갈 때, 다음을 고른다<end>"),
            qa_layout_drift.semantic_spacing_key("평온하게 살아갈 때,다음을 고른다<end>"),
        )


class SyncLayoutContentTests(unittest.TestCase):
    def test_sentence_punctuation_stays_before_existing_line_break(self):
        korean = "문장이 끝났다. 다음 문장이다<end>"
        layout = "문장이 끝났다<line-break>다음 문장이다<end>"
        self.assertEqual(
            qa_layout_drift.sync_layout_content(korean, layout),
            "문장이 끝났다.<line-break>다음 문장이다<end>",
        )

    def test_boundary_comma_insertion_stays_before_existing_line_break(self):
        korean = "평온하게 살아갈 때, 다음을 고른다<end>"
        layout = "평온하게 살아갈 때<line-break>다음을 고른다<end>"
        self.assertEqual(
            qa_layout_drift.sync_layout_content(korean, layout),
            "평온하게 살아갈 때,<line-break>다음을 고른다<end>",
        )

    def test_boundary_closing_quote_stays_before_existing_line_break(self):
        korean = "그가 말했다」 다음을 고른다<end>"
        layout = "그가 말했다<line-break>다음을 고른다<end>"
        self.assertEqual(
            qa_layout_drift.sync_layout_content(korean, layout),
            "그가 말했다」<line-break>다음을 고른다<end>",
        )

    def test_mixed_punctuation_text_insertion_at_boundary_fails_closed(self):
        korean = "그가 말했다.새 문장 다음을 고른다<end>"
        layout = "그가 말했다<line-break>다음을 고른다<end>"
        with self.assertRaisesRegex(ValueError, "requires reflow"):
            qa_layout_drift.sync_layout_content(korean, layout)

    def test_multiple_sentence_boundaries_keep_punctuation_on_previous_line(self):
        korean = "자여. 응답하라. 마지막 말이다<end>"
        layout = "자여<line-break>응답하라<line-break>마지막 말이다<end>"
        self.assertEqual(
            qa_layout_drift.sync_layout_content(korean, layout),
            "자여.<line-break>응답하라.<line-break>마지막 말이다<end>",
        )

    def test_boundary_comma_removal_preserves_line_break(self):
        korean = "평온하게 살아갈 때 다음을 고른다<end>"
        layout = "평온하게 살아갈 때,<line-break>다음을 고른다<end>"
        self.assertEqual(
            qa_layout_drift.sync_layout_content(korean, layout),
            "평온하게 살아갈 때<line-break>다음을 고른다<end>",
        )

    def test_lexical_replacement_and_sentence_boundary_can_coexist(self):
        korean = "석양이 수평선 너머로 저문다. 그때 생각한다<end>"
        layout = "석양이 지평선 너머로 저문다<line-break>그때 생각한다<end>"
        self.assertEqual(
            qa_layout_drift.sync_layout_content(korean, layout),
            "석양이 수평선 너머로 저문다.<line-break>그때 생각한다<end>",
        )

    def test_postcondition_rejects_new_line_start_comma(self):
        with self.assertRaisesRegex(ValueError, "prohibited line-start"):
            qa_layout_drift.assert_layout_postconditions(
                "앞<line-break>뒤<end>",
                "앞<line-break>,뒤<end>",
            )

    def test_postcondition_rejects_new_empty_display_line(self):
        with self.assertRaisesRegex(ValueError, "empty display line"):
            qa_layout_drift.assert_layout_postconditions(
                "앞<line-break>중간<line-break>뒤<end>",
                "앞<line-break><line-break>뒤<end>",
            )


if __name__ == "__main__":
    unittest.main()
