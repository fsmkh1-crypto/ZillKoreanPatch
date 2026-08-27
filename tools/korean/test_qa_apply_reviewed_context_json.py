import importlib.util
from pathlib import Path
import unittest

MODULE_PATH = Path(__file__).with_name("qa-apply-reviewed-context-json.py")
spec = importlib.util.spec_from_file_location("qa_apply_reviewed_context_json", MODULE_PATH)
mod = importlib.util.module_from_spec(spec)
assert spec.loader is not None
spec.loader.exec_module(mod)


class ReviewedContextTomlStringTest(unittest.TestCase):
    def round_trip(self, value: str) -> None:
        line = mod.encode_korean_line(value)
        self.assertEqual(mod.decode_korean_line(line), value)

    def test_plain_hangul(self):
        self.round_trip("루루안타야.<end>")

    def test_ascii_quotes(self):
        self.round_trip('그는 "안 돼"라고 말했다.<end>')

    def test_backslash(self):
        self.round_trip(r"경로 C:\\TEMP 같은 표기<end>")

    def test_control_escapes(self):
        self.round_trip("첫째\n둘째\t끝<end>")

    def test_decode_existing_escaped_quote(self):
        self.assertEqual(
            mod.decode_korean_line('korean = "그는 \\"간다\\"고 했다.<end>"'),
            '그는 "간다"고 했다.<end>',
        )

    def test_reject_non_string(self):
        with self.assertRaises(ValueError):
            mod.decode_korean_line("korean = 123")


if __name__ == "__main__":
    unittest.main()
